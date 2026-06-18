package main

import (
	"fmt"
	"os"

	"github.com/Ddhjx-code/AgentHub/internal/config"
	"github.com/Ddhjx-code/AgentHub/internal/database"
	adminHandler "github.com/Ddhjx-code/AgentHub/internal/handler/admin"
	agentHandler "github.com/Ddhjx-code/AgentHub/internal/handler/agent"
	chatHandler "github.com/Ddhjx-code/AgentHub/internal/handler/chat"
	userHandler "github.com/Ddhjx-code/AgentHub/internal/handler/user"
	"github.com/Ddhjx-code/AgentHub/internal/llm"
	agentRepo "github.com/Ddhjx-code/AgentHub/internal/repository/agent"
	convRepo "github.com/Ddhjx-code/AgentHub/internal/repository/conversation"
	msgRepo "github.com/Ddhjx-code/AgentHub/internal/repository/message"
	txRepo "github.com/Ddhjx-code/AgentHub/internal/repository/transaction"
	userRepo "github.com/Ddhjx-code/AgentHub/internal/repository/user"
	walletRepo "github.com/Ddhjx-code/AgentHub/internal/repository/wallet"
	"github.com/Ddhjx-code/AgentHub/internal/router"
	agentSvc "github.com/Ddhjx-code/AgentHub/internal/service/agent"
	chatSvc "github.com/Ddhjx-code/AgentHub/internal/service/chat"
	userSvc "github.com/Ddhjx-code/AgentHub/internal/service/user"
	"github.com/Ddhjx-code/AgentHub/internal/tool"
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
	cr := convRepo.NewRepository(db)
	mr := msgRepo.NewRepository(db)
	tr := txRepo.NewRepository(db)

	llmClient := llm.NewClient()
	toolExec := tool.NewExecutor(cfg.Coze.BaseURL, cfg.N8N.DefaultTimeout)

	us := userSvc.NewService(ur, wr, cfg.JWT, lg)
	as := agentSvc.NewService(ar, lg)
	cs := chatSvc.NewService(ar, cr, mr, wr, tr, llmClient, toolExec, lg)

	uh := userHandler.NewHandler(us)
	agH := agentHandler.NewHandler(as)
	adH := adminHandler.NewHandler(as)
	chH := chatHandler.NewHandler(cs)

	if cfg.Server.Mode != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.Default()
	router.Setup(engine, cfg.JWT.Secret, ur, uh, agH, adH, chH)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	lg.Info("AgentHub server starting", "addr", addr)
	if err := engine.Run(addr); err != nil {
		lg.Error("server failed", "error", err)
		os.Exit(1)
	}
}
