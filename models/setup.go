package models

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	// DBConn database access
	DBConn *gorm.DB
	// RedisClient redis cache access
	RedisClient redis.Client
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		panic("Error loading .env file")
	}
}

// Config retrieves env vars
func Config(key string) string {
	return os.Getenv(key)
}

// InitDatabase configures database connection from environment variables
func InitDatabase() {
	var err error
	dbpass := Config("POSTGRES_PASSWORD")
	dbuser := Config("POSTGRES_USER")
	dbport := Config("POSTGRES_PORT")
	dbhost := Config("POSTGRES_HOST")
	dbname := Config("POSTGRES_DB")

	dsn := fmt.Sprintf("host=" + dbhost + " user=" + dbuser + " password=" + dbpass + " dbname=" + dbname + " port=" + dbport)
	DBConn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("Could not connect.")
	}
	fmt.Println("Connected to database.")

	RedisClient = *redis.NewClient(&redis.Options{
		Addr:     Config("REDIS_HOST") + ":" + Config("REDIS_PORT"),
		Password: "",
		DB:       0,
	})
	status := RedisClient.Ping(context.Background())
	fmt.Println("====REDIS PING: ", status.Val())

	DBConn.AutoMigrate(&Forum{}, &Thread{}, &Post{}, &User{}, &Member{}, &PendingMember{}, &Image{})
	fmt.Println("Database migrated.")
}
