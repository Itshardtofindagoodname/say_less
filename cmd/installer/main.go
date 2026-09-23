//go:build windows && amd64

// Command installer builds say_less.exe, a Windows setup wizard (like Python's
// installer) that installs the Say Less language and adds it to PATH.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

const version = "0.1.0"

// ---- Win32 constants ----

const (
	wmDestroy        = 0x0002
	wmCommand        = 0x0111
	wmClose          = 0x0010
	wmCtlColorStatic = 0x0138
	wmCtlColorBtn    = 0x0135
	wmSetfont        = 0x0030
	wmSettingChange  = 0x001A

	wmUser    = 0x0400
	wmRefresh = wmUser + 1
	wmFinish  = wmUser + 2

	wsOverlappedWindow = 0x00CF0000
	wsChild            = 0x40000000
	wsVisible          = 0x10000000
	wsTabstop          = 0x00010000
	wsBorder           = 0x00800000
	wsExClientEdge     = 0x00000200

	ssLeft        = 0x00000000
	esAutohscroll = 0x0080
	bsPushButton  = 0x00000000
	bsAcheckbox   = 0x00000003
	bmGetCheck    = 0x00F0
	bmSetCheck    = 0x00F1
	bSTChecked    = 0x0001
	pbmSetRange32 = wmUser + 6
	pbmSetPos     = wmUser + 2

	btnClicked = 0
	swShow     = 1
	swHidden   = 0

	seeMaskNocloseprocess = 0x00000040
	tokenQuery            = 0x0008
	tokenElevation        = 20
)

// Control identifiers
const (
	idcBanner   = 101
	idcTitle    = 102
	idcSubtitle = 103
	idcLabelDir = 104
	idcEditDir  = 105
	idcBrowse   = 106
	idcChkPath  = 107
	idcStatus   = 108
	idcProgress = 109
	idcInstall  = 110
	idcCancel   = 111
)

// Registry constants
const (
	hKEYCurrentUser   = 0x80000001
	hKEYLocalMachine  = 0x80000002
	kEYRead           = 0x00020019
	kEYWrite          = 0x00020006
	rEGExpandSz       = uint32(2)
	eRRorFileNotFound = syscall.Errno(2)
)

// System PATH registry key
const environmentKey = `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`

type wndClassEx struct {
	CbSize, Style uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     syscall.Handle
	HIcon         syscall.Handle
	HCursor       syscall.Handle
	HbrBackground syscall.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       syscall.Handle
}

type msg struct {
	HWnd    uintptr
	Message uint32
	_       uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

type logFont struct {
	LfHeight, LfWidth, LfEscapement, LfOrientation, LfWeight                                                    int32
	LfItalic, LfUnderline, LfStrikeOut, LfCharSet, LfOutPrecision, LfClipPrecision, LfQuality, LfPitchAndFamily byte
	LfFaceName                                                                                                  [32]uint16
}

type browseInfo struct {
	HWndOwner      uintptr
	PIDLRoot       uintptr
	PszDisplayName *uint16
	LpszTitle      *uint16
	UlFlags        uint32
	Lpfn           uintptr
	LParam         uintptr
	IImage         int32
}

type initCommonControlsEx struct {
	DwSize uint32
	DwICC  uint32
}

type shellExecuteInfo struct {
	CbSize, FMask uint32
	HWnd          uintptr
	LpVerb        *uint16
	LpFile        *uint16
	LpParameters  *uint16
	LpDirectory   *uint16
	NShow         int32
	HInstApp      uintptr
	LpIDList      uintptr
	LpClass       *uint16
	HKeyClass     uintptr
	DHotKey       uint32
	HMonitor      uintptr
	HProcess      uintptr
}

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	comctl32 = syscall.NewLazyDLL("comctl32.dll")
	ole32    = syscall.NewLazyDLL("ole32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")

	procGetModuleHandleW     = kernel32.NewProc("GetModuleHandleW")
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
	procMessageBoxW          = user32.NewProc("MessageBoxW")
	procEnableWindow         = user32.NewProc("EnableWindow")
	procLoadCursorW          = user32.NewProc("LoadCursorW")
	procLoadIconW            = user32.NewProc("LoadIconW")
	procGetSysColorBrush     = user32.NewProc("GetSysColorBrush")
	procSendMessageTimeoutW  = user32.NewProc("SendMessageTimeoutW")

	procShellExecuteExW = shell32.NewProc("ShellExecuteExW")

	procCloseHandle         = kernel32.NewProc("CloseHandle")
	procWaitForSingleObject = kernel32.NewProc("WaitForSingleObject")
	procGetExitCodeProcess  = kernel32.NewProc("GetExitCodeProcess")
	procGetCurrentProcess   = kernel32.NewProc("GetCurrentProcess")

	procCreateSolidBrush    = gdi32.NewProc("CreateSolidBrush")
	procGetStockObject      = gdi32.NewProc("GetStockObject")
	procSetTextColor        = gdi32.NewProc("SetTextColor")
	procSetBkMode           = gdi32.NewProc("SetBkMode")
	procCreateFontIndirectW = gdi32.NewProc("CreateFontIndirectW")

	procInitCommonControlsEx = comctl32.NewProc("InitCommonControlsEx")
	procSHBrowseForFolderW   = shell32.NewProc("SHBrowseForFolderW")
	procSHGetPathFromIDListW = shell32.NewProc("SHGetPathFromIDListW")
	procCoTaskMemFree        = ole32.NewProc("CoTaskMemFree")

	procRegOpenKeyExW       = advapi32.NewProc("RegOpenKeyExW")
	procRegQueryValueExW    = advapi32.NewProc("RegQueryValueExW")
	procRegSetValueExW      = advapi32.NewProc("RegSetValueExW")
	procRegCloseKey         = advapi32.NewProc("RegCloseKey")
	procOpenProcessToken    = advapi32.NewProc("OpenProcessToken")
	procGetTokenInformation = advapi32.NewProc("GetTokenInformation")
)

