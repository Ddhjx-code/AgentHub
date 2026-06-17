package main

import (
	"fmt"
	"os"

	"github.com/Ddhjx-code/AgentHub/internal/config"
	"github.com/Ddhjx-code/AgentHub/internal/database"
	adminHandler "github.com/Ddhjx-code/AgentHub/internal/handler/admin"
	agentHandler "github.com/Ddhjx-code/AgentHub/internal/handler/agent"
	userHandler "github.com/Ddhjx-code/AgentHub/internal/handler/user"
	agentRepo "github.com/Ddhjx-code/AgentHub/internal/repository/agent"
	userRepo "github.com/Ddhjx-code/AgentHub/internal/repository/user"
	walletRepo "github.com/Ddhjx-code/AgentHub/internal/repository/wallet"
	"github.com/Ddhjx-code/AgentHub/internal/router"
	agentSvc "github.com/Ddhjx-code/AgentHub/internal/service/agent"
	userSvc "github.com/Ddhjx-code/AgentHub/internal/service/user"
	"github.com/Ddhjx-code/AgentHub/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	lg := logger.New(cfg.Server.Mode)

	db, err := database.New(cfg.Database.DSN)
	if err != nil {
		lg.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		lg.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}
	lg.Info("database migrated successfully")

	ur := userRepo.NewRepository(db)
	wr := walletRepo.NewRepository(db)
	ar := agentRepo.NewRepository(db)

	us := userSvc.NewService(ur, wr, cfg.JWT, lg)
	as := agentSvc.NewService(ar, lg)

	uh := userHandler.NewHandler(us)
	agH := agentHandler.NewHandler(as)
	adH := adminHandler.NewHandler(as)

	if cfg.Server.Mode != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.Default()
	router.Setup(engine, cfg.JWT.Secret, ur, uh, agH, adH)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	lg.Info("AgentHub server starting", "addr", addr)
	if err := engine.Run(addr); err != nil {
		lg.Error("server failed", "error", err)
		os.Exit(1)
	}
}
