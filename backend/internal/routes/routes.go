package routes

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"sales-module/internal/controllers"
	"sales-module/internal/middleware"
	"sales-module/internal/services"
)

// Setup registers all routes on the Gin engine
func Setup(r *gin.Engine, db *sql.DB) {
	// Services
	salesService := services.NewSalesService(db)
	invoiceService := services.NewInvoiceService(db)
	reportService := services.NewReportService(db)
	productService := services.NewProductService(db)

	// Controllers
	authCtrl := controllers.NewAuthController(db)
	healthCtrl := controllers.NewHealthController()
	salesCtrl := controllers.NewSalesController(salesService, invoiceService)
	invoiceCtrl := controllers.NewInvoiceController(invoiceService)
	customerCtrl := controllers.NewCustomerController(db)
	reportCtrl := controllers.NewReportController(reportService)
	productCtrl := controllers.NewProductController(productService)

	api := r.Group("/api")
	{
		// Public routes
		api.GET("/health", healthCtrl.Health)
		api.POST("/auth/login", authCtrl.Login)

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// Sales
			protected.GET("/sales", salesCtrl.List)
			protected.POST("/sales", salesCtrl.Create)
			protected.GET("/sales/:id", salesCtrl.Get)
			protected.PUT("/sales/:id", salesCtrl.Update)
			protected.PATCH("/sales/:id/status", salesCtrl.UpdateStatus)
			protected.DELETE("/sales/:id", salesCtrl.Delete)

			// Invoices
			protected.GET("/invoices", invoiceCtrl.List)
			protected.GET("/invoices/:id", invoiceCtrl.Get)
			protected.POST("/invoices/generate/:saleId", invoiceCtrl.Generate)
                        protected.PATCH("/invoices/:id/status", invoiceCtrl.UpdateStatus)
                        protected.DELETE("/invoices/:id", invoiceCtrl.Delete)

			// Customers
			protected.GET("/customers", customerCtrl.List)
                        protected.POST("/customers", customerCtrl.Create)
                        protected.PUT("/customers/:id", customerCtrl.Update)
                        protected.DELETE("/customers/:id", customerCtrl.Delete)
			protected.GET("/customers/:id/history", customerCtrl.GetHistory)

			// Reports
			protected.GET("/reports/dashboard", reportCtrl.Dashboard)
			protected.GET("/reports/summary", reportCtrl.Summary)
			protected.GET("/reports/top-products", reportCtrl.TopProducts)
			protected.GET("/reports/revenue", reportCtrl.Revenue)
			protected.POST("/reports/export", reportCtrl.Export)

			// Products
			protected.GET("/products", productCtrl.List)
		}
	}
}
