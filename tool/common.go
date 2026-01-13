package tool

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
)

// 创建文件夹
func MakeDir(path string) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// 创建文件夹
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Fatal("创建文件失败" + err.Error())
		}
	}
}

// InArray 判断数据是否在切片数组中
func InArray[T int | string](value T, array []T) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}

func Md5(str string) string {
	h := md5.New()
	_, err := io.WriteString(h, str)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
