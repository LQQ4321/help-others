package client

import (
	"net/http"

	"github.com/LQQ4321/help-others/db"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type jsonFunc func([]string, *gin.Context)

var (
	DB          *gorm.DB
	logger      *zap.SugaredLogger
	jsonFuncMap map[string]jsonFunc
)

func ClientInit(loggerInstance *zap.SugaredLogger) {
	DB = db.DB
	logger = loggerInstance
	jsonFuncMap = make(map[string]jsonFunc)
	jsonFuncMap = map[string]jsonFunc{
		"login": login,
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
