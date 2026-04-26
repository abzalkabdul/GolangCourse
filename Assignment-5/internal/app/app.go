package app

import (
	"fmt"
	"os"
	"strings"

	v1 "assignment_5/internal/controller/http/v1"
	"assignment_5/internal/entity"
	"assignment_5/internal/usecase"
	"assignment_5/internal/usecase/repo"
	"assignment_5/pkg/logger"
	"assignment_5/pkg/postgres"

	"github.com/gin-gonic/gin"
)

func Run() error {
	if err := loadEnv(".env"); err != nil {
		fmt.Println("Warning: could not load .env file:", err)
	}

	pg, err := postgres.New()
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}

	if err := pg.Conn.AutoMigrate(&entity.User{}); err != nil {
		return fmt.Errorf("automigrate: %w", err)
	}

	l := logger.New()
	userRepo := repo.NewUserRepo(pg)
	userUC := usecase.NewUserUseCase(userRepo)

	router := gin.Default()
	v1.NewRouter(router, userUC, l)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return router.Run(":" + port)
}

// loadEnv reads KEY=VALUE pairs from a file and sets them as environment variables.
func loadEnv(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
	return nil
}
