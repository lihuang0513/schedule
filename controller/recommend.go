package controller

import (
	"net/http"
	"schedule-api/services"
	"schedule-api/tool"
	"schedule-api/validate"

	"github.com/gin-gonic/gin"
)

// RecommendList 全民赛事推荐接口
// 客户端可通过 pgame_league_ids 参数过滤指定联赛
func RecommendList(c *gin.Context) {

	var req validate.PgameRecommendRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusOK, []validate.PgameRecommendDateItem{})
		return
	}

	// 校验签名
	if !tool.VerifySign("", req.Sign, req.Time, req.AppName) {
		//c.JSON(http.StatusOK, []validate.PgameRecommendDateItem{})
		//return
	}

	// 获取格式化后的数据，支持按联赛ID过滤
	list := services.GetPgameRecommendFormatted(req.PgameLeagueIds)

	c.JSON(http.StatusOK, list)
}
