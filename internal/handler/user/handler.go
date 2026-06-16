package user

import (
	"regexp"

	"github.com/Ddhjx-code/AgentHub/internal/middleware"
	userSvc "github.com/Ddhjx-code/AgentHub/internal/service/user"
	"github.com/Ddhjx-code/AgentHub/pkg/errcode"
	"github.com/Ddhjx-code/AgentHub/pkg/response"
	"github.com/gin-gonic/gin"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type Handler struct {
	userSvc userSvc.Service
}

func NewHandler(s userSvc.Service) *Handler {
	return &Handler{userSvc: s}
}

type registerReq struct {
	Email    string `json:"email" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	if !emailRegex.MatchString(req.Email) {
		response.ErrorWithMsg(c, errcode.ErrInvalidParam, "invalid email format")
		return
	}

	if len(req.Password) < 6 || len(req.Password) > 72 {
		response.ErrorWithMsg(c, errcode.ErrInvalidParam, "password must be 6-72 characters")
		return
	}

	user, err := h.userSvc.Register(c.Request.Context(), userSvc.RegisterRequest{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
	})
	if err != nil {
		if ec, ok := err.(*errcode.ErrCode); ok {
			response.Error(c, ec)
			return
		}
		response.Error(c, errcode.ErrInternalServer)
		return
	}

	response.Created(c, user)
}

func (h *Handler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	token, user, err := h.userSvc.Login(c.Request.Context(), userSvc.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if ec, ok := err.(*errcode.ErrCode); ok {
			response.Error(c, ec)
			return
		}
		response.Error(c, errcode.ErrInternalServer)
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"user":  user,
	})
}

func (h *Handler) Profile(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(int64)

	user, err := h.userSvc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if ec, ok := err.(*errcode.ErrCode); ok {
			response.Error(c, ec)
			return
		}
		response.Error(c, errcode.ErrInternalServer)
		return
	}

	response.Success(c, user)
}
