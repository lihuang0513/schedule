package data

import (
	"log"
	c "schedule-api/conf"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

// Rdb redis客户端
var Rdb *redis.Client

func initRedis(ws *sync.WaitGroup) {
	defer ws.Done()
	Rdb = redis.NewClient(&redis.Options{
		Addr:         c.Config.Redis.Host + ":" + c.Config.Redis.Port,
		Password:     c.Config.Redis.Password,
		PoolSize:     100,             // 连接池大小
		MinIdleConns: 10,              // 最小空闲连接
		DialTimeout:  5 * time.Second, // 连接超时
		ReadTimeout:  3 * time.Second, // 读取超时
		WriteTimeout: 3 * time.Second, // 写入超时
	})
	_, err := Rdb.Ping().Result()

	if err != nil {
		log.Fatalf("第一台Redis connection failed：%q", err.Error())
	}
	c.Config.Redis.Enable = true
	log.Println("初始化redis, 完成", c.Config.Redis.Host)
}
