package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/config"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/es2"
	handler "github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/handler"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
	"github.com/rs/zerolog"
)

func main() {
	handler.Logger = zerolog.New(os.Stdout).With().
		Timestamp().
		Str("service", "carrier-connector").
		Logger()

	if err := godotenv.Load(); err != nil {
		handler.Logger.Info().Msg("No .env file found, using system environment")
	}

	client := es2.NewES2Client(&config.ES2Config{
		BaseURL:                  handler.GetEnv("SMDP_URL", "https://smdp.example.com"),
		APIKey:                   handler.GetEnv("SMDP_API_KEY", "test-api-key"),
		FunctionalityRequesterID: handler.GetEnv("FUNCTIONALITY_REQUESTER_ID", "carrier-connector"),
		InsecureSkipVerify:       handler.GetEnv("INSECURE_SKIP_VERIFY", "true") == "true",
	})
	port := handler.GetEnv("PORT", "8080")

	router := gin.Default()
	router.Use(gin.LoggerWithFormatter(func(p gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			p.ClientIP, p.TimeStamp.Format(time.RFC1123),
			p.Method, p.Path, p.Request.Proto,
			p.StatusCode, p.Latency, p.Request.UserAgent(), p.ErrorMessage,
		)
	}))
	router.Use(gin.Recovery())

	dsn := handler.GetEnv("DATABASE_DSN", "")
	if dsn == "" {
		handler.Logger.Fatal().Msg("DATABASE_DSN is required (no in-memory fallback)")
	}
	postgresRepo, err := repository.NewPostgresProfileStore(dsn)
	if err != nil {
		handler.Logger.Fatal().Err(err).Msg("Failed to connect to Postgres")
	}
	defer postgresRepo.Close()

	// Wrap with cache (5 minute TTL)
	profileRepo := repository.NewCachedProfileStore(postgresRepo, 5*time.Minute)
	defer profileRepo.Close()

	setupRoutes(router, client, profileRepo)

	handler.Logger.Info().Str("port", port).Msg("Carrier Connector API server starting")
	if err := router.Run(":" + port); err != nil {
		handler.Logger.Fatal().Err(err).Msg("Failed to start server")
	}
}
