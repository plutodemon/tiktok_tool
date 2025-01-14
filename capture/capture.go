package capture

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"os"
	"regexp"
	"slices"
	"strings"
	"tiktok_tool/config"
)

// CheckNpcapInstalled 检查是否安装了Npcap
func CheckNpcapInstalled() bool {
	_, err := os.Stat("C:\\Windows\\System32\\Npcap")
	return err == nil
}

// StopCapturing 停止抓包
func StopCapturing() {
	if config.IsCapturing {
		config.IsCapturing = false
		close(config.StopCapture)

		config.HandleMutex.Lock()
		for _, handle := range config.Handles {
			handle.Close()
		}
		config.Handles = nil
		config.HandleMutex.Unlock()
	}
}

// StartCapture 开始抓包
func StartCapture(onServerFound func(string), onStreamKeyFound func(string), onError func(error)) {
	allDevices, err := pcap.FindAllDevs()
	if err != nil {
		onError(err)
		return
	}

	devices := make([]pcap.Interface, 0)
	for _, device := range allDevices {
		if len(config.CurrentSettings.NetworkInterfaces) == 0 {
			if strings.Contains(device.Description, "Bluetooth") ||
				strings.Contains(device.Description, "loopback") {
				continue
			}
			devices = append(devices, device)
			continue
		}
		if slices.Contains(config.CurrentSettings.NetworkInterfaces, device.Description) {
			devices = append(devices, device)
		}
	}

	if len(devices) == 0 {
		onError(fmt.Errorf("未找到可用的网络接口"))
		return
	}

	for _, device := range devices {
		if config.IsDebug {
			fmt.Printf("正在监听网络接口: %s (%s)\n", device.Name, device.Description)
		}
		go captureDevice(device.Name, onServerFound, onStreamKeyFound)
	}
}

func captureDevice(deviceName string, onServerFound func(string), onStreamKeyFound func(string)) {
	handle, err := pcap.OpenLive(deviceName, 65535, true, pcap.BlockForever)
	if err != nil {
		return
	}

	config.HandleMutex.Lock()
	config.Handles = append(config.Handles, handle)
	config.HandleMutex.Unlock()

	defer func() {
		config.HandleMutex.Lock()
		handle.Close()
		config.HandleMutex.Unlock()
	}()

	err = handle.SetBPFFilter("tcp")
	if err != nil {
		return
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for {
		select {
		case <-config.StopCapture:
			return
		default:
			packet, err := packetSource.NextPacket()
			if err != nil {
				continue
			}

			appLayer := packet.ApplicationLayer()
			if appLayer == nil {
				continue
			}

			payload := string(appLayer.Payload())

			if strings.Contains(strings.ToLower(payload), "rtmp://") {
				if config.IsDebug {
					fmt.Printf("发现包含 rtmp:// 的数据包: %s\n", payload)
				}
				serverRegexCompile := config.CurrentSettings.ServerRegex
				if len(serverRegexCompile) == 0 {
					serverRegexCompile = config.DefaultSettings.ServerRegex
				}
				serverRegex := regexp.MustCompile(serverRegexCompile)
				matches := serverRegex.FindStringSubmatch(payload)

				if config.IsDebug {
					fmt.Printf("服务器地址匹配结果: %v\n", matches)
				}

				if len(matches) >= 1 {
					serverUrl := matches[1]
					onServerFound(serverUrl)
					if config.IsDebug {
						fmt.Printf("找到服务器地址: %s\n", serverUrl)
					}
				}

				streamPatterns := []string{config.DefaultSettings.StreamKeyRegex}
				if len(config.CurrentSettings.StreamKeyRegex) > 0 &&
					config.CurrentSettings.StreamKeyRegex != config.DefaultSettings.StreamKeyRegex {
					streamPatterns = append(streamPatterns, config.CurrentSettings.StreamKeyRegex)
				}

				for _, pattern := range streamPatterns {
					streamRegex := regexp.MustCompile(pattern)
					matches = streamRegex.FindStringSubmatch(payload)

					if len(matches) >= 2 {
						streamKey := matches[1]
						onStreamKeyFound(streamKey)
						if config.IsDebug {
							fmt.Printf("找到推流码: %s\n", streamKey)
						}
						break
					}
				}
			}
		}
	}
}
