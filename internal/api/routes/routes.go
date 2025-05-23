package routes

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/bernardofernandezz/scheduling-api/internal/api/handlers"
	"github.com/bernardofernandezz/scheduling-api/internal/api/middleware"
	"github.com/bernardofernandezz/scheduling-api/internal/config"
	"github.com/bernardofernandezz/scheduling-api/internal/repository"
	"github.com/bernardofernandezz/scheduling-api/internal/service"
	"github.com/bernardofernandezz/scheduling-api/pkg/auth"
)

// SetupRouter configures and returns the API router
func SetupRouter(repos *repository.Repositories, cfg *config.Config) *gin.Engine {
	// Set Gin mode based on configuration
	gin.SetMode(cfg.Server.Mode)

	// Initialize router with recovery and logging
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(middleware.SecurityHeaders())

	// Configure CORS with environment settings
	corsOrigins := strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",")
	if len(corsOrigins) == 0 || (len(corsOrigins) == 1 && corsOrigins[0] == "") {
		corsOrigins = []string{"*"} // Default to all origins if not specified
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Configure rate limits from environment
	reqLimit, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_REQUESTS"))
	if reqLimit <= 0 {
		reqLimit = 60 // Default to 60 requests
	}
	
	rateDuration := os.Getenv("RATE_LIMIT_DURATION")
	var duration time.Duration
	if rateDuration == "" {
		duration = time.Minute // Default to 1 minute
	} else {
		var err error
		duration, err = time.ParseDuration(rateDuration)
		if err != nil {
			duration = time.Minute // Default to 1 minute on parse error
		}
	}

	// Create services
	userService := service.NewUserService(repos.UserRepo, cfg)
	appointmentService := service.NewAppointmentService(
		repos.AppointmentRepo,
		repos.EmployeeRepo,
		repos.SupplierRepo,
		repos.OperationRepo,
		repos.ProductRepo,
	)

	// Create JWT manager
	jwtManager := auth.NewJWTManager(
		cfg.Auth.JWTSecret,
		time.Duration(cfg.Auth.ExpireTime)*time.Hour,
	)

	// Create handlers
	authHandler := handlers.NewAuthHandler(userService, jwtManager)
	appointmentHandler := handlers.NewAppointmentHandler(appointmentService)

	// Create authentication middleware
	authMiddleware := auth.AuthMiddleware(userService)

	// Rate limiters with different configurations for public and protected routes
	publicLimiter := middleware.RateLimiter(reqLimit, duration)
	protectedLimiter := middleware.RateLimiter(reqLimit*5, duration) // 5x more for authenticated users

	// API group
	api := router.Group("/api")
	{
		// Public authentication routes
		authRoutes := api.Group("/auth")
		authRoutes.Use(publicLimiter)
		{
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/refresh", authHandler.RefreshToken)
			authRoutes.POST("/password-reset", authHandler.RequestPasswordReset)
		}

		// Protected routes requiring authentication
		protected := api.Group("/")
		protected.Use(authMiddleware, protectedLimiter)
		{
			// User routes
			userRoutes := protected.Group("/users")
			{
				userRoutes.GET("/profile", authHandler.Profile)
				userRoutes.POST("/change-password", authHandler.ChangePassword)
			}

			// Appointment routes
			appointmentRoutes := protected.Group("/appointments")
			{
				// Basic CRUD operations
				appointmentRoutes.POST("", appointmentHandler.Create)
				appointmentRoutes.GET("", appointmentHandler.List)
				appointmentRoutes.GET("/:id", appointmentHandler.Get)
				appointmentRoutes.PUT("/:id", appointmentHandler.Update)
				appointmentRoutes.DELETE("/:id", appointmentHandler.Delete)

				// Status management
				appointmentRoutes.POST("/:id/status", appointmentHandler.UpdateStatus)

				// Availability checking
				appointmentRoutes.POST("/check-availability", appointmentHandler.CheckAvailability)

				// Specialized queries
				appointmentRoutes.GET("/upcoming", appointmentHandler.GetUpcoming)
				appointmentRoutes.GET("/by-date-range", appointmentHandler.GetByDateRange)
				appointmentRoutes.GET("/by-supplier/:supplier_id", appointmentHandler.GetBySupplier)
				appointmentRoutes.GET("/by-employee/:employee_id", appointmentHandler.GetByEmployee)
				appointmentRoutes.GET("/by-operation/:operation_id", appointmentHandler.GetByOperation)
			}

			// Admin routes (requires admin role)
			adminRoutes := protected.Group("/admin")
			adminRoutes.Use(auth.RoleMiddleware("admin"))
			{
				adminRoutes.GET("/statistics/appointments", appointmentHandler.GetStatistics)
			}
		}
	}

	// Health check endpoint for container orchestration
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
			"time":   time.Now().UTC().Format(time.RFC3339),
			"mode":   cfg.Server.Mode,
			"version": "1.0.0",
		})
	})

	// Readiness probe for Kubernetes
	router.GET("/ready", func(c *gin.Context) {
		// Check if database is accessible
		db := repos.GetDB()
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "Database connection not available",
			})
			return
		}
		
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "Failed to get database connection",
			})
			return
		}
		
		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "Database connection failed: " + err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
			"database": "connected",
		})
	})

	// Handle 404 Not Found
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Endpoint not found",
			"path":  c.Request.URL.Path,
		})
	})

	return router
}

