package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"schedule-api/services"
	"schedule-api/tool"
	"schedule-api/validate"
	"time"

	"github.com/gin-gonic/gin"
)

// MatchRecordList 完赛列表接口
// 从内存缓存读取完赛数据（已整合全民赛程），根据用户兴趣过滤
// 请求参数：next_date(日期), usersports(兴趣标签), callback(JSONP)
func MatchRecordList(c *gin.Context) {
	// 绑定请求参数
	var req validate.MatchRecordListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respond(c, req.Callback, emptyResponse("参数错误"))
		return
	}

	// 校验签名（暂时关闭）
	if !tool.VerifySign(req.NextDate, req.Sign, req.Time, req.AppName) {
		//respond(c, req.Callback, emptyResponse("签名错误"))
		//return
	}

	// 处理日期参数，默认当天
	if req.NextDate == "" || req.NextDate == "__DATE-__" {
		req.NextDate = time.Now().Format("2006-01-02")
	}

	// 从内存缓存获取完赛数据（已整合全民赛程 + 兴趣过滤 + 过滤联赛）
	matchRecordResult := services.GetMatchRecordList(req)

	// 返回响应
	if matchRecordResult == nil {
		respond(c, req.Callback, emptyResponse(""))
		return
	}
	respond(c, req.Callback, *matchRecordResult)
}

// emptyResponse 构造空数据响应
func emptyResponse(msg string) validate.MatchRecordResponse {
	return validate.MatchRecordResponse{
		Msg:    msg,
		NoData: 1,
		List:   []interface{}{},
	}
}

// respond 统一响应函数，固定返回 MatchRecordResponse 格式
func respond(c *gin.Context, callback string, resp validate.MatchRecordResponse) {
	if callback != "" {
		c.Header("Content-Type", "application/javascript; charset=utf-8")
		jsonData, _ := json.Marshal(resp)
		c.String(http.StatusOK, fmt.Sprintf("%s(%s);", callback, string(jsonData)))
		return
	}
	c.JSON(http.StatusOK, resp)
}
