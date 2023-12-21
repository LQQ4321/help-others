package client

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/LQQ4321/help-others/config"
	"github.com/LQQ4321/help-others/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func lendAHand(c *gin.Context) {
	var response struct {
		Status         string      `json:"status"`
		SingleLendHand db.LendHand `json:"singleLendHand"`
	}
	response.Status = config.RETURN_FAIL
	files := []multipart.File{}
	for i := 1; i < 2; i++ {
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
	paramFields := []string{"remark", "codeType",
		"seekHelpId", "date", "userId", "uploadTime"}
	paramValues := []string{}
	for _, v := range paramFields {
		paramValues = append(paramValues, c.Request.FormValue(v))
	}
	for i, v := range paramFields {
		if len(paramValues[i]) == 0 {
			logger.Errorln("parse " + v + " param fail")
			c.JSON(http.StatusOK, response)
			return
		}
	}
	seekHelpId, err := strconv.Atoi(paramValues[2])
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	var lendHandCount int64
	err = DB.Model(&db.LendHand{}).
		Where(&db.LendHand{SeekHelpId: seekHelpId}).
		Count(&lendHandCount).Error
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	lendHandCount++
	seekHelps := []db.SeekHelp{}
	err = DB.Model(&db.SeekHelp{}).
		Where("upload_time LIKE ?", paramValues[3]+"%").
		Find(&seekHelps).Error
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	seekHelpCount := -1
	for i, _ := range seekHelps {
		if seekHelps[i].ID == seekHelpId {
			seekHelpCount = i + 1
			break
		}
	}
	if seekHelpCount == -1 {
		logger.Errorln(fmt.Errorf("Not found seekHelpCount"))
		c.JSON(http.StatusOK, response)
		return
	}
	// diff -U NUM origin.txt copyId.txt > diffId.txt
	codeFilePath := config.USER_UPLOAD_FOLDER + paramValues[3] + "/" +
		strconv.Itoa(seekHelpCount) + "/" + "copy" +
		strconv.Itoa(int(lendHandCount)) + ".txt"
	err = SaveAFile(codeFilePath, files[0])
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	originFilePath := config.USER_UPLOAD_FOLDER + paramValues[3] + "/" +
		strconv.Itoa(seekHelpCount) + "/" + "origin.txt"
	diffFilePath := config.USER_UPLOAD_FOLDER + paramValues[3] + "/" +
		strconv.Itoa(seekHelpCount) + "/" + "diff" +
		strconv.Itoa(int(lendHandCount)) + ".txt"
	cmd := exec.Command("sh", "-c", "diff -U 9999 "+originFilePath+" "+codeFilePath+" > "+diffFilePath)
	// cmd := exec.Command("diff", "-U", "9999", originFilePath, codeFilePath, ">", diffFilePath)
	// 这里是同步，有点耗时间，可以考虑Start和Wait的异步结合
	// 虽然报错exit status 1，但是结果还是好的，所以感觉不用管这里的报错
	cmd.Run()
	userId, err := strconv.Atoi(paramValues[4])
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	var websiteConfig []int
	keys := []string{"MaxComment", "LendHandBan"}
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
		err := tx.Model(&db.User{ID: userId}).First(&user).Error
		if err != nil {
			return err
		}
		response.SingleLendHand = db.LendHand{
			SeekHelpId:     seekHelpId,
			LendHanderId:   paramValues[4],
			LendHanderName: user.Name,
			UploadTime:     paramValues[5],
			Remark:         paramValues[0],
			CodePath:       codeFilePath,
			DiffPath:       diffFilePath,
			MaxComment:     websiteConfig[0],
			Ban:            websiteConfig[1],
		}
		err = tx.Model(&db.LendHand{}).Create(&response.SingleLendHand).Error
		if err != nil {
			return err
		}
		if len(user.LendHand) != 0 {
			user.LendHand += "#"
		}
		user.LendHand += strconv.Itoa(response.SingleLendHand.ID)
		return tx.Model(&db.User{}).Where(&db.User{ID: user.ID}).Updates(&user).Error
	})
	if err != nil {
		logger.Errorln(err)
	} else {
		response.Status = config.RETURN_SUCCEED
	}
	c.JSON(http.StatusOK, response)
}

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
