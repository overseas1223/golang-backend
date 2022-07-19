package routes

import (
	"server/controllers"
	"server/middlewares"
	"server/services"

	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine) {
	router.POST("/sign-up", controllers.SignUp())
	router.POST("/sign-in", controllers.SignIn())
	router.GET("/google/:email", services.HandleGoogleLogin())
	router.GET("/signin-google", services.CallBackFromGoogle())
}

func DriveRoutes(router *gin.Engine) {
	router.POST("/register-car", controllers.CreateCar())
	router.POST("/add-order", controllers.CreateOrder())
	router.POST("/add-review", controllers.CreateReview())
	router.POST("/cancel-order", controllers.CancelOrder())
	router.GET("/profile-business-car", controllers.GetProfileCarByEmail())
	router.GET("/profile-business-home", controllers.GetProfileBusinessOrderByEmail())
	router.GET("/profile-business-review", controllers.GetProfileReviewByEmail())
	router.GET("/profile-business-revenue", controllers.GetProfileRevenueByEmail())
	router.GET("/user-orders", controllers.GetProfileUserOrderByEmail())
	router.GET("/user-order-details", controllers.GetOrderDetails())
}

func DriveRoute(router *gin.Engine) {
	UserRoutes(router)
	router.Use(middlewares.DeserializeUser())
	DriveRoutes(router)
}
