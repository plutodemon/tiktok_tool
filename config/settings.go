package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Settings 配置结构体
type Settings struct {
	NetworkInterfaces []string `toml:"network_interfaces"` // 网卡名称列表
	ServerRegex       string   `toml:"server_regex"`       // 服务器地址正则表达式
	StreamKeyRegex    string   `toml:"stream_key_regex"`   // 推流码正则表达式
}

// DefaultSettings 默认配置
var DefaultSettings = Settings{
	NetworkInterfaces: []string{},
	ServerRegex:       `(rtmp://push-rtmp-[^\.]+\.douyincdn\.com/thirdgame)`,
	StreamKeyRegex:    `(stream-\d+\?expire=\d+&sign=[a-f0-9]+(?:&volcSecret=[a-f0-9]+&volcTime=\d+)?)`,
}

var CurrentSettings = DefaultSettings

// LoadSettings 加载配置文件
func LoadSettings() error {
	// 尝试从config目录加载
	configPath := "config/tiktok_tool_cfg.toml"
	if _, err := os.Stat(configPath); err != nil {
		// 尝试从当前目录加载
		configPath = "tiktok_tool_cfg.toml"
		if _, err := os.Stat(configPath); err != nil {
			// 配置文件不存在，使用默认配置
			return nil
		}
	}

	_, err := toml.DecodeFile(configPath, &CurrentSettings)
	return err
}

// SaveSettings 保存配置文件
func SaveSettings(settings Settings) error {
	// 优先保存到config目录
	configDir := "config"
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// config目录不存在，创建它
		//if err := os.MkdirAll(configDir, 0755); err != nil {
		// 如果创建失败，保存到当前目录
		configDir = "."
		//}
	}

	configPath := filepath.Join(configDir, "tiktok_tool_cfg.toml")
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	return encoder.Encode(settings)
}
