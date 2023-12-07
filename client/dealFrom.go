package client

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/LQQ4321/help-others/config"
	"github.com/LQQ4321/help-others/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func seekAHelp(c *gin.Context) {
	var response struct {
		Status         string      `json:"status"`
		SingleSeekHelp db.SeekHelp `json:"singleSeekHelp"`
	}
	response.Status = config.RETURN_FAIL
	logger.Infoln(c.Request)
	files := c.Request.MultipartForm.File["files"] //还是说是 files[]
	_score := c.Request.FormValue("score")
	imageType := c.Request.FormValue("imageType")
	codeType := c.Request.FormValue("codeType")
	user_id := c.Request.FormValue("userId")
	uploadTime := c.Request.FormValue("uploadTime")
	problemLink := c.Request.FormValue("problemLink")
	remark := c.Request.FormValue("remark")
	if len(files) != 2 || len(_score) == 0 || len(imageType) == 0 ||
		len(codeType) == 0 || len(user_id) == 0 || len(uploadTime) == 0 ||
		len(problemLink) == 0 || len(remark) == 0 {
		logger.Errorln(fmt.Errorf("parse fail"))
		c.JSON(http.StatusOK, response)
		return
	}
	date := strings.Split(uploadTime, " ")[0]
	var curDateId int64
	err := DB.Model(&db.SeekHelp{}).
		Where("upload_time LIKE ?", date+"%").Count(&curDateId).Error
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	curDateId++
	// 创建当天的文件夹
	dirName := config.USER_UPLOAD_FOLDER + date + "/" +
		strconv.FormatInt(curDateId, 10) + "/"
	err = os.MkdirAll(dirName, 0755)
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	codeFilePath := dirName + "origin.txt"
	if err := c.SaveUploadedFile(files[0], codeFilePath); err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	imageFilePath := dirName + config.PROBLEM_PICTURE_NAME + "." + imageType
	if err := c.SaveUploadedFile(files[1], imageFilePath); err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	userId, err := strconv.Atoi(user_id)
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	score, err := strconv.Atoi(_score)
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	var websiteConfig []int
	keys := []string{"MaxHelp", "MaxComment", "SeekHelpBan"}
	for _, v := range keys {
		result, err := RDB.Get(context.Background(), v).Result()
		if err != nil {
			logger.Errorln(err)
			c.JSON(http.StatusOK, response)
			return
		}
		num, err := strconv.Atoi(result)
		if err != nil {
			logger.Errorln(err)
			c.JSON(http.StatusOK, response)
			return
		}
		websiteConfig = append(websiteConfig, num)
	}
	err = DB.Transaction(func(tx *gorm.DB) error {
		response.SingleSeekHelp = db.SeekHelp{
			ProblemLink: problemLink,
			ImagePath:   imageFilePath,
			TopicRemark: remark,
			UploadTime:  date,
			CodePath:    codeFilePath,
			Language:    config.Language[codeType],
			MaxHelp:     websiteConfig[0],
			MaxComment:  websiteConfig[1],
			Score:       score,
			Ban:         websiteConfig[2],
		}
		err := tx.Model(&db.SeekHelp{}).Create(&response.SingleSeekHelp).Error
		if err != nil {
			return err
		}
		user := db.User{}
		err = tx.Model(&db.User{}).
			Where(&db.User{ID: userId}).First(&user).Error
		if err != nil {
			return err
		}
		user.Score -= score
		if len(user.SeekHelp) != 0 {
			user.SeekHelp += "#"
		}
		user.SeekHelp += strconv.Itoa(response.SingleSeekHelp.ID)
		// 这里的updates可能有bug,有ID，应该不用使用Where了吧
		return tx.Model(&db.User{}).Updates(&user).Error
	})
	if err != nil {
		logger.Errorln(err)
	} else {
		response.Status = config.RETURN_SUCCEED
	}
	c.JSON(http.StatusOK, response)
}
