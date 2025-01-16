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
func StartCapture(onServerFound func(string), onStreamKeyFound func(string), onError func(error), onGetAll func()) {
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
			fmt.Printf("正在监听网络接口: %s \n", device.Description)
		}
		go captureDevice(device.Name, onServerFound, onStreamKeyFound, onGetAll)
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
					allReadyGetServer = true
					if config.IsDebug {
						fmt.Printf("找到服务器地址: %s\n", serverUrl)
					}
				}
			}

			if !allReadyGetStream {
				streamKeyRegexCompile := config.CurrentSettings.StreamKeyRegex
				if len(streamKeyRegexCompile) == 0 {
					streamKeyRegexCompile = config.DefaultSettings.StreamKeyRegex
				}

				streamRegex := regexp.MustCompile(streamKeyRegexCompile)
				matches := streamRegex.FindStringSubmatch(payload)

				if len(matches) >= 1 {
					streamStr := matches[1]
					onStreamKeyFound(streamStr)
					allReadyGetStream = true
					if config.IsDebug {
						fmt.Printf("找到推流码字符串: %s\n", streamStr)
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
