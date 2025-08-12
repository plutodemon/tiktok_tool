package lkit

import (
	"fmt"
	"strings"
	"time"
	"unsafe"
)

const (
	// 鼠标输入相关常量
	INPUT_MOUSE           = 0
	MOUSEEVENTF_LEFTDOWN  = 0x0002
	MOUSEEVENTF_LEFTUP    = 0x0004
	MOUSEEVENTF_RIGHTDOWN = 0x0008
	MOUSEEVENTF_RIGHTUP   = 0x0010
)

// INPUT 结构体用于SendInput API
type INPUT struct {
	Type uint32
	Mi   MouseInput
}

// MouseInput 结构体定义鼠标输入
type MouseInput struct {
	Dx          int32
	Dy          int32
	MouseData   uint32
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

// SimulateMouseClick 模拟鼠标点击指定坐标
// x: 屏幕X坐标
// y: 屏幕Y坐标
// button: 鼠标按钮类型 ("left" 或 "right")
// 返回值: 成功返回nil，失败返回错误信息
func SimulateMouseClick(x, y int, button string) error {
	if x < 0 || y < 0 {
		return fmt.Errorf("坐标不能为负数: x=%d, y=%d", x, y)
	}

	// 移动鼠标到指定位置
	ret, _, err := procSetCursorPos.Call(uintptr(x), uintptr(y))
	if ret == 0 {
		return fmt.Errorf("移动鼠标失败: %v", err)
	}

	// 短暂延迟确保鼠标移动完成
	time.Sleep(5 * time.Millisecond)

	// 根据按钮类型设置鼠标事件标志
	var downFlag, upFlag uint32
	switch strings.ToLower(button) {
	case "left", "":
		downFlag = MOUSEEVENTF_LEFTDOWN
		upFlag = MOUSEEVENTF_LEFTUP
	case "right":
		downFlag = MOUSEEVENTF_RIGHTDOWN
		upFlag = MOUSEEVENTF_RIGHTUP
	default:
		return fmt.Errorf("不支持的鼠标按钮类型: %s (支持 'left' 或 'right')", button)
	}

	// 创建鼠标按下事件
	inputDown := INPUT{
		Type: INPUT_MOUSE,
		Mi: MouseInput{
			Dx:      0,
			Dy:      0,
			DwFlags: downFlag,
		},
	}

	// 创建鼠标释放事件
	inputUp := INPUT{
		Type: INPUT_MOUSE,
		Mi: MouseInput{
			Dx:      0,
			Dy:      0,
			DwFlags: upFlag,
		},
	}

	// 发送鼠标按下事件
	ret, _, err = procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&inputDown)),
		uintptr(unsafe.Sizeof(inputDown)),
	)
	if ret == 0 {
		return fmt.Errorf("发送鼠标按下事件失败: %v", err)
	}

	// 短暂延迟模拟真实点击
	time.Sleep(10 * time.Millisecond)

	// 发送鼠标释放事件
	ret, _, err = procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&inputUp)),
		uintptr(unsafe.Sizeof(inputUp)),
	)
	if ret == 0 {
		return fmt.Errorf("发送鼠标释放事件失败: %v", err)
	}

	return nil
}

// SimulateLeftClick 模拟鼠标左键点击指定坐标的便捷方法
// x: 屏幕X坐标
// y: 屏幕Y坐标
// 返回值: 成功返回nil，失败返回错误信息
func SimulateLeftClick(x, y int) error {
	return SimulateMouseClick(x, y, "left")
}

// SimulateRightClick 模拟鼠标右键点击指定坐标的便捷方法
// x: 屏幕X坐标
// y: 屏幕Y坐标
// 返回值: 成功返回nil，失败返回错误信息
func SimulateRightClick(x, y int) error {
	return SimulateMouseClick(x, y, "right")
}
