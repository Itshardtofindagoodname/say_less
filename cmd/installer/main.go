//go:build windows && amd64

package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

const version = "0.1.0"

const (
	wmDestroy         = 0x0002
	wmClose           = 0x0010
	wmEraseBackground = 0x0014
	wmPaint           = 0x000F
	wmCommand         = 0x0111
	wmSetFont         = 0x0030
	wmSettingChange   = 0x001A
	wmCtlColorEdit    = 0x0133
	wmCtlColorStatic  = 0x0138
	wmCtlColorBtn     = 0x0135

	wmUser    = 0x0400
	wmRefresh = wmUser + 1
	wmFinish  = wmUser + 2

	wsOverlappedWindow = 0x00CF0000
	wsChild            = 0x40000000
	wsVisible          = 0x10000000
	wsTabstop          = 0x00010000
	wsBorder           = 0x00800000
	wsClipChildren     = 0x02000000
	wsClipSiblings     = 0x04000000
	wsExClientEdge     = 0x00000200

	ssLeft          = 0x00000000
	ssBitmap        = 0x0000000E
	ssCenterImage   = 0x00000200
	ssPathEllipsis  = 0x00000800
	stmSetImage     = 0x017E
	esAutohscroll   = 0x0080
	bsPushButton    = 0x00000000
	bsDefaultButton = 0x00000001
	bsAutoCheckBox  = 0x00000003
	bmGetCheck      = 0x00F0
	bmSetCheck      = 0x00F1
	bstChecked      = 0x0001
	pbmSetRange32   = wmUser + 6
	pbmSetPos       = wmUser + 2

	btnClicked       = 0
	swHide           = 0
	swShow           = 1
	seeMaskNoProcess = 0x00000040
	dibRgbColors     = 0
	frPrivate        = 0x00000010
	transparent      = 1
	colorWindowText  = 8
	colorButtonFace  = 15
)

const (
	stepIntro = iota
	stepLocation
	stepInstall
)

const (
	installReady = iota
	installRunning
	installDone
	installFailed
)

const (
	idcLeftPanel = iota + 100
	idcLeftEyebrow
	idcPortrait
	idcLeftTitle
	idcLeftVersion
	idcRightPanel
	idcLogo
	idcLogoFallback
	idcStepIndicator
	idcTitle
	idcDescription
	idcHeaderDivider
	idcAboutHeading
	idcAboutBody
	idcIncludedPanel
	idcIncludedText
	idcPathCheck
	idcLocationHeading
	idcLocationLabel
	idcLocationEdit
	idcBrowse
	idcValidation
	idcLocationNote
	idcProgressHeading
	idcProgress
	idcPercent
	idcProgressDetail
	idcInstallPath
	idcSuccessHeading
	idcSuccessBody
	idcSuccessPath
	idcErrorDetail
	idcFooterDivider
	idcBack
	idcAction
)

const environmentKey = `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`

const (
	hkeyCurrentUser   = 0x80000001
	hkeyLocalMachine  = 0x80000002
	keyRead           = 0x00020019
	keyWrite          = 0x00020006
	regExpandSz       = uint32(2)
	errorFileNotFound = syscall.Errno(2)
)

type windowClassEx struct {
	cbSize, style uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     syscall.Handle
	hIcon         syscall.Handle
	hCursor       syscall.Handle
	hbrBackground syscall.Handle
	pszMenuName   *uint16
	pszClassName  *uint16
	hIconSm       syscall.Handle
}

type message struct {
	hwnd    uintptr
	message uint32
	_       uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	point   struct{ x, y int32 }
}

type point struct{ x, y int32 }

type rect struct{ left, top, right, bottom int32 }

type logFont struct {
	lfHeight, lfWidth, lfEscapement, lfOrientation, lfWeight     int32
	lfItalic, lfUnderline, lfStrikeOut, lfCharSet                uint8
	lfOutPrecision, lfClipPrecision, lfQuality, lfPitchAndFamily uint8
	lfFaceName                                                   [32]uint16
}

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

type initCommonControlsEx struct {
	dwSize uint32
	dwICC  uint32
}

type shellExecuteInfo struct {
	cbSize, fMask uint32
	hwnd          uintptr
	lpVerb        *uint16
	lpFile        *uint16
	lpParameters  *uint16
	lpDirectory   *uint16
	nShow         int32
	hInstApp      uintptr
	lpIDList      uintptr
	lpClass       *uint16
	hkeyClass     uintptr
	dHotKey       uint32
	hMonitor      uintptr
	hProcess      uintptr
}

type bitmapInfoHeader struct {
	size            uint32
	width           int32
	height          int32
	planes          uint16
	bitCount        uint16
	compression     uint32
	sizeImage       uint32
	xPelsPerMeter   int32
	yPelsPerMeter   int32
	colorsUsed      uint32
	colorsImportant uint32
}

type bitmapInfo struct {
	header bitmapInfoHeader
	colors [1]uint32
}

type paintStruct struct {
	hDC       uintptr
	erase     uint32
	repaint   rect
	restore   uint32
	incUpdate uint32
}

type payloadSource struct {
	reader io.ReadCloser
	size   int64
}

type progressWriter struct {
	destination io.Writer
	total       int64
	written     int64
	update      func(int64, int64)
}

