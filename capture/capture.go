package capture

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"tiktok_tool/config"
	"tiktok_tool/lkit"
	"tiktok_tool/llog"
)

var (
	handles     []*pcap.Handle
	handleMutex sync.Mutex

	IsCapturing bool
	StopCapture chan struct{}

	allReadyGetServer = false
	allReadyGetStream = false

	SrcIPAddr     = ""
	SrcIPPort     uint16
	DstIPAddr     = ""
	DstIPPort     uint16
	SrcIPAddrPort = ""
	DstIPAddrPort = ""
)

// CheckNpcapInstalled 检查是否安装了Npcap
func CheckNpcapInstalled() bool {
	_, err := os.Stat("C:\\Windows\\System32\\Npcap")
	return err == nil
}

// StopCapturing 停止抓包
func StopCapturing() {
	if IsCapturing == false {
		return
	}
	llog.Debug("停止抓包")
	IsCapturing = false
	close(StopCapture)

	// 重置状态变量
	allReadyGetServer = false
	allReadyGetStream = false

	handleMutex.Lock()
	for _, handle := range handles {
		handle.Close()
	}
	handles = nil
	handleMutex.Unlock()
}

// StartCapture 开始抓包
func StartCapture(onServerFound, onStreamKeyFound, onStreamIpFound func(string), onError func(error), onGetAll func()) {
	// 重置状态变量
	llog.Debug("开始抓包")
	allReadyGetServer = false
	allReadyGetStream = false
	baseCfg := config.GetConfig().BaseSettings
	allDevices, err := pcap.FindAllDevs()
	if err != nil {
		onError(err)
		return
	}

	devices := make([]pcap.Interface, 0)
	for _, device := range allDevices {
		if len(baseCfg.NetworkInterfaces) == 0 {
			if strings.Contains(device.Description, "Bluetooth") ||
				strings.Contains(device.Description, "loopback") {
				continue
			}
			devices = append(devices, device)
			continue
		}
		if slices.Contains(baseCfg.NetworkInterfaces, device.Description) {
			devices = append(devices, device)
		}
	}

	if len(devices) == 0 {
		onError(fmt.Errorf("未找到可用的网络接口"))
		return
	}

	llog.Info("正在监听网络接口:", devices)

	for _, device := range devices {
		lkit.SafeGo(func() {
			captureDevice(device.Name, onServerFound, onStreamKeyFound, onStreamIpFound, onGetAll)
		})
	}
}

func captureDevice(deviceName string, onServerFound, onStreamKeyFound, onStreamIpFound func(string), onGetAll func()) {
	handle, err := pcap.OpenLive(deviceName, 65535, true, pcap.BlockForever)
	if err != nil {
		return
	}

	handleMutex.Lock()
	handles = append(handles, handle)
	handleMutex.Unlock()

	defer func() {
		handleMutex.Lock()
		handle.Close()
		handleMutex.Unlock()
	}()

	err = handle.SetBPFFilter("tcp")
	if err != nil {
		return
	}

	baseCfg := config.GetConfig().BaseSettings
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for {
		select {
		case <-StopCapture:
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
				serverRegex := regexp.MustCompile(baseCfg.ServerRegex)
				matches := serverRegex.FindStringSubmatch(payload)

				if len(matches) >= 1 {
					serverUrl := matches[0]
					onServerFound(serverUrl)
					allReadyGetServer = true
					llog.InfoF("找到服务器地址: %s", serverUrl)
				}
			}

			if !allReadyGetStream {
				streamRegex := regexp.MustCompile(baseCfg.StreamKeyRegex)
				matches := streamRegex.FindStringSubmatch(payload)

				if len(matches) >= 1 {
					streamStr := matches[0]
					onStreamKeyFound(streamStr)
					allReadyGetStream = true
					llog.InfoF("找到推流码字符串: %s", streamStr)
					lkit.SafeGo(func() {
						getDstInfo(packet, onStreamIpFound)
					})
				}
			}

			if allReadyGetServer && allReadyGetStream {
				llog.Debug("已找到服务器地址和推流码字符串, 停止抓包")
				onGetAll()
				StopCapturing()
			}
		}
	}
}

func getDstInfo(packet gopacket.Packet, onStreamIpFound func(string)) {
	ip4Layer := packet.Layer(layers.LayerTypeIPv4)
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if ip4Layer == nil || tcpLayer == nil {
		return
	}
	ipv4, ok1 := ip4Layer.(*layers.IPv4)
	tcp, ok2 := tcpLayer.(*layers.TCP)
	if !ok1 || !ok2 {
		return
	}

	SrcIPAddr = ipv4.SrcIP.String()
	SrcIPPort = uint16(tcp.SrcPort)
	DstIPAddr = ipv4.DstIP.String()
	DstIPPort = uint16(tcp.DstPort)
	SrcIPAddrPort = lkit.GetAddr(ipv4.SrcIP, tcp.SrcPort)
	DstIPAddrPort = lkit.GetAddr(ipv4.DstIP, tcp.DstPort)
	onStreamIpFound(DstIPAddrPort)

	llog.Info("本地IP: ", SrcIPAddrPort)
	llog.Info("推流目标IP: ", DstIPAddrPort)
}
