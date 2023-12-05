package db

// help others 网站
// 这是一个debug网站，想要参与的用户应该首先注册一个用户名唯一的账号。
// 网站的迭代是按照天来计算的，每一天都有一定的求助上限(源源不断的求助会减弱帮助者的成就感)，
// 管理员可以设置网站的权限，然后控制用户的操作行为，

// 1.这是一个帮助新手查找代码中未知错误的平台，所以想要帮助新手的用户提交的代码应该是在原本的代码的基础上进行修改，
// 而不是重新写一份正确的代码。

// 2.每一份求助代码的悬赏分数只能够提供给一名解决问题的施助者，这取决于求助者，悬赏分数已经给予，不能再撤销

// 3.每名用户一天只能发布一次求助，但是可以发布多次求助。对每个求助只能发布一次代码

// 4.用户发布了一次求助，不管有没有任帮助，他的总分数都要减去本次求助的悬赏分数,
// 这是为了避免个别求助者不将悬赏分数下发给帮助者。当然，求助者不能自己帮助自己

// 5.关于权限的统一说明，1 表示禁止 0 表示允许,下面的权限默认由低位到高位，最后用户具有的权限取并集
// 5.1 用户权限 发布求助，发布帮助，查看求助内容，评论
// 5.2 求助权限 发布帮助，查看求助内容，评论
// 5.3 帮助权限 查看帮助内容，评论

// 6.为了保证用户的唯一性，用户的名称应该唯一，并且不能携带空格

type User struct {
	ID        int    `gorm:"primaryKey"`
	Name      string //用户名
	Password  string //登录密码
	IsManager bool   //是否是管理员
	SeekHelp  string //求助他人的列表
	LendHand  string //帮助他人的列表(每个人对于一次求助只能发布一份代码)
	Ban       int    //用户权限
	Score     int    //当前用户拥有的分值，初始是 3，分值为零时不允许发布求助
}

// 求助
type SeekHelp struct {
	ID          int    `gorm:"primaryKey"`
	ProblemLink string //题目链接(ProblemLink和ImagePath两者至少要有一个)
	ImagePath   string //题目的截图地址(目前仅限上传一张图片好了)
	TopicRemark string //题目文字描述或备注(不能为空，信息当然是越多越好)
	UploadTime  string //提交日期 2023-01-10 09:00
	CodePath    string //代码保留的位置 files/date/id
	Language    string //代码的语言类型
	MaxHelp     int    //最大帮助数量
	MaxComment  int    //最大评论数量
	Score       int    //悬赏的分数
	Like        int    //点赞数
	Ban         int    //该条求助的权限
	Status      int    //当前代码的状态，如 0 未debug 1 成功debug 2 debug失败 等
}

// 帮助
type LendHand struct {
	ID         int    `gorm:"primaryKey"`
	SeekHelpId int    //被帮助的是哪条求助
	UploadTime string //发布时间
	Remark     string //附带的解释信息
	CodePath   string //代码地址
	MaxComment int    //最大评论数量
	Like       int    //点赞数
	Ban        int    //该条帮助的权限
	Status     int    //当前代码的状态，如 0 没有评价信息，只是上传成功 1 得到求助者的肯定 2 得到求助者的否定
}

type Comment struct {
	ID          int    `gorm:"primaryKey"`
	Text        string //评论内容
	SendTime    string //发布时间
	Type        int    //评论的地方 0 求助页面 1 帮助页面
	HelpId      int    //评论的是那一条求组或帮助
	Publisher   string //发布人姓名
	PublisherId int    //发布人在数据库表User中的id
	Like        int    //点赞数
	Ban         int    //评论的权限 0 正常显示 1 折叠该评论
}
