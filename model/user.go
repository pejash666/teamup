package model

// BasicUser 预计在中间件读取数据库加载用户基本信息
type BasicUser struct {
	OpenID string `json:"open_id"` // 在一个小程序下，一个用户的唯一标识
	//UserID     uint   `json:"user_id"`  // 自行维护的User库主键 （一个用户openid会对应多个sport_type的多条记录）
	UnionID    string `json:"union_id"` // 同一个微信开放平台账号，一个用户的唯一标识
	SessionKey string `json:"session_key"`
}
