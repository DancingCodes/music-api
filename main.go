package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	_ = godotenv.Load()

	initCOS()
	initDB()
	r := setupRouter()

	slog.Info("服务已启动(Port:8081)")
	if err := r.Run(":8081"); err != nil {
		slog.Error("服务启动失败", "error", err)
		os.Exit(1)
	}
}
