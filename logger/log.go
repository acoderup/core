package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

var Logger *logrus.Logger
var debug = true // 控制是否输出到控制台
func init() {
	logs := logrus.New()

	// 创建日志目录
	if err := os.MkdirAll("./logs", 0755); err != nil {
		logrus.Fatal("Failed to create log directory:", err)
	}

	// 配置日志文件路径
	logFile := "./logs/app.log"

	// 如果日志文件已存在
	if _, err := os.Stat(logFile); err == nil {
		// 首先检查并清理旧的备份文件
		files, err := os.ReadDir("./logs")
		if err != nil {
			logrus.Fatalf("读取日志目录失败: %v", err)
		}

		// 收集所有备份文件
		var backupFiles []string
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if matched, _ := filepath.Match("app_*.log", file.Name()); matched {
				backupFiles = append(backupFiles, file.Name())
			}
		}

		// 如果备份文件超过3个，按修改时间排序并删除最旧的
		if len(backupFiles) >= 3 {
			// 按修改时间排序
			sort.Slice(backupFiles, func(i, j int) bool {
				info1, _ := os.Stat(filepath.Join("./logs", backupFiles[i]))
				info2, _ := os.Stat(filepath.Join("./logs", backupFiles[j]))
				return info1.ModTime().Before(info2.ModTime())
			})

			// 删除最旧的文件，只保留最新的3个
			for i := 0; i < len(backupFiles)-2; i++ {
				if err := os.Remove(filepath.Join("./logs", backupFiles[i])); err != nil {
					logrus.Warnf("删除旧备份文件失败: %v", err)
				}
			}
		}

		// 创建新的备份
		backupFile := fmt.Sprintf("./logs/app_%s.log", time.Now().Format("20060102_150405"))
		if err := os.Rename(logFile, backupFile); err != nil {
			logrus.Fatalf("日志备份失败: %v", err)
		}
	}

	// 配置日志轮转
	fileWriter := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     30,
		Compress:   true,
	}

	// 完全禁用logrus默认输出
	logs.SetOutput(io.Discard)

	// 添加自定义hook处理所有输出
	logs.AddHook(&customHook{
		fileWriter: fileWriter,
		debug:      debug,
	})

	Logger = logs
}
