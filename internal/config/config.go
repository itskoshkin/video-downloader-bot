package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"

	"video-downloader-bot/internal/utils/text"
	"video-downloader-bot/pkg/postgres"
)

const (
	LogLevel       = "app.log.level"           // string ("DEBUG", "INFO", "WARN", "ERROR")
	LogFormat      = "app.log.log_format"      // string ("text" or "json")
	LogToConsole   = "app.log.log2console"     // bool
	LogToFile      = "app.log.log2file"        // bool
	LogFilePath    = "app.log.file_path"       // string (path)
	LogFileMode    = "app.log.file_mode"       // string ("append", "overwrite", "rotate")
	LogFilesFolder = "app.log.old_logs_folder" // string (path)

	DatabaseHost     = "app.database.host"          // string
	DatabasePort     = "app.database.port"          // int
	DatabaseUser     = "app.database.user"          // string
	DatabasePassword = "app.database.password"      // string
	DatabaseName     = "app.database.database_name" // string
	DatabaseSslMode  = "app.database.ssl_mode"      // string

	TelegramBotLibDebug             = "app.telegram.gotgbot_debug"             // bool
	TelegramBotToken                = "app.telegram.bot.token"                 // string
	TelegramBotDropPendingUpdates   = "app.telegram.bot.drop_pending_updates"  // bool
	TelegramBotLongPollingTimeout   = "app.telegram.bot.get_updates_timeout"   // int
	TelegramBotHttpClientTimeout    = "app.telegram.bot.http_client_timeout"   // int
	TelegramBotVideoDumpChatID      = "app.telegram.bot.video_dump_chat_id"    // int
	TelegramBotVideoDownloadFolder  = "app.telegram.bot.download_folder"       // string (path)
	TelegramBotVideoConvertedFolder = "app.telegram.bot.convert_folder"        // string (path)
	TelegramBotRateLimitPerMinute   = "app.telegram.bot.rate_limit_per_minute" // int
	TelegramBotRateLimitBurst       = "app.telegram.bot.rate_limit_burst"      // int
	TelegramBotRateLimitPerDay      = "app.telegram.bot.rate_limit_per_day"    // int
	TelegramBotInlineCacheTime      = "app.telegram.bot.inline_cache_time"     // int (seconds)
	TelegramBotMaxFileSizeMB        = "app.telegram.bot.max_file_size_mb"      // int
	TelegramBotMaxVideoDuration     = "app.telegram.bot.max_video_duration"    // int (seconds)

	YtDlpBinary      = "app.yt-dlp.binary_file"  // string (path)
	YtDlpDebug       = "app.yt-dlp.debug"        // bool
	YtDlpUseCookies  = "app.yt-dlp.use_cookies"  // bool
	YtDlpCookiesFile = "app.yt-dlp.cookies_file" // string (path)

	FfmpegDebug = "app.ffmpeg.debug" // bool

	ProvidersChainDefault   = "app.providers.chains.default"            // []string (ordered provider names)
	PreviewInstagramDomains = "app.providers.preview.instagram_domains" // []string (embed-fix domains, primary first)
	PreviewTiktokDomains    = "app.providers.preview.tiktok_domains"    // []string (embed-fix domains, primary first)
	HikerApiKey             = "app.providers.hikerapi.api_key"          // string
	AiograpiBaseURL         = "app.providers.aiograpi.base_url"         // string
	AiograpiSessionID       = "app.providers.aiograpi.session_id"       // string (X-Session-ID header for the sidecar)
)

// ProvidersChainKey builds the config key for a platform's provider chain, e.g. "app.providers.chains.instagram".
func ProvidersChainKey(platform string) string {
	return "app.providers.chains." + platform
}

