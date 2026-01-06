package data

import (
	c "app/conf"
	"github.com/go-redis/redis"
	"log"
	"sync"
)

// Rdb redis客户端
var Rdb *redis.Client

func initRedis(ws *sync.WaitGroup) {
	defer ws.Done()
	Rdb = redis.NewClient(&redis.Options{
		Addr:     c.Config.Redis.Host + ":" + c.Config.Redis.Port,
		Password: c.Config.Redis.Password,
	})
	_, err := Rdb.Ping().Result()

	if err != nil {
		log.Fatalf("第一台Redis connection failed：%q", err.Error())
	}
	c.Config.Redis.Enable = true
	log.Println("初始化redis, 完成", c.Config.Redis.Host)
}
