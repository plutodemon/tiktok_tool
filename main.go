package main

import (
	_ "embed"
)
import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

var Debug = ""

var isDebug = false

func init() {
	if Debug == "true" {
		isDebug = true
	}
}

// 添加全局变量来控制抓包
var (
	isCapturing bool
	stopCapture chan struct{}
	handles     []*pcap.Handle
	handleMutex sync.Mutex
)

// 添加停止抓包的函数
func stopCapturing() {
	if isCapturing {
		isCapturing = false
		close(stopCapture)

		// 关闭所有handle
		handleMutex.Lock()
		for _, handle := range handles {
			handle.Close()
		}
		handles = nil
		handleMutex.Unlock()
	}
}

func checkNpcapInstalled() bool {
	_, err := os.Stat("C:\\Windows\\System32\\Npcap")
	return err == nil
}

func showInstallDialog(window fyne.Window) {
	content := container.NewVBox(
		widget.NewLabel("需要安装 Npcap 才能使用抓包功能："),
		widget.NewLabel("1. 点击下面的按钮下载 Npcap, 选择\"Npcap 1.80 installer for Windows即可\""),
		widget.NewButton("下载 Npcap", func() {
			url := "https://npcap.com/#download"
			if err := exec.Command("cmd", "/c", "start", url).Start(); err != nil {
				dialog.ShowError(err, window)
			}
		}),
		widget.NewLabel("2. 安装完成后重启程序"),
		widget.NewLabel("注意：安装时请勾选\"Install Npcap in WinPcap API-compatible Mode\""),
	)

	customDialog := dialog.NewCustom("安装提示", "关闭", content, window)
	// 设置对话框大小
	customDialog.Resize(fyne.NewSize(600, 500))
	customDialog.Show()

	// 添加关闭对话框时退出程序的回调
	customDialog.SetOnClosed(func() {
		window.Close()
	})
}

//go:embed tiktok.png
var iconBytes []byte

