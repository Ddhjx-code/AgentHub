package middleware

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Ddhjx-code/AgentHub/internal/model"
	userRepo "github.com/Ddhjx-code/AgentHub/internal/repository/user"
	"github.com/Ddhjx-code/AgentHub/pkg/errcode"
	"github.com/Ddhjx-code/AgentHub/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const UserIDKey = "user_id"

func JWTAuth(secret string, ur userRepo.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

		token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			response.Error(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

		subject, err := token.Claims.GetSubject()
		if err != nil || subject == "" {
			response.Error(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

		userID, err := strconv.ParseInt(subject, 10, 64)
		if err != nil {
			response.Error(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

		user, err := ur.GetByID(c.Request.Context(), userID)
		if err != nil || user == nil {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}
		if user.Status == model.UserStatusBanned {
			response.Error(c, errcode.ErrUserBanned)
			c.Abort()
			return
		}

		c.Set(UserIDKey, userID)
		c.Next()
	}
}