var (
	hInstance   syscall.Handle
	mainHwnd    syscall.Handle
	classFont   syscall.Handle
	titleFont   syscall.Handle
	bannerBrush syscall.Handle
	controls    map[int]syscall.Handle

	statusMtx sync.Mutex
	statusLog []string

	state int // 0=ready, 1=installing, 2=done
)

func utf16(s string) *uint16 {
	p, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		return nil
	}
	return p
}

// ---- window proc ----

func wndProc(hwnd syscall.Handle, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case wmDestroy:
		procPostQuitMessage.Call(0)
		return 0
	case wmClose:
		procDestroyWindow.Call(uintptr(hwnd))
		return 0
	case wmCommand:
		code := uint32(wParam >> 16)
		id := uint32(wParam & 0xFFFF)
		if code == btnClicked {
			switch id {
			case idcInstall:
				onInstall(hwnd)
			case idcCancel:
				onCancel(hwnd)
			case idcBrowse:
				onBrowse(hwnd)
			}
		}
		return 0
	case wmCtlColorStatic, wmCtlColorBtn:
		return colorHandler(uintptr(hwnd), lParam)
	case wmRefresh:
		return refreshHandler(hwnd, wParam)
	case wmFinish:
		return finishHandler(hwnd, wParam)
	}
	r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(message), wParam, lParam)
	return r
}

func colorHandler(hwnd, lParam uintptr) uintptr {
	procSetBkMode.Call(lParam, 1) // TRANSPARENT
	if hwnd == uintptr(controls[idcBanner]) ||
		hwnd == uintptr(controls[idcTitle]) ||
		hwnd == uintptr(controls[idcSubtitle]) {
		procSetTextColor.Call(lParam, 0x00FFFFFF)
		return uintptr(bannerBrush)
	}
	procSetTextColor.Call(lParam, 0x00000000)
	b, _, _ := procGetSysColorBrush.Call(15) // COLOR_BTNFACE
	return b
}

func refreshHandler(hwnd syscall.Handle, wParam uintptr) uintptr {
	p := int(wParam)
	procSendMessageW.Call(uintptr(controls[idcProgress]), pbmSetPos, uintptr(p), 0)
	statusMtx.Lock()
	txt := strings.Join(statusLog, "\r\n")
	statusMtx.Unlock()
	procSetWindowTextW.Call(uintptr(controls[idcStatus]), uintptr(unsafe.Pointer(utf16(txt))))
	return 0
}