func main() {
	myApp := app.New()
	myApp.SetIcon(&fyne.StaticResource{
		StaticName:    "icon",
		StaticContent: iconBytes,
	})
	window := myApp.NewWindow("抖音直播推流配置抓取")
	// 设置主窗口大小
	window.Resize(fyne.NewSize(600, 300))
	window.SetFixedSize(true)
	window.CenterOnScreen()

	if !checkNpcapInstalled() {
		showInstallDialog(window)
		window.ShowAndRun()
		return
	}

	status := widget.NewLabel("等待开始抓包...")
	serverAddr := widget.NewEntry()
	serverAddr.SetPlaceHolder("服务器地址")
	serverAddr.Resize(fyne.NewSize(200, serverAddr.MinSize().Height))
	streamKey := widget.NewEntry()
	streamKey.SetPlaceHolder("推流码")
	streamKey.Resize(fyne.NewSize(200, streamKey.MinSize().Height))

	copyServerBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		window.Clipboard().SetContent(serverAddr.Text)
		status.SetText("已复制服务器地址")
	})

	copyStreamBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		window.Clipboard().SetContent(streamKey.Text)
		status.SetText("已复制推流码")
	})

	// 创建抓包按钮（合并开始和停止功能）
	captureBtn := widget.NewButton("开始抓包", nil)
	captureBtn.Importance = widget.HighImportance // 设置为绿色
	captureBtn.Resize(captureBtn.MinSize())       // 设置为自适应宽度

	// 创建重启按钮
	restartBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		confirmDialog := dialog.NewCustomConfirm("确认重启", "确定", "取消",
			widget.NewLabel("确定要重启程序吗? "),
			func(ok bool) {
				if ok {
					if isCapturing {
						stopCapturing()
					}

					exe, err := os.Executable()
					if err != nil {
						dialog.ShowError(err, window)
						return
					}

					cmd := exec.Command(exe)
					err = cmd.Start()
					if err != nil {
						dialog.ShowError(err, window)
						return
					}

					myApp.Quit()
				}
			}, window)
		confirmDialog.Show()
	})
	restartBtn.Importance = widget.WarningImportance // 设置为黄色
	restartBtn.Resize(restartBtn.MinSize())          // 设置为自适应宽度

	// 设置抓包按钮的回调函数
	captureBtn.OnTapped = func() {
		if !isCapturing {
			// 开始抓包
			isCapturing = true
			stopCapture = make(chan struct{})
			status.SetText("正在查找网络接口...")

			// 更改按钮样式为停止状态
			captureBtn.SetText("停止抓包")
			captureBtn.Importance = widget.DangerImportance // 设置为红色

			// 在开始新的抓包时清空数据
			serverAddr.SetText("")
			streamKey.SetText("")

			allDevices, err := pcap.FindAllDevs()
			if err != nil {
				status.SetText("错误: " + err.Error())
				return
			}

			devices := make([]pcap.Interface, 0)
			for _, device := range allDevices {
				if strings.Contains(device.Description, "Bluetooth") ||
					strings.Contains(device.Description, "loopback") {
					continue
				}
				devices = append(devices, device)
			}

			for _, device := range devices {
				if isDebug {
					fmt.Printf("正在监听网络接口: %s (%s)\n", device.Name, device.Description)
				}
				go func(deviceName string) {
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

					status.SetText("正在抓包...")

					packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
					for {
						select {
						case <-stopCapture:
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
								if isDebug {
									fmt.Printf("发现包含 rtmp:// 的数据包: %s\n", payload)
								}
								serverRegex := regexp.MustCompile(`(rtmp://push-rtmp-[^\.]+\.douyincdn\.com/thirdgame)`)
								matches := serverRegex.FindStringSubmatch(payload)

								if isDebug {
									fmt.Printf("服务器地址匹配结果: %v\n", matches)
								}

								if len(matches) >= 1 {
									serverUrl := matches[1]
									serverAddr.SetText(serverUrl)
									status.SetText("已找到推流服务器地址")
									if isDebug {
										fmt.Printf("找到服务器地址: %s\n", serverUrl)
									}
								}
							}

							streamPatterns := []string{
								`(stream-\d+\?expire=\d+&sign=[a-f0-9]+(?:&volcSecret=[a-f0-9]+&volcTime=\d+)?)`, // 匹配两种格式的推流码
							}

							for _, pattern := range streamPatterns {
								streamRegex := regexp.MustCompile(pattern)

								matches := streamRegex.FindStringSubmatch(payload)

								if len(matches) >= 2 {
									streamKey.SetText(matches[1])
									status.SetText("已找到推流码")
									fmt.Printf("找到推流码: %s\n", matches[1])
									break
								}
							}
						}
					}
				}(device.Name)
			}
		} else {
			// 停止抓包
			stopCapturing()
			status.SetText("已停止抓包")

			// 更改按钮样式为开始状态
			captureBtn.SetText("开始抓包")
			captureBtn.Importance = widget.HighImportance // 设置为绿色
		}
	}

	// 创建固定宽度的标签
	serverLabel := widget.NewLabel("服务器地址:")
	streamLabel := widget.NewLabel("推 流 码 :") // 使用空格补齐，使其与"服务器地址"等宽

	// 创建分组：推流配置
	configTitle := container.NewBorder(
		nil, nil,
		widget.NewRichTextFromMarkdown("## 推流配置"),
		container.NewHBox(
			captureBtn,
			restartBtn,
		),
	)

	configGroup := widget.NewCard("", "", container.NewVBox(
		configTitle,
		container.NewBorder(nil, nil, serverLabel, copyServerBtn,
			serverAddr,
		),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, streamLabel, copyStreamBtn,
			streamKey,
		),
	))

	// 创建使用说明按钮
	helpBtn := widget.NewButtonWithIcon("", theme.HelpIcon(), func() {
		// 创建可滚动的说明内容
		scroll := container.NewScroll(
			container.NewVBox(
				widget.NewLabel("1. 点击\"开始抓包\"按钮"),
				widget.NewLabel("2. 打开抖音直播"),
				widget.NewLabel("3. 等待自动获取推流配置"),
				widget.NewLabel("4. 如需停止请点击\"停止抓包\"按钮"),
			),
		)

		var helpDialog dialog.Dialog
		// 创建固定布局，底部放置关闭按钮
		content := container.NewBorder(
			nil,    // 顶部无内容
			nil,    // 底部无内容
			nil,    // 左侧无内容
			nil,    // 右侧无内容
			scroll, // 中间放置可滚动内容
		)

		helpDialog = dialog.NewCustom("使用说明", "关闭", content, window) // 不使用默认的关闭按钮
		helpDialog.Resize(fyne.NewSize(300, 220))                    // 设置对话框大小
		helpDialog.Show()
		//helpDialog.Hide()
	})

	// 创建状态栏
	statusBar := container.NewHBox(
		widget.NewIcon(theme.InfoIcon()),
		status,
		helpBtn,
	)

	content := container.NewVBox(
		configGroup,
		widget.NewSeparator(),
		statusBar,
	)

	// 添加内边距
	paddedContent := container.NewPadded(content)

	window.SetContent(paddedContent)
	window.Resize(fyne.NewSize(600, 200))
	window.CenterOnScreen()
	window.ShowAndRun()
}
