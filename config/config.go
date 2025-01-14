package config

import (
	"github.com/google/gopacket/pcap"
	"sync"
)

var Debug = ""

var IsDebug = false

func init() {
	if Debug == "true" {
		IsDebug = true
	}
}

// 全局变量
var (
	IsCapturing bool
	StopCapture chan struct{}
	Handles     []*pcap.Handle
	HandleMutex sync.Mutex
)
