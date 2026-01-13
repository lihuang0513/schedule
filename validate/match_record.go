package validate

import "time"

// MatchRecordListRequest 完赛列表请求参数
type MatchRecordListRequest struct {
	NextDate       string `form:"next_date"`        // 日期，默认当天
	Sign           string `form:"sign"`             // 签名
	Time           int64  `form:"time"`             // 时间戳
	UserSports     string `form:"usersports"`       // 用户兴趣运动标签，如 "1,2,3"
	AppName        string `form:"appname"`          // 应用名称
	Callback       string `form:"callback"`         // JSONP 回调函数名
	PgameLeagueIds string `form:"pgame_league_ids"` // 联赛ID列表，逗号分隔
}

// DayMatchRecordCache 单日完赛缓存数据结构
type DayMatchRecordCache struct {
	Date     string        `json:"date"`
	DateStr  string        `json:"date_str"`
	List     []interface{} `json:"list"`
	UpdateAt time.Time     `json:"update_at"`
}

// MatchRecordResponse 完赛列表响应
type MatchRecordResponse struct {
	Msg      string        `json:"msg,omitempty"`
	Date     string        `json:"date"`
	DateStr  string        `json:"date_str"`
	NoData   int           `json:"no_data,omitempty"`
	NextDate string        `json:"next_date"`
	List     []interface{} `json:"list"`
}

// DayScheduleData 每日赛程静态数据结构
type DayScheduleData struct {
	Date    string                   `json:"date"`
	DateStr string                   `json:"date_str"`
	List    []map[string]interface{} `json:"list"`
}

// PgameLeagueData Redis 中联赛数据结构
type PgameLeagueData struct {
	Date    string                   `json:"date"`
	DateStr string                   `json:"date_str"`
	List    []map[string]interface{} `json:"list"`
}

// MatchListFilter 赛事列表过滤参数
type MatchListFilter struct {
	UserSportsLabels []string            // 用户兴趣标签
	LeagueIdSet      map[string]struct{} // 联赛ID集合
}
