package client

import (
	"time"

	"github.com/LQQ4321/help-others/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func init() {

	// 强制日志颜色化
	gin.ForceConsoleColor()

	router := gin.Default()
	// TODO 再服务器上调试应该修改的地方
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "Access-Control-Allow-Origin"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.MaxMultipartMemory = 8 << 20 //8MB

	// 定义静态文件路由
	// TODO flutter build 有些组件显示错误，与flutter run不同
	// 猜测一：router.Static的问题
	// 猜测二：flutter build的问题
	// 个人感觉偏向后者
	router.Static("/", "./assets/web")

	router.POST("/requestJson", jsonRequest)
	router.POST("/requestForm", formRequest)

	// 启动服务器
	go router.Run(config.URL_PORT)
}
