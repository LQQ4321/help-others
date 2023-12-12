package client

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/LQQ4321/help-others/config"
	"github.com/LQQ4321/help-others/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// {filePath}
func downloadFile(info []string, c *gin.Context) {
	c.File(info[0])
}

// {"login",name,password}
func verifyUser(info []string, c *gin.Context) {
	var response struct {
		Status string   `json:"status"`
		Info   []string `json:"info"`
		User   db.User  `json:"user"`
	}
	response.Status = config.RETURN_FAIL
	response.Info = make([]string, 0)
	if info[0] == "login" {
		var response struct {
			Status     string   `json:"status"`
			ConfigData []int    `json:"configData"`
			User       db.User  `json:"user"`
			Info       []string `json:"info"`
		}
		response.Status = config.RETURN_FAIL
		response.ConfigData = make([]int, 0)
		response.Info = make([]string, 0)
		err := DB.Model(&db.User{}).
			Where(&db.User{Name: info[1], Password: info[2]}).
			First(&response.User).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Info = append(response.Info,
					"The user name does not exist or the password is incorrect.")
			} else {
				response.Info = append(response.Info, config.INTERNAL_ERROR)
				logger.Errorln(err)
			}
		} else {
			configs := []string{"MaxSeekHelpPerDay", "LoginDuration"}
			for i, v := range configs {
				result, err := RDB.Get(context.Background(), v).Result()
				if err != nil {
					logger.Errorln(err)
					response.Info = append(response.Info, config.INTERNAL_ERROR)
					break
				}
				num, err := strconv.Atoi(result)
				if err != nil {
					logger.Errorln(err)
					response.Info = append(response.Info, config.INTERNAL_ERROR)
					break
				}
				response.ConfigData = append(response.ConfigData, num)
				if i+1 == len(configs) {
					response.Status = config.RETURN_SUCCEED
				}
			}
		}
		c.JSON(http.StatusOK, response)
		return
	} else if info[0] == "register" {

	} else if info[0] == "retrievePassword" {

	}
	c.JSON(http.StatusOK, response)
}

// {"seekHelp",date}
// {"lendHand",seekHelpId}
// {"comment",helpType,helpId}
func requestList(info []string, c *gin.Context) {
	if info[0] == "seekHelp" {
		var response struct {
			Status       string        `json:"status"`
			SeekHelpList []db.SeekHelp `json:"seekHelpList"`
			Info         []string      `json:"info"`
		}
		response.Status = config.RETURN_FAIL
		response.SeekHelpList = make([]db.SeekHelp, 0)
		response.Info = make([]string, 0)
		err := DB.Model(&db.SeekHelp{}).
			Where("upload_time LIKE ?", info[1]+"%").
			Find(&response.SeekHelpList).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			response.Info = append(response.Info, config.INTERNAL_ERROR)
			logger.Errorln(err)
		} else {
			// 感觉可以不用从数据库中读这些数据出来
			for i, _ := range response.SeekHelpList {
				response.SeekHelpList[i].ProblemLink = ""
				response.SeekHelpList[i].ImagePath = ""
				response.SeekHelpList[i].CodePath = ""
				response.SeekHelpList[i].TopicRemark = ""
			}
			response.Status = config.RETURN_SUCCEED
		}
		c.JSON(http.StatusOK, response)
	} else if info[0] == "lendHand" {
		var response struct {
			Status       string        `json:"status"`
			LendHandList []db.LendHand `json:"lendHandList"`
			Info         []string      `json:"info"`
		}
		response.Status = config.RETURN_FAIL
		response.LendHandList = make([]db.LendHand, 0)
		response.Info = make([]string, 0)
		seekHelpId, err := strconv.Atoi(info[1])
		if err != nil {
			response.Info = append(response.Info, config.INTERNAL_ERROR)
			logger.Errorln(err)
		} else {
			err = DB.Model(&db.LendHand{}).
				Where(&db.LendHand{SeekHelpId: seekHelpId}).
				Find(&response.LendHandList).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				response.Info = append(response.Info, config.INTERNAL_ERROR)
				logger.Error(err)
			} else {
				for i, _ := range response.LendHandList {
					response.LendHandList[i].Remark = ""
					response.LendHandList[i].CodePath = ""
				}
				response.Status = config.RETURN_SUCCEED
			}
		}
		c.JSON(http.StatusOK, response)
	} else if info[0] == "comment" {
		var response struct {
			Status      string       `json:"status"`
			CommentList []db.Comment `json:"commentList"`
			Info        []string     `json:"info"`
		}
		response.Status = config.RETURN_FAIL
		response.CommentList = make([]db.Comment, 0)
		response.Info = make([]string, 0)
		helpType, err := strconv.Atoi(info[1])
		if err != nil {
			response.Info = append(response.Info, config.INTERNAL_ERROR)
			logger.Errorln(err)
			c.JSON(http.StatusOK, response)
			return
		}
		helpId, err := strconv.Atoi(info[2])
		if err != nil {
			response.Info = append(response.Info, config.INTERNAL_ERROR)
			logger.Errorln(err)
			c.JSON(http.StatusOK, response)
			return
		}
		err = DB.Model(&db.Comment{}).
			Where(&db.Comment{Type: helpType, HelpId: helpId}).
			Find(&response.CommentList).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			response.Info = append(response.Info, config.INTERNAL_ERROR)
			logger.Errorln(err)
		} else {
			response.Status = config.RETURN_SUCCEED
		}
		c.JSON(http.StatusOK, response)
	}
}

