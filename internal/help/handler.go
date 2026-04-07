package help

import (
	"net/http"

	"azadi-go/internal/httputil"
)

type Handler struct {
	render httputil.RenderFunc
}

func NewHandler(render httputil.RenderFunc) *Handler {
	return &Handler{render: render}
}

func (h *Handler) FAQs(w http.ResponseWriter, r *http.Request)       { h.render(w, r, "help/faqs.html", nil) }
func (h *Handler) WaysToPay(w http.ResponseWriter, r *http.Request)  { h.render(w, r, "help/ways-to-pay.html", nil) }
func (h *Handler) ContactUs(w http.ResponseWriter, r *http.Request)  { h.render(w, r, "help/contact-us.html", nil) }
func (h *Handler) Cookies(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "legal/cookies.html", nil)
}
func (h *Handler) Privacy(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "legal/privacy.html", nil)
}
func (h *Handler) Terms(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "legal/terms.html", nil)
}