type readerOnly struct{ io.Reader }

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	comctl32 = syscall.NewLazyDLL("comctl32.dll")
	ole32    = syscall.NewLazyDLL("ole32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")
	ntdll    = syscall.NewLazyDLL("ntdll.dll")

	procGetModuleHandleW     = kernel32.NewProc("GetModuleHandleW")
	procCloseHandle          = kernel32.NewProc("CloseHandle")
	procGetCurrentProcess    = kernel32.NewProc("GetCurrentProcess")
	procWaitForSingleObject  = kernel32.NewProc("WaitForSingleObject")
	procGetExitCodeProcess   = kernel32.NewProc("GetExitCodeProcess")
	procMoveFileExW          = kernel32.NewProc("MoveFileExW")
	procRegisterClassExW     = user32.NewProc("RegisterClassExW")
	procCreateWindowExW      = user32.NewProc("CreateWindowExW")
	procDefWindowProcW       = user32.NewProc("DefWindowProcW")
	procGetMessageW          = user32.NewProc("GetMessageW")
	procTranslateMessage     = user32.NewProc("TranslateMessage")
	procDispatchMessageW     = user32.NewProc("DispatchMessageW")
	procShowWindow           = user32.NewProc("ShowWindow")
	procUpdateWindow         = user32.NewProc("UpdateWindow")
	procDestroyWindow        = user32.NewProc("DestroyWindow")
	procPostQuitMessage      = user32.NewProc("PostQuitMessage")
	procSetWindowTextW       = user32.NewProc("SetWindowTextW")
	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	procSendMessageW         = user32.NewProc("SendMessageW")
	procPostMessageW         = user32.NewProc("PostMessageW")
	procEnableWindow         = user32.NewProc("EnableWindow")
	procSetFocus             = user32.NewProc("SetFocus")
	procLoadCursorW          = user32.NewProc("LoadCursorW")
	procLoadIconW            = user32.NewProc("LoadIconW")
	procGetSysColorBrush     = user32.NewProc("GetSysColorBrush")
	procSendMessageTimeoutW  = user32.NewProc("SendMessageTimeoutW")
	procGetDlgCtrlID         = user32.NewProc("GetDlgCtrlID")
	procSetBkMode            = gdi32.NewProc("SetBkMode")
	procSetBkColor           = gdi32.NewProc("SetBkColor")
	procSetTextColor         = gdi32.NewProc("SetTextColor")
	procCreateFontIndirectW  = gdi32.NewProc("CreateFontIndirectW")
	procCreateSolidBrush     = gdi32.NewProc("CreateSolidBrush")
	procDeleteObject         = gdi32.NewProc("DeleteObject")
	procCreateDIBSection     = gdi32.NewProc("CreateDIBSection")
	procRtlMoveMemory        = ntdll.NewProc("RtlMoveMemory")
	procSetWindowSubclass    = comctl32.NewProc("SetWindowSubclass")
	procDefSubclassProc      = comctl32.NewProc("DefSubclassProc")
	procBeginPaint           = user32.NewProc("BeginPaint")
	procEndPaint             = user32.NewProc("EndPaint")
	procGetClientRect        = user32.NewProc("GetClientRect")
	procCreateCompatibleDC   = gdi32.NewProc("CreateCompatibleDC")
	procSelectObject         = gdi32.NewProc("SelectObject")
	procBitBlt               = gdi32.NewProc("BitBlt")
	procAddFontMemResourceEx = gdi32.NewProc("AddFontMemResourceEx")
	procGetDC                = user32.NewProc("GetDC")
	procReleaseDC            = user32.NewProc("ReleaseDC")
	procShellExecuteExW      = shell32.NewProc("ShellExecuteExW")
	procInitCommonControlsEx = comctl32.NewProc("InitCommonControlsEx")
	procSHBrowseForFolderW   = shell32.NewProc("SHBrowseForFolderW")
	procSHGetPathFromIDListW = shell32.NewProc("SHGetPathFromIDListW")
	procCoTaskMemFree        = ole32.NewProc("CoTaskMemFree")
	procRegOpenKeyExW        = advapi32.NewProc("RegOpenKeyExW")
	procRegQueryValueExW     = advapi32.NewProc("RegQueryValueExW")
	procRegSetValueExW       = advapi32.NewProc("RegSetValueExW")
	procRegCloseKey          = advapi32.NewProc("RegCloseKey")
	procOpenProcessToken     = advapi32.NewProc("OpenProcessToken")
	procGetTokenInformation  = advapi32.NewProc("GetTokenInformation")
	procAdjustWindowRectEx   = user32.NewProc("AdjustWindowRectEx")
	procGetSystemMetrics     = user32.NewProc("GetSystemMetrics")
	procSetProcessDPIAware   = user32.NewProc("SetProcessDPIAware")
	procGetDpiForSystem      = user32.NewProc("GetDpiForSystem")
)

var (
	hInstance      syscall.Handle
	mainHwnd       syscall.Handle
	controls       map[int]syscall.Handle
	fontRegular    syscall.Handle
	fontSemibold   syscall.Handle
	leftBrush      syscall.Handle
	rightBrush     syscall.Handle
	surfaceBrush   syscall.Handle
	dividerBrush   syscall.Handle
	whiteBrush     syscall.Handle
	logoBitmap     syscall.Handle
	portraitBitmap syscall.Handle
	fontScale      = 1

	currentStep     = stepIntro
	installStatus   = installReady
	selectedDir     string
	addToPath       bool
	installError    string
	validationError bool

	statusMutex      sync.Mutex
	statusText       string
	detailText       string
	statusPercent    int
	progressPosition int
)

func utf16(s string) *uint16 {
	ptr, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		return nil
	}
	return ptr
}

func scaled(value int) int {
	return (value*fontScale + 48) / 96
}

func setText(id int, value string) {
	if control := controls[id]; control != 0 {
		procSetWindowTextW.Call(uintptr(control), uintptr(unsafe.Pointer(utf16(value))))
	}
}

func setVisible(id int, visible bool) {
	if control := controls[id]; control != 0 {
		command := uintptr(swHide)
		if visible {
			command = swShow
		}
		procShowWindow.Call(uintptr(control), command)
	}
}

func setEnabled(id int, enabled bool) {
	if control := controls[id]; control != 0 {
		value := uintptr(0)
		if enabled {
			value = 1
		}
		procEnableWindow.Call(uintptr(control), value)
	}
}

func getEditText(id int) string {
	control := uintptr(controls[id])
	length, _, _ := procGetWindowTextLengthW.Call(control)
	if length == 0 {
		return ""
	}
	buffer := make([]uint16, length+2)
	procGetWindowTextW.Call(control, uintptr(unsafe.Pointer(&buffer[0])), length+1)
	return syscall.UTF16ToString(buffer)
}

func getCheck(id int) bool {
	result, _, _ := procSendMessageW.Call(uintptr(controls[id]), bmGetCheck, 0, 0)
	return result == bstChecked
}

func windowProc(hwnd syscall.Handle, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case wmDestroy:
		procPostQuitMessage.Call(0)
		return 0
	case wmClose:
		if installStatus == installRunning {
			setInstallerStatus("Installation in progress", "Please wait for the installation to finish before closing.")
			updateProgressUI(progressPosition)
			return 0
		}
		procDestroyWindow.Call(uintptr(hwnd))
		return 0
	case wmCommand:
		if uint32(wParam>>16) == btnClicked {
			switch uint32(wParam & 0xFFFF) {
			case idcBack:
				onBack()
			case idcAction:
				onAction()
			case idcBrowse:
				onBrowse(hwnd)
			}
		}
		return 0
	case wmEraseBackground:
		return uintptr(rightBrush)
	case wmCtlColorEdit:
		procSetBkColor.Call(lParam, 0x00FFFFFF)
		procSetTextColor.Call(lParam, 0x00251F35)
		return uintptr(whiteBrush)
	case wmCtlColorStatic:
		return staticColorHandler(lParam)
	case wmCtlColorBtn:
		procSetTextColor.Call(lParam, 0x00251F35)
		brush, _, _ := procGetSysColorBrush.Call(colorButtonFace)
		return brush
	case wmRefresh:
		refreshInstaller(wParam)
		return 0
	case wmFinish:
		finishInstaller(wParam != 0)
		return 0
	}
	result, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(message), wParam, lParam)
	return result
}

