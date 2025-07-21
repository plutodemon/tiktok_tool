package capture

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"

	"tiktok_tool/config"
	"tiktok_tool/lkit"
	"tiktok_tool/llog"
)

var (
	allReadyGetServer = false
	allReadyGetStream = false
)

// CheckNpcapInstalled 检查是否安装了Npcap
func CheckNpcapInstalled() bool {
	_, err := os.Stat("C:\\Windows\\System32\\Npcap")
	return err == nil
}

// StopCapturing 停止抓包
func StopCapturing() {
	if config.IsCapturing == false {
		return
	}
	config.IsCapturing = false
	close(config.StopCapture)

	config.HandleMutex.Lock()
	for _, handle := range config.Handles {
		handle.Close()
	}
	config.Handles = nil
	config.HandleMutex.Unlock()
}

// StartCapture 开始抓包
func StartCapture(onServerFound func(string), onStreamKeyFound func(string), onError func(error), onGetAll func()) {
	allDevices, err := pcap.FindAllDevs()
	if err != nil {
		onError(err)
		return
	}

	devices := make([]pcap.Interface, 0)
	for _, device := range allDevices {
		if len(config.GetConfig().NetworkInterfaces) == 0 {
			if strings.Contains(device.Description, "Bluetooth") ||
				strings.Contains(device.Description, "loopback") {
				continue
			}
			devices = append(devices, device)
			continue
		}
		if slices.Contains(config.GetConfig().NetworkInterfaces, device.Description) {
			devices = append(devices, device)
		}
	}

	if len(devices) == 0 {
		onError(fmt.Errorf("未找到可用的网络接口"))
		return
	}

	for _, device := range devices {
		if config.IsDebug {
			llog.DebugF("正在监听网络接口: %s", device.Description)
		}
		lkit.SafeGo(func() {
			captureDevice(device.Name, onServerFound, onStreamKeyFound, onGetAll)
		})
	}
}

func captureDevice(deviceName string, onServerFound func(string), onStreamKeyFound func(string), onGetAll func()) {
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

			if !allReadyGetServer && strings.Contains(strings.ToLower(payload), "rtmp://") {
				serverRegexCompile := config.GetConfig().ServerRegex
				if len(serverRegexCompile) == 0 {
					serverRegexCompile = config.DefaultConfig.ServerRegex
				}

				serverRegex := regexp.MustCompile(serverRegexCompile)
				matches := serverRegex.FindStringSubmatch(payload)

				if config.IsDebug {
					llog.DebugF("服务器地址匹配结果: %v", matches)
				}

				if len(matches) >= 1 {
					serverUrl := matches[1]
					onServerFound(serverUrl)
					allReadyGetServer = true
					if config.IsDebug {
						llog.DebugF("找到服务器地址: %s", serverUrl)
					}
				}
			}

			if !allReadyGetStream {
				streamKeyRegexCompile := config.GetConfig().StreamKeyRegex
				if len(streamKeyRegexCompile) == 0 {
					streamKeyRegexCompile = config.DefaultConfig.StreamKeyRegex
				}

				streamRegex := regexp.MustCompile(streamKeyRegexCompile)
				matches := streamRegex.FindStringSubmatch(payload)

				if len(matches) >= 1 {
					streamStr := matches[1]
					onStreamKeyFound(streamStr)
					allReadyGetStream = true
					if config.IsDebug {
						llog.DebugF("找到推流码字符串: %s", streamStr)
					}
					break
				}
			}

			if allReadyGetServer && allReadyGetStream {
				onGetAll()
				StopCapturing()
			}
		}
	}
}
