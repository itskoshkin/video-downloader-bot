package cleanup

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"

	"video-downloader-bot/internal/config"
	"video-downloader-bot/internal/logger"
)

const (
	fileTTL       = 30 * time.Minute // a file lingering this long is orphaned (downloads finish in seconds)
	sweepInterval = 15 * time.Minute
)

// StartSweeper deletes stale files from the download/convert folders — an orphan safety net for
// leaked yt-dlp fragments (.part, .fdash-*) and files left behind by failed conversions. Successful
// downloads are removed right after sending; anything older than fileTTL is stale. files/static
// (cookies) is never touched. Runs one sweep now, then every sweepInterval in the background.
func StartSweeper() {
	folders := []string{
		viper.GetString(config.TelegramBotVideoDownloadFolder),
		viper.GetString(config.TelegramBotVideoConvertedFolder),
	}
	sweep(folders)
	go func() {
		ticker := time.NewTicker(sweepInterval)
		defer ticker.Stop()
		for range ticker.C {
			sweep(folders)
		}
	}()
}

func sweep(folders []string) {
	for _, dir := range folders {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue // skip subdirs and dotfiles (.DS_Store, etc.)
			}
			info, err := e.Info()
			if err != nil || time.Since(info.ModTime()) <= fileTTL {
				continue
			}
			path := filepath.Join(dir, e.Name())
			if os.Remove(path) == nil {
				logger.Debug("cleanup: removed stale file \"%s\"", path)
			}
		}
	}
}
