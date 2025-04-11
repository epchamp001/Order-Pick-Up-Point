package main

import (
	"context"
	_ "order-pick-up-point/docs"
	"order-pick-up-point/internal/app"
	"order-pick-up-point/internal/config"
	"order-pick-up-point/pkg/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title Order Pick-Up Point
// @version 1.0
// @description Service for processing orders at Pick-Up Point. Allow registration and login by email/password as well as dummy login using user roles (client, employee, moderator).

// @contact.name Egor Ponyaev
// @contact.url https://github.com/epchamp001
// @contact.email epchamp001@gmail.com

// @license.name MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT token. Obtain the token via /login (using email and password) or via /dummy Login (passing desired role).
func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	cfg, err := config.LoadConfig("configs/", ".env")
	if err != nil {
		panic(err)
	}

	log := logger.NewLogger(cfg.Env)
	defer log.Sync()

	log.Infow("config", cfg)

	server := app.NewServer(cfg, log)

	if err := server.Run(ctx); err != nil {
		log.Fatalw("Failed to start server",
			"error", err,
		)
	}

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(cfg.PublicServer.ShutdownTimeout)*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Errorw("Shutdown failed",
			"error", err,
		)
		os.Exit(1)
	}

	log.Info("Application stopped gracefully")
}
