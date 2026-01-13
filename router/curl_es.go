package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	config "schedule-api/conf"
	"schedule-api/data"
	"time"
)

// LogEntry 用于记录日志条目，包含时间戳、日志级别、消息内容和项目信息。
type LogEntry struct {
	// Timestamp 日志条目的时间戳
	Timestamp string `json:"timestamp"`
	// Level 日志级别，如 "INFO", "ERROR" 等
	Level string `json:"level"`
	// Message 日志消息内容
	Message string `json:"message"`
	// Project 项目名称
	Project string `json:"project"`
}

// bugToES 将错误日志发送到Elasticsearch
func bugToES(errLog string) {
	logEntry := LogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "ERROR",
		Message:   fmt.Sprintf("%v", errLog),
		Project:   "SCHEDULE",
	}
	// 替换为你的Elasticsearch URL 和 索引名
	sendLogToES(logEntry, config.Config.ElasticSearch.Host+":"+config.Config.ElasticSearch.Port,
		"bug-text", config.Config.ElasticSearch.User, config.Config.ElasticSearch.Password)
}

// sendLogToES 将日志条目发送到Elasticsearch。
// logEntry 是要发送的日志条目。
// esURL 是Elasticsearch服务器的URL。
// index 是要写入的日志索引。
// username 和 password 用于Elasticsearch的认证。
func sendLogToES(logEntry LogEntry, esURL, index, username, password string) {
	// 将日志转换为JSON
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		return
	}

	// 构建请求
	url := "http://" + fmt.Sprintf("%s/%s/_doc", esURL, index)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/json")
	// 添加认证信息
	request.SetBasicAuth(username, password)

	// 发送请求
	client := &http.Client{
		Timeout: time.Second * 3, // 设置超时时间为3秒
	}
	response, err := client.Do(request)
	if err != nil {
		data.Logger.Println("Elasticsearch信息发送失败: " + err.Error() + "\n")
		return
	}
	defer response.Body.Close()
}
