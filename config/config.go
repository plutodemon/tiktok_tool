package config

import (
	"os"
	"path/filepath"
	"sync"

	"tiktok_tool/llog"

	"github.com/BurntSushi/toml"
	"github.com/google/gopacket/pcap"
)

var (
	Debug   string
	IsDebug bool

	IsCapturing bool
	StopCapture chan struct{}
	Handles     []*pcap.Handle
	HandleMutex sync.Mutex

	currentConfig *Config
	configPath    string
)

func init() {
	if Debug == "true" {
		IsDebug = true
	}

	configPath = filepath.Join("config", "tiktok_tool_cfg.toml")
}

type Config struct {
	NetworkInterfaces []string `toml:"network_interfaces"`  // 网卡名称列表
	ServerRegex       string   `toml:"server_regex"`        // 服务器地址正则表达式
	StreamKeyRegex    string   `toml:"stream_key_regex"`    // 推流码正则表达式
	OBSConfigPath     string   `toml:"obs_config_path"`     // OBS配置文件路径
	LiveCompanionPath string   `toml:"live_companion_path"` // 直播伴侣启动路径
	PluginScriptPath  string   `toml:"plugin_script_path"`  // 自动化插件脚本路径

	// 日志设置
	LogConfig *llog.LogSetting `toml:"log"`
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	NetworkInterfaces: []string{},
	ServerRegex:       `(rtmp://push-rtmp-[a-zA-Z0-9\-]+\.douyincdn\.com/thirdgame)`,
	StreamKeyRegex:    `"(stream-\d+\?(?:[^&]+=[^&]*&)*expire=\d{10}&sign=[^&]+)"`,
	OBSConfigPath:     "", // 默认为空，需要用户手动配置
	LiveCompanionPath: "", // 默认为空，需要用户手动配置
	PluginScriptPath:  "", // 默认为空，需要用户手动配置
	LogConfig:         llog.DefaultConfig,
}

// LoadConfig 加载配置文件
func LoadConfig() error {
	// 尝试从config目录加载
	if _, err := os.Stat(configPath); err != nil {
		// 尝试从当前目录加载
		configPath = "tiktok_tool_cfg.toml"
		if _, err := os.Stat(configPath); err != nil {
			// 配置文件不存在，使用默认配置
			return nil
		}
	}
	currentConfig = &Config{}
	_, err := toml.DecodeFile(configPath, currentConfig)
	return err
}

// SaveSettings 保存配置文件
func SaveSettings(settings Config) error {
	// 优先保存到config目录
	configDir := "config"
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// config目录不存在，创建它
		// if err := os.MkdirAll(configDir, 0755); err != nil {
		// 如果创建失败，保存到当前目录
		configDir = "."
		// }
	}
	configPath = filepath.Join(configDir, "tiktok_tool_cfg.toml")
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	return encoder.Encode(settings)
}

func GetConfig() Config {
	if currentConfig == nil {
		currentConfig = &DefaultConfig
	}
	return *currentConfig
}

func SetConfig(cfg Config) {
	currentConfig = &cfg
}
