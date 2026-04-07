package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"azadi-go/internal/agreement"
	"azadi-go/internal/audit"
	"azadi-go/internal/auth"
	"azadi-go/internal/bank"
	"azadi-go/internal/config"
	"azadi-go/internal/contact"
	"azadi-go/internal/document"
	"azadi-go/internal/email"
	"azadi-go/internal/help"
	"azadi-go/internal/payment"
	"azadi-go/internal/paymentdate"
	"azadi-go/internal/seed"
	"azadi-go/internal/server"
	"azadi-go/internal/settlement"
	"azadi-go/internal/statement"

	"cloud.google.com/go/datastore"
)

func main() {
	cfg := config.Load()

	// Configure logging
	var handler slog.Handler
	if cfg.IsDev() {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	slog.SetDefault(slog.New(handler))

	ctx := context.Background()

	// Create Datastore client
	dsClient, err := datastore.NewClient(ctx, cfg.GCPProjectID)
	if err != nil {
		slog.Error("failed to create datastore client", "error", err)
		os.Exit(1)
	}
	defer func() { _ = dsClient.Close() }()

	// Repositories
	customerRepo := auth.NewCustomerRepository(dsClient)
	agreementRepo := agreement.NewRepository(dsClient)
	paymentRepo := payment.NewRepository(dsClient)
	bankRepo := bank.NewRepository(dsClient)
	settlementRepo := settlement.NewRepository(dsClient)
	statementRepo := statement.NewRepository(dsClient)
	documentRepo := document.NewRepository(dsClient)
	auditRepo := audit.NewRepository(dsClient)

	// Core services
	auditService := audit.NewService(auditRepo)
	loginTracker := auth.NewLoginAttemptTracker(ctx)
	rateLimiter := auth.NewRateLimiter(ctx)

	// Session store
	sessions, err := auth.NewSessionStore(ctx, cfg.EncryptionKey, !cfg.IsDev())
	if err != nil {
		slog.Error("failed to create session store", "error", err)
		os.Exit(1)
	}
	flashes := auth.NewFlashStore()

	// Encryptor
	encryptor, err := bank.NewEncryptor(cfg.EncryptionKey)
	if err != nil {
		slog.Error("failed to create encryptor", "error", err)
		os.Exit(1)
	}

	// Email service
	emailService := email.NewService(cfg.ResendAPIKey, cfg.FromEmail, customerRepo)

	// Domain services
	agreementService := agreement.NewService(agreementRepo)
	contactService := contact.NewService(customerRepo, auditService)
	bankService := bank.NewService(bankRepo, encryptor, auditService, emailService)
	paymentDateService := paymentdate.NewService(agreementService, agreementRepo, auditService)
	settlementService := settlement.NewService(settlementRepo, agreementService)
	statementService := statement.NewService(statementRepo, auditService, agreementService)
	documentService := document.NewService(documentRepo)

	// Stripe
	stripeService := payment.NewStripeService(cfg.StripeAPIKey, cfg.StripeWebhookSecret)
	paymentService := payment.NewService(stripeService, paymentRepo, auditService, customerRepo, agreementService)
	webhookHandler := payment.NewWebhookHandler(paymentRepo, auditService, stripeService, emailService)

	// Auth
	authProvider := auth.NewProvider(agreementRepo, customerRepo, loginTracker)

	// Build server — handlers use renderFn which is a late-binding closure
	srv, err := server.New(server.Deps{
		Cfg:         cfg,
		Sessions:    sessions,
		Flashes:     flashes,
		RateLimiter: rateLimiter,
	})
	if err != nil {
		slog.Error("failed to create server", "error", err)
		os.Exit(1)
	}

	renderFn := srv.RenderFunc()

	// Create and register handlers
	srv.SetHandlers(
		auth.NewHandler(authProvider, sessions, flashes, cfg.SeedData, renderFn),
		agreement.NewHandler(agreementService, renderFn),
		payment.NewHandler(paymentService, webhookHandler, agreementService, sessions, cfg.StripePublishableKey, renderFn),
		paymentdate.NewHandler(paymentDateService, agreementService, sessions, flashes, renderFn),
		bank.NewHandler(bankService, sessions, flashes, renderFn),
		contact.NewHandler(contactService, sessions, flashes, renderFn),
		settlement.NewHandler(settlementService, agreementService, renderFn),
		statement.NewHandler(statementService, agreementService, contactService, sessions, flashes, renderFn),
		document.NewHandler(documentService, renderFn),
		help.NewHandler(renderFn),
	)

	// Seed data in dev mode
	if cfg.SeedData {
		seeder := seed.NewSeeder(dsClient, encryptor)
		if err := seeder.Seed(ctx, "seed/customers.json"); err != nil {
			slog.Error("failed to seed data", "error", err)
		}
	}

	// Start HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: srv.Handler(),
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		slog.Info("server starting", "port", cfg.Port, "env", cfg.Environment)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			stop() // trigger shutdown path instead of os.Exit
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
