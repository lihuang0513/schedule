package validate

// CacheRefreshResult 缓存刷新结果
type CacheRefreshResult struct {
	Code    int    `json:"-"` // HTTP 状态码，不输出到 JSON
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Date    string `json:"date,omitempty"`
}
