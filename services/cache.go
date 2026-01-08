package services

import (
	"app/data"
	"app/validate"
	"net/http"
	"time"
)

// RefreshCache 刷新指定类型的缓存
// cacheType: recommend（推荐）、finished（完赛）
// date: 完赛类型必传，格式 2006-01-02
func RefreshCache(cacheType, date string) validate.CacheRefreshResult {
	switch cacheType {
	case "recommend":
		RefreshPgameRecommendCache()
		return validate.CacheRefreshResult{Code: http.StatusOK, Success: true, Msg: "推荐缓存刷新成功"}
	case "finished":
		if date == "" {
			return validate.CacheRefreshResult{Code: http.StatusBadRequest, Success: false, Msg: "日期错误"}
		}
		if _, err := time.Parse("2006-01-02", date); err != nil {
			return validate.CacheRefreshResult{Code: http.StatusBadRequest, Success: false, Msg: "日期错误."}
		}
		cacheData := FetchDayMatchRecord(date)
		data.SetMatchRecordCache(date, cacheData)
		return validate.CacheRefreshResult{Code: http.StatusOK, Success: true, Msg: "完赛缓存刷新成功", Date: date}
	default:
		return validate.CacheRefreshResult{Code: http.StatusBadRequest, Success: false, Msg: "参数错误"}
	}
}
