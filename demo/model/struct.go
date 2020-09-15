package model

// 用户信息
type User struct {
	Id          int    `bson:"_id" json:"id"`                 // 用戶id,自增
	Phone       string `bson:"phone" json:"phone"`            // 手机(唯一,用于登录)
	UserName    string `bson:"user_name" json:"user_name"`    // 用户名(唯一,可用于登录)
	NickName    string `bson:"nick_name" json:"nick_name"`    // 昵称(显示)
	RealName    string `bson:"real_name" json:"real_name"`    // 真实姓名
	Password    string `bson:"password" json:"-"`             // 密码hex
	Salt        string `bson:"salt" json:"-"`                 // 密码盐
	Email       string `bson:"email" json:"email"`            // 邮箱
	Remark      string `bson:"remark" json:"remark"`          // 备注
	LoginFailed int    `bson:"login_failed" json:"-"`         // 错误登录次数
	LastLogin   int    `bson:"last_login" json:"last_login"`  // 最后登录时间
	LastIp      string `bson:"last_login_ip" json:"last_ip" ` // 最后登录ip
	Created     int    `bson:"created" json:"created"`        // 创建时间
}

// 用户接口操作日志
type ApiLog struct {
	Id        int         `bson:"_id" json:"id"` // id 自增
	Uid       int         `json:"uid" bson:"uid"`
	Ip        string      `json:"ip" bson:"ip"`
	ApiPath   string      `json:"api_path" bson:"api_path"`
	ApiMethod string      `json:"api_method" bson:"api_method"`
	ApiName   string      `json:"api_name" bson:"api_name"`
	Token     string      `json:"token" bson:"token"`         // token
	ReqQuery  interface{} `json:"req_query" bson:"req_query"` // request query
	ReqBody   interface{} `json:"req_body" bson:"req_body"`   // request body
	OldData   interface{} `json:"old_data" bson:"old_data"`   // 老数据,修改情况
	RespData  interface{} `json:"resp_data" bson:"resp_data"` // 响应数据
	Remark    string      `json:"remark" bson:"remark"`
	Success   bool        `json:"success" bson:"success"`
	Created   int         `json:"created" bson:"created"`
}
