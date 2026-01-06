package validate

// MatchRecordListRequest 完赛列表请求参数
type MatchRecordListRequest struct {
	NextDate   string `form:"next_date"`  // 日期，默认当天
	Sign       string `form:"sign"`       // 签名
	Time       int64  `form:"time"`       // 时间戳
	UserSports string `form:"usersports"` // 用户兴趣运动标签，如 "1,2,3"
	AppName    string `form:"appname"`    // 应用名称
	Callback   string `form:"callback"`   // JSONP 回调函数名
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

