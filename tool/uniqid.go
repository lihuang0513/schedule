package tool

// 导入所需的包
import (
	// 使用 crypto/rand 以确保随机数的安全性
	"crypto/rand"
	// 用于格式化输出
	"fmt"
	// 导入时间库，用于处理时间相关操作
	"time"
)

// Uniqid 生成一个唯一的标识符字符串。
// 该函数使用当前时间戳和随机字节生成一个唯一标识符，并将其截取为指定长度。
//
// 返回值:
//
//	uniqid: 生成的唯一标识符字符串
func Uniqid() (string, error) {
	const uniqidLength = 23 // 使用常量定义唯一标识符的长度

	// 获取当前时间戳
	timestamp := time.Now().UnixNano()

	// 生成随机字节
	randomBytes := make([]byte, 10)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("生成随机字节时出错: %w", err)
	}

	// 将随机字节转换为十六进制字符串
	randomStr := fmt.Sprintf("%x", randomBytes)

	// 组合时间戳和随机字符串
	uniqid := fmt.Sprintf("%x%x", timestamp, randomStr)

	// 如果唯一标识符的长度大于指定长度，则截取指定长度
	if len(uniqid) > uniqidLength {
		uniqid = uniqid[:uniqidLength]
	}

	return uniqid, nil
}
