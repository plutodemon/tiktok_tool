package config

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"path/filepath"
)

// Settings 配置结构体
type Settings struct {
	NetworkInterfaces []string `toml:"network_interfaces"` // 网卡名称列表
	ServerRegex       string   `toml:"server_regex"`       // 服务器地址正则表达式
	StreamKeyRegex    string   `toml:"stream_key_regex"`   // 推流码正则表达式
	OBSConfigPath     string   `toml:"obs_config_path"`    // OBS配置文件路径
}

// DefaultSettings 默认配置
var DefaultSettings = Settings{
	NetworkInterfaces: []string{},
	ServerRegex:       `(rtmp://push-rtmp-[a-zA-Z0-9\-]+\.douyincdn\.com/thirdgame)`,
	StreamKeyRegex:    `(stream-\d+\?expire=\d+&sign=[a-f0-9]+(?:&volcSecret=[a-f0-9]+&volcTime=\d+)?)`,
	OBSConfigPath:     "", // 默认为空，需要用户手动配置
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

// WriteOBSConfig 将推流配置写入OBS配置文件(service.json)
func WriteOBSConfig(configPath, server, key string) error {
	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 读取JSON配置文件
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析JSON
	var config map[string]interface{}
	err = json.Unmarshal(content, &config)
	if err != nil {
		return fmt.Errorf("解析JSON配置文件失败: %v", err)
	}

	// 确保settings字段存在
	if config["settings"] == nil {
		config["settings"] = make(map[string]interface{})
	}

	// 获取settings对象
	settings, ok := config["settings"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("配置文件格式错误: settings字段不是对象")
	}

	// 更新server和key字段
	settings["server"] = server
	settings["key"] = key

	// 将修改后的配置转换回JSON
	newContent, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化JSON配置失败: %v", err)
	}

	// 写回文件
	err = os.WriteFile(configPath, newContent, 0644)
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// GetDefaultOBSConfigPath 获取默认的OBS配置文件路径(service.json)
func GetDefaultOBSConfigPath() string {
	// Windows系统下OBS service.json配置文件的常见路径
	possiblePaths := []string{
		filepath.Join(os.Getenv("APPDATA"), "obs-studio", "basic", "profiles", "Untitled", "service.json"),
		filepath.Join(os.Getenv("APPDATA"), "obs-studio", "basic", "profiles", "Default", "service.json"),
	}

	// 检查每个可能的路径
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 尝试查找其他配置文件
	profilesDir := filepath.Join(os.Getenv("APPDATA"), "obs-studio", "basic", "profiles")
	if entries, err := os.ReadDir(profilesDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				configPath := filepath.Join(profilesDir, entry.Name(), "service.json")
				if _, err := os.Stat(configPath); err == nil {
					return configPath
				}
			}
		}
	}

	return ""
}