func staticColorHandler(child uintptr) uintptr {
	procSetBkMode.Call(child, transparent)
	id, _, _ := procGetDlgCtrlID.Call(child)
	switch int(id) {
	case idcLeftPanel, idcPortrait, idcLeftEyebrow, idcLeftTitle, idcLeftVersion:
		if id == idcLeftEyebrow {
			procSetTextColor.Call(child, 0x00CF4A68)
		} else if id == idcLeftVersion {
			procSetTextColor.Call(child, 0x00726A7D)
		} else {
			procSetTextColor.Call(child, 0x00251F35)
		}
		return uintptr(leftBrush)
	case idcStepIndicator:
		procSetTextColor.Call(child, 0x00CF4A68)
	case idcHeaderDivider, idcFooterDivider:
		return uintptr(dividerBrush)
	case idcIncludedPanel:
		return uintptr(surfaceBrush)
	case idcValidation:
		if validationError {
			procSetTextColor.Call(child, 0x000020C0)
		} else {
			procSetTextColor.Call(child, 0x00726A7D)
		}
	case idcDescription, idcLocationNote, idcProgressDetail, idcInstallPath, idcSuccessBody, idcSuccessPath, idcErrorDetail:
		procSetTextColor.Call(child, 0x00726A7D)
	default:
		procSetTextColor.Call(child, 0x00251F35)
	}
	return uintptr(rightBrush)
}

func loadFonts() {
	fontData := [][]byte{bundledFontRegular, bundledFontSemibold}
	for _, data := range fontData {
		if len(data) > 0 {
			var count uint32
			procAddFontMemResourceEx.Call(uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)), frPrivate, uintptr(unsafe.Pointer(&count)))
		}
	}
	fontRegular = createFont(12, 400)
	fontSemibold = createFont(12, 600)
}

func createFont(size, weight int32) syscall.Handle {
	var value logFont
	value.lfHeight = -int32(scaled(int(size)))
	value.lfWeight = weight
	value.lfCharSet = 1
	value.lfOutPrecision = 3
	value.lfClipPrecision = 2
	value.lfQuality = 5
	value.lfPitchAndFamily = 0
	for i, char := range "GoogleSansCode NFP" {
		value.lfFaceName[i] = uint16(char)
	}
	font, _, _ := procCreateFontIndirectW.Call(uintptr(unsafe.Pointer(&value)))
	return syscall.Handle(font)
}

func createBitmap(data []byte, width, height int) syscall.Handle {
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return 0
	}
	var info bitmapInfo
	info.header.size = uint32(unsafe.Sizeof(info.header))
	info.header.width = int32(width)
	info.header.height = -int32(height)
	info.header.planes = 1
	info.header.bitCount = 32
	info.header.compression = 0
	bounds := decoded.Bounds()
	scaledImage := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		sourceY := bounds.Min.Y + y*bounds.Dy()/height
		for x := 0; x < width; x++ {
			sourceX := bounds.Min.X + x*bounds.Dx()/width
			scaledImage.SetNRGBA(x, y, color.NRGBAModel.Convert(decoded.At(sourceX, sourceY)).(color.NRGBA))
		}
	}
	pixels := make([]byte, width*height*4)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := scaledImage.NRGBAAt(x, y)
			red := uint8((uint32(pixel.R)*uint32(pixel.A) + 255*(255-uint32(pixel.A))) / 255)
			green := uint8((uint32(pixel.G)*uint32(pixel.A) + 255*(255-uint32(pixel.A))) / 255)
			blue := uint8((uint32(pixel.B)*uint32(pixel.A) + 255*(255-uint32(pixel.A))) / 255)
			offset := (y*width + x) * 4
			pixels[offset] = blue
			pixels[offset+1] = green
			pixels[offset+2] = red
			pixels[offset+3] = 0xFF
		}
	}
	screenDC, _, _ := procGetDC.Call(0)
	if screenDC == 0 {
		return 0
	}
	var bits uintptr
	bitmap, _, _ := procCreateDIBSection.Call(screenDC, uintptr(unsafe.Pointer(&info)), dibRgbColors, uintptr(unsafe.Pointer(&bits)), 0, 0)
	procReleaseDC.Call(0, screenDC)
	if bitmap == 0 || bits == 0 {
		if bitmap != 0 {
			procDeleteObject.Call(bitmap)
		}
		return 0
	}
	procRtlMoveMemory.Call(bits, uintptr(unsafe.Pointer(&pixels[0])), uintptr(len(pixels)))
	return syscall.Handle(bitmap)
}

var imageSubclassCallback = syscall.NewCallback(imageSubclass)

func imageSubclass(hwnd uintptr, message uint32, wParam, lParam uintptr, subclassID uintptr, refData uintptr) uintptr {
	if message == wmPaint {
		id, _, _ := procGetDlgCtrlID.Call(hwnd)
		bitmap := portraitBitmap
		if int(id) == idcLogo {
			bitmap = logoBitmap
		}
		if bitmap != 0 {
			var paint paintStruct
			target, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&paint)))
			if target != 0 {
				var client rect
				procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&client)))
				memory, _, _ := procCreateCompatibleDC.Call(target)
				if memory != 0 {
					old, _, _ := procSelectObject.Call(memory, uintptr(bitmap))
					procBitBlt.Call(target, 0, 0, uintptr(client.right), uintptr(client.bottom), memory, 0, 0, 0x00CC0020)
					procSelectObject.Call(memory, old)
					procDeleteObject.Call(memory)
				}
				procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&paint)))
				return 1
			}
		}
	}
	result, _, _ := procDefSubclassProc.Call(hwnd, uintptr(message), wParam, lParam)
	return result
}

func subclassImage(control syscall.Handle, id int) {
	if control != 0 {
		procSetWindowSubclass.Call(uintptr(control), imageSubclassCallback, uintptr(id), 0)
	}
}

