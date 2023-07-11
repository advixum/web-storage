package main

import (
	db "spa-api/database"
	"spa-api/handlers"
	"spa-api/logging"
	"spa-api/middleware"
	"spa-api/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var log = logging.Config

func main() {
	// Database initial
	db.Connect()
	db.C.AutoMigrate(&models.User{}, &models.File{})

	// Run router
	r := router()
	r.Run("127.0.0.1:8080")
}

func router() *gin.Engine {
	// RSA keys
	middleware.Keys()

	// Gin settings
	r := gin.New()
	r.SetTrustedProxies([]string{"127.0.0.1"})
	r.Use(gin.LoggerWithWriter(log.WriterLevel(logrus.InfoLevel)))
	r.Use(gin.RecoveryWithWriter(log.WriterLevel(logrus.ErrorLevel)))
	r.Use(secure.Secure(middleware.Security()))
	r.Use(cors.New(middleware.CORS()))
	authJWT := middleware.JWT()
	r.MaxMultipartMemory = 8 << 20

	// Public routes
	pub := r.Group("/api/pub")
	pub.POST("/login", authJWT.LoginHandler)
	pub.POST("/signup", handlers.SignUp)

	// Authenticated routes
	auth := r.Group("/api/auth")
	auth.Use(authJWT.MiddlewareFunc())
	auth.GET("/files", handlers.List)
	auth.GET("/download", handlers.Download)
	auth.POST("/upload", handlers.Upload)
	auth.POST("/rename", handlers.Rename)
	auth.POST("/delete", handlers.Delete)
	return r
}
