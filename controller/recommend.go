package controller

import (
	"app/services"
	"app/tool"
	"app/validate"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RecommendList 全民赛事推荐接口
func RecommendList(c *gin.Context) {
	var req validate.PgameRecommendRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusOK, validate.PgameRecommendResponse{Status: 0, Msg: "参数错误"})
		return
	}

	// 校验签名
	if !tool.VerifySign("", req.Sign, req.Time, req.AppName) {
		c.JSON(http.StatusOK, validate.PgameRecommendResponse{Status: 0, Msg: "签名错误"})
		return
	}

	list := services.GetPgameRecommend()

	resp := validate.PgameRecommendResponse{
		Status: 1,
		Msg:    "成功",
		List:   list,
	}

	c.JSON(http.StatusOK, resp)
}
