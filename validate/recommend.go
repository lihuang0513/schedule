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
	Sign           string `form:"sign"`             // 签名
	Time           int64  `form:"time"`             // 时间戳
	AppName        string `form:"appname"`          // 应用名称
	PgameLeagueIds string `form:"pgame_league_ids"` // 联赛ID列表，逗号分隔
}

// PgameRecommendCache 推荐数据缓存结构
type PgameRecommendCache struct {
	Data     map[string]interface{} `json:"data"`
	UpdateAt time.Time              `json:"update_at"`
}

// PgameRecommendDateItem 按日期分组的赛事数据
type PgameRecommendDateItem struct {
	FormatDate string        `json:"formatDate"` // 格式化日期 2026-01-12
	Date       string        `json:"date"`       // 显示日期 01月12日 星期一
	List       []interface{} `json:"list"`       // 赛事列表
}
