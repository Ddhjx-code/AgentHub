package router

import (
	"github.com/Ddhjx-code/AgentHub/internal/handler/user"
	"github.com/Ddhjx-code/AgentHub/internal/middleware"
	userRepo "github.com/Ddhjx-code/AgentHub/internal/repository/user"
	"github.com/gin-gonic/gin"
)

func Setup(engine *gin.Engine, jwtSecret string, ur userRepo.Repository, userHandler *user.Handler) {
	engine.Use(middleware.CORS())

	api := engine.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		userGroup := api.Group("/user")
		userGroup.Use(middleware.JWTAuth(jwtSecret, ur))
		{
			userGroup.GET("/profile", userHandler.Profile)
		}
	}
}
