package client

import (
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
		Status string `json:"status"`
	}
	response.Status = config.RETURN_FAIL
	files := c.Request.MultipartForm.File["files"] //还是说是 files[]
	_score := c.Request.FormValue("score")
	imageType := c.Request.FormValue("imageType")
	codeType := c.Request.FormValue("codeType")
	user_id := c.Request.FormValue("userId")
	date := c.Request.FormValue("date")
	language := c.Request.FormValue("language")
	problemLink := c.Request.FormValue("problemLink")
	remark := c.Request.FormValue("remark")
	if len(files) != 2 || len(_score) == 0 || len(imageType) == 0 ||
		len(codeType) == 0 || len(user_id) == 0 || len(date) == 0 ||
		len(problemLink) == 0 || len(remark) == 0 {
		logger.Errorln(fmt.Errorf("parse fail"))
		c.JSON(http.StatusOK, response)
		return
	}
	// 创建当天的文件夹
	dirName := strings.Split(date, " ")[0]
	err := os.MkdirAll(config.CODE_FOLDER+dirName, 0755)
	if err != nil {
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
	err = DB.Transaction(func(tx *gorm.DB) error {
		seekHelp := db.SeekHelp{
			ProblemLink: problemLink,
			ImagePath:   "",
			TopicRemark: remark,
			UploadTime:  date,
			CodePath:    "",
			Language:    language,
			MaxHelp:     0,
			MaxComment:  0,
			Score:       score,
			Ban:         0,
		}
		err := tx.Model(&db.SeekHelp{}).Create(&seekHelp).Error
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
		user.SeekHelp += strconv.Itoa(seekHelp.ID)
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
