package router

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/minisource/feedback/config"
	"github.com/minisource/feedback/internal/database"
	"github.com/minisource/feedback/internal/handler"
	"github.com/minisource/feedback/internal/middleware"
	"github.com/minisource/feedback/internal/repository"
	"github.com/minisource/feedback/internal/usecase"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/go-sdk/auth"
	"github.com/minisource/go-sdk/comment"
)

// Router holds the fiber app and dependencies
type Router struct {
	app    *fiber.App
	db     *database.MongoDB
	cfg    *config.Config
	logger logging.Logger
}

// NewRouter creates a new router
func NewRouter(db *database.MongoDB, cfg *config.Config, logger logging.Logger, authClient *auth.Client) *Router {
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorHandler: customErrorHandler,
	})

	r := &Router{
		app:    app,
		db:     db,
		cfg:    cfg,
		logger: logger,
	}

	r.setupMiddleware(authClient)
	r.setupRoutes(authClient)

	return r
}

// setupMiddleware sets up global middleware
func (r *Router) setupMiddleware(authClient *auth.Client) {
	// Recovery middleware
	r.app.Use(recover.New())

	// CORS middleware
	r.app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Tenant-ID",
		AllowMethods: "GET, POST, PUT, DELETE, PATCH, OPTIONS",
	}))

	// Logging middleware
	r.app.Use(middleware.LoggingMiddleware(r.logger))

	// Tenant middleware
	r.app.Use(middleware.TenantMiddleware())
}

// setupRoutes sets up all routes
func (r *Router) setupRoutes(authClient *auth.Client) {
	// Initialize repositories
	feedbackRepo := repository.NewFeedbackRepository(r.db)
	voteRepo := repository.NewVoteRepository(r.db)
	categoryRepo := repository.NewCategoryRepository(r.db)
	settingRepo := repository.NewSettingRepository(r.db)
	subscriptionRepo := repository.NewSubscriptionRepository(r.db)

	// Initialize comment client (uses comment microservice via go-sdk)
	commentClient := comment.NewClient(comment.ClientConfig{
		BaseURL:  r.cfg.Comment.ServiceURL,
		Timeout:  time.Duration(r.cfg.Comment.Timeout) * time.Second,
		Logger:   r.logger,
		GetToken: authClient.GetToken,
	})

	// Initialize usecases
	feedbackUsecase := usecase.NewFeedbackUsecase(feedbackRepo, voteRepo, categoryRepo, settingRepo, subscriptionRepo)
	subscriptionUsecase := usecase.NewSubscriptionUsecase(subscriptionRepo, feedbackRepo, categoryRepo)
	adminUsecase := usecase.NewAdminUsecase(feedbackRepo, categoryRepo, settingRepo)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler()
	feedbackHandler := handler.NewFeedbackHandler(feedbackUsecase)
	commentHandler := handler.NewCommentHandler(commentClient)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionUsecase)
	adminHandler := handler.NewAdminHandler(adminUsecase, feedbackUsecase)

	// Swagger route
	r.app.Get("/swagger/*", swagger.HandlerDefault)

	// Health routes (no auth required)
	r.app.Get("/health", healthHandler.Health)
	r.app.Get("/ready", healthHandler.Ready)

	// Auth middleware
	authMiddleware := middleware.AuthMiddleware(middleware.AuthConfig{
		AuthClient: authClient,
		SkipPaths:  []string{"/health", "/ready"},
		RequireAdmin: []string{
			"/api/v1/admin",
		},
	})

	// API v1 routes
	api := r.app.Group("/api/v1", authMiddleware)

	// Rate limiting for API routes
	api.Use(middleware.RateLimitMiddleware(middleware.RateLimitConfig{
		Max:     r.cfg.Feedback.RateLimitRequests,
		Window:  time.Duration(r.cfg.Feedback.RateLimitWindow) * time.Second,
		KeyFunc: middleware.DefaultRateLimitKeyFunc,
	}))

	// Public routes
	r.setupPublicRoutes(api, feedbackHandler, commentHandler, adminHandler)

	// Protected routes (require auth)
	requireAuth := middleware.RequireAuthMiddleware()
	r.setupProtectedRoutes(api, feedbackHandler, commentHandler, subscriptionHandler, requireAuth)

	// Admin routes
	r.setupAdminRoutes(api, adminHandler, feedbackHandler, commentHandler)
}

