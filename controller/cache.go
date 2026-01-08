package controller

import (
	"app/data"
	"app/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CacheStats 缓存统计接口
func CacheStats(c *gin.Context) {
	if c.Query("key") != "hbafbasljfkmg" {
		c.JSON(http.StatusForbidden, gin.H{"msg": "forbidden"})
		return
	}
	c.JSON(http.StatusOK, data.GetCacheStats())
}

// CacheRefresh 手动刷新缓存接口
func CacheRefresh(c *gin.Context) {

	if c.Query("key") != "hbafbasljfkmg" {
		c.JSON(http.StatusForbidden, gin.H{"msg": "forbidden"})
		return
	}

	result := services.RefreshCache(c.Query("type"), c.Query("date"))
	c.JSON(result.Code, result)
}

// CacheData 获取指定日期的完赛缓存数据
func CacheData(c *gin.Context) {
	if c.Query("key") != "hbafbasljfkmg" {
		c.JSON(http.StatusForbidden, gin.H{"msg": "forbidden"})
		return
	}

	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "日期参数错误"})
		return
	}

	cacheData, needReload := data.GetMatchRecordCache(date)
	if needReload || cacheData == nil {
		c.JSON(http.StatusOK, gin.H{"msg": "缓存不存在", "date": date})
		return
	}

	c.JSON(http.StatusOK, cacheData)
}