func finishHandler(hwnd syscall.Handle, wParam uintptr) uintptr {
	state = 2
	procEnableWindow.Call(uintptr(controls[idcInstall]), 1)
	procEnableWindow.Call(uintptr(controls[idcCancel]), 1)
	if wParam == 0 {
		procSetWindowTextW.Call(uintptr(controls[idcInstall]), uintptr(unsafe.Pointer(utf16("Close"))))
		procMessageBoxW.Call(
			uintptr(hwnd),
			uintptr(unsafe.Pointer(utf16("Say Less installed successfully.\r\n\r\nOpen a new terminal and run:\r\n    sale --version"))),
			uintptr(unsafe.Pointer(utf16("Say Less"))),
			0x40, // MB_ICONINFORMATION
		)
	} else {
		state = 0
		procSetWindowTextW.Call(uintptr(controls[idcInstall]), uintptr(unsafe.Pointer(utf16("Install"))))
		procMessageBoxW.Call(
			uintptr(hwnd),
			uintptr(unsafe.Pointer(utf16("Say Less installation failed.\r\nSee the message in the window for details."))),
			uintptr(unsafe.Pointer(utf16("Say Less"))),
			0x10, // MB_ICONERROR
		)
	}
	return 0
}

func onCancel(hwnd syscall.Handle) {
	procDestroyWindow.Call(uintptr(hwnd))
}

func onBrowse(hwnd syscall.Handle) {
	var bi browseInfo
	bi.HWndOwner = uintptr(hwnd)
	var disp [260]uint16
	bi.PszDisplayName = &disp[0]
	bi.LpszTitle = utf16("Select the folder where Say Less should be installed")
	bi.UlFlags = 0x0001 | 0x0040 // BIF_RETURNONLYFSDIRS | BIF_NEWDIALOGSTYLE
	pidl, _, _ := procSHBrowseForFolderW.Call(uintptr(unsafe.Pointer(&bi)))
	if pidl == 0 {
		return
	}
	var path [260]uint16
	r, _, _ := procSHGetPathFromIDListW.Call(pidl, uintptr(unsafe.Pointer(&path[0])))
	if r != 0 {
		procSetWindowTextW.Call(uintptr(controls[idcEditDir]), uintptr(unsafe.Pointer(&path[0])))
	}
	procCoTaskMemFree.Call(pidl)
}

func onInstall(hwnd syscall.Handle) {
	switch state {
	case 0:
		startInstall()
	case 2:
		procDestroyWindow.Call(uintptr(hwnd))
	}
}

func startInstall() {
	state = 1
	procEnableWindow.Call(uintptr(controls[idcInstall]), 0)
	procEnableWindow.Call(uintptr(controls[idcCancel]), 0)
	procEnableWindow.Call(uintptr(controls[idcEditDir]), 0)
	procEnableWindow.Call(uintptr(controls[idcBrowse]), 0)
	procEnableWindow.Call(uintptr(controls[idcChkPath]), 0)

	dir := strings.TrimSpace(getEditText(idcEditDir))
	if dir == "" {
		dir = defaultInstallDir()
	}
	addPath := getCheck(idcChkPath) == bSTChecked
	go runInstall(dir, addPath)
}

func setText(id int, s string) {
	if h := controls[id]; h != 0 {
		procSetWindowTextW.Call(uintptr(h), uintptr(unsafe.Pointer(utf16(s))))
	}
}

func appendStatus(line string) {
	statusMtx.Lock()
	statusLog = append(statusLog, line)
	statusMtx.Unlock()
}

func postProgress(p int, line string) {
	appendStatus(line)
	procPostMessageW.Call(uintptr(mainHwnd), wmRefresh, uintptr(p), 0)
}

func getEditText(id int) string {
	h := uintptr(controls[id])
	if h == 0 {
		return ""
	}
	n, _, _ := procGetWindowTextLengthW.Call(h)
	if n == 0 {
		return ""
	}
	buf := make([]uint16, n+2)
	procGetWindowTextW.Call(h, uintptr(unsafe.Pointer(&buf[0])), uintptr(n+1))
	return syscall.UTF16ToString(buf)
}

func getCheck(id int) uintptr {
	r, _, _ := procSendMessageW.Call(uintptr(controls[id]), bmGetCheck, 0, 0)
	return r
}

func isElevated() bool {
	// TOKEN_QUERY | GetTokenInformation(TokenElevation)
	tokenQuery := uint32(0x0008)
	tokenElevation := uint32(20)
	curProc, _, _ := procGetCurrentProcess.Call()
	var tok syscall.Handle
	if r, _, _ := procOpenProcessToken.Call(curProc, uintptr(tokenQuery), uintptr(unsafe.Pointer(&tok))); r == 0 {
		return false
	}
	defer procCloseHandle.Call(uintptr(tok))
	var elev uint32
	var sz uint32
	if r, _, _ := procGetTokenInformation.Call(uintptr(tok), uintptr(tokenElevation), uintptr(unsafe.Pointer(&elev)), 4, uintptr(unsafe.Pointer(&sz))); r == 0 {
		return false
	}
	return elev != 0
}

