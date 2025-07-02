package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

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

// FindLiveCompanionPath 在所有盘符中搜索直播伴侣启动文件
func FindLiveCompanionPath() string {
	possibleFileNames := []string{
		"直播伴侣 Launcher.exe",
		"直播伴侣.exe",
	}

	// 依次搜索每个可能的文件名
	for _, fileName := range possibleFileNames {
		if path := FindFileInAllDrives(fileName); path != "" {
			return path
		}
	}

	return ""
}
