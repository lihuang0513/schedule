package controller

import (
	"app/services"
	"app/validate"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RecommendList 全民赛事推荐接口
func RecommendList(c *gin.Context) {
	list := services.GetPgameRecommend()

	resp := validate.PgameRecommendResponse{
		Status: 1,
		Msg:    "成功",
		List:   list,
	}

	c.JSON(http.StatusOK, resp)
}
