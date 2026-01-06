package tool

import (
	config "app/conf"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
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

func Replace(a1, a2 []string) []string {
	r := make([]string, 2*len(a1))
	for i, e := range a1 {
		r[i*2] = e
		r[i*2+1] = a2[i]
	}
	return r
}

// GetDeviceParam 获取设备参数
func GetDeviceParam(zbbDid string) (map[string]any, error) {
	// 初始化返回参数
	param := make(map[string]any)
	// 设备信息 base64解码
	decodeByte, err := base64.StdEncoding.DecodeString(zbbDid)
	if err != nil {
		return param, err
	}
	// 对设备信息进行aes解码
	dec, err := AesDecrypt(decodeByte, []byte(config.ZbbDidSalt))
	if err != nil {
		return param, err
	}
	// 参数解密
	dec = PKCS7UnPadding(dec)
	// 对解密后得到大参数做转换
	err = json.Unmarshal(dec, &param)
	// 返回设备信息
	return param, err
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

// ArrayToMap 数组转Map对象
func ArrayToMap(sliceData []string) map[string]int {
	// 初始化切片标签
	mapData := make(map[string]int)
	// 循环设置用户的兴趣标签对象
	for index, item := range sliceData {
		if _, ok := mapData[item]; !ok {
			mapData[item] = index
		}
	}
	return mapData
}

// ConvertVersionToInt 用于版本号转int
func ConvertVersionToInt(version string) int {
	// 移除字符串中的所有"."字符
	cleanVersion := strings.Replace(version, ".", "", -1)

	// 将清洁后的字符串转换为int
	intVersion, err := strconv.Atoi(cleanVersion)
	if err != nil {
		return 0
	}

	return intVersion
}

func Md5(str string) string {
	h := md5.New()
	_, err := io.WriteString(h, str)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// 获取感兴趣运动标签对象
func getUserSportLabelMap(userSportsStr string) map[string]int {
	// 根据"，"把用户的兴趣字符串转成切片
	userSportSlice := strings.Split(userSportsStr, ",")
	// 循环设置添加用户的兴趣标签
	var labelArr []string
	for _, userSport := range userSportSlice {
		switch userSport {
		case "1":
			labelArr = append(labelArr, "足球")
		case "2":
			labelArr = append(labelArr, "篮球")
		case "3":
			labelArr = append(labelArr, "NBA")
		case "4":
			labelArr = append(labelArr, "电竞")
		case "41":
			labelArr = append(labelArr, "英雄联盟")
		case "42":
			labelArr = append(labelArr, "DOTA2")
		case "43":
			labelArr = append(labelArr, "绝地求生")
		case "44":
			labelArr = append(labelArr, "王者荣耀")
		case "45":
			labelArr = append(labelArr, "无畏契约")
		case "51":
			labelArr = append(labelArr, "F1")
		case "52":
			labelArr = append(labelArr, "网球")
		case "53":
			labelArr = append(labelArr, "斯诺克")
		case "54":
			labelArr = append(labelArr, "NFL")
		case "55":
			labelArr = append(labelArr, "MLB")
		case "56":
			labelArr = append(labelArr, "NHL")
		case "57":
			labelArr = append(labelArr, "拳击")
		case "58":
			labelArr = append(labelArr, "UFC")
		case "59":
			labelArr = append(labelArr, "高尔夫")
		case "60":
			labelArr = append(labelArr, "田径")
		case "61":
			labelArr = append(labelArr, "排球")
		case "62":
			labelArr = append(labelArr, "羽毛球")
		case "63":
			labelArr = append(labelArr, "乒乓球")
		default:
			labelArr = append(labelArr, "其他", "综合")
		}
	}
	// 转换成标签map并返回
	return ArrayToMap(labelArr)
}
