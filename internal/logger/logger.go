package logger

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"

	"video-downloader-bot/internal/config"
	"video-downloader-bot/internal/telegram/middlewares/requests"
	"video-downloader-bot/internal/utils/text"
)

var (
	logFile      *os.File
	currentLevel = LevelInfo
	consoleLog   *log.Logger
	fileLog      *log.Logger
	jsonFileLog  *log.Logger

	// Consecutive-message dedup: collapses identical back-to-back lines (e.g. a reconnect error
	// repeating while the network is down) into one line + a "repeated N times" summary on the next
	// distinct line. Pairs with the backoff in UpdaterErrorHandler to kill offline log spam.
	dedupMu      sync.Mutex
	dedupText    string
	dedupColored string
	dedupPlain   string
	dedupCount   int
)

type Level int

const (
	LevelError Level = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

type jsonEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

func SetupLogger() {
	fmt.Print("Setting up logger...")

	switch viper.GetString(config.LogLevel) {
	case "DEBUG":
		currentLevel = LevelDebug
	case "ERROR":
		currentLevel = LevelError
	case "WARN":
		currentLevel = LevelWarn
	default:
		currentLevel = LevelInfo
	}

	if viper.GetBool(config.LogToFile) {
		filePath := viper.GetString(config.LogFilePath)

		flags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
		if viper.GetString(config.LogFileMode) == "overwrite" {
			flags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
		}
		if viper.GetString(config.LogFileMode) == "rotate" {
			rotateLogFile(filePath, viper.GetString(config.LogFilesFolder))
		}

		var err error
		logFile, err = os.OpenFile(filePath, flags, 0666)
		if err != nil {
			fmt.Println()
			log.Fatalf("Fatal: failed to open log file: %v", err)
		}

		if viper.GetString(config.LogFormat) == "json" {
			jsonFileLog = log.New(logFile, "", 0)
			b, _ := json.Marshal(jsonEntry{Time: timestamp(), Level: "INFO", Message: "=== new run ==="})
			jsonFileLog.Println(string(b))
		} else {
			fileLog = log.New(logFile, "", 0)
			_, _ = logFile.WriteString("\n\n==== New run at " + time.Now().Format("2006/01/02 15:04:05") + " ====\n")
		}
	}

	if viper.GetBool(config.LogToConsole) {
		consoleLog = log.New(os.Stdout, "", 0)
	}

	fmt.Println(text.Green("      Done."))
}

func rotateLogFile(filePath, logsFolder string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return
	}

	if err := os.MkdirAll(logsFolder, 0755); err != nil {
		log.Printf("failed to create logs folder: %v", err)
		return
	}

	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	newName := filepath.Join(logsFolder, name+"_"+time.Now().Format("2006-01-02_15-04-05")+ext)
	if err := os.Rename(filePath, newName); err != nil {
		log.Printf("failed to rotate log file: %v", err)
		return
	}

	if viper.GetBool(config.LogGzipOldLogs) {
		if err := gzipFile(newName); err != nil {
			log.Printf("failed to gzip rotated log: %v", err)
		}
	}

	pruneOldLogs(logsFolder)
}

// pruneOldLogs enforces retention on rotated archives in logsFolder: it keeps the newest ones within
// the configured count / total-size / age limits (0 disables that limit) and deletes the oldest beyond them.
func pruneOldLogs(logsFolder string) {
	maxCount := viper.GetInt(config.LogMaxOldFiles)
	maxSize := int64(viper.GetInt(config.LogMaxOldSizeMB)) * 1024 * 1024
	maxAge := time.Duration(viper.GetInt(config.LogMaxOldAgeDays)) * 24 * time.Hour
	if maxCount <= 0 && maxSize <= 0 && maxAge <= 0 {
		return
	}

	entries, err := os.ReadDir(logsFolder)
	if err != nil {
		return
	}

	type archive struct {
		path string
		size int64
		mod  time.Time
	}
	var archives []archive
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if n := e.Name(); !strings.HasSuffix(n, ".log") && !strings.HasSuffix(n, ".log.gz") {
			continue // only touch rotated log archives
		}
		if info, err := e.Info(); err == nil {
			archives = append(archives, archive{filepath.Join(logsFolder, e.Name()), info.Size(), info.ModTime()})
		}
	}
	sort.Slice(archives, func(i, j int) bool { return archives[i].mod.After(archives[j].mod) }) // newest first

	now := time.Now()
	var kept int64
	for i, a := range archives {
		drop := (maxCount > 0 && i >= maxCount) ||
			(maxAge > 0 && now.Sub(a.mod) > maxAge) ||
			(maxSize > 0 && kept+a.size > maxSize)
		if drop {
			_ = os.Remove(a.path)
		} else {
			kept += a.size
		}
	}
}

