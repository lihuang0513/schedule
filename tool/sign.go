package tool

import (
	config "app/conf"
	"crypto/sha256"
	"fmt"
	"time"
)

// VerifySign 验证签名
// nextDate: 日期参数
// sign: 客户端传递的签名
// timestamp: 时间戳
// appName: 应用名称 (zhibo8 或 dongqiudi)
func VerifySign(nextDate, sign string, timestamp int64, appName string) bool {
	// 检查签名是否过期
	if time.Now().Unix()-timestamp > config.SignExpireSec {
		return false
	}

	// 生成签名并比较
	expectedSign := GenerateSign(nextDate, timestamp, appName)
	return expectedSign == sign
}

// GenerateSign 生成签名
func GenerateSign(nextDate string, timestamp int64, appName string) string {
	timeStr := fmt.Sprintf("%d", timestamp)

	if appName == config.DQD {
		// dongqiudi 使用 SHA256
		data := nextDate + config.DQDSignKey + timeStr
		hash := sha256.Sum256([]byte(data))
		return fmt.Sprintf("%x", hash)
	}

	// zhibo8 使用 MD5
	return Md5(nextDate + config.SignKey + timeStr)
}

