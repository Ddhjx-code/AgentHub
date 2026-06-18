package router

import (
	"github.com/Ddhjx-code/AgentHub/internal/handler/admin"
	agentHandler "github.com/Ddhjx-code/AgentHub/internal/handler/agent"
	"github.com/Ddhjx-code/AgentHub/internal/handler/user"
	"github.com/Ddhjx-code/AgentHub/internal/middleware"
	userRepo "github.com/Ddhjx-code/AgentHub/internal/repository/user"
	"github.com/gin-gonic/gin"
)

func Setup(
	engine *gin.Engine,
	jwtSecret string,
	ur userRepo.Repository,
	userHandler *user.Handler,
	agentH *agentHandler.Handler,
	adminH *admin.Handler,
) {
	engine.Use(middleware.CORS())

	api := engine.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		protected := api.Group("")
		protected.Use(middleware.JWTAuth(jwtSecret, ur))
		{
			userGroup := protected.Group("/user")
			{
				userGroup.GET("/profile", userHandler.Profile)
			}

			agentGroup := protected.Group("/agents")
			{
				agentGroup.GET("", agentH.List)
				agentGroup.GET("/:id", agentH.Detail)
			}

			adminGroup := protected.Group("/admin")
			adminGroup.Use(middleware.RequireAdmin())
			{
				adminAgents := adminGroup.Group("/agents")
				{
					adminAgents.POST("", adminH.CreateAgent)
					adminAgents.GET("", adminH.ListAgents)
					adminAgents.GET("/:id", adminH.GetAgent)
					adminAgents.PUT("/:id", adminH.UpdateAgent)
					adminAgents.DELETE("/:id", adminH.DeleteAgent)
					adminAgents.PUT("/:id/toggle", adminH.ToggleAgentStatus)
				}
			}
		}
	}
}