func LoadConfig() {
	fmt.Print("Loading configuration...")

	viper.SetConfigFile("./config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println()
		log.Fatalf("Fatal: failed to read configuration: %v", err)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := ValidateConfigFields(); err != nil {
		fmt.Println()
		log.Fatalf("Fatal: failed to load configuration: %v", err)
	}

	fmt.Println(text.Green("  Done."))
}

func ValidateConfigFields() error {
	var required = []string{ // Must be present and non-empty
		DatabaseHost, DatabasePort, DatabaseUser, DatabasePassword,
		TelegramBotToken, TelegramBotDropPendingUpdates, TelegramBotHttpClientTimeout,
		TelegramBotVideoDumpChatID, TelegramBotVideoDownloadFolder, TelegramBotVideoConvertedFolder,
	}
	var dependent = map[string][]string{ // If A=true => must be non-empty B (, C...)
		LogToFile: {LogFilePath},
	}
	var possibleValues = map[string][]string{ // If present, must be one of these values
		LogLevel:    {"DEBUG", "INFO", "WARN", "ERROR"},
		LogFormat:   {"text", "json"},
		LogFileMode: {"append", "overwrite", "rotate"},
	}
	var defaults = map[string]any{ // Will be set if not present, overwrites above required/dependent
		/* Log */ LogLevel: "INFO", LogFormat: "text", LogToConsole: true, LogToFile: true, LogFilePath: "application.log", LogFileMode: "append",
		/* Postgres */ DatabaseHost: "localhost", DatabasePort: 5432, DatabaseUser: "postgres", DatabaseName: "video-downloader-bot", DatabaseSslMode: "disable",
		/* Telegram */ TelegramBotLongPollingTimeout: 9, TelegramBotHttpClientTimeout: 10,
		TelegramBotVideoDownloadFolder: "./files/downloads", TelegramBotVideoConvertedFolder: "./files/converted",
		TelegramBotRateLimitPerMinute: 10, TelegramBotRateLimitBurst: 3, TelegramBotRateLimitPerDay: 100,
		TelegramBotMaxFileSizeMB: 20, TelegramBotMaxVideoDuration: 300, TelegramBotInlineCacheTime: 0,
		/* External tools */ YtDlpDebug: false, FfmpegDebug: false,
		/* Providers */
		ProvidersChainDefault:          []string{"yt-dlp", "preview"},
		ProvidersChainKey("instagram"): []string{"yt-dlp", "aiograpi", "hikerapi", "preview"},
		PreviewInstagramDomains:        []string{"vxinstagram.com", "eeinstagram.com", "uuinstagram.com", "zzinstagram.com"},
		PreviewTiktokDomains:           []string{"tnktok.com"},
	}

	for k, v := range defaults {
		if !viper.IsSet(k) {
			viper.Set(k, v)
		}
	}

	var missing []string
	for _, key := range required {
		if isEmptyValue(key) {
			missing = append(missing, key)
		}
	}
	if viper.GetString(LogFileMode) == "rotate" {
		if isEmptyValue(LogFilesFolder) {
			missing = append(missing, fmt.Sprintf("%s (required when %s=rotate)", LogFilesFolder, LogFileMode))
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields/values in config: %s", strings.Join(missing, ", "))
	}

	for triggerKey, requiredKeys := range dependent {
		if viper.GetBool(triggerKey) || (viper.IsSet(triggerKey) && viper.GetString(triggerKey) != "") {
			for _, key := range requiredKeys {
				if isEmptyValue(key) {
					missing = append(missing, fmt.Sprintf("%s (%s is set)", key, triggerKey))
				}
			}
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields/values in config: %s", strings.Join(missing, ", "))
	}

	var invalid []string
	for key, allowed := range possibleValues {
		if !viper.IsSet(key) {
			continue
		}
		val := strings.TrimSpace(viper.GetString(key))
		if val == "" {
			continue
		}
		found := false
		for _, a := range allowed {
			if val == a {
				found = true
				break
			}
		}
		if !found {
			invalid = append(invalid, fmt.Sprintf("'%s' for '%s' (must be one of [%s])", val, key, strings.Join(allowed, ", ")))
		}
	}
	if len(invalid) > 0 {
		return fmt.Errorf("invalid config values: %s", strings.Join(invalid, ", "))
	}

	return nil
}

func isEmptyValue(key string) bool {
	if !viper.IsSet(key) {
		return true
	}

	switch value := viper.Get(key).(type) {
	case string:
		return strings.TrimSpace(value) == ""
	case []string:
		return len(value) == 0
	case []any:
		return len(value) == 0
	default:
		return strings.TrimSpace(viper.GetString(key)) == ""
	}
}

func PostgresConfig() postgres.Config {
	return postgres.Config{
		Host:     viper.GetString(DatabaseHost),
		Port:     viper.GetString(DatabasePort),
		User:     viper.GetString(DatabaseUser),
		Password: viper.GetString(DatabasePassword),
		Database: viper.GetString(DatabaseName),
		SSLMode:  viper.GetString(DatabaseSslMode),
		LogLevel: viper.GetString(LogLevel),
	}
}

func EnsureWorkingDirs() error {
	if err := os.MkdirAll(viper.GetString(TelegramBotVideoDownloadFolder), 0o755); err != nil {
		return fmt.Errorf("failed to create yt-dlp working directory: %w", err)
	}

	if err := os.MkdirAll(viper.GetString(TelegramBotVideoConvertedFolder), 0o755); err != nil {
		return fmt.Errorf("failed to create ffmpeg working directory: %w", err)
	}

	return nil
}