// runElevatedPath relaunches this same exe as administrator with --worker so it
// can update the SYSTEM PATH. Returns true if the elevated copy succeeded.
func runElevatedPath(binDir string) bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	params := "--worker \"" + binDir + "\""

	var sei shellExecuteInfo
	sei.CbSize = uint32(unsafe.Sizeof(sei))
	sei.FMask = 0x00000040 // SEE_MASK_NOCLOSEPROCESS (keeps us a handle to wait on)
	sei.HWnd = uintptr(mainHwnd)
	sei.LpVerb = utf16("runas")
	sei.LpFile = utf16(exe)
	sei.LpParameters = utf16(params)
	sei.NShow = swHidden

	ok, _, _ := procShellExecuteExW.Call(uintptr(unsafe.Pointer(&sei)))
	if ok == 0 {
		return false
	}
	if sei.HProcess == 0 {
		return false
	}
	procWaitForSingleObject.Call(sei.HProcess, 300000)
	var code uint32
	procGetExitCodeProcess.Call(sei.HProcess, uintptr(unsafe.Pointer(&code)))
	procCloseHandle.Call(sei.HProcess)
	return code == 0
}

func defaultInstallDir() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		base = os.Getenv("USERPROFILE")
	}
	return filepath.Join(base, "Programs", "SayLess")
}

// ---- installation ----

func runInstall(installDir string, addPath bool) {
	postProgress(2, "Preparing to install Say Less...")

	payload, err := resolvePayload()
	if err != nil {
		postProgress(0, "Error: "+err.Error())
		procPostMessageW.Call(uintptr(mainHwnd), wmFinish, 1, 0)
		return
	}

	binDir := filepath.Join(installDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		postProgress(0, "Error: "+err.Error())
		procPostMessageW.Call(uintptr(mainHwnd), wmFinish, 1, 0)
		return
	}

	salePath := filepath.Join(binDir, "sale.exe")
	postProgress(30, "Copying sale.exe to "+binDir+" ...")
	if err := os.WriteFile(salePath, payload, 0755); err != nil {
		postProgress(0, "Error: "+err.Error())
		procPostMessageW.Call(uintptr(mainHwnd), wmFinish, 1, 0)
		return
	}

	if addPath {
		postProgress(60, "Adding "+binDir+" to your system PATH...")
		if isElevated() {
			if err := addDirToSystemPath(binDir); err != nil {
				postProgress(0, "Error: "+err.Error())
				procPostMessageW.Call(uintptr(mainHwnd), wmFinish, 1, 0)
				return
			}
		} else {
			postProgress(60, "Requesting administrator permission to update your PATH...")
			if runElevatedPath(binDir) {
				postProgress(80, "PATH updated by an elevated Say Less copy.")
			} else {
				postProgress(80, "Admin prompt declined — adding to your personal PATH instead.")
				if err := addDirToUserPath(binDir); err != nil {
					postProgress(0, "Error: "+err.Error())
					procPostMessageW.Call(uintptr(mainHwnd), wmFinish, 1, 0)
					return
				}
			}
		}
	}

	postProgress(100, "Installation complete. You can now run 'sale' from any terminal.")
	procPostMessageW.Call(uintptr(mainHwnd), wmFinish, 0, 0)
}

func resolvePayload() ([]byte, error) {
	if len(bundledPayload) > 0 {
		return bundledPayload, nil
	}
	exe, err := os.Executable()
	if err == nil {
		nextTo := filepath.Join(filepath.Dir(exe), "sale.exe")
		if b, err := os.ReadFile(nextTo); err == nil {
			return b, nil
		}
	}
	return nil, fmt.Errorf("no bundled sale.exe found; release builds embed the language binary")
}

func addDirToUserPath(dir string) error {
	k, err := regOpen(hKEYCurrentUser, "Environment", kEYRead|kEYWrite)
	if err != nil {
		return err
	}
	defer procRegCloseKey.Call(k)
	return appendToPath(k, dir)
}

func addDirToSystemPath(dir string) error {
	k, err := regOpen(hKEYLocalMachine, environmentKey, kEYRead|kEYWrite)
	if err != nil {
		return err
	}
	defer procRegCloseKey.Call(k)
	return appendToPath(k, dir)
}

