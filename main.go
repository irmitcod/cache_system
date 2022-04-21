package main

import (
	"argos/config"
	"argos/src/controller"
	"argos/src/models/image"
	"argos/src/repository"
	"argos/src/utils/lfu"
	"argos/src/utils/workerpool"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"log"
)

// @title    AISA API
// @version  1.0
func main() {

	//config logger
	entry := config.NewLogger()

	// Setup Configuration
	configuration := config.GetConfig()
	database := config.NewMemoryClient(configuration)

	// Setup Repository
	productRepository := repository.NewImageRepository(database)

	// setup lfu cache
	lfuCach := lfu.New()

	// setup worker pool
	totalWorker := 5
	wp := workerpool.NewWorkerPool(totalWorker)
	wp.Run()

	//setup localache
	cache := config.NewCache()
	// Setup Service
	imageService := image.NewService(&productRepository, wp, configuration.MaxHeight, configuration.MaxWidth, cache, lfuCach, entry)

	//setup controller
	handler := controller.NewHandler(imageService)

	// Start server
	router := gin.Default()
	router.Use(cors.Default())
	prefix := router.Group(configuration.Prefix)
	handler.Route(prefix)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	serverAddr := fmt.Sprintf("%s:%d", configuration.Host, configuration.Port)
	log.Panic(router.Run(serverAddr))
}