func newControl(id int, class, value string, style uintptr, x, y, width, height int, extended uintptr) {
	control, _, _ := procCreateWindowExW.Call(
		extended,
		uintptr(unsafe.Pointer(utf16(class))),
		uintptr(unsafe.Pointer(utf16(value))),
		style|wsChild|wsVisible|wsClipSiblings,
		uintptr(scaled(x)), uintptr(scaled(y)), uintptr(scaled(width)), uintptr(scaled(height)),
		uintptr(mainHwnd), uintptr(id), uintptr(hInstance), 0,
	)
	controls[id] = syscall.Handle(control)
	if control == 0 {
		return
	}
	font := fontRegular
	if id == idcTitle || id == idcLeftTitle || id == idcAboutHeading || id == idcLocationHeading ||
		id == idcProgressHeading || id == idcSuccessHeading || id == idcStepIndicator || id == idcLocationLabel {
		font = fontSemibold
	}
	if id == idcTitle {
		procSendMessageW.Call(control, wmSetFont, uintptr(createFont(28, 600)), 1)
	} else if id == idcLeftTitle || id == idcLogoFallback {
		procSendMessageW.Call(control, wmSetFont, uintptr(createFont(19, 600)), 1)
	} else if id == idcStepIndicator || id == idcLeftEyebrow || id == idcLocationLabel {
		procSendMessageW.Call(control, wmSetFont, uintptr(createFont(10, 600)), 1)
	} else if id == idcAboutHeading || id == idcLocationHeading || id == idcProgressHeading || id == idcSuccessHeading {
		procSendMessageW.Call(control, wmSetFont, uintptr(createFont(13, 600)), 1)
	} else if id == idcIncludedPanel {
		procSendMessageW.Call(control, wmSetFont, uintptr(createFont(12, 400)), 1)
	} else {
		procSendMessageW.Call(control, wmSetFont, uintptr(font), 1)
	}
}

func createControls() {
	newControl(idcLeftPanel, "STATIC", "", ssLeft, 0, 0, 340, 580, 0)
	newControl(idcLeftEyebrow, "STATIC", "LANGUAGE INSTALLER", ssLeft, 30, 30, 250, 20, 0)
	newControl(idcPortrait, "STATIC", "", ssBitmap|ssCenterImage, 26, 88, 288, 360, 0)
	if portraitBitmap != 0 {
		subclassImage(controls[idcPortrait], idcPortrait)
	}
	newControl(idcLeftTitle, "STATIC", "say_less", ssLeft, 30, 482, 250, 32, 0)
	newControl(idcLeftVersion, "STATIC", "Version "+version, ssLeft, 30, 516, 250, 20, 0)

	newControl(idcRightPanel, "STATIC", "", ssLeft, 340, 0, 580, 580, 0)
	newControl(idcLogo, "STATIC", "", ssBitmap|ssCenterImage, 372, 25, 116, 58, 0)
	if logoBitmap != 0 {
		subclassImage(controls[idcLogo], idcLogo)
		setVisible(idcLogo, true)
	} else {
		setVisible(idcLogo, false)
	}
	newControl(idcLogoFallback, "STATIC", "say_less", ssLeft, 372, 36, 160, 36, 0)
	setVisible(idcLogoFallback, logoBitmap == 0)
	newControl(idcStepIndicator, "STATIC", "STEP 1 / 3", ssPathEllipsis, 742, 43, 146, 24, 0)
	newControl(idcTitle, "STATIC", "", ssPathEllipsis, 372, 126, 516, 38, 0)
	newControl(idcDescription, "STATIC", "", ssLeft, 372, 172, 516, 42, 0)
	newControl(idcHeaderDivider, "STATIC", "", ssLeft, 372, 220, 516, 1, 0)

	newControl(idcAboutHeading, "STATIC", "A simpler way to build", ssLeft, 372, 250, 516, 24, 0)
	newControl(idcAboutBody, "STATIC", "say_less is a small, fast general-purpose language designed for clear, expressive programs.\r\n\r\nThis installer adds the language compiler and its command-line tools to your computer.", ssLeft, 372, 284, 516, 88, 0)
	newControl(idcIncludedPanel, "STATIC", "WHAT WILL BE INSTALLED\r\n\r\n  •  The say_less compiler and sale command\r\n  •  Runtime files in your selected folder", ssLeft, 372, 376, 516, 78, 0)
	newControl(idcPathCheck, "BUTTON", "Add say_less to PATH", bsAutoCheckBox|wsTabstop, 372, 470, 300, 24, 0)
	procSendMessageW.Call(uintptr(controls[idcPathCheck]), bmSetCheck, bstChecked, 0)

	newControl(idcLocationHeading, "STATIC", "Select a destination", ssLeft, 372, 250, 516, 24, 0)
	newControl(idcLocationLabel, "STATIC", "INSTALLATION LOCATION", ssLeft, 372, 288, 516, 20, 0)
	newControl(idcLocationEdit, "EDIT", defaultInstallDir(), esAutohscroll|wsBorder|wsTabstop, 372, 316, 420, 32, wsExClientEdge)
	newControl(idcBrowse, "BUTTON", "Browse", bsPushButton|wsTabstop, 804, 316, 84, 32, 0)
	newControl(idcValidation, "STATIC", "Choose a writable folder on this computer.", ssPathEllipsis, 372, 356, 516, 22, 0)
	newControl(idcLocationNote, "STATIC", "say_less will be installed in a bin folder inside this location.", ssPathEllipsis, 372, 402, 516, 22, 0)

	newControl(idcProgressHeading, "STATIC", "Ready to install", ssLeft, 372, 252, 380, 28, 0)
	newControl(idcProgress, "msctls_progress32", "", 0, 372, 298, 516, 22, 0)
	procSendMessageW.Call(uintptr(controls[idcProgress]), pbmSetRange32, 0, 100)
	procSendMessageW.Call(uintptr(controls[idcProgress]), pbmSetPos, 0, 0)
	newControl(idcPercent, "STATIC", "0%", ssPathEllipsis, 818, 330, 70, 24, 0)
	newControl(idcProgressDetail, "STATIC", "The bundled language runtime will be written to the selected folder.", ssPathEllipsis, 372, 360, 516, 24, 0)
	newControl(idcInstallPath, "STATIC", "", ssPathEllipsis, 372, 402, 516, 24, 0)

	newControl(idcSuccessHeading, "STATIC", "say_less has been installed successfully.", ssLeft, 372, 250, 516, 32, 0)
	newControl(idcSuccessBody, "STATIC", "The language compiler is ready to use. Open a new terminal to run sale --version.", ssLeft, 372, 296, 516, 48, 0)
	newControl(idcSuccessPath, "STATIC", "", ssPathEllipsis, 372, 360, 516, 24, 0)
	newControl(idcErrorDetail, "STATIC", "", ssLeft, 372, 344, 516, 96, 0)

	newControl(idcFooterDivider, "STATIC", "", ssLeft, 372, 500, 516, 1, 0)
	newControl(idcBack, "BUTTON", "Back", bsPushButton|wsTabstop, 688, 525, 92, 32, 0)
	newControl(idcAction, "BUTTON", "Next  →", bsPushButton|bsDefaultButton|wsTabstop, 796, 525, 92, 32, 0)
}

