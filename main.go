package main

import (
	"server/configs"
	"server/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	configs.ConnectDB()
	routes.DriveRoute(router)
	router.Run("localhost:5000")
}
