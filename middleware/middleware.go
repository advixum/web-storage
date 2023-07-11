package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"spa-api/handlers"
	"spa-api/logging"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/secure"
)

var log = logging.Config

// Security parameters
func Security() secure.Options {
	return secure.Options{
		AllowedHosts:          []string{"127.0.0.1:8080", "ssl.example.com"},
		SSLRedirect:           false, // false for dev | true for prod
		SSLHost:               "ssl.example.com",
		SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "http"},
		STSSeconds:            315360000,
		STSIncludeSubdomains:  true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'",
	}
}

// CORS parameters
func CORS() cors.Config {
	config := cors.DefaultConfig()
	config.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	config.AllowHeaders = append(config.AllowHeaders, "Authorization")
	config.AllowOrigins = []string{
		//"http://localhost:3000",
		//"http://127.0.0.1:3000",
		"https://ssl.example.com",
	}
	return config
}

// JWT parameters
func JWT() jwt.GinJWTMiddleware {
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:            "Application API",
		SigningAlgorithm: "RS256", // USE ASYMMETRIC ENCRYPTION ONLY!!!
		PrivKeyFile:      "keys/priv.pem",
		PubKeyFile:       "keys/pub.pem",
		Timeout:          time.Hour,
		MaxRefresh:       time.Hour,
		PayloadFunc:      handlers.Payload,
		Authenticator:    handlers.LogIn,
		TokenLookup:      "header: Authorization, query: token, cookie: jwt",
		TokenHeadName:    "Bearer",
		TimeFunc:         time.Now,
	})
	if err != nil {
		log.Fatal(logging.F()+"() error:", err)
	}
	return *authMiddleware
}

// Generates middleware key files for the asymmetric algorithm (RS256),
// if they do not exist.
func Keys() {
	check := 0
	_, err := os.Stat("keys/priv.pem")
	if err != nil {
		if os.IsNotExist(err) {
			check++
		} else {
			log.Error(logging.F()+"() priv.pem checking error:", err)
		}
	}
	_, err = os.Stat("keys/pub.pem")
	if err != nil {
		if os.IsNotExist(err) {
			check++
		} else {
			log.Error(logging.F()+"() pub.pem checking error:", err)
		}
	}
	if check == 0 {
		return
	} else {
		os.Remove("keys/priv.pem")
		os.Remove("keys/pub.pem")
		priv, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			log.Error(logging.F()+"() cannot generate privave key:", err)
			return
		}
		privBytes := x509.MarshalPKCS1PrivateKey(priv)
		privPEM := pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: privBytes,
			},
		)
		pub := &priv.PublicKey
		pubBytes, err := x509.MarshalPKIXPublicKey(pub)
		if err != nil {
			log.Error(logging.F()+"() cannot convert public key to PEM:", err)
			return
		}
		pubPEM := pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PUBLIC KEY",
				Bytes: pubBytes,
			},
		)
		os.WriteFile("keys/priv.pem", privPEM, 0700)
		os.WriteFile("keys/pub.pem", pubPEM, 0755)
	}
}
