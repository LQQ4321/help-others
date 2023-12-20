package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/LQQ4321/help-others/config"
	"github.com/LQQ4321/help-others/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// {type,seekOrLendId,text,sendTime,publisher,publisherId}
func sendAComment(info []string, c *gin.Context) {
	var response struct {
		Status  string     `json:"status"`
		Comment db.Comment `json:"comment"`
	}
	response.Status = config.RETURN_SUCCEED
	sendType, err := strconv.Atoi(info[0])
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	seekOrLendId, err := strconv.Atoi(info[1])
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	publisherId, err := strconv.Atoi(info[5])
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	response.Comment = db.Comment{
		Text:         info[2],
		SendTime:     info[3],
		Type:         sendType,
		SeekOrLendId: seekOrLendId,
		Publisher:    info[4],
		PublisherId:  publisherId,
	}
	err = DB.Model(&db.Comment{}).Create(&response.Comment).Error
	if err != nil {
		logger.Errorln(err)
	} else {
		response.Status = config.RETURN_SUCCEED
	}
	c.JSON(http.StatusOK, response)
}

// 点赞后不能取消
// {seekHelp,seekHelpId,userId} 前端负责检验seekHelperId和userId是不是同一个人
// {lendHand,lendHandId,userId} 前后端一起负责检验seekHelperId和userId是不是同一个人(防止数据不同步)
// TODO 如果是求助者点的赞，应该把score给对应的帮助者
func likeOperate(info []string, c *gin.Context) {
	var response struct {
		Status string `json:"status"`
	}
	response.Status = config.RETURN_FAIL
	dbId, err := strconv.Atoi(info[1])
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	userId, err := strconv.Atoi(info[2])
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	user := db.User{}
	err = DB.Model(&db.User{}).Where(&db.User{ID: userId}).
		First(&user).Error
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, response)
		return
	}
	if info[0] == "seekHelp" {
		err = DB.Transaction(func(tx *gorm.DB) error {
			seekHelp := db.SeekHelp{}
			err := tx.Model(&db.SeekHelp{}).
				Where(&db.SeekHelp{ID: dbId}).First(&seekHelp).Error
			if err != nil {
				return err
			}
			seekHelp.Like++
			if len(user.SeekHelpLikeList) != 0 {
				user.SeekHelpLikeList += "#"
			}
			user.SeekHelpLikeList += info[1]
			err = tx.Model(&db.SeekHelp{}).
				Where(&db.SeekHelp{ID: seekHelp.ID}).Updates(&seekHelp).Error
			if err != nil {
				return err
			}
			return tx.Model(&db.User{}).
				Where(&db.User{ID: user.ID}).Updates(&user).Error
		})
		if err != nil {
			logger.Errorln(err)
		} else {
			response.Status = config.RETURN_SUCCEED
		}
	} else if info[0] == "lendHand" {
		err = DB.Transaction(func(tx *gorm.DB) error {
			lendHand := db.LendHand{}
			err := tx.Model(&db.LendHand{}).
				Where(&db.LendHand{ID: dbId}).First(&lendHand).Error
			if err != nil {
				return err
			}
			seekHelp := db.SeekHelp{}
			err = tx.Model(&db.SeekHelp{}).
				Where(&db.SeekHelp{ID: lendHand.SeekHelpId}).First(&seekHelp).Error
			if err != nil {
				return err
			}
			// 点赞的人和求助的人是同一个人，说明求助者采纳了这份建议
			if seekHelp.SeekHelperId == info[2] {
				seekHelp.Status = 1
				lendHand.Status = 1
				lendHander := db.User{}
				// 这里的错误不是数据库操作引起的，会不会有什么问题？
				lendHanderId, err := strconv.Atoi(lendHand.LendHanderId)
				if err != nil {
					return err
				}
				err = tx.Model(&db.User{}).
					Where(&db.User{ID: lendHanderId}).First(&lendHander).Error
				if err != nil {
					return err
				}
				lendHander.Score += seekHelp.Score
				err = tx.Model(&db.SeekHelp{}).
					Where(&db.SeekHelp{ID: seekHelp.ID}).Updates(&seekHelp).Error
				if err != nil {
					return err
				}
				err = tx.Model(&db.User{}).
					Where(&db.User{ID: lendHander.ID}).Updates(&lendHander).Error
				if err != nil {
					return err
				}
			}
			lendHand.Like++
			if len(user.LendHandLikeList) != 0 {
				user.LendHandLikeList += "#"
			}
			user.LendHandLikeList += info[1]
			err = tx.Model(&db.LendHand{}).
				Where(&db.LendHand{ID: lendHand.ID}).Updates(&lendHand).Error
			if err != nil {
				return err
			}
			return tx.Model(&db.User{}).
				Where(&db.User{ID: user.ID}).Updates(&user).Error
		})
		if err != nil {
			logger.Errorln(err)
		} else {
			response.Status = config.RETURN_SUCCEED
		}
	}
	c.JSON(http.StatusOK, response)
}

// {filePath}
func downloadFile(info []string, c *gin.Context) {
	c.File(info[0])
}

// {"login",name,password}
func verifyUser(info []string, c *gin.Context) {
	if info[0] == "login" {
		var response struct {
			Status       string        `json:"status"`
			ConfigData   []int         `json:"configData"`
			User         db.User       `json:"user"`
			UnsolvedList []db.SeekHelp `json:"unsolvedList"`
		}
		response.Status = config.RETURN_FAIL
		response.ConfigData = make([]int, 0)
		response.UnsolvedList = make([]db.SeekHelp, 0)
		err := DB.Model(&db.User{}).
			Where(&db.User{Name: info[1], Password: info[2]}).
			First(&response.User).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Errorln(fmt.
					Errorf("The user name does not exist or the password is incorrect"))
			} else {
				logger.Errorln(err)
			}
		} else {
			configs := []string{"MaxSeekHelpPerDay", "LoginDuration"}
			for _, v := range configs {
				result, err := RDB.Get(context.Background(), v).Result()
				if err != nil {
					logger.Errorln(err)
					break
				}
				num, err := strconv.Atoi(result)
				if err != nil {
					logger.Errorln(err)
					break
				}
				response.ConfigData = append(response.ConfigData, num)
			}
			err = DB.Model(&db.SeekHelp{}).Where("status = ?", 0).
				Limit(50).Find(&response.UnsolvedList).Error
			if err != nil {
				logger.Errorln(err)
			} else {
				response.Status = config.RETURN_SUCCEED
			}
		}
		c.JSON(http.StatusOK, response)
	} else if info[0] == "register" {

	} else if info[0] == "retrievePassword" {

	}
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
		}
		response.Status = config.RETURN_FAIL
		response.CommentList = make([]db.Comment, 0)
		helpType, err := strconv.Atoi(info[1])
		if err != nil {
			logger.Errorln(err)
			c.JSON(http.StatusOK, response)
			return
		}
		seekOrLendId, err := strconv.Atoi(info[2])
		if err != nil {
			logger.Errorln(err)
			c.JSON(http.StatusOK, response)
			return
		}
		err = DB.Model(&db.Comment{}).
			Where("type = ? AND seek_or_lend_id = ?", helpType, seekOrLendId).
			Find(&response.CommentList).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
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
				response.CodePath = lendHand.DiffPath
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
