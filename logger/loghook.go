package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// customHook 自定义hook处理所有日志输出
type customHook struct {
	fileWriter io.Writer
	debug      bool
}

func (h *customHook) Fire(entry *logrus.Entry) error {
	// 文件输出 - JSON格式
	jsonFormatter := &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := filepath.Base(f.File)
			return "", fmt.Sprintf("%s:%d", filename, f.Line)
		},
	}
	if jsonData, err := jsonFormatter.Format(entry); err == nil {
		h.fileWriter.Write(jsonData)
	}

	// 控制台输出 - 彩色文本格式(仅在debug模式下)
	if h.debug {
		textFormatter := &logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := filepath.Base(f.File)
				return "", fmt.Sprintf("%s:%d", filename, f.Line)
			},
		}
		if textData, err := textFormatter.Format(entry); err == nil {
			os.Stdout.Write(textData)
		}
	}

	return nil
}

func (h *customHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
