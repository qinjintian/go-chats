package routers

import (
	"go-chats/app/http/controller"
	"github.com/gin-gonic/gin"
	"go-chats/app/http/middleware"
)

func InitRouter(r *gin.Engine) {

	r.Any("/test", (&controller.PublicController{}).Test)                    // 测试
	r.Any("/login", (&controller.PublicController{}).Login)                  // 登录
	r.Any("/register", (&controller.PublicController{}).Register)            // 注册
	r.Any("/reset-password", (&controller.PublicController{}).ResetPassword) // 找回密码

	authorized := r.Group("/")
	authorized.Use(middleware.Auth())
	{
		r.GET("logout", (&controller.PublicController{}).Logout)                 // 登录
		r.GET("index", middleware.Auth(), (&controller.IndexController{}).Index) // 主页
	}
}
