package controller

import (
	"app/services"
	"app/tool"
	"app/validate"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// MatchRecordList 完赛列表接口
// GET /match_record/list
func MatchRecordList(c *gin.Context) {
	// 绑定请求参数
	var req validate.MatchRecordListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respondWithError(c, req.Callback, "参数错误")
		return
	}

	// 处理日期参数
	if req.NextDate == "" || req.NextDate == "__DATE-__" {
		req.NextDate = time.Now().Format("2006-01-02")
	}

	// 校验签名
	if !tool.VerifySign(req.NextDate, req.Sign, req.Time, req.AppName) {
		respondWithError(c, req.Callback, "err 3")
		return
	}

	// 获取完赛列表数据
	result := services.GetMatchRecordList(req)

	// 返回结果
	respondWithJSON(c, req.Callback, result)
}

// respondWithJSON 返回JSON响应，支持JSONP
func respondWithJSON(c *gin.Context, callback string, data interface{}) {
	if callback != "" {
		// JSONP 响应
		c.Header("Content-Type", "application/javascript; charset=utf-8")
		c.String(http.StatusOK, fmt.Sprintf("%s(%s);", callback, toJSON(data)))
		return
	}

	// 普通 JSON 响应
	c.JSON(http.StatusOK, data)
}

// respondWithError 返回错误响应
func respondWithError(c *gin.Context, callback string, msg string) {
	resp := validate.MatchRecordResponse{
		Msg:      msg,
		Date:     "",
		DateStr:  "",
		NoData:   1,
		NextDate: "",
		List:     []interface{}{},
	}
	respondWithJSON(c, callback, resp)
}

// toJSON 将数据转换为JSON字符串
func toJSON(data interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(jsonData)
}