func registerClass() error {
	class := windowClassEx{
		cbSize:       uint32(unsafe.Sizeof(windowClassEx{})),
		style:        0x0003,
		lpfnWndProc:  syscall.NewCallback(windowProc),
		hInstance:    hInstance,
		pszClassName: utf16("SayLessSetup"),
	}
	icon, _, _ := procLoadIconW.Call(0, 32512)
	cursor, _, _ := procLoadCursorW.Call(0, 32512)
	class.hIcon = syscall.Handle(icon)
	class.hIconSm = syscall.Handle(icon)
	class.hCursor = syscall.Handle(cursor)
	result, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&class)))
	if result == 0 {
		return syscall.GetLastError()
	}
	return nil
}

func createWindow() error {
	var controlsInit initCommonControlsEx
	controlsInit.dwSize = uint32(unsafe.Sizeof(controlsInit))
	controlsInit.dwICC = 0x00000020
	procInitCommonControlsEx.Call(uintptr(unsafe.Pointer(&controlsInit)))

	controls = make(map[int]syscall.Handle)
	loadFonts()
	if len(bundledLogo) > 0 {
		logoBitmap = createBitmap(bundledLogo, scaled(116), scaled(58))
	}
	if len(bundledPortrait) > 0 {
		portraitBitmap = createBitmap(bundledPortrait, scaled(288), scaled(360))
	}

	brushes := []struct {
		color uintptr
		brush *syscall.Handle
	}{
		{0x00FBF4F6, &leftBrush},
		{0x00FFFFFF, &rightBrush},
		{0x00FCFAFD, &surfaceBrush},
		{0x00EAE5F0, &dividerBrush},
		{0x00FFFFFF, &whiteBrush},
	}
	for _, item := range brushes {
		brush, _, _ := procCreateSolidBrush.Call(item.color)
		*item.brush = syscall.Handle(brush)
	}

	style := uintptr(wsOverlappedWindow &^ 0x00070000)
	windowRect := rect{0, 0, int32(scaled(920)), int32(scaled(580))}
	procAdjustWindowRectEx.Call(uintptr(unsafe.Pointer(&windowRect)), style, 0, 0)
	width := windowRect.right - windowRect.left
	height := windowRect.bottom - windowRect.top
	screenWidth, _, _ := procGetSystemMetrics.Call(0)
	screenHeight, _, _ := procGetSystemMetrics.Call(1)
	x := (int(screenWidth) - int(width)) / 2
	y := (int(screenHeight) - int(height)) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	window, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16("SayLessSetup"))),
		uintptr(unsafe.Pointer(utf16("say_less "+version+" Setup"))),
		style,
		uintptr(x), uintptr(y), uintptr(width), uintptr(height),
		0, 0, uintptr(hInstance), 0,
	)
	if window == 0 {
		return syscall.GetLastError()
	}
	mainHwnd = syscall.Handle(window)
	createControls()
	setStep(stepIntro)
	procShowWindow.Call(window, swShow)
	procUpdateWindow.Call(window)
	procSetFocus.Call(uintptr(controls[idcAction]))
	return nil
}

func initializeDPI() {
	procSetProcessDPIAware.Call()
	fontScale = 96
	if procGetDpiForSystem.Find() != nil {
		dpi, _, _ := procGetDpiForSystem.Call()
		if dpi >= 96 {
			fontScale = int(dpi)
		}
	}
	screenWidth, _, _ := procGetSystemMetrics.Call(0)
	screenHeight, _, _ := procGetSystemMetrics.Call(1)
	if screenWidth > 0 {
		limit := (int(screenWidth) - 40) * 96 / 940
		if limit < fontScale {
			fontScale = limit
		}
	}
	if screenHeight > 0 {
		limit := (int(screenHeight) - 40) * 96 / 620
		if limit < fontScale {
			fontScale = limit
		}
	}
	if fontScale < 48 {
		fontScale = 48
	}
}

func allTransientControls() []int {
	return []int{
		idcAboutHeading, idcAboutBody, idcIncludedPanel, idcPathCheck,
		idcLocationHeading, idcLocationLabel, idcLocationEdit, idcBrowse, idcValidation, idcLocationNote,
		idcProgressHeading, idcProgress, idcPercent, idcProgressDetail, idcInstallPath,
		idcSuccessHeading, idcSuccessBody, idcSuccessPath, idcErrorDetail,
	}
}

func showControls(ids ...int) {
	for _, id := range ids {
		setVisible(id, true)
	}
}

func hideControls(ids ...int) {
	for _, id := range ids {
		setVisible(id, false)
	}
}

func setStep(step int) {
	currentStep = step
	hideControls(allTransientControls()...)
	setEnabled(idcBack, true)
	setEnabled(idcAction, true)
	switch step {
	case stepIntro:
		setText(idcStepIndicator, "STEP 1 / 3")
		setText(idcTitle, "Welcome to say_less")
		setText(idcDescription, "Install a small, fast language built for expressive and useful programs.")
		showControls(idcAboutHeading, idcAboutBody, idcIncludedPanel, idcPathCheck)
		setVisible(idcBack, false)
		setText(idcAction, "Next  →")
	case stepLocation:
		validationError = false
		setText(idcValidation, "Choose a writable folder on this computer.")
		setText(idcStepIndicator, "STEP 2 / 3")
		setText(idcTitle, "Choose where say_less lives")
		setText(idcDescription, "Select the folder where the language compiler and command-line tools will be installed.")
		showControls(idcLocationHeading, idcLocationLabel, idcLocationEdit, idcBrowse, idcValidation, idcLocationNote)
		setText(idcBack, "Back")
		setText(idcAction, "Next  →")
	case stepInstall:
		showReadyState()
	}
}

func showReadyState() {
	installStatus = installReady
	statusMutex.Lock()
	installError = ""
	statusMutex.Unlock()
	setText(idcStepIndicator, "STEP 3 / 3")
	setText(idcTitle, "Install say_less")
	setText(idcDescription, "The language runtime will be installed in the same window. No additional setup is required.")
	setText(idcProgressHeading, "Ready to install")
	setText(idcProgressDetail, "The bundled language runtime will be written to the selected folder.")
	setText(idcInstallPath, "Destination: "+filepath.Join(selectedDir, "bin"))
	showControls(idcProgressHeading, idcProgress, idcPercent, idcProgressDetail, idcInstallPath)
	setVisible(idcBack, true)
	setText(idcBack, "Back")
	setText(idcAction, "Install")
	setEnabled(idcBack, true)
	setEnabled(idcAction, true)
	updateProgressUI(0)
}

