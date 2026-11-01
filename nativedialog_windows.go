package main

import (
	"syscall"
	"unsafe"
)

var (
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	getOpenFileNameW = comdlg32.NewProc("GetOpenFileNameW")
	browseForFolder  = shell32.NewProc("SHBrowseForFolderW")
	shGetPathFromIDListW = shell32.NewProc("SHGetPathFromIDListW")
	localFree        = kernel32.NewProc("LocalFree")
)

var kernel32 = syscall.NewLazyDLL("kernel32.dll")

const (
	ofnFileMustExist    = 0x00001000
	ofnHideReadOnly     = 0x00000004
	ofnNoChangeDir      = 0x00000008
	ofnExplorer         = 0x00080000
	ofnLongNames        = 0x00200000
	bifUseNewUI         = 0x00000040
	bifReturnOnlyFSDirs = 0x00000001
)

type openFilename struct {
	lStructSize    uint32
	hwndOwner      uintptr
	hInstance      uintptr
	lpstrFilter    *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter uint32
	nFilterIndex   uint32
	lpstrFile      *uint16
	nMaxFile       uint32
	lpstrFileTitle *uint16
	nMaxFileTitle  uint32
	lpstrInitialDir *uint16
	lpstrTitle     *uint16
	flags          uint32
	nFileOffset    uint16
	nFileExtension uint16
	lpstrDefExt    *uint16
	lCustData      uintptr
	lpfnHook       uintptr
	lpTemplateName *uint16
	pvReserved     uintptr
	dwReserved     uint32
	flagsEx        uint32
}

func nativeSelectPDF() (string, error) {
	// Build UTF-16 filter buffer manually: "PDF\0*.pdf\0\0"
	// UTF16PtrFromString stops at first \x00, so we must build the slice directly.
	filterUTF16 := append(
		syscall.StringToUTF16("PDF Files (*.pdf)"),
		syscall.StringToUTF16("*.pdf")...,
	)
	// append adds a trailing 0 from StringToUTF16, giving us the double-null terminator
	buf := make([]uint16, 260)
	title, _ := syscall.UTF16PtrFromString("选择 PDF 文件")

	ofn := openFilename{
		lStructSize:  uint32(unsafe.Sizeof(openFilename{})),
		lpstrFilter:  &filterUTF16[0],
		nFilterIndex: 1,
		lpstrFile:    &buf[0],
		nMaxFile:     uint32(len(buf)),
		lpstrTitle:   title,
		flags:        ofnFileMustExist | ofnHideReadOnly | ofnNoChangeDir | ofnExplorer | ofnLongNames,
	}

	ret, _, _ := getOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if ret == 0 {
		return "", nil // user cancelled
	}
	return syscall.UTF16ToString(buf), nil
}

func nativeSelectFolder() (string, error) {
	type browseInfo struct {
		hwndOwner      uintptr
		pidlRoot       uintptr
		pszDisplayName *uint16
		lpszTitle      *uint16
		ulFlags        uint32
		lpfn           uintptr
		lParam         uintptr
		iImage         int32
	}

	title, _ := syscall.UTF16PtrFromString("选择保存目录")
	displayName := make([]uint16, 260)

	bi := browseInfo{
		pszDisplayName: &displayName[0],
		lpszTitle:      title,
		ulFlags:        bifUseNewUI | bifReturnOnlyFSDirs,
	}

	pidl, _, _ := browseForFolder.Call(uintptr(unsafe.Pointer(&bi)))
	if pidl == 0 {
		return "", nil // user cancelled
	}
	defer localFree.Call(pidl)

	pathBuf := make([]uint16, 260)
	shGetPathFromIDListW.Call(pidl, uintptr(unsafe.Pointer(&pathBuf[0])))
	return syscall.UTF16ToString(pathBuf), nil
}