// gzipFile compresses path to path+".gz" and removes the original on success.
func gzipFile(path string) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	out, err := os.Create(path + ".gz")
	if err != nil {
		_ = in.Close()
		return err
	}

	gz := gzip.NewWriter(out)
	_, copyErr := io.Copy(gz, in)
	gzErr := gz.Close()
	_ = out.Close()
	_ = in.Close()

	if copyErr != nil {
		return copyErr
	}
	if gzErr != nil {
		return gzErr
	}
	return os.Remove(path)
}

func GetWriters() io.Writer {
	if jsonFileLog != nil {
		if consoleLog != nil {
			return os.Stdout
		}
		return io.Discard
	}
	switch {
	case consoleLog != nil && fileLog != nil:
		return io.MultiWriter(os.Stdout, logFile)
	case fileLog != nil:
		return logFile
	case consoleLog != nil:
		return os.Stdout
	default:
		return io.Discard
	}
}

func GetLevel() Level { return currentLevel }

func SetLevel(l Level) { currentLevel = l }

func Debug(format string, args ...any) {
	if currentLevel < LevelDebug {
		return
	}
	write(text.Purple("DEBUG"), "DEBUG", fmt.Sprintf(format, args...))
}

func Info(format string, args ...any) {
	if currentLevel < LevelInfo {
		return
	}
	write(text.Green("INFO "), "INFO", fmt.Sprintf(format, args...))
}

func Warn(format string, args ...any) {
	if currentLevel < LevelWarn {
		return
	}
	write(text.Yellow("WARN "), "WARN", fmt.Sprintf(format, args...))
}

func Error(format string, args ...any) {
	write(text.Red("ERROR"), "ERROR", fmt.Sprintf(format, args...))
}

func DebugWithID(ctx context.Context, format string, args ...any) {
	Debug(prefixWithID(ctx, format), args...)
}

func InfoWithID(ctx context.Context, format string, args ...any) {
	Info(prefixWithID(ctx, format), args...)
}

func WarnWithID(ctx context.Context, format string, args ...any) {
	Warn(prefixWithID(ctx, format), args...)
}

func ErrorWithID(ctx context.Context, format string, args ...any) {
	Error(prefixWithID(ctx, format), args...)
}

func ErrorWithFileID(ctx context.Context, format string, args ...any) {
	writeFile("ERROR", fmt.Sprintf(prefixWithID(ctx, format), args...))
}

func Fatal(v ...any) {
	write(text.Bold(text.Red("FATAL")), "FATAL", fmt.Sprint(v...))
	os.Exit(1)
}

func Fatalf(format string, args ...any) {
	write(text.Bold(text.Red("FATAL")), "FATAL", fmt.Sprintf(format, args...))
	os.Exit(1)
}

func prefixWithID(ctx context.Context, format string) string {
	if id := req.FromContext(ctx); id != "" {
		return "[" + id + "] " + format
	}
	return format
}

func write(coloredLevel, plainLevel, message string) {
	dedupMu.Lock()
	if message == dedupText {
		dedupCount++
		dedupMu.Unlock()
		return
	}
	sumColored, sumPlain, repeats := dedupColored, dedupPlain, dedupCount
	dedupText, dedupColored, dedupPlain, dedupCount = message, coloredLevel, plainLevel, 0
	dedupMu.Unlock()

	if repeats > 0 {
		emit(sumColored, sumPlain, fmt.Sprintf("(previous message repeated %d more time(s))", repeats))
	}
	emit(coloredLevel, plainLevel, message)
}

func emit(coloredLevel, plainLevel, message string) {
	if consoleLog != nil {
		consoleLog.Printf("%s %s %s", timestamp(), coloredLevel, message)
	}
	writeFile(plainLevel, message)
}

func writeFile(plainLevel, text string) {
	if fileLog != nil {
		fileLog.Printf("%s %s %s", timestamp(), plainLevel, text)
	}
	if jsonFileLog != nil {
		b, _ := json.Marshal(jsonEntry{Time: timestamp(), Level: plainLevel, Message: text})
		jsonFileLog.Println(string(b))
	}
}

func timestamp() string { return time.Now().Format("2006/01/02 15:04:05") }

type GlobalLogger struct{} // Wraps package-level functions for use as an injected Logger dependency in services

func (GlobalLogger) Error(format string, v ...any) { Error(format, v...) }

func Close() {
	// Flush any pending dedup summary before the file is closed.
	dedupMu.Lock()
	sumColored, sumPlain, repeats := dedupColored, dedupPlain, dedupCount
	dedupCount = 0
	dedupMu.Unlock()
	if repeats > 0 {
		emit(sumColored, sumPlain, fmt.Sprintf("(previous message repeated %d more time(s))", repeats))
	}

	if logFile != nil {
		_ = logFile.Close()
	}
}
