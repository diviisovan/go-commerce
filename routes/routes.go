package routes

import (
	"time"

	"go-ecommerce/config"
	"go-ecommerce/controllers"
	"go-ecommerce/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all routes for the application.
//
// It returns an error because the auth controller validates its configuration
// (JWT secret strength) while being built, and a misconfigured deployment must
// fail at startup rather than on the first login attempt.
func SetupRoutes(cfg *config.Config) (*gin.Engine, error) {
	router := gin.Default()

	// Add request logging middleware
	router.Use(middleware.RequestLogger())

	// Configure CORS. Note: AllowAllOrigins is fine for local development, but
	// for production list the exact front-end origins instead.
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(corsConfig))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger documentation. See swagger.go: the custom index adds the "Bearer "
	// prefix for you, so the Authorize dialog takes a bare access token.
	router.GET("/swagger/*any", swaggerHandler("doc.json"))

	// Auth controller is built once and shared: it owns the TokenManager that
	// both issues tokens and, via the middleware below, verifies them.
	authController, err := controllers.NewAuthController(cfg.Auth)
	if err != nil {
		return nil, err
	}
	requireAuth := middleware.RequireAuth(authController.Service().TokenManager())

	// API routes
	api := router.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			// Public endpoints are rate limited per IP. Account lockout stops
			// an attacker hammering one account; this stops them spreading the
			// same attempt across many accounts from one source.
			public := auth.Group("", middleware.RateLimit(20, time.Minute))
			{
				public.POST("/signup", authController.Signup)
				public.POST("/login", authController.Login)
				public.POST("/refresh", authController.Refresh)
				public.POST("/logout", authController.Logout)
			}

			// Protected endpoints require a valid access token.
			protected := auth.Group("", requireAuth)
			{
				protected.GET("/me", authController.Me)
				protected.POST("/change-password", authController.ChangePassword)
				protected.POST("/logout-all", authController.LogoutAll)
			}
		}

		// Product routes
		productController := controllers.NewProductController()
		products := api.Group("/products")
		{
			products.POST("", productController.CreateProduct)
			products.GET("", productController.GetAllProducts)
			products.GET("/:id", productController.GetProduct)
			products.PUT("/:id", productController.UpdateProduct)
			products.DELETE("/:id", productController.DeleteProduct)

			// To lock writes down to admins, move them into a group like this:
			//
			//   admin := products.Group("", requireAuth,
			//       middleware.RequireRole(models.RoleAdmin))
			//   admin.POST("", productController.CreateProduct)
			//
			// Left open for now so existing clients keep working.
		}

		// Cart routes
		cartController := controllers.NewCartController()
		carts := api.Group("/carts")
		{
			carts.GET("/user/:user_id", cartController.GetOrCreateCart)
			carts.GET("/:cart_id", cartController.GetCart)
			carts.POST("/:cart_id/items", cartController.AddItemToCart)
			carts.DELETE("/:cart_id/items/:item_id", cartController.RemoveItemFromCart)
		}

		// Order routes
		orderController := controllers.NewOrderController()
		orders := api.Group("/orders")
		{
			orders.POST("", orderController.CreateOrder)
			orders.GET("/:id", orderController.GetOrder)
			orders.GET("/user/:user_id", orderController.GetOrdersByUser)
			orders.PUT("/:id/status", orderController.UpdateOrderStatus)
			orders.POST("/:id/cancel", orderController.CancelOrder)
		}
	}

	return router, nil
}