func onBack() {
	if installStatus == installRunning {
		return
	}
	switch currentStep {
	case stepLocation:
		setStep(stepIntro)
	case stepInstall:
		if installStatus == installDone {
			procDestroyWindow.Call(uintptr(mainHwnd))
			return
		}
		if installStatus == installFailed {
			procDestroyWindow.Call(uintptr(mainHwnd))
			return
		}
		setStep(stepLocation)
	}
	procSetFocus.Call(uintptr(controls[idcAction]))
}

func onAction() {
	if installStatus == installRunning {
		return
	}
	switch currentStep {
	case stepIntro:
		addToPath = getCheck(idcPathCheck)
		setStep(stepLocation)
		procSetFocus.Call(uintptr(controls[idcLocationEdit]))
	case stepLocation:
		dir, err := validateInstallDir(getEditText(idcLocationEdit))
		if err != nil {
			validationError = true
			setText(idcValidation, err.Error())
			procSetFocus.Call(uintptr(controls[idcLocationEdit]))
			return
		}
		validationError = false
		selectedDir = dir
		setText(idcLocationEdit, dir)
		setStep(stepInstall)
		procSetFocus.Call(uintptr(controls[idcAction]))
	case stepInstall:
		switch installStatus {
		case installDone:
			procDestroyWindow.Call(uintptr(mainHwnd))
		case installReady, installFailed:
			startInstallation()
		}
	}
}

func onBrowse(hwnd syscall.Handle) {
	var info browseInfo
	info.hwndOwner = uintptr(hwnd)
	info.lpszTitle = utf16("Select where say_less should be installed")
	info.ulFlags = 0x0001 | 0x0040
	pidl, _, _ := procSHBrowseForFolderW.Call(uintptr(unsafe.Pointer(&info)))
	if pidl == 0 {
		return
	}
	defer procCoTaskMemFree.Call(pidl)
	buffer := make([]uint16, 32768)
	result, _, _ := procSHGetPathFromIDListW.Call(pidl, uintptr(unsafe.Pointer(&buffer[0])))
	if result != 0 {
		setText(idcLocationEdit, syscall.UTF16ToString(buffer))
		validationError = false
		setText(idcValidation, "Choose a writable folder on this computer.")
	}
}

func validateInstallDir(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("Choose an installation folder.")
	}
	if !filepath.IsAbs(value) {
		return "", fmt.Errorf("Use a complete local path, such as C:\\SayLess.")
	}
	clean := filepath.Clean(value)
	if filepath.VolumeName(clean) == "" || strings.HasPrefix(clean, `\\`) {
		return "", fmt.Errorf("Use a complete local path, such as C:\\SayLess.")
	}
	if filepath.Dir(clean) == clean {
		return "", fmt.Errorf("Choose a folder inside your user or program directories.")
	}
	writablePath := clean
	if info, err := os.Stat(clean); err == nil {
		if !info.IsDir() {
			return "", fmt.Errorf("This path points to a file. Choose a folder instead.")
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("This folder cannot be used: %v", err)
	} else {
		parent := filepath.Dir(clean)
		for {
			info, statErr := os.Stat(parent)
			if statErr == nil {
				if !info.IsDir() {
					return "", fmt.Errorf("A parent of this path points to a file.")
				}
				writablePath = parent
				break
			}
			if !os.IsNotExist(statErr) {
				return "", fmt.Errorf("This folder cannot be used: %v", statErr)
			}
			next := filepath.Dir(parent)
			if next == parent {
				break
			}
			parent = next
		}
	}
	probe, err := os.CreateTemp(writablePath, ".sayless-write-test-*")
	if err != nil {
		return "", fmt.Errorf("This folder is not writable: %v", err)
	}
	probePath := probe.Name()
	if err := probe.Close(); err != nil {
		_ = os.Remove(probePath)
		return "", fmt.Errorf("This folder is not writable: %v", err)
	}
	if err := os.Remove(probePath); err != nil {
		return "", fmt.Errorf("This folder is not writable: %v", err)
	}
	return clean, nil
}

func startInstallation() {
	installStatus = installRunning
	statusMutex.Lock()
	installError = ""
	statusMutex.Unlock()
	hideControls(idcSuccessHeading, idcSuccessBody, idcSuccessPath, idcErrorDetail)
	showControls(idcProgressHeading, idcProgress, idcPercent, idcProgressDetail, idcInstallPath)
	setText(idcTitle, "Installing say_less")
	setText(idcDescription, "Please keep this window open while say_less is installed.")
	setText(idcProgressHeading, "Installing say_less")
	setText(idcProgressDetail, "Preparing the bundled language runtime...")
	setInstallerStatus("Preparing installation...", "Preparing the bundled language runtime...")
	setText(idcAction, "Installing...")
	setEnabled(idcAction, false)
	setEnabled(idcBack, false)
	updateProgressUI(0)
	go runInstallation(selectedDir, addToPath)
}

func (writer *progressWriter) Write(data []byte) (int, error) {
	written, err := writer.destination.Write(data)
	writer.written += int64(written)
	if written > 0 && writer.update != nil {
		writer.update(writer.written, writer.total)
	}
	return written, err
}

func openPayload() (payloadSource, error) {
	if len(bundledPayload) > 0 {
		return payloadSource{reader: io.NopCloser(bytes.NewReader(bundledPayload)), size: int64(len(bundledPayload))}, nil
	}
	executable, err := os.Executable()
	if err != nil {
		return payloadSource{}, err
	}
	path := filepath.Join(filepath.Dir(executable), "sale.exe")
	file, err := os.Open(path)
	if err != nil {
		return payloadSource{}, fmt.Errorf("no bundled sale.exe found; release builds include it inside the installer")
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return payloadSource{}, err
	}
	return payloadSource{reader: file, size: info.Size()}, nil
}

