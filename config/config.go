package config

import (
	"os"
	"path/filepath"

	"tiktok_tool/llog"

	"github.com/BurntSushi/toml"
)

const (
	CfgFilePath = "config"               // 配置文件目录
	CfgFileName = "tiktok_tool_cfg.toml" // 配置文件名
)

var (
	Debug   string
	IsDebug bool

	currentConfig *Config
	configPath    string
)

func init() {
	if Debug == "true" {
		IsDebug = true
	}

	configPath = filepath.Join(CfgFilePath, CfgFileName)
}

type Config struct {
	BaseSettings   *BaseSettings    `toml:"base"`   // 基础设置
	PathSettings   *PathSettings    `toml:"path"`   // 路径设置
	ScriptSettings *ScriptSettings  `toml:"script"` // 脚本设置
	LogConfig      *llog.LogSetting `toml:"log"`    // 日志配置
}

type BaseSettings struct {
	NetworkInterfaces []string `toml:"network_interfaces"`   // 网卡名称列表
	ServerRegex       string   `toml:"server_regex"`         // 服务器地址正则表达式
	StreamKeyRegex    string   `toml:"stream_key_regex"`     // 推流码正则表达式
	MinimizeOnClose   bool     `toml:"minimize_on_close"`    // 关闭窗口时最小化到系统托盘而不退出
	OpenLiveWhenStart bool     `toml:"open_live_when_start"` // 启动时自动打开直播伴侣
}

type PathSettings struct {
	OBSLaunchPath     string `toml:"obs_launch_path"`     // OBS启动路径
	OBSConfigPath     string `toml:"obs_config_path"`     // OBS配置文件路径
	LiveCompanionPath string `toml:"live_companion_path"` // 直播伴侣启动路径
	PluginScriptPath  string `toml:"plugin_script_path"`  // 自动化插件脚本路径
}

type ScriptSettings struct {
	PluginCheckInterval  int32 `toml:"plugin_check_interval"`   // 插件检查间隔（秒）
	PluginWaitAfterFound int32 `toml:"plugin_wait_after_found"` // 插件找到后等待时间（秒）
	PluginTimeout        int32 `toml:"plugin_timeout"`          // 插件超时时间（秒）
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	BaseSettings: &BaseSettings{
		NetworkInterfaces: make([]string, 0),
		ServerRegex:       `(rtmp://push-rtmp[^ ]*?\.douyincdn\.com[^\x00\r\n ]*)`,
		StreamKeyRegex:    `(stream-[^\s]*?expire=\d{10}&sign=[^\s]+[^\x00\r\n ]*)`,
		MinimizeOnClose:   false,
		OpenLiveWhenStart: true,
	},
	PathSettings: &PathSettings{},
	ScriptSettings: &ScriptSettings{
		PluginCheckInterval:  1,
		PluginWaitAfterFound: 5,
		PluginTimeout:        20,
	},
	LogConfig: llog.DefaultConfig,
}

// LoadConfig 加载配置文件
func LoadConfig() error {
	if _, err := os.Stat(configPath); err != nil {
		configPath = filepath.Join(CfgFileName)
		if _, err = os.Stat(configPath); err != nil {
			return nil
		}
	}
	currentConfig = &Config{}
	_, err := toml.DecodeFile(configPath, currentConfig)
	return err
}

// SaveSettings 保存配置文件
func SaveSettings(settings *Config) error {
	configDir := CfgFilePath
	if err := os.MkdirAll(configDir, 0755); err != nil {
		configDir = "."
	}

	savePath := filepath.Join(configDir, CfgFileName)
	file, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return toml.NewEncoder(file).Encode(settings)
}

func GetConfig() *Config {
	if currentConfig == nil {
		currentConfig = &DefaultConfig
	}
	return currentConfig
}
