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
			response.Status = config.RETURN_SUCCEED
		}
		c.JSON(http.StatusOK, response)
		logger.Infoln(response.Status)
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
