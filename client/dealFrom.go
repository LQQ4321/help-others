package client

import (
	"context"
	"io"
	"mime/multipart"
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
	files := []multipart.File{}
	for i := 1; i < 3; i++ {
		file, _, err := c.Request.FormFile("file" + strconv.Itoa(i))
		if err != nil {
			logger.Errorln(err)
			c.JSON(http.StatusOK, response)
			return
		}
		// 这里关闭的文件，会不会一直是最后一个
		defer file.Close()
		files = append(files, file)
	}
	paramFields := []string{"problemLink", "remark", "score",
		"imageType", "codeType", "userId", "uploadTime"}
	paramValues := []string{}
	for _, v := range paramFields {
		paramValues = append(paramValues, c.Request.FormValue(v))
	}
	for i, v := range paramFields {
		if i > 0 && len(paramValues[i]) == 0 {
			logger.Errorln("parse " + v + " param fail")
			c.JSON(http.StatusOK, response)
			return
		}
	}
	date := strings.Split(paramValues[6], " ")[0]
	var curDateId int64
	err := DB.Model(&db.SeekHelp{}).
		Where("upload_time LIKE ?", date+"%").Count(&curDateId).Error
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	// 注意，如果后面的事务没有成功，那么下一次得到的count还是一样的
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
	filePaths := []string{}
	for i := 0; i < 2; i++ {
		if i == 0 {
			filePaths = append(filePaths, dirName+config.ORIGIN_CODE_NAME)
		} else {
			filePaths = append(filePaths, dirName+config.PROBLEM_PICTURE_NAME+"."+paramValues[3])
		}
		err = SaveAFile(filePaths[i], files[i])
		if err != nil {
			logger.Errorln(err)
			c.JSON(http.StatusOK, response)
			return
		}
	}
	userId, err := strconv.Atoi(paramValues[5])
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	score, err := strconv.Atoi(paramValues[2])
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
		user := db.User{}
		err := tx.Model(&db.User{}).
			Where(&db.User{ID: userId}).First(&user).Error
		if err != nil {
			return err
		}
		response.SingleSeekHelp = db.SeekHelp{
			ProblemLink:    paramValues[0],
			SeekHelperId:   paramValues[5],
			SeekHelperName: user.Name,
			ImagePath:      filePaths[1],
			TopicRemark:    paramValues[1],
			UploadTime:     paramValues[6],
			CodePath:       filePaths[0],
			Language:       config.Language[paramValues[4]],
			MaxHelp:        websiteConfig[0],
			MaxComment:     websiteConfig[1],
			Score:          score,
			Ban:            websiteConfig[2],
		}
		err = tx.Model(&db.SeekHelp{}).Create(&response.SingleSeekHelp).Error
		if err != nil {
			return err
		}
		user.Score -= score
		if len(user.SeekHelp) != 0 {
			user.SeekHelp += "#"
		}
		user.SeekHelp += strconv.Itoa(response.SingleSeekHelp.ID)
		// 这里的updates可能有bug,有ID，应该不用使用Where了吧,还是要的，要不然就使用Save吧,还没试过
		return tx.Model(&db.User{}).Where(&db.User{ID: user.ID}).Updates(&user).Error
	})
	if err != nil {
		logger.Errorln(err)
	} else {
		response.Status = config.RETURN_SUCCEED
	}
	c.JSON(http.StatusOK, response)
}

func SaveAFile(savePath string, file multipart.File) error {
	out, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	return err
}
