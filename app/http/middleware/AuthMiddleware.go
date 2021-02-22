package middleware

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := sessions.Default(c).Get("user")
		fmt.Println(user)
		if user == nil {
			c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("login?rand=%d", time.Now().UnixNano()))
		}
	}
}
