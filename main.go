package main

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"log"
	math_rand "math/rand"
	"os"
	"time"

	"github.com/LQQ4321/help-others/client"
	"github.com/LQQ4321/help-others/config"
	"github.com/LQQ4321/help-others/db"
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
	time.Sleep(time.Hour * 3)
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
	if err := os.MkdirAll(config.CODE_FOLDER, 0755); err != nil {
		logger.Fatal(config.CODE_FOLDER+" create fail :", zap.Error(err))
	}
}
