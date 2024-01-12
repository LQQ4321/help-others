package main

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"log"
	math_rand "math/rand"
	"os"

	"github.com/LQQ4321/help-others/client"
	"github.com/LQQ4321/help-others/config"
	"github.com/LQQ4321/help-others/db"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var logger *zap.Logger

func main() {
	var err error
	logger, err = zap.NewProduction()
	defer logger.Sync()
	if err != nil {
		log.Fatalln("init logger fail :", err)
	}
	db.MysqlInit(logger)
	client.ClientInit(logger.Sugar())
	initRand()
	initFiles()
	initRouter()
}

func initRouter() {
	// 强制日志颜色化
	gin.ForceConsoleColor()

	router := gin.Default()
	// TODO 再服务器上调试应该修改的地方
	// router.Use(cors.New(cors.Config{
	// 	AllowOrigins:     []string{"*"},
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
	// 	AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "Access-Control-Allow-Origin"},
	// 	ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
	// 	AllowCredentials: true,
	// 	MaxAge:           12 * time.Hour,
	// }))
	router.MaxMultipartMemory = 8 << 20 //8MB

	// 定义静态文件路由
	// TODO flutter build 有些组件显示错误，与flutter run不同
	// 猜测一：router.Static的问题
	// 猜测二：flutter build的问题
	// 个人感觉偏向后者
	router.Static("/", "./assets/web")

	router.POST("/requestJson", client.JsonRequest)
	router.POST("/requestForm", client.FormRequest)

	// 启动服务器
	router.Run(config.URL_PORT)
}

func initRand() {
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	if err != nil {
		logger.Fatal("random generator init failed :", zap.Error(err))
	}
	sd := int64(binary.LittleEndian.Uint64(b[:]))
	logger.Sugar().Infof("random seed : %d", sd)
	math_rand.Seed(sd)
}

func initFiles() {
	if err := os.MkdirAll(config.USER_UPLOAD_FOLDER, 0755); err != nil {
		logger.Fatal(config.USER_UPLOAD_FOLDER+" create fail :", zap.Error(err))
	}
}