func appendToPath(k uintptr, dir string) error {
	cur, err := regQuery(k, "Path")
	if err == eRRorFileNotFound {
		cur = ""
	} else if err != nil {
		return err
	}

	if !pathInList(cur, dir) {
		if cur != "" && !strings.HasSuffix(cur, ";") {
			cur += ";"
		}
		cur += dir
		if err := regSet(k, "Path", cur); err != nil {
			return err
		}
	}

	var result uintptr
	procSendMessageTimeoutW.Call(
		0xFFFF, // HWND_BROADCAST
		wmSettingChange,
		0,
		uintptr(unsafe.Pointer(utf16("Environment"))),
		0x0002, // SMTO_ABORTIFHUNG
		5000,
		uintptr(unsafe.Pointer(&result)),
	)
	return nil
}

func pathInList(list, dir string) bool {
	dl := strings.ToLower(dir)
	for _, p := range strings.Split(list, ";") {
		if strings.ToLower(strings.TrimSpace(p)) == dl {
			return true
		}
	}
	return false
}

func regOpen(root uintptr, subkey string, access uint32) (uintptr, error) {
	var k uintptr
	r, _, _ := procRegOpenKeyExW.Call(
		root,
		uintptr(unsafe.Pointer(utf16(subkey))),
		0,
		uintptr(access),
		uintptr(unsafe.Pointer(&k)),
	)
	if r != 0 {
		return 0, fmt.Errorf("registry open failed (%v)", r)
	}
	return k, nil
}

func regQuery(k uintptr, name string) (string, error) {
	var size uint32
	r, _, _ := procRegQueryValueExW.Call(
		k,
		uintptr(unsafe.Pointer(utf16(name))),
		0, 0, 0,
		uintptr(unsafe.Pointer(&size)),
	)
	if r == uintptr(eRRorFileNotFound) {
		return "", eRRorFileNotFound
	}
	if r != 0 {
		return "", fmt.Errorf("registry query failed (%v)", r)
	}
	buf := make([]uint16, size/2+2)
	r2, _, _ := procRegQueryValueExW.Call(
		k,
		uintptr(unsafe.Pointer(utf16(name))),
		0, 0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if r2 != 0 {
		return "", fmt.Errorf("registry query failed (%v)", r2)
	}
	return syscall.UTF16ToString(buf), nil
}

func regSet(k uintptr, name, value string) error {
	u, _ := syscall.UTF16FromString(value)
	data := make([]byte, len(u)*2)
	for i, c := range u {
		data[i*2] = byte(c)
		data[i*2+1] = byte(c >> 8)
	}
	r, _, _ := procRegSetValueExW.Call(
		k,
		uintptr(unsafe.Pointer(utf16(name))),
		0,
		uintptr(rEGExpandSz),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
	)
	if r != 0 {
		return fmt.Errorf("registry set failed (%v)", r)
	}
	return nil
}

// ---- window setup ----

var className = utf16("SayLessSetupW")

func registerClass() error {
	wc := wndClassEx{}
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	wc.LpfnWndProc = syscall.NewCallback(wndProc)
	wc.HInstance = hInstance
	hIcon, _, _ := procLoadIconW.Call(0, 32512)     // IDI_APPLICATION
	hCursor, _, _ := procLoadCursorW.Call(0, 32512) // IDC_ARROW
	brush, _, _ := procGetSysColorBrush.Call(15)    // COLOR_BTNFACE
	wc.HIcon = syscall.Handle(hIcon)
	wc.HCursor = syscall.Handle(hCursor)
	wc.HbrBackground = syscall.Handle(brush)
	wc.LpszClassName = className
	r, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if r == 0 {
		return syscall.GetLastError()
	}
	return nil
}

func createWindow() error {
	var icc initCommonControlsEx
	icc.DwSize = uint32(unsafe.Sizeof(icc))
	icc.DwICC = 0x00000020 // ICC_PROGRESS_CLASS
	procInitCommonControlsEx.Call(uintptr(unsafe.Pointer(&icc)))

	h, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16("Say Less "+version+" Setup"))),
		wsOverlappedWindow|wsVisible,
		160, 120, 540, 400,
		0, 0, uintptr(hInstance), 0,
	)
	if h == 0 {
		return syscall.GetLastError()
	}
	mainHwnd = syscall.Handle(h)

	classFontH, _, _ := procGetStockObject.Call(17) // DEFAULT_GUI_FONT
	classFont = syscall.Handle(classFontH)
	titleFont = createTitleFont()
	bannerBrushH, _, _ := procCreateSolidBrush.Call(0x00FF1C58) // Say Less purple RGB(88,28,255)
	bannerBrush = syscall.Handle(bannerBrushH)

	createControls()

	procShowWindow.Call(h, swShow)
	procUpdateWindow.Call(h)
	return nil
}

