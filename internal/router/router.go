package router

import (
	"github.com/Ddhjx-code/AgentHub/internal/handler/admin"
	agentHandler "github.com/Ddhjx-code/AgentHub/internal/handler/agent"
	chatHandler "github.com/Ddhjx-code/AgentHub/internal/handler/chat"
	kbHandler "github.com/Ddhjx-code/AgentHub/internal/handler/knowledge"
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
	chatH *chatHandler.Handler,
	kbH *kbHandler.Handler,
) {
	engine.Use(middleware.CORS())

	api := engine.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		publicAgents := api.Group("/agents")
		{
			publicAgents.GET("", agentH.List)
			publicAgents.GET("/:id", agentH.Detail)
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
				agentGroup.POST("/:id/chat", chatH.SendMessage)
			}

			convGroup := protected.Group("/conversations")
			{
				convGroup.GET("", chatH.ListConversations)
				convGroup.GET("/:id/messages", chatH.GetMessages)
				convGroup.DELETE("/:id", chatH.DeleteConversation)
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
					adminAgents.POST("/:id/knowledge-bases", kbH.BindAgentKB)
					adminAgents.GET("/:id/knowledge-bases", kbH.ListAgentKBs)
					adminAgents.DELETE("/:id/knowledge-bases/:kb_id", kbH.UnbindAgentKB)
				}

				adminKBs := adminGroup.Group("/knowledge-bases")
				{
					adminKBs.POST("", kbH.CreateKB)
					adminKBs.GET("", kbH.ListKBs)
					adminKBs.GET("/:id", kbH.GetKB)
					adminKBs.PUT("/:id", kbH.UpdateKB)
					adminKBs.DELETE("/:id", kbH.DeleteKB)
					adminKBs.POST("/:id/documents", kbH.UploadDocument)
					adminKBs.GET("/:id/documents", kbH.ListDocuments)
					adminKBs.DELETE("/:id/documents/:doc_id", kbH.DeleteDocument)
				}
			}
		}
	}
}