func runInstallation(installDir string, addPath bool) {
	payload, err := openPayload()
	if err != nil {
		postInstallFailure(err)
		return
	}
	defer payload.reader.Close()

	binDir := filepath.Join(installDir, "bin")
	postInstallProgress(2, "Preparing installation folder...", "Creating "+binDir)
	if err := os.MkdirAll(binDir, 0755); err != nil {
		postInstallFailure(fmt.Errorf("could not create %s: %v", binDir, err))
		return
	}
	if payload.size <= 0 {
		postInstallFailure(fmt.Errorf("the bundled sale.exe is empty"))
		return
	}

	tempPath := filepath.Join(binDir, "sale.exe.installing")
	_ = os.Remove(tempPath)
	file, err := os.OpenFile(tempPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0755)
	if err != nil {
		postInstallFailure(fmt.Errorf("could not start writing sale.exe: %v", err))
		return
	}
	writer := &progressWriter{destination: file, total: payload.size}
	writer.update = func(written, total int64) {
		percent := 5 + int(written*74/total)
		if percent > 79 {
			percent = 79
		}
		postInstallProgress(percent, "Writing sale.exe...", fmt.Sprintf("%.1f MB of %.1f MB", float64(written)/(1024*1024), float64(total)/(1024*1024)))
	}
	_, copyErr := io.CopyBuffer(writer, readerOnly{payload.reader}, make([]byte, 128*1024))
	if copyErr == nil && writer.written != payload.size {
		copyErr = fmt.Errorf("expected %d bytes but wrote %d", payload.size, writer.written)
	}
	if copyErr == nil {
		copyErr = file.Sync()
	}
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(tempPath)
		postInstallFailure(fmt.Errorf("could not write sale.exe: %v", copyErr))
		return
	}
	if closeErr != nil {
		_ = os.Remove(tempPath)
		postInstallFailure(fmt.Errorf("could not finish writing sale.exe: %v", closeErr))
		return
	}
	tempInfo, err := os.Stat(tempPath)
	if err != nil || tempInfo.Size() != payload.size {
		_ = os.Remove(tempPath)
		postInstallFailure(fmt.Errorf("the installed file did not pass its size check"))
		return
	}
	postInstallProgress(82, "Verifying the installed file...", fmt.Sprintf("%d bytes written", writer.written))
	finalPath := filepath.Join(binDir, "sale.exe")
	moved, _, moveError := procMoveFileExW.Call(uintptr(unsafe.Pointer(utf16(tempPath))), uintptr(unsafe.Pointer(utf16(finalPath))), 0x00000001|0x00000008)
	if moved == 0 {
		_ = os.Remove(tempPath)
		postInstallFailure(fmt.Errorf("could not finalize sale.exe: %v", moveError))
		return
	}
	finalInfo, err := os.Stat(finalPath)
	if err != nil || finalInfo.Size() != payload.size {
		postInstallFailure(fmt.Errorf("the finalized sale.exe could not be verified"))
		return
	}
	postInstallProgress(88, "Finalizing installation...", finalPath)

	if addPath {
		postInstallProgress(90, "Updating PATH...", "Adding "+binDir)
		scope, err := addDirToPath(binDir)
		if err != nil {
			postInstallFailure(fmt.Errorf("could not update PATH: %v", err))
			return
		}
		postInstallProgress(96, "Finalizing installation...", scope+" PATH updated")
	} else {
		postInstallProgress(96, "Finalizing installation...", "PATH was left unchanged")
	}
	postInstallProgress(100, "Installation complete", finalPath)
	procPostMessageW.Call(uintptr(mainHwnd), wmFinish, 0, 0)
}

func addDirToPath(dir string) (string, error) {
	if isElevated() {
		return "System", addDirToSystemPath(dir)
	}
	updated, err := runElevatedPath(dir)
	if err != nil {
		return "", err
	}
	if updated {
		return "System", nil
	}
	return "User", addDirToUserPath(dir)
}

func isElevated() bool {
	process, _, _ := procGetCurrentProcess.Call()
	var token syscall.Handle
	result, _, _ := procOpenProcessToken.Call(process, 0x0008, uintptr(unsafe.Pointer(&token)))
	if result == 0 {
		return false
	}
	defer procCloseHandle.Call(uintptr(token))
	var elevated uint32
	var size uint32
	result, _, _ = procGetTokenInformation.Call(uintptr(token), 20, uintptr(unsafe.Pointer(&elevated)), 4, uintptr(unsafe.Pointer(&size)))
	return result != 0 && elevated != 0
}

func runElevatedPath(binDir string) (bool, error) {
	executable, err := os.Executable()
	if err != nil {
		return false, err
	}
	var info shellExecuteInfo
	info.cbSize = uint32(unsafe.Sizeof(info))
	info.fMask = seeMaskNoProcess
	info.hwnd = uintptr(mainHwnd)
	info.lpVerb = utf16("runas")
	info.lpFile = utf16(executable)
	info.lpParameters = utf16("--worker " + quoteWindowsArgument(binDir))
	info.nShow = swHide
	result, _, _ := procShellExecuteExW.Call(uintptr(unsafe.Pointer(&info)))
	if result == 0 || info.hProcess == 0 {
		return false, nil
	}
	waitResult, _, _ := procWaitForSingleObject.Call(info.hProcess, 0xFFFFFFFF)
	if waitResult != 0 {
		procCloseHandle.Call(info.hProcess)
		return false, fmt.Errorf("the elevated PATH update could not be completed")
	}
	var exitCode uint32
	exitResult, _, _ := procGetExitCodeProcess.Call(info.hProcess, uintptr(unsafe.Pointer(&exitCode)))
	procCloseHandle.Call(info.hProcess)
	if exitResult == 0 {
		return false, fmt.Errorf("the elevated PATH update result could not be read")
	}
	if exitCode != 0 {
		return false, fmt.Errorf("the elevated PATH update failed")
	}
	return true, nil
}

func quoteWindowsArgument(value string) string {
	return "\"" + strings.ReplaceAll(value, "\"", "\\\"") + "\""
}

func defaultInstallDir() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		base = os.Getenv("USERPROFILE")
	}
	if base == "" {
		return filepath.Clean("say_less")
	}
	return filepath.Join(base, "Programs", "SayLess")
}

func addDirToUserPath(dir string) error {
	key, err := registryOpen(hkeyCurrentUser, "Environment", keyRead|keyWrite)
	if err != nil {
		return err
	}
	defer procRegCloseKey.Call(key)
	return appendToPath(key, dir)
}

func addDirToSystemPath(dir string) error {
	key, err := registryOpen(hkeyLocalMachine, environmentKey, keyRead|keyWrite)
	if err != nil {
		return err
	}
	defer procRegCloseKey.Call(key)
	return appendToPath(key, dir)
}

func appendToPath(key uintptr, dir string) error {
	current, err := registryQuery(key, "Path")
	if err == errorFileNotFound {
		current = ""
	} else if err != nil {
		return err
	}
	if !pathInList(current, dir) {
		if current != "" && !strings.HasSuffix(current, ";") {
			current += ";"
		}
		if err := registrySet(key, "Path", current+dir); err != nil {
			return err
		}
	}
	var result uintptr
	procSendMessageTimeoutW.Call(0xFFFF, wmSettingChange, 0, uintptr(unsafe.Pointer(utf16("Environment"))), 0x0002, 5000, uintptr(unsafe.Pointer(&result)))
	return nil
}

