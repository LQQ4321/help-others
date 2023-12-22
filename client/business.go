package client

import (
	"net/http"

	"github.com/LQQ4321/help-others/db"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type jsonFunc func([]string, *gin.Context)
type formFunc func(*gin.Context)

var (
	DB          *gorm.DB
	RDB         *redis.Client
	logger      *zap.SugaredLogger
	jsonFuncMap map[string]jsonFunc
	formFuncMap map[string]formFunc
)

func ClientInit(loggerInstance *zap.SugaredLogger) {
	DB = db.DB
	RDB = db.RDB
	logger = loggerInstance
	jsonFuncMap = make(map[string]jsonFunc)
	jsonFuncMap = map[string]jsonFunc{
		"sendVerificationCode": sendVerificationCode,
		"login":                login,
		"register":             register,
		"forgotPassword":       forgotPassword,
		"requestList":          requestList,
		"requestShowDataOne":   requestShowDataOne,
		"downloadFile":         downloadFile,
		"likeOperate":          likeOperate,
		"sendAComment":         sendAComment,
	}
	formFuncMap = make(map[string]formFunc)
	formFuncMap = map[string]formFunc{
		"seekAHelp": seekAHelp,
		"lendAHand": lendAHand,
	}
}

// 关于cookie的使用，可以在总入口进行校验，如果不能读取到cookie或者cookie不存在与redis的set中，
// (有两个set，一个用户，一个管理员)，那么就需要重新登录，该次请求就没有必要继续下去了

func jsonRequest(c *gin.Context) {
	var request struct {
		RequestType string   `json:"requestType"`
		Info        []string `json:"info"`
	}
	if err := c.BindJSON(&request); err != nil {
		logger.Error("parse request data fail :", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if v, ok := jsonFuncMap[request.RequestType]; ok {
		v(request.Info, c)
	} else {
		c.JSON(http.StatusNotFound, nil)
	}
}

func formRequest(c *gin.Context) {
	requestType := c.Request.FormValue("requestType")
	if v, ok := formFuncMap[requestType]; ok {
		v(c)
	} else {
		c.JSON(http.StatusNotFound, nil)
	}
}
