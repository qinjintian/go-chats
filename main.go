package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-ini/ini"
	"go-chats/bootstrap"
)

func main() {
	// 加载配置文件
	cfg, err := ini.Load("config/app.ini")
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to load config/app.ini file, error: %v", err))
		return
	}

	// 设置GIN运行模式，默认是 debug 开发模式，release 为生产模式, test 为测试模式
	gin.SetMode(cfg.Section(ini.DefaultSection).Key("RUN_MODE").MustString(""))

	// 使用Logger和Recovery中间件的Engine实例
	r := gin.Default()

	// 自定义日志输出格式和
	bootstrap.Init(r, cfg)
}
