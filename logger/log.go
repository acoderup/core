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
var debug = false

func init() {
	debug = os.Getenv("LOGDEBUG") != ""
	logs := logrus.New()
	logs.SetReportCaller(true)

	// 创建日志目录
	if err := ensureLogDirectory(); err != nil {
		logrus.Fatalf("无法创建日志目录: %v", err)
	}

	// 配置日志文件路径
	logFile := "./logs/app.log"
	errorLogFile := "./logs/error.log"

	// 处理日志文件备份
	if err := handleLogBackups(logFile, errorLogFile); err != nil {
		logrus.Fatalf("日志备份处理失败: %v", err)
	}

	// 配置日志轮转
	fileWriter := newLogWriter(logFile)
	errorFileWriter := newLogWriter(errorLogFile)

	// 完全禁用logrus默认输出
	logs.SetOutput(io.Discard)

	// 添加自定义hook处理所有输出
	logs.AddHook(&customHook{
		fileWriter:      fileWriter,
		errorFileWriter: errorFileWriter,
		debug:           debug,
	})

	Logger = logs
}

// ensureLogDirectory 确保日志目录存在
func ensureLogDirectory() error {
	return os.MkdirAll("./logs", 0755)
}

// newLogWriter 创建新的日志写入器
func newLogWriter(filename string) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    100, // MB
		MaxBackups: 3,
		MaxAge:     30, // 天
		Compress:   true,
	}
}

// handleLogBackups 处理日志文件备份
func handleLogBackups(logFile, errorLogFile string) error {
	nowTime := time.Now()
	// 备份主日志文件
	if err := backupIfExists(nowTime, logFile, "app"); err != nil {
		return err
	}

	// 备份错误日志文件
	if err := backupIfExists(nowTime, errorLogFile, "error"); err != nil {
		return err
	}

	// 清理旧备份文件
	return cleanOldBackups()
}

// backupIfExists 如果文件存在则进行备份
func backupIfExists(nowTime time.Time, filePath, prefix string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	backupFile := fmt.Sprintf("./logs/%s_%s.log", prefix, nowTime.Format("20060102_150405"))
	if err := os.Rename(filePath, backupFile); err != nil {
		return fmt.Errorf("无法备份文件 %s: %v", filePath, err)
	}
	return nil
}

// cleanOldBackups 清理旧的备份文件
func cleanOldBackups() error {
	files, err := os.ReadDir("./logs")
	if err != nil {
		return fmt.Errorf("无法读取日志目录: %v", err)
	}

	// 分类收集备份文件
	var appBackups, errorBackups []backupFileInfo

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			logrus.Warnf("无法获取文件信息 %s: %v", file.Name(), err)
			continue
		}

		if matched, _ := filepath.Match("app_*.log", file.Name()); matched {
			appBackups = append(appBackups, backupFileInfo{name: file.Name(), modTime: info.ModTime()})
		} else if matched, _ := filepath.Match("error_*.log", file.Name()); matched {
			errorBackups = append(errorBackups, backupFileInfo{name: file.Name(), modTime: info.ModTime()})
		}
	}

	// 清理旧备份
	if err := cleanBackupGroup(appBackups, "应用日志"); err != nil {
		return err
	}
	if err := cleanBackupGroup(errorBackups, "错误日志"); err != nil {
		return err
	}

	return nil
}

type backupFileInfo struct {
	name    string
	modTime time.Time
}

// cleanBackupGroup 清理一组备份文件
func cleanBackupGroup(files []backupFileInfo, logType string) error {
	if len(files) <= 3 {
		return nil
	}

	// 按修改时间排序(从旧到新)
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	// 保留最新的3个，删除其余的
	for i := 0; i < len(files)-3; i++ {
		if err := os.Remove(filepath.Join("./logs", files[i].name)); err != nil {
			logrus.Warnf("无法删除旧的%s备份 %s: %v", logType, files[i].name, err)
		}
	}

	return nil
}
