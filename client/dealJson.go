package client

import (
	"context"
	"errors"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/LQQ4321/help-others/config"
	"github.com/LQQ4321/help-others/db"
	"github.com/gin-gonic/gin"
	"github.com/jordan-wright/email"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 1.发送验证码邮件到用户，并将验证码保存到redis中，设置过期时间
// 2.用户打开邮件，拿到验证码，发送到后台进行验证
// 3.后台接收到用户发送过来的验证码，将验证码与redis中暂存的验证码比较
// 4.不存在则表示过期，存在但是不匹配则表示错误，一致则表示验证成功,证明该用户确实是注册账户时拥有该邮箱账号的人
// 5.返回状态码给前端，失败则要继续验证；成功则可以注册账户或者输入新密码了

// {register，mailbox}前端应该检验一下邮箱的格式,或者可能出现的sql注入
// {forgotPassword,mailbox}
func sendVerificationCode(info []string, c *gin.Context) {
	var response struct {
		Status string `json:"status"`
		// 1内部错误 2邮箱已存在(想对于注册) 3邮箱不存在(相对于忘记密码) 4太频繁的请求
		ErrorCode int `json:"errorCode"`
	}
	response.Status = config.RETURN_FAIL
	response.ErrorCode = 1
	_, err := RDB.Get(context.Background(), info[1]).Result()
	if err == redis.Nil {
		var count int64
		err = DB.Model(&db.User{}).Where(&db.User{Mailbox: info[1]}).
			Count(&count).Error
		if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
			logger.Errorln(err)
		} else if info[0] == "register" && count > 0 {
			response.ErrorCode = 2
		} else if info[0] == "forgotPassword" && count <= 0 {
			response.ErrorCode = 3
		} else {
			verificationPassword := db.GenerateRandomString(6)
			err = sendCodeToUser(info[1], verificationPassword)
			if err != nil {
				logger.Errorln(err)
			} else {
				_, err = RDB.Set(context.Background(), info[1], verificationPassword,
					time.Minute*config.VERIFICATION_CODE_VAILD_TIME).Result()
				if err != nil {
					logger.Errorln(err)
				} else {
					response.Status = config.RETURN_SUCCEED
				}
			}
		}
	} else if err != nil {
		logger.Errorln(err)
	} else {
		response.ErrorCode = 4
	}
	c.JSON(http.StatusOK, response)
}

// 向指定邮箱发送验证码
func sendCodeToUser(receivingMailbox, verificationPassword string) error {
	e := email.NewEmail()
	//设置发送方的邮箱
	e.From = config.SENDER_MAILBOX
	// 设置接收方的邮箱（这里是一个字符串数组，什么是可以群发的）
	e.To = []string{receivingMailbox}
	//设置主题
	e.Subject = "You have been received a verification password from help-others"
	//设置文件发送的内容
	e.Text = []byte(verificationPassword)
	//设置服务器相关的配置
	// ""，账号密码，授权码，""
	return e.Send(config.SMTP_SERVER_PATH+config.SMTP_SERVER_PORT, smtp.PlainAuth("",
		config.SENDER_MAILBOX, config.SMTP_SERVER_VERIFICATION_CODE, config.SMTP_SERVER_PATH))
}

// {"username",name,password}
// {"mailbox","mailbox path","password"}
// {authcode,mailbox path,auth code}
func login(info []string, c *gin.Context) {
	var response struct {
		Status       string        `json:"status"`
		ConfigData   []int         `json:"configData"`
		User         db.User       `json:"user"`
		UnsolvedList []db.SeekHelp `json:"unsolvedList"`
		// 1 内部错误 2 验证码过期 3 验证码或密码错误 4 邮箱地址或者用户名不存在
		ErrorCode int `json:"errorCode"`
	}
	response.Status = config.RETURN_FAIL
	response.ConfigData = make([]int, 0)
	response.UnsolvedList = make([]db.SeekHelp, 0)
	response.ErrorCode = 1
	flag := false
	var err error
	if info[0] == "username" {
		err = DB.Model(&db.User{}).Where(&db.User{Name: info[1]}).
			First(&response.User).Error
	} else {
		err = DB.Model(&db.User{}).Where(&db.User{Mailbox: info[1]}).
			First(&response.User).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.ErrorCode = 4
		} else {
			logger.Errorln(err)
		}
	} else {
		if info[0] == "authcode" {
			val, err := RDB.Get(context.Background(), info[1]).Result()
			if err == redis.Nil {
				response.ErrorCode = 2
			} else if err != nil {
				logger.Errorln(err)
			} else if val != info[2] {
				response.ErrorCode = 3
			} else {
				flag = true
			}
		} else {
			if response.User.Password != info[2] {
				response.ErrorCode = 3
			} else {
				flag = true
			}
		}
	}
	if flag {
		flag = false
		configs := []string{"LoginDuration", "MaxUploadFileSize", "MaxUploadImageSize"}
		for i, v := range configs {
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
			if i == len(configs)-1 {
				flag = true
			}
		}
		if flag {
			err := DB.Model(&db.SeekHelp{}).Where("status = ?", 0).
				Limit(50).Find(&response.UnsolvedList).Error
			if err != nil {
				logger.Errorln(err)
			} else {
				response.User.SeekHelp = ""
				response.User.LendHand = ""
				response.Status = config.RETURN_SUCCEED
			}
		}
	}
	c.JSON(http.StatusOK, response)
}

