package response

import (
	"net/http"

	"github.com/Ddhjx-code/AgentHub/pkg/errcode"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func SuccessWithMeta(c *gin.Context, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
		Meta:    meta,
	})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func Error(c *gin.Context, err *errcode.ErrCode) {
	status := errCodeToHTTPStatus(err.Code)
	c.JSON(status, Response{
		Code:    err.Code,
		Message: err.Message,
	})
}

func ErrorWithMsg(c *gin.Context, err *errcode.ErrCode, msg string) {
	status := errCodeToHTTPStatus(err.Code)
	c.JSON(status, Response{
		Code:    err.Code,
		Message: msg,
	})
}

func errCodeToHTTPStatus(code int) int {
	switch code {
	case 400:
		return http.StatusBadRequest
	case 401:
		return http.StatusUnauthorized
	case 403:
		return http.StatusForbidden
	case 404:
		return http.StatusNotFound
	case 500:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}
