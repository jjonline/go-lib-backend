package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"
)

type dailySizeRotateWriter struct {
	mu          sync.Mutex
	currentFile *os.File
	currentDate string
	currentSize int64
	currentSeq  int
	maxSize     int64  // 单个日志文件体积
	maxDays     int64  // 最大保留天数
	basePath    string // 日志文件路径
}

func newDailySizeRotateWriter(basePath string, maxSize, maxDays int64) *dailySizeRotateWriter {
	return &dailySizeRotateWriter{
		basePath: basePath,
		maxSize:  maxSize,
		maxDays:  maxDays,
	}
}

func (w *dailySizeRotateWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	today := now.Format("20060102")

	rotate := false
	if w.currentFile == nil {
		rotate = true
	} else if today != w.currentDate {
		rotate = true
	} else if w.maxSize > 0 && (w.currentSize+int64(len(p))) > w.maxSize {
		// maxSize为0表示不切割
		rotate = true
	}

	if rotate {
		if err := w.rotate(today); err != nil {
			return 0, err
		}
	}

	n, err = w.currentFile.Write(p)
	if err != nil {
		return n, err
	}
	w.currentSize += int64(n)
	return n, nil
}

// rotate 轮转日志文件，变更日志文件名 & 清理历史日志文件
//   - newDate 20250527
func (w *dailySizeRotateWriter) rotate(newDate string) error {
	if w.currentFile != nil {
		if err := w.currentFile.Close(); err != nil {
			return err
		}
	}

	var seq int
	var err error

	if newDate != w.currentDate {
		seq, err = findMaxSeq(newDate, w.basePath)
		if err != nil {
			return err
		}

		// +++++++++++++++++++++++++++++++++++
		// 刚启动时，找到的最大seg可能没写满继续使用
		// +++++++++++++++++++++++++++++++++++
		if w.currentFile == nil {
			existOne := w.genFileName(newDate, seq)
			if checkFileExist(existOne) {
				f, err := os.OpenFile(existOne, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
				if err != nil {
					return err
				}
				if info, err := f.Stat(); err == nil && (w.maxSize == 0 || info.Size() < w.maxSize) {
					w.currentFile = f
					w.currentDate = newDate
					w.currentSize = info.Size() // reset exist size
					w.currentSeq = seq
					return nil
				}

				// 满了或其他缘故，关闭这个文件递增seq
				_ = f.Close()
			}
		}

		seq += 1
	} else {
		seq = w.currentSeq + 1
	}

	var (
		filename = w.genFileName(newDate, seq)
		dir      = filepath.Dir(filename)
	)

	if err = os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 日志轮转新建文件时，触发历史文件删除
	if w.maxDays > 0 && !checkFileExist(filename) {
		go func() {
			_ = w.removeHistoryFile()
		}()
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	w.currentFile = f
	w.currentDate = newDate
	w.currentSize = 0
	w.currentSeq = seq

	return nil
}

// genFileName 生成文件名
func (w *dailySizeRotateWriter) genFileName(ymdDate string, seq int) string {
	return fmt.Sprintf("%s/%s-%d.log", w.basePath, ymdDate, seq)
}

// removeHistoryFile 清理历史日志文件
func (w *dailySizeRotateWriter) removeHistoryFile() error {
	var (
		pattern  = fmt.Sprintf("%s/*-*.log", w.basePath)
		re       = regexp.MustCompile(`^(\d+)-(\d+)\.log$`)
		bDate, _ = strconv.Atoi(time.Now().AddDate(0, 0, -int(w.maxDays)).Format("20060102"))
	)

	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		matches := re.FindStringSubmatch(filepath.Base(file))
		// {{20250527-9.log , 20250527 ,9}}
		if matches == nil {
			continue
		}
		cDate, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}

		// 文件名 Ymd < 最大保留日的Ymd，执行删除
		if cDate < bDate {
			slog.Info("remove out of maximum retention date history file", slog.String("file", file))
			_ = os.Remove(file)
		}
	}

	return nil
}

// findMaxSeq 当前日志存储文件系统里获取到指定日期的最大seg
//   - date      20250527
//   - basePath  ./runtime/
func findMaxSeq(date, basePath string) (int, error) {
	pattern := fmt.Sprintf("%s/%s-*.log", basePath, date)
	files, err := filepath.Glob(pattern)
	if err != nil {
		return 0, err
	}

	maxSeq := 0
	for _, file := range files {
		seq, err := parseFileName2Seq(file, date)
		if err != nil {
			continue
		}
		if seq > maxSeq {
			maxSeq = seq
		}
	}

	return maxSeq, nil
}

// parseFileName2Seq 解析文件名为指定日期下递增序列的seq，例如 20250527-1.log 中的横杠后的1
//   - filename runtime/20250527-1.log
//   - date     20250527
func parseFileName2Seq(filename, date string) (int, error) {
	escapedDate := regexp.QuoteMeta(date)
	re := regexp.MustCompile(fmt.Sprintf(`^%s-(\d+)\.log$`, escapedDate))
	matches := re.FindStringSubmatch(filepath.Base(filename))
	if matches == nil {
		return 0, fmt.Errorf("filename format invalid")
	}

	// matches {{ 20250527-1.log, 1 }}
	seq, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}
	return seq, nil
}

// checkFileExist 判斷文件是否存在  存在返回 true 不存在返回false
func checkFileExist(filePath string) bool {
	var exist = true
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
