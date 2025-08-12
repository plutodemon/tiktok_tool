package ui

import (
	"testing"
	"time"

	"tiktok_tool/lkit"
)

func TestName(t *testing.T) {
	get := make(chan struct{})

	go func() {
		time.Sleep(5 * time.Second)
		close(get)
	}()

	timeout := time.After(2 * time.Second)
	select {
	case <-get:
		t.Log("Received signal")
	case <-timeout:
		t.Log("timeout")
	}

	t.Log("finish")
}

func TestClick(t *testing.T) {
	if err := lkit.SimulateRightClick(15, 15); err != nil {
		t.Errorf("模拟鼠标左键点击失败: %v", err)
	} else {
		t.Log("模拟鼠标左键点击成功")
	}
}

func TestUI(t *testing.T) {
	NewMainWindow()
}
