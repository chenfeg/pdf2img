# PDF 转图片工具 (pdf2img)

一款跨平台（Windows / Linux / macOS）的 PDF 转图片工具，支持将多页 PDF 合并为单张长图，自动去除页面间留白。

## 功能特性

- **多页合并** — 将 PDF 所有页面拼接为一张完整长图
- **智能去白边** — 自动检测并裁剪页面顶部/底部的空白区域（99.8% 背景阈值，保留 10px 安全边距）
- **多格式输出** — 支持 PNG / JPG / JPEG 格式，JPEG 质量 95
- **高分辨率渲染** — 300 DPI 输出，保证文字和图像清晰
- **原生文件对话框** — Windows 下使用系统原生对话框（comdlg32 / shell32），不受应用窗口尺寸限制
- **炫酷进度条** — 渐变色分段动画进度条（青色 → 紫色），带微光效果
- **卡片式 UI** — Fyne 深色主题，分组对齐，视觉层次清晰
- **应用图标** — 窗口图标 + Windows .exe 图标

## 界面预览

```
┌──────────────────────────────────────────┐
│  ┌─ PDF 文件 ──────────────────────────┐ │
│  │ [选择PDF文件]  document.pdf         │ │
│  └─────────────────────────────────────┘ │
│                                          │
│  ┌─ 保存位置 ──────────────────────────┐ │
│  │ [选择保存目录]  D:\Documents        │ │
│  └─────────────────────────────────────┘ │
│                                          │
│  ┌─ 输出设置 ──────────────────────────┐ │
│  │ 格式:    [PNG ▼]                    │ │
│  │ 文件名:  [document___________]      │ │
│  │ * 同名图片文件将被直接覆盖           │ │
│  └─────────────────────────────────────┘ │
│                                          │
│       [======== 开始转换 ========]       │
│       [████████████░░░░  75%████]       │
│       正在处理第 3/4 页...              │
└──────────────────────────────────────────┘
```

## 系统要求

- Go 1.24+
- Windows / Linux / macOS
- `libmupdf.dll`（Windows）或对应平台的 MuPDF 动态链接库

## 构建方法

### Windows

项目使用 `purego`（无 CGO）方式调用 MuPDF，需要外部 DLL。

```bash
# 1. 安装依赖
go mod tidy

# 2. 获取 libmupdf.dll
#    从 PyMuPDF wheel 包中提取 mupdfcpp64.dll，重命名为 libmupdf.dll
#    放置到项目根目录或编译后的 exe 同目录

# 3. 生成应用图标资源（可选）
fyne bundle -package main -name appIcon appicon.png > bundled.go
# 生成 Windows .exe 图标:
# rsrc -ico appicon.ico -o icon_windows.syso

# 4. 编译（必须使用 nocgo 标签）
go build -tags nocgo -o pdf2img.exe .
```

> **为什么用 `-tags nocgo`？**
> go-fitz 默认使用 CGO 编译，需要完整的 C 编译器和 MuPDF 静态库。
> 使用 `nocgo` 标签会启用 `purego` 模式，通过 `dlopen` / `LoadLibrary` 动态加载 DLL，
> 无需 CGO 即可调用 MuPDF，大幅简化跨平台编译。

### Linux / macOS

```bash
# 安装 MuPDF 共享库
# Ubuntu/Debian:
sudo apt install libmupdf-dev

# macOS:
brew install mupdf

# 编译
go build -tags nocgo -o pdf2img .
```

> Linux/macOS 下文件对话框使用 Fyne 内置对话框。

## 使用方法

1. 启动程序
2. 点击 **「选择PDF文件」** 选择要转换的 PDF 文件
3. 保存位置默认与 PDF 文件相同目录，可点击 **「选择保存目录」** 更改
4. 选择输出格式（PNG / JPG / JPEG）
5. 输入输出文件名（默认使用 PDF 文件名）
6. 点击 **「开始转换」**
7. 等待进度条完成，图片自动保存

## 项目结构

```
pdf2img/
├── main.go                     # 主入口，GUI 布局
├── bundled.go                  # 应用图标（fyne bundle 生成）
├── nativedialog_windows.go     # Windows 原生文件/目录对话框
├── converter/
│   └── converter.go            # PDF 渲染、白边裁剪、页面拼接、图片编码
├── ui/
│   └── progress.go             # 炫酷渐变进度条组件
├── appicon.png                 # 应用图标 PNG（256×256）
├── appicon.ico                 # 应用图标 ICO（多尺寸）
├── icon_windows.syso           # Windows 资源文件（rsrc 生成）
├── libmupdf.dll                # MuPDF 动态库（需自行获取）
├── go.mod
└── go.sum
```

## 核心处理流程

```
PDF 文件
  │
  ▼
go-fitz 打开文档（300 DPI 渲染）
  │
  ▼
逐页渲染为 *image.RGBA
  │
  ▼
TrimWhitespace — 裁剪每页顶部/底部空白
  │  • 采样四角像素计算背景色
  │  • 99.8% 背景占比阈值 + 12 色宽容
  │  • 保留 10px 安全边距
  ▼
StitchImages — 纵向拼接所有页面
  │  • 取最大页宽，窄页居中
  │  • 白色背景填充
  ▼
EncodeImage — 编码为 PNG / JPEG（质量 95）
  │
  ▼
写入文件
```

## 技术栈

| 组件 | 技术 |
|------|------|
| GUI 框架 | [Fyne v2](https://fyne.io/) v2.8.2（深色主题） |
| PDF 渲染 | [go-fitz](https://github.com/gen2brain/go-fitz) v1.28.2（MuPDF 绑定） |
| 动态加载 | [purego](https://github.com/ebitengine/purego)（无 CGO 调用系统 DLL） |
| 原生对话框 | Win32 API `GetOpenFileNameW` / `SHBrowseForFolderW`（syscall 直接调用） |
| 图片处理 | Go 标准库 `image/draw`, `image/png`, `image/jpeg` |

## 注意事项

- **libmupdf.dll** 必须与 `pdf2img.exe` 放在同一目录下，程序启动时会检测并提示
- 同名输出文件会被直接覆盖（界面有提示）
- 白边裁剪仅处理顶部和底部，不裁剪左右边缘（避免误裁文字）
- 对于深色背景的 PDF 页面，不会进行裁剪（自动检测背景色亮度）

## License

MIT
