package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
)

func CustomRecovery() gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultWriter, func(c *gin.Context, recovered any) {
		// Handle panic
		//msg := "Unhandled Error:"

		if err, hasErr := recovered.(error); hasErr {
			_ = c.Error(err.(error))
			msg = fmt.Sprintf("Unhandled Error: %v", err.(error).Error())
		}
		//response.SystemError(c, errors.New(msg))
	})
}
