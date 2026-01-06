package controller

import (
	"app/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RecommendList(c *gin.Context) {
	data, err := services.GetPGameList(c)
	if err != nil {
		c.JSON(http.StatusOK, err.Error())
		return
	}

	// 返回偏好推荐结果
	c.JSON(http.StatusOK, data)

	return
}
