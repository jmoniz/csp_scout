package api

import (
	"csp_report_backend/internal/db"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// SetupRouter initializes and configures the Chi HTTP router.
func SetupRouter(database *db.DB) *chi.Mux {
	r := chi.NewRouter()

	// Global middlewares
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(CorsMiddleware())

	// Handlers
	reportHandler := NewReportHandler(database)
	sessionHandler := NewSessionHandler(database)
	violationHandler := NewViolationHandler(database)
	policyHandler := NewPolicyHandler(database)

	// Healthcheck
	r.Get("/health", HandleHealth)

	// API routes
	r.Route("/api", func(api chi.Router) {
		// Reports ingestion
		api.Options("/reports/{sessionId}", reportHandler.HandleOptions)
		api.Post("/reports/{sessionId}", reportHandler.IngestReport)

		// Sessions
		api.Get("/sessions", sessionHandler.ListSessions)
		api.Post("/sessions", sessionHandler.CreateSession)

		api.Route("/sessions/{id}", func(session chi.Router) {
			session.Get("/", sessionHandler.GetSession)
			session.Delete("/", sessionHandler.DeleteSession)

			// Policy & Settings
			session.Get("/policy", policyHandler.GetPolicy)
			session.Get("/export", policyHandler.GetPolicy)
			session.Put("/settings", policyHandler.UpdateSettings)

			// Violations
			session.Get("/violations", violationHandler.ListViolations)
			session.Post("/violations/bulk", violationHandler.BulkUpdateViolations)
			session.Post("/violations/approve-self", violationHandler.BulkApproveSelf)
			session.Get("/violations/{violationId}/samples", violationHandler.GetViolationSamples)
			session.Patch("/violations/{violationId}", violationHandler.UpdateViolationStatus)
		})
	})

	// Root aliases for ingestion (/sessionId)
	r.Options("/{sessionId}", reportHandler.HandleOptions)
	r.Post("/{sessionId}", reportHandler.IngestReport)

	return r
}
