package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"azadi-go/internal/agreement"
	"azadi-go/internal/auth"
	"azadi-go/internal/bank"
	"azadi-go/internal/config"
	"azadi-go/internal/contact"
	"azadi-go/internal/document"
	"azadi-go/internal/help"
	"azadi-go/internal/httputil"
	"azadi-go/internal/middleware"
	"azadi-go/internal/payment"
	"azadi-go/internal/paymentdate"
	"azadi-go/internal/settlement"
	"azadi-go/internal/statement"
)

type Server struct {
	cfg                *config.Config
	sessions           *auth.SessionStore
	flashes            *auth.FlashStore
	templates          map[string]*template.Template
	viteAssets         map[string]string
	authHandler        *auth.Handler
	agreementHandler   *agreement.Handler
	paymentHandler     *payment.Handler
	paymentDateHandler *paymentdate.Handler
	bankHandler        *bank.Handler
	contactHandler     *contact.Handler
	settlementHandler  *settlement.Handler
	statementHandler   *statement.Handler
	documentHandler    *document.Handler
	helpHandler        *help.Handler
	rateLimiter        *auth.RateLimiter
}

type Deps struct {
	Cfg         *config.Config
	Sessions    *auth.SessionStore
	Flashes     *auth.FlashStore
	RateLimiter *auth.RateLimiter
}

func New(deps Deps) (*Server, error) {
	s := &Server{
		cfg:         deps.Cfg,
		sessions:    deps.Sessions,
		flashes:     deps.Flashes,
		rateLimiter: deps.RateLimiter,
	}

	if err := s.loadTemplates(); err != nil {
		return nil, fmt.Errorf("loading templates: %w", err)
	}
	s.loadViteManifest()

	return s, nil
}

func (s *Server) SetHandlers(
	authH *auth.Handler,
	agreementH *agreement.Handler,
	paymentH *payment.Handler,
	paymentDateH *paymentdate.Handler,
	bankH *bank.Handler,
	contactH *contact.Handler,
	settlementH *settlement.Handler,
	statementH *statement.Handler,
	documentH *document.Handler,
	helpH *help.Handler,
) {
	s.authHandler = authH
	s.agreementHandler = agreementH
	s.paymentHandler = paymentH
	s.paymentDateHandler = paymentDateH
	s.bankHandler = bankH
	s.contactHandler = contactH
	s.settlementHandler = settlementH
	s.statementHandler = statementH
	s.documentHandler = documentH
	s.helpHandler = helpH
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	})
	mux.HandleFunc("GET /login", s.authHandler.LoginPage)
	mux.HandleFunc("POST /login", s.authHandler.Login)
	mux.HandleFunc("GET /login-error", s.authHandler.LoginError)
	mux.HandleFunc("GET /logout", s.authHandler.Logout)
	mux.HandleFunc("POST /api/stripe/webhook", s.paymentHandler.HandleWebhook)
	mux.HandleFunc("GET /api/csp-report", handleCSPReport)
	mux.HandleFunc("GET /health", healthCheck)

	// Help/legal pages (public, but show sidebar if logged in)
	optionalAuth := func(h http.HandlerFunc) http.Handler {
		return auth.OptionalAuth(s.sessions, h)
	}
	mux.Handle("GET /help/faqs", optionalAuth(s.helpHandler.FAQs))
	mux.Handle("GET /help/ways-to-pay", optionalAuth(s.helpHandler.WaysToPay))
	mux.Handle("GET /help/contact-us", optionalAuth(s.helpHandler.ContactUs))
	mux.Handle("GET /cookies", optionalAuth(s.helpHandler.Cookies))
	mux.Handle("GET /privacy", optionalAuth(s.helpHandler.Privacy))
	mux.Handle("GET /terms", optionalAuth(s.helpHandler.Terms))

	// Authenticated routes
	requireAuth := func(h http.HandlerFunc) http.Handler {
		return auth.RequireAuth(s.sessions, h)
	}
	mux.Handle("GET /my-account", requireAuth(s.agreementHandler.MyAccount))
	mux.Handle("GET /agreements/{id}", requireAuth(s.agreementHandler.AgreementDetail))
	mux.Handle("GET /my-contact-details", requireAuth(s.contactHandler.ContactPage))
	mux.Handle("POST /my-contact-details", requireAuth(s.contactHandler.UpdateContact))
	mux.Handle("GET /my-documents", requireAuth(s.documentHandler.DocumentsPage))
	mux.Handle("GET /documents/{id}/download", requireAuth(s.documentHandler.Download))
	mux.Handle("GET /finance/make-a-payment", requireAuth(s.paymentHandler.PaymentPage))
	mux.Handle("POST /finance/make-a-payment", requireAuth(s.paymentHandler.MakePayment))
	mux.Handle("GET /finance/change-payment-date", requireAuth(s.paymentDateHandler.ChangeDatePage))
	mux.Handle("POST /finance/change-payment-date", requireAuth(s.paymentDateHandler.ChangeDate))
	mux.Handle("GET /finance/update-bank-details", requireAuth(s.bankHandler.BankDetailsPage))
	mux.Handle("POST /finance/update-bank-details", requireAuth(s.bankHandler.UpdateBankDetails))
	mux.Handle("GET /finance/settlement-figure", requireAuth(s.settlementHandler.SettlementPage))
	mux.Handle("POST /finance/settlement-figure", requireAuth(s.settlementHandler.CalculateSettlement))
	mux.Handle("GET /finance/request-a-statement", requireAuth(s.statementHandler.StatementPage))
	mux.Handle("POST /finance/request-a-statement", requireAuth(s.statementHandler.RequestStatement))

	// Static assets
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("frontend/dist"))))

	// Middleware chain
	var handler http.Handler = mux
	handler = middleware.CSRF(handler)
	handler = middleware.CSP(s.cfg.ViteDevURL)(handler)
	handler = middleware.SecurityHeaders(handler)
	handler = s.rateLimiter.Middleware(handler)
	handler = middleware.Logging(handler)

	return handler
}

