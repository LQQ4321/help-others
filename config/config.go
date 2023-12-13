package config

// 网络配置
const (
	// TODO 再服务器上调试应该修改的地方
	// URL_PORT   = ":80"
	// MYSQL_PORT = ":3306"
	// REDIS_PORT = ":6379"
	// MYSQL_PATH = "help-others-mysql"
	// REDIS_PATH = "help-others-redis"
	URL_PORT   = "127.0.0.1:8080"
	MYSQL_PORT = ":5053"
	REDIS_PORT = ":5054"
	MYSQL_PATH = "175.178.57.154"
	REDIS_PATH = "175.178.57.154"
	MYSQL_DSN  = "root:3515063609563648226@tcp(" +
		MYSQL_PATH + MYSQL_PORT +
		")/?charset=utf8mb4&parseTime=True&loc=Local"
)

// 文件配置
const (
	USER_UPLOAD_FOLDER   = "files/user_upload/" //保存所有的代码
	ORIGIN_CODE_NAME     = "origin.txt"
	PROBLEM_PICTURE_NAME = "problem"
)

// 网站配置
var (
	WebsiteConfig = map[string]int{
		"MaxPublish":        10,     //最大发布数量(如果修改该值后，今天的发布数量已经超额，那么不作删除多余求助的处理)
		"MaxHelp":           10,     //对于每个求助的最大帮助数量
		"MaxSeekHelpPerDay": 3,      //每名用户每天的求助数量
		"MaxComment":        100,    //最大评论数量(对于每个求助和帮助而言)
		"UserBan":           0b1111, //用户权限
		"SeekHelpBan":       0b111,  //求助权限
		"LendHandBan":       0b11,   //帮助权限
		"InitScore":         3,      //用户的初始分数
		"LoginDuration":     24,     //每次的登录时长,单位是小时
	}
)

var (
	Language = map[string]string{
		"c":   "C",
		"cpp": "C++",
		"go":  "Golang",
	}
)

const (
	RETURN_SUCCEED = "succeed"
	RETURN_FAIL    = "fail"
)

const (
	INTERNAL_ERROR = "Server internal error"
)
