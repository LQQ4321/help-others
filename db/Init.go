package db

import (
	"context"
	"math/rand"

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
	logger.Sugar().Infoln("help-others-mysql init succeed !")

	// 下面是redis的连接配置
	RDB = redis.NewClient(&redis.Options{
		Addr:     config.REDIS_PATH + config.REDIS_PORT,
		Password: "",
		DB:       0,
	})
	// TODO 下面应该使用事务，不然锁是加上了，但是值写入的时候可能出现错误，从而导致后续重启也无法写入数据
	// TODO 每次都重新加载配置，因为配置可能会修改，方便调试
	err = RDB.Del(context.Background(), "websiteConfig").Err()
	if err != nil {
		logger.Fatal("release lock fail :", zap.Error(err))
	}
	result, err := RDB.SetNX(context.Background(), "websiteConfig", 1, 0).Result()
	if err != nil {
		logger.Fatal("set website config fail :", zap.Error(err))
	} else if result { //websiteConfig键还未存在
		for k, v := range config.WebsiteConfig {
			err = RDB.Set(context.Background(), k, v, 0).Err()
			if err != nil {
				logger.Fatal("set website config fail :", zap.Error(err))
			}
		}
	}
	logger.Sugar().Infoln("help-others-redis init succeed !")
}

func GenerateRandomString(length int) string {
	charset := "0123456789"
	charsetLength := len(charset)

	randomString := make([]byte, length)
	for i := 0; i < length; i++ {
		randomString[i] = charset[rand.Intn(charsetLength)]
	}

	return string(randomString)
}