// {seekHelp,seekHelpId}
// {lendHand,lendHandId}
func requestShowDataOne(info []string, c *gin.Context) {
	var response struct {
		Status      string `json:"status"`
		CodePath    string `json:"codePath"`
		ImagePath   string `json:"imagePath"`
		Remark      string `json:"remark"`
		ProblemLink string `json:"problemLink"`
	}
	response.Status = config.RETURN_FAIL
	if info[0] == "seekHelp" {
		seekHelpId, err := strconv.Atoi(info[1])
		if err != nil {
			logger.Errorln(err)
		} else {
			seekHelp := db.SeekHelp{}
			err = DB.Model(&db.SeekHelp{}).Where(&db.SeekHelp{ID: seekHelpId}).
				First(&seekHelp).Error
			if err != nil {
				logger.Errorln(err)
			} else {
				response.CodePath = seekHelp.CodePath
				response.ImagePath = seekHelp.ImagePath
				response.Remark = seekHelp.TopicRemark
				response.ProblemLink = seekHelp.ProblemLink
				response.Status = config.RETURN_SUCCEED
			}
		}
	} else if info[0] == "lendHand" {
		lendHandId, err := strconv.Atoi(info[1])
		if err != nil {
			logger.Errorln(err)
		} else {
			lendHand := db.LendHand{}
			err = DB.Model(&db.LendHand{}).Where(&db.LendHand{ID: lendHandId}).
				First(&lendHand).Error
			if err != nil {
				logger.Errorln(err)
			} else {
				response.CodePath = lendHand.CodePath
				response.Remark = lendHand.Remark
				response.Status = config.RETURN_SUCCEED
			}
		}
	}
	c.JSON(http.StatusOK, response)
}

// {"seekHelp",seekHelpId}
// {"lendHand",lendHandId}
// func requestShowData(info []string, c *gin.Context) {
// 	if info[0] == "seekHelp" {
// 		seekHelpId, err := strconv.Atoi(info[1])
// 		if err != nil {
// 			logger.Errorln(err)
// 		}
// 	} else if info[0] == "lendHand" {

// 	}
// }

// 删除一条求助信息，和该条求助信息有关的所有数据都会被删除
// 为了保证数据同步，最好前端若干次操作后进行一次网络请求(寻找一个平衡点)
// TODO 等到项目上线的时候，浏览器的同源策略保证请求是manager调用的，就不需要认证了
// {seekHelpId}
// func deleteASeekHelp(info []string, c *gin.Context) {
// 	var response struct {
// 		Status string `json:"status"`
// 	}

// }
