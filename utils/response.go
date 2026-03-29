package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse is the standard envelope for all responses
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Success sends a 200 OK with data payload
func RespondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: data})
}

// Created sends a 201 Created with data payload
func RespondCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{Success: true, Data: data})
}

// RespondMessage sends a success response with only a message
func RespondMessage(c *gin.Context, code int, message string) {
	c.JSON(code, APIResponse{Success: true, Message: message})
}

// RespondError sends a structured error response
func RespondError(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(code, APIResponse{Success: false, Error: message})
}
