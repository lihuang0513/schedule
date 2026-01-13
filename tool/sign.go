package tool

import (
	"crypto/sha256"
	"fmt"
	config "schedule-api/conf"
	"time"
)

// VerifySign 验证签名
// data: 签名数据（可为空）
// sign: 客户端传递的签名
// timestamp: 时间戳
// appName: 应用名称 (zhibo8 或 dongqiudi)
func VerifySign(data, sign string, timestamp int64, appName string) bool {
	// 检查签名是否过期
	if time.Now().Unix()-timestamp > config.SignExpireSec {
		return false
	}

	// 生成签名并比较
	expectedSign := GenerateSign(data, timestamp, appName)
	return expectedSign == sign
}

// GenerateSign 生成签名
func GenerateSign(data string, timestamp int64, appName string) string {
	timeStr := fmt.Sprintf("%d", timestamp)

	if appName == config.DQD {
		// dongqiudi 使用 SHA256
		str := data + config.DQDSignKey + timeStr
		hash := sha256.Sum256([]byte(str))
		return fmt.Sprintf("%x", hash)
	}

	// zhibo8 使用 MD5
	return Md5(data + config.SignKey + timeStr)
}
