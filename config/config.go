package config

// 网络配置
const (
	URL_PORT   = "127.0.0.1:8080" //":80"
	MYSQL_PORT = ":5053"          //":3306"
	REDIS_PORT = ":5054"          //":6379"
	MYSQL_PATH = "175.178.57.154" //"help-others-mysql"
	REDIS_PATH = "175.178.57.154" //"help-others-redis"
	MYSQL_DSN  = "root:3515063609563648226@tcp(" +
		MYSQL_PATH + MYSQL_PORT +
		")/?charset=utf8mb4&parseTime=True&loc=Local"
)

// 文件配置
const (
	CODE_FOLDER = "files/codes/" //保存所有的代码
)

// 网站配置
var (
	WebsiteConfig = map[string]int{
		"MaxPublish":           10,   //最大发布数量(如果修改该值后，今天的发布数量已经超额，那么不作删除多余求助的处理)
		"MaxHelp":              10,   //对于每个求助的最大帮助数量
		"MaxComment":           100,  //最大评论数量
		"UserBan":              0b00, //用户权限
		"SeekHelpBan":          0b00, //求助权限
		"LendHandBan":          0b00, //帮助权限
		"InitScore":            3,    //用户的初始分数
		"UserLoginDuration":    1,    //每次的登录时长(用户),单位是小时
		"ManagerLoginDuration": 1,    //每次的登录时长(管理员)
	}
)

const (
	RETURN_SUCCEED = "succeed"
	RETURN_FAIL    = "fail"
)

const (
	INTERNAL_ERROR = "Server internal error"
)
