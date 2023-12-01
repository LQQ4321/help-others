package client

import (
	"github.com/LQQ4321/help-others/config"
	"github.com/gin-gonic/gin"
)

func init() {

	// 强制日志颜色化
	gin.ForceConsoleColor()

	router := gin.Default()

	// 定义静态文件路由
	router.Static("/", "./assets/web")

	// 启动服务器
	go router.Run(config.URL_PORT)
}
