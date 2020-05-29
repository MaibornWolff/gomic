package main

import (
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func createHTTPRouter(logLevel zerolog.Level) *gin.Engine {
	if logLevel == zerolog.DebugLevel {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Use structured logging
	router.Use(logger.SetLogger())

	// Recover from panics
	router.Use(gin.Recovery())

	return router
}
