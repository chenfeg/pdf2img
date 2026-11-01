//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	user32                   = syscall.NewLazyDLL("user32.dll")
	procSystemParametersInfo = user32.NewProc("SystemParametersInfoW")
	procGetSystemMetrics     = user32.NewProc("GetSystemMetrics")
	procGetDpiForSystem      = user32.NewProc("GetDpiForSystem")
)

type winRect struct {
	Left, Top, Right, Bottom int32
}

// usableScreenSize 返回主显示器工作区尺寸（不含任务栏），单位为物理像素
func usableScreenSize() (width, height int32) {
	const spiGetWorkArea = 0x0030
	var r winRect
	ok, _, _ := procSystemParametersInfo.Call(spiGetWorkArea, 0, uintptr(unsafe.Pointer(&r)), 0)
	if ok != 0 && r.Right > r.Left && r.Bottom > r.Top {
		return r.Right - r.Left, r.Bottom - r.Top
	}
	w, _, _ := procGetSystemMetrics.Call(uintptr(0)) // SM_CXSCREEN
	h, _, _ := procGetSystemMetrics.Call(uintptr(1)) // SM_CYSCREEN
	return int32(w), int32(h)
}

// systemDPIScale 返回系统 DPI 缩放比例（96 DPI 为 1.0），用于窗口创建前估算缩放
func systemDPIScale() float32 {
	if procGetDpiForSystem.Find() != nil {
		return 1
	}
	dpi, _, _ := procGetDpiForSystem.Call()
	if dpi == 0 {
		return 1
	}
	return float32(dpi) / 96
}