func createTitleFont() syscall.Handle {
	var lf logFont
	lf.LfHeight = -24
	lf.LfWeight = 700
	lf.LfCharSet = 1 // DEFAULT_CHARSET
	for i, c := range "Segoe UI" {
		if i >= 31 {
			break
		}
		lf.LfFaceName[i] = uint16(c)
	}
	f, _, _ := procCreateFontIndirectW.Call(uintptr(unsafe.Pointer(&lf)))
	return syscall.Handle(f)
}

func newControl(id int, class, text string, style uintptr, x, y, w, h int, extra uintptr) {
	hCtrl, _, _ := procCreateWindowExW.Call(
		extra,
		uintptr(unsafe.Pointer(utf16(class))),
		uintptr(unsafe.Pointer(utf16(text))),
		style|wsChild|wsVisible,
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		uintptr(mainHwnd),
		uintptr(id),
		uintptr(hInstance), 0,
	)
	controls[id] = syscall.Handle(hCtrl)

	font := classFont
	if id == idcTitle || id == idcSubtitle {
		font = titleFont
	}
	procSendMessageW.Call(hCtrl, wmSetfont, uintptr(font), 1)
}

func createControls() {
	controls = make(map[int]syscall.Handle)

	// Banner band (Say Less purple with white title)
	newControl(idcBanner, "STATIC", "", ssLeft, 0, 0, 540, 70, 0)
	newControl(idcTitle, "STATIC", "Say Less", ssLeft, 18, 8, 320, 34, 0)
	newControl(idcSubtitle, "STATIC", "Tiny, fast general-purpose language", ssLeft, 18, 44, 400, 20, 0)

	// Form area
	newControl(idcLabelDir, "STATIC", "Installation directory:", ssLeft, 96, 96, 320, 18, 0)
	newControl(idcEditDir, "EDIT", defaultInstallDir(), esAutohscroll|wsBorder, 96, 116, 380, 24, wsExClientEdge)
	newControl(idcBrowse, "BUTTON", "Browse...", bsPushButton|wsTabstop, 486, 116, 40, 24, 0)
	newControl(idcChkPath, "BUTTON", "Add sale to your PATH (recommended)", bsAcheckbox|wsTabstop, 96, 152, 380, 20, 0)
	procSendMessageW.Call(uintptr(controls[idcChkPath]), bmSetCheck, bSTChecked, 0)

	newControl(idcStatus, "STATIC", "Click Install to start.", ssLeft, 96, 190, 430, 84, 0)
	newControl(idcProgress, "msctls_progress32", "", 0, 96, 280, 430, 16, 0)
	procSendMessageW.Call(uintptr(controls[idcProgress]), pbmSetRange32, 0, 100)
	procSendMessageW.Call(uintptr(controls[idcProgress]), pbmSetPos, 0, 0)

	newControl(idcInstall, "BUTTON", "Install", bsPushButton|wsTabstop, 336, 320, 88, 28, 0)
	newControl(idcCancel, "BUTTON", "Cancel", bsPushButton|wsTabstop, 436, 320, 88, 28, 0)
}

// Worker mode: invoked elevated (via UAC) to update the system PATH on the
// primary installer's behalf. This keeps PATH elevation internal to one exe.
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

	hm, _, _ := procGetModuleHandleW.Call(0)
	hInstance = syscall.Handle(hm)

	if err := registerClass(); err != nil {
		procMessageBoxW.Call(0, uintptr(unsafe.Pointer(utf16("Failed to initialize installer: "+err.Error()))), uintptr(unsafe.Pointer(utf16("Say Less"))), 0x10)
		os.Exit(1)
	}
	if err := createWindow(); err != nil {
		procMessageBoxW.Call(0, uintptr(unsafe.Pointer(utf16("Failed to open installer window: "+err.Error()))), uintptr(unsafe.Pointer(utf16("Say Less"))), 0x10)
		os.Exit(1)
	}

	var m msg
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if r == 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}
