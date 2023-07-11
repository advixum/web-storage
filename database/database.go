package database

import (
	"fmt"
	"os"
	"spa-api/logging"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var C *gorm.DB
var log = logging.Config

func Connect() {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	if gin.Mode() == gin.TestMode {
		dbName = "test"
	}
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, pass, dbName, port,
	)
	var err error
	C, err = gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{Logger: logging.GL(log)},
	)
	if err != nil {
		log.Fatal(logging.F()+"() failed to initialize database:", err)
	}
}