// setupPublicRoutes sets up public routes
func (r *Router) setupPublicRoutes(api fiber.Router, feedbackHandler *handler.FeedbackHandler, commentHandler *handler.CommentHandler, adminHandler *handler.AdminHandler) {
	// Feedback (read)
	api.Get("/feedback", feedbackHandler.List)
	api.Get("/feedback/trending", feedbackHandler.GetTrending)
	api.Get("/feedback/stats", feedbackHandler.GetStats)
	api.Get("/feedback/:id", feedbackHandler.GetByID)

	// Comments (read)
	api.Get("/feedback/:feedback_id/comments", commentHandler.List)
	api.Get("/feedback/:feedback_id/comments/stats", commentHandler.GetStats)
	api.Get("/comments/:id", commentHandler.GetByID)
	api.Get("/comments/:id/replies", commentHandler.GetReplies)

	// Categories (read)
	api.Get("/categories", adminHandler.ListCategories)
	api.Get("/categories/:id", adminHandler.GetCategory)

	// Settings (public only)
	api.Get("/settings", adminHandler.GetPublicSettings)
}

// setupProtectedRoutes sets up routes that require authentication
func (r *Router) setupProtectedRoutes(api fiber.Router, feedbackHandler *handler.FeedbackHandler, commentHandler *handler.CommentHandler, subscriptionHandler *handler.SubscriptionHandler, requireAuth fiber.Handler) {
	// Feedback (write)
	api.Post("/feedback", requireAuth, feedbackHandler.Create)
	api.Put("/feedback/:id", requireAuth, feedbackHandler.Update)
	api.Delete("/feedback/:id", requireAuth, feedbackHandler.Delete)
	api.Post("/feedback/:id/vote", requireAuth, feedbackHandler.Vote)

	// Comments (write)
	api.Post("/feedback/:feedback_id/comments", requireAuth, commentHandler.Create)
	api.Put("/comments/:id", requireAuth, commentHandler.Update)
	api.Delete("/comments/:id", requireAuth, commentHandler.Delete)
	api.Post("/comments/:id/reactions", requireAuth, commentHandler.AddReaction)
	api.Delete("/comments/:id/reactions", requireAuth, commentHandler.RemoveReaction)

	// Subscriptions
	api.Get("/subscriptions", requireAuth, subscriptionHandler.ListByUser)
	api.Post("/subscriptions", requireAuth, subscriptionHandler.Subscribe)
	api.Put("/subscriptions/:id", requireAuth, subscriptionHandler.UpdatePreferences)
	api.Delete("/subscriptions/:id", requireAuth, subscriptionHandler.Unsubscribe)
	api.Get("/feedback/:feedback_id/subscription", requireAuth, subscriptionHandler.CheckSubscription)
	api.Delete("/feedback/:feedback_id/unsubscribe", requireAuth, subscriptionHandler.UnsubscribeFromFeedback)
}

// setupAdminRoutes sets up admin routes
func (r *Router) setupAdminRoutes(api fiber.Router, adminHandler *handler.AdminHandler, feedbackHandler *handler.FeedbackHandler, commentHandler *handler.CommentHandler) {
	admin := api.Group("/admin")

	// Categories
	admin.Post("/categories", adminHandler.CreateCategory)
	admin.Put("/categories/:id", adminHandler.UpdateCategory)
	admin.Delete("/categories/:id", adminHandler.DeleteCategory)

	// Settings
	admin.Get("/settings", adminHandler.GetSettings)
	admin.Put("/settings", adminHandler.UpdateSetting)
	admin.Post("/settings/initialize", adminHandler.InitializeSettings)

	// Moderation
	admin.Get("/feedback/pending", adminHandler.GetPendingFeedback)
	admin.Post("/feedback/:id/approve", adminHandler.ApproveFeedback)
	admin.Post("/feedback/:id/reject", adminHandler.RejectFeedback)
	admin.Put("/feedback/:id/status", adminHandler.UpdateStatus)
	admin.Post("/feedback/:id/response", adminHandler.AddOfficialResponse)

	// Stats
	admin.Get("/stats", adminHandler.GetStats)
}

// App returns the fiber app
func (r *Router) App() *fiber.App {
	return r.app
}

// customErrorHandler handles errors
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error":   true,
		"message": message,
	})
}
