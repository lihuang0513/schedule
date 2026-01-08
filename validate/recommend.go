package validate

import "time"

type PGameListRequestParams struct {
	DeviceId    string `json:"device_id,omitempty"`
	Platform    string `json:"os,omitempty"`
	VersionCode int    `json:"version_code,omitempty"`
	UserSports  string `json:"usersports,omitempty"`
	IsDebug     bool   `json:"is_debug,omitempty"`
	RankKeys    string `json:"rank_keys,omitempty"`
}

// PgameRecommendRequest 推荐接口请求参数
type PgameRecommendRequest struct {
	Sign    string `form:"sign"`    // 签名
	Time    int64  `form:"time"`    // 时间戳
	AppName string `form:"appname"` // 应用名称
}

// PgameRecommendCache 推荐数据缓存结构
type PgameRecommendCache struct {
	Data     map[string]interface{} `json:"data"`
	UpdateAt time.Time              `json:"update_at"`
}

// PgameRecommendResponse 全民赛事推荐响应
type PgameRecommendResponse struct {
	Status int                    `json:"status"`
	Msg    string                 `json:"msg"`
	List   map[string]interface{} `json:"list"`
}