func (s *Server) render(w http.ResponseWriter, r *http.Request, name string, data map[string]any) {
	if data == nil {
		data = make(map[string]any)
	}

	// Inject common template data
	if customer := auth.CustomerFromContext(r.Context()); customer != nil {
		data["Customer"] = customer
		data["CustomerName"] = customer.CustomerName
	}
	data["CSPNonce"] = middleware.NonceFromContext(r.Context())
	data["CSRFToken"] = middleware.CSRFTokenFromRequest(r)
	data["RequestURI"] = r.URL.Path
	data["ViteDevURL"] = s.cfg.ViteDevURL
	data["ViteAssets"] = s.viteAssets

	// Flash messages
	if sessionID := auth.SessionIDFromContext(r.Context()); sessionID != "" {
		if msg := s.flashes.Get(sessionID, "success"); msg != "" {
			data["Success"] = msg
		}
		if msg := s.flashes.Get(sessionID, "error"); msg != "" {
			data["Error"] = msg
		}
	}

	tmpl, ok := s.templates[name]
	if !ok {
		slog.Error("template not found", "template", name)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		slog.Error("template render failed", "template", name, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	_, _ = buf.WriteTo(w)
}

func (s *Server) loadTemplates() error {
	funcMap := template.FuncMap{
		"nonce": func() string { return "" },
		"formatDate": func(t interface{}) string {
			switch v := t.(type) {
			case string:
				return v
			default:
				return fmt.Sprintf("%v", v)
			}
		},
		"hasPrefix": strings.HasPrefix,
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s) //nolint:gosec
		},
		"seq": func(start, end int) []int {
			var s []int
			for i := start; i <= end; i++ {
				s = append(s, i)
			}
			return s
		},
		"fileType": func(filename string) string {
			ext := strings.TrimPrefix(filepath.Ext(filename), ".")
			switch ext {
			case "pdf":
				return "pdf"
			default:
				return "doc"
			}
		},
		"seqReverse": func(start, end int) []int {
			var s []int
			for i := start; i >= end; i-- {
				s = append(s, i)
			}
			return s
		},
	}

	// Parse shared templates (layout + fragments) into a base set
	base := template.New("").Funcs(funcMap)
	for _, pattern := range []string{"templates/layout.html", "templates/fragments/*.html"} {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			content, err := os.ReadFile(match)
			if err != nil {
				return fmt.Errorf("reading template %s: %w", match, err)
			}
			if _, err := base.New(match).Parse(string(content)); err != nil {
				return fmt.Errorf("parsing template %s: %w", match, err)
			}
		}
	}

	// For each page template, clone the base set and add the page
	s.templates = make(map[string]*template.Template)
	for _, dir := range []string{"", "finance", "help", "legal"} {
		pattern := filepath.Join("templates", dir, "*.html")
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			rel, _ := filepath.Rel("templates", match)
			// Skip layout and fragments (already in base)
			if rel == "layout.html" || filepath.Dir(rel) == "fragments" {
				continue
			}
			content, err := os.ReadFile(match)
			if err != nil {
				return fmt.Errorf("reading template %s: %w", match, err)
			}
			// Clone the base template set and add this page
			pageSet, err := base.Clone()
			if err != nil {
				return fmt.Errorf("cloning base for %s: %w", rel, err)
			}
			if _, err := pageSet.New(rel).Parse(string(content)); err != nil {
				return fmt.Errorf("parsing template %s: %w", rel, err)
			}
			s.templates[rel] = pageSet
		}
	}
	return nil
}

func (s *Server) loadViteManifest() {
	s.viteAssets = make(map[string]string)
	manifestPath := "frontend/dist/.vite/manifest.json"
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		slog.Info("vite manifest not found (expected in dev)", "path", manifestPath)
		return
	}

	var manifest map[string]struct {
		File string   `json:"file"`
		CSS  []string `json:"css"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		slog.Warn("failed to parse vite manifest", "error", err)
		return
	}

	for key, entry := range manifest {
		s.viteAssets[key] = "/assets/" + entry.File
		for _, css := range entry.CSS {
			s.viteAssets[key+".css"] = "/assets/" + css
		}
	}
	slog.Info("loaded vite manifest", "entries", len(manifest))
}

func (s *Server) RenderFunc() httputil.RenderFunc {
	return s.render
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func handleCSPReport(w http.ResponseWriter, r *http.Request) {
	slog.Info("CSP violation report received")
	w.WriteHeader(http.StatusOK)
}
