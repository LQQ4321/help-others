package db

import (
	"context"
	"errors"

	"github.com/LQQ4321/help-others/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	logger *zap.Logger
	DB     *gorm.DB
	RDB    *redis.Client
)

func MysqlInit(loggerInstance *zap.Logger) {
	logger = loggerInstance
	var err error
	DB, err = gorm.Open(mysql.Open(config.MYSQL_DSN), &gorm.Config{})
	if err != nil {
		logger.Fatal("connect database fail :", zap.Error(err))
	}
	err = DB.Exec("CREATE DATABASE IF NOT EXISTS help_others").Error
	if err != nil {
		logger.Fatal("create database help_others fail :", zap.Error(err))
	}
	err = DB.Exec("USE help_others").Error
	if err != nil {
		logger.Fatal("unable to use the database help_others :", zap.Error(err))
	}
	err = DB.AutoMigrate(&User{}, &SeekHelp{}, &LendHand{}, &Comment{})
	if err != nil {
		logger.Fatal("create tables fail : ", zap.Error(err))
	}
	err = DB.Model(&User{}).Where(&User{IsManager: true}).First(&User{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = DB.Model(&User{}).
				Create(&User{Name: "root", Password: "root", IsManager: true}).Error
			if err != nil {
				logger.Fatal("create manager role fail : ", zap.Error(err))
			}
		} else {
			logger.Fatal("create manager role fail : ", zap.Error(err))
		}
	}
	logger.Sugar().Infoln("help-others-mysql init succeed !")

	// 下面是redis的连接配置
	RDB = redis.NewClient(&redis.Options{
		Addr:     config.REDIS_PATH + config.REDIS_PORT,
		Password: "",
		DB:       0,
	})
	cmdResult := RDB.SetNX(context.Background(), "websiteConfig", 1, 0)
	if cmdResult.Err() != nil {
		logger.Fatal("set website config fail :", zap.Error(cmdResult.Err()))
	} else if cmdResult.Val() { //websiteConfig键还未存在
		for k, v := range config.WebsiteConfig {
			err = RDB.Set(context.Background(), k, v, 0).Err()
			if err != nil {
				logger.Fatal("set website config fail :", zap.Error(err))
			}
		}
	}
	logger.Sugar().Infoln("help-others-redis init succeed !")
}