func pathInList(list, dir string) bool {
	target := strings.TrimSpace(dir)
	for _, item := range strings.Split(list, ";") {
		if strings.EqualFold(target, strings.TrimSpace(item)) {
			return true
		}
	}
	return false
}

func registryOpen(root uintptr, subkey string, access uint32) (uintptr, error) {
	var key uintptr
	result, _, _ := procRegOpenKeyExW.Call(root, uintptr(unsafe.Pointer(utf16(subkey))), 0, uintptr(access), uintptr(unsafe.Pointer(&key)))
	if result != 0 {
		return 0, fmt.Errorf("registry open failed (%v)", result)
	}
	return key, nil
}

func registryQuery(key uintptr, name string) (string, error) {
	var size uint32
	result, _, _ := procRegQueryValueExW.Call(key, uintptr(unsafe.Pointer(utf16(name))), 0, 0, 0, uintptr(unsafe.Pointer(&size)))
	if result == uintptr(errorFileNotFound) {
		return "", errorFileNotFound
	}
	if result != 0 {
		return "", fmt.Errorf("registry query failed (%v)", result)
	}
	buffer := make([]uint16, size/2+2)
	result, _, _ = procRegQueryValueExW.Call(key, uintptr(unsafe.Pointer(utf16(name))), 0, 0, uintptr(unsafe.Pointer(&buffer[0])), uintptr(unsafe.Pointer(&size)))
	if result != 0 {
		return "", fmt.Errorf("registry query failed (%v)", result)
	}
	return syscall.UTF16ToString(buffer), nil
}

func registrySet(key uintptr, name, value string) error {
	utf16Value, _ := syscall.UTF16FromString(value)
	data := make([]byte, len(utf16Value)*2)
	for i, value := range utf16Value {
		data[i*2] = byte(value)
		data[i*2+1] = byte(value >> 8)
	}
	result, _, _ := procRegSetValueExW.Call(key, uintptr(unsafe.Pointer(utf16(name))), 0, uintptr(regExpandSz), uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)))
	if result != 0 {
		return fmt.Errorf("registry write failed (%v)", result)
	}
	return nil
}

func setInstallerStatus(status, detail string) {
	statusMutex.Lock()
	statusText = status
	detailText = detail
	statusMutex.Unlock()
}

func postInstallProgress(percent int, status, detail string) {
	statusMutex.Lock()
	statusText = status
	detailText = detail
	statusPercent = percent
	statusMutex.Unlock()
	procPostMessageW.Call(uintptr(mainHwnd), wmRefresh, 0, 0)
}

func postInstallFailure(err error) {
	statusMutex.Lock()
	installError = err.Error()
	statusText = "Installation stopped"
	detailText = err.Error()
	statusPercent = 0
	statusMutex.Unlock()
	procPostMessageW.Call(uintptr(mainHwnd), wmRefresh, 0, 0)
	procPostMessageW.Call(uintptr(mainHwnd), wmFinish, 1, 0)
}

func updateProgressUI(percent int) {
	statusMutex.Lock()
	status := statusText
	detail := detailText
	statusMutex.Unlock()
	renderProgress(percent, status, detail)
}

func renderProgress(percent int, status, detail string) {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	progressPosition = percent
	procSendMessageW.Call(uintptr(controls[idcProgress]), pbmSetPos, uintptr(percent), 0)
	setText(idcPercent, fmt.Sprintf("%d%%", percent))
	if status != "" {
		setText(idcProgressHeading, status)
	}
	if detail != "" {
		setText(idcProgressDetail, detail)
	}
}

func refreshInstaller(uintptr) {
	statusMutex.Lock()
	percent := statusPercent
	status := statusText
	detail := detailText
	statusMutex.Unlock()
	renderProgress(percent, status, detail)
}

func finishInstaller(failed bool) {
	statusMutex.Lock()
	errorText := installError
	statusMutex.Unlock()
	if failed {
		installStatus = installFailed
		setText(idcTitle, "Installation couldn't finish")
		setText(idcDescription, "Review the error below, then retry or close the installer. No separate dialog is required.")
		hideControls(idcSuccessHeading, idcSuccessBody, idcSuccessPath)
		hideControls(idcProgress, idcProgressDetail, idcInstallPath)
		showControls(idcProgressHeading, idcPercent, idcErrorDetail)
		setText(idcProgressHeading, "Installation stopped")
		setText(idcErrorDetail, errorText)
		setText(idcBack, "Close")
		setText(idcAction, "Retry")
		setEnabled(idcBack, true)
		setEnabled(idcAction, true)
		return
	}

	installStatus = installDone
	setText(idcTitle, "Installation complete")
	setText(idcDescription, "say_less is ready to use on this computer.")
	hideControls(idcProgressHeading, idcProgress, idcPercent, idcProgressDetail, idcInstallPath)
	showControls(idcSuccessHeading, idcSuccessBody, idcSuccessPath)
	if addToPath {
		setText(idcSuccessBody, "The language compiler is ready to use. Open a new terminal to run sale --version.")
	} else {
		setText(idcSuccessBody, "The language compiler is ready to use. Run \""+filepath.Join(selectedDir, "bin", "sale.exe")+"\" --version from the installed folder.")
	}
	setText(idcSuccessPath, "Installed to: "+filepath.Join(selectedDir, "bin"))
	setVisible(idcBack, false)
	setText(idcAction, "Close")
	setEnabled(idcAction, true)
	procSetFocus.Call(uintptr(controls[idcAction]))
}

func workerMain() int {
	if len(os.Args) < 3 || os.Args[1] != "--worker" {
		return 1
	}
	if err := addDirToSystemPath(os.Args[2]); err != nil {
		return 1
	}
	return 0
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--worker" {
		os.Exit(workerMain())
	}
	initializeDPI()
	module, _, _ := procGetModuleHandleW.Call(0)
	hInstance = syscall.Handle(module)
	if err := registerClass(); err != nil {
		procMessageBox.Call(0, uintptr(unsafe.Pointer(utf16("Failed to initialize installer: "+err.Error()))), uintptr(unsafe.Pointer(utf16("say_less"))), 0x10)
		os.Exit(1)
	}
	if err := createWindow(); err != nil {
		procMessageBox.Call(0, uintptr(unsafe.Pointer(utf16("Failed to open installer window: "+err.Error()))), uintptr(unsafe.Pointer(utf16("say_less"))), 0x10)
		os.Exit(1)
	}
	var event message
	for {
		result, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&event)), 0, 0, 0)
		if int32(result) == -1 {
			break
		}
		if result == 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&event)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&event)))
	}
}

var procMessageBox = user32.NewProc("MessageBoxW")