// 忘记密码(不能从前端获取到原本的密码，只能重新弄一个了[还顺便把修改密码的功能间接实现了^_^])
// {mailbox,password,verificationPassword}
func forgotPassword(info []string, c *gin.Context) {
	var response struct {
		Status    string  `json:"status"`
		ErrorCode int     `json:"errorCode"` //1内部错误 2验证码过期 3验证码错误 4邮箱不存在
		User      db.User `json:"user"`
	}
	response.Status = config.RETURN_FAIL
	response.ErrorCode = 1
	val, err := RDB.Get(context.Background(), info[0]).Result()
	if err == redis.Nil {
		response.ErrorCode = 2
	} else if err != nil {
		logger.Errorln(err)
	} else if val != info[2] {
		response.ErrorCode = 3
	} else {
		var count int64
		err = DB.Model(&db.User{}).
			Where(&db.User{Mailbox: info[0]}).Count(&count).Error
		if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
			logger.Errorln(err)
		} else if count <= 0 {
			response.ErrorCode = 4
		} else {
			response.User = db.User{Password: info[1]}
			err = DB.Model(&db.User{}).Where(&db.User{Mailbox: info[0]}).
				Updates(&response.User).Error
			if err != nil {
				logger.Errorln(err)
			} else {
				response.Status = config.RETURN_SUCCEED
			}
		}
	}
	c.JSON(http.StatusOK, response)
}

// 注册账户
// {name,mailbox,password,verificationPassword,registerTime}
func register(info []string, c *gin.Context) {
	var response struct {
		Status    string  `json:"status"`
		ErrorCode int     `json:"errorCode"` //1内部错误 2验证码过期 3验证码错误 4用户名或邮箱已存在
		User      db.User `json:"user"`
	}
	response.Status = config.RETURN_FAIL
	response.ErrorCode = 1
	// 首先应该验证mailbox和对应的verificationPassword在redis中是否存在(验证码过期或错误)
	val, err := RDB.Get(context.Background(), info[1]).Result()
	if err == redis.Nil {
		response.ErrorCode = 2
	} else if err != nil {
		logger.Errorln(err)
	} else if val != info[3] {
		response.ErrorCode = 3
	} else {
		var count int64
		err = DB.Model(&db.User{}).Where(&db.User{Name: info[0]}).
			Or(&db.User{Mailbox: info[1]}).Count(&count).Error
		if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
			logger.Errorln(err)
		} else if count > 0 {
			response.ErrorCode = 4
		} else {
			response.User = db.User{
				Name:     info[0],
				Mailbox:  info[1],
				Password: info[2],
				// TODO fixme 应该从redis中动态获取，而不是从配置文件中静态获取
				Ban:          config.WebsiteConfig["UserBan"],
				Score:        config.WebsiteConfig["InitScore"],
				RegisterTime: info[4],
			}
			err = DB.Model(&db.User{}).Create(&response.User).Error
			if err != nil {
				logger.Errorln(err)
			} else {
				response.Status = config.RETURN_SUCCEED
			}
		}
	}
	c.JSON(http.StatusOK, response)
}

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

// {userId}
func getContributions(info []string, c *gin.Context) {
	var response struct {
		Status        string        `json:"status"`
		SeekHelpList  []db.SeekHelp `json:"seekHelpList"`
		LendHandList  []db.LendHand `json:"lendHandList"`
		SeekHelpList2 []db.SeekHelp `json:"seekHelpList2"`
	}
	response.Status = config.RETURN_FAIL
	response.SeekHelpList = make([]db.SeekHelp, 0)
	response.LendHandList = make([]db.LendHand, 0)
	response.SeekHelpList2 = make([]db.SeekHelp, 0)
	userId, err := strconv.Atoi(info[0])
	if err != nil {
		logger.Errorln(err)
	} else {
		user := db.User{}
		err = DB.Model(&db.User{}).
			Where(&db.User{ID: userId}).First(&user).Error
		if err != nil {
			logger.Errorln(err)
		} else {
			if len(user.SeekHelp) > 0 {
				seekHelpList := strings.Split(user.SeekHelp, "#")
				list := []int{}
				for _, v := range seekHelpList {
					num, err := strconv.Atoi(v)
					if err != nil {
						logger.Errorln(err)
						c.JSON(http.StatusOK, response)
						return
					}
					list = append(list, num)
				}
				err = DB.Model(&db.SeekHelp{}).Where(list).
					Select([]string{"id", "seek_helper_id", "seek_helper_name",
						"upload_time", "language", "score", "like", "status"}).
					Find(&response.SeekHelpList).Error
				if err != nil {
					logger.Errorln(err)
					c.JSON(http.StatusOK, response)
					return
				}
			}
			if len(user.LendHand) > 0 {
				lendHandList := strings.Split(user.LendHand, "#")
				list := []int{}
				for _, v := range lendHandList {
					num, err := strconv.Atoi(v)
					if err != nil {
						logger.Errorln(err)
						c.JSON(http.StatusOK, response)
						return
					}
					list = append(list, num)
				}
				err = DB.Model(&db.LendHand{}).Where(list).
					Select([]string{"id", "seek_help_id", "lend_hander_id",
						"lend_hander_name", "upload_time", "like", "status"}).
					Find(&response.LendHandList).Error
				if err != nil {
					logger.Errorln(err)
					c.JSON(http.StatusOK, response)
					return
				}
			}
			list := []int{}
			for _, v := range response.LendHandList {
				list = append(list, v.SeekHelpId)
			}
			if len(list) > 0 {
				err = DB.Model(&db.SeekHelp{}).Where(list).
					Select([]string{"id", "seek_helper_id", "seek_helper_name",
						"upload_time", "language", "score", "like", "status"}).
					Find(&response.SeekHelpList2).Error
				if err != nil {
					logger.Errorln(err)
					c.JSON(http.StatusOK, response)
					return
				}
			}
			response.Status = config.RETURN_SUCCEED
		}
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
