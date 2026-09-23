//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

// ---- Win32 declarations (same pattern as cmd/installer) ----

const (
	hKEYLocalMachine = 0x80000002
	hKEYCurrentUser  = 0x80000001
	kEYRead          = 0x00020019
	kEYWrite         = 0x00020006
	rEGExpandSz      = uint32(2)

	environmentSystemKey = `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`
	environmentUserKey   = `Environment`

	wmSettingChange = 0x001A
)

var (
	advapi32 = syscall.NewLazyDLL("advapi32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procRegOpenKeyExW       = advapi32.NewProc("RegOpenKeyExW")
	procRegQueryValueExW    = advapi32.NewProc("RegQueryValueExW")
	procRegSetValueExW      = advapi32.NewProc("RegSetValueExW")
	procRegCloseKey         = advapi32.NewProc("RegCloseKey")
	procOpenProcessToken    = advapi32.NewProc("OpenProcessToken")
	procGetTokenInformation = advapi32.NewProc("GetTokenInformation")
	procCloseHandle         = kernel32.NewProc("CloseHandle")
	procGetCurrentProcess   = kernel32.NewProc("GetCurrentProcess")
	procShellExecuteExW     = shell32.NewProc("ShellExecuteExW")
	procWaitForSingleObject = kernel32.NewProc("WaitForSingleObject")
	procGetExitCodeProcess  = kernel32.NewProc("GetExitCodeProcess")
	procSendMessageTimeoutW = user32.NewProc("SendMessageTimeoutW")
)

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

func utf16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func isLocalElevated() bool {
	const (
		tokenQuery    = 0x0008
		tokenElevation = 20
	)
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

// relaunchElevated re-launches this same executable as administrator via UAC
// and waits for it to finish. Returns true if the elevated uninstall succeeded.
func relaunchElevated() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	params := "-g uninstall --elevated"
	var sei shellExecuteInfo
	sei.CbSize = uint32(unsafe.Sizeof(sei))
	sei.FMask = 0x00000040 // SEE_MASK_NOCLOSEPROCESS
	sei.LpVerb = utf16Ptr("runas")
	sei.LpFile = utf16Ptr(exe)
	sei.LpParameters = utf16Ptr(params)
	sei.NShow = 0 // SW_HIDE
	ok, _, _ := procShellExecuteExW.Call(uintptr(unsafe.Pointer(&sei)))
	if ok == 0 || sei.HProcess == 0 {
		return false
	}
	procWaitForSingleObject.Call(sei.HProcess, 600000)
	var code uint32
	procGetExitCodeProcess.Call(sei.HProcess, uintptr(unsafe.Pointer(&code)))
	procCloseHandle.Call(sei.HProcess)
	return code == 0
}

// ---- registry helpers ----

func regOpen(root uintptr, subkey string, access uint32) (uintptr, error) {
	var k uintptr
	r, _, _ := procRegOpenKeyExW.Call(root, uintptr(unsafe.Pointer(utf16Ptr(subkey))), 0, uintptr(access), uintptr(unsafe.Pointer(&k)))
	if r != 0 {
		return 0, fmt.Errorf("registry open failed (%v)", r)
	}
	return k, nil
}

func regQuery(k uintptr, name string) (string, error) {
	var size uint32
	r, _, _ := procRegQueryValueExW.Call(k, uintptr(unsafe.Pointer(utf16Ptr(name))), 0, 0, 0, uintptr(unsafe.Pointer(&size)))
	if r != 0 {
		return "", fmt.Errorf("registry query failed (%v)", r)
	}
	buf := make([]uint16, size/2+2)
	r2, _, _ := procRegQueryValueExW.Call(k, uintptr(unsafe.Pointer(utf16Ptr(name))), 0, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
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
	r, _, _ := procRegSetValueExW.Call(k, uintptr(unsafe.Pointer(utf16Ptr(name))), 0, uintptr(rEGExpandSz), uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)))
	if r != 0 {
		return fmt.Errorf("registry set failed (%v)", r)
	}
	return nil
}

func readPath(root uintptr, subkey string) (string, error) {
	k, err := regOpen(root, subkey, kEYRead)
	if err != nil {
		return "", err
	}
	defer procRegCloseKey.Call(k)
	return regQuery(k, "Path")
}

func writePath(root uintptr, subkey, value string) error {
	k, err := regOpen(root, subkey, kEYRead|kEYWrite)
	if err != nil {
		return err
	}
	defer procRegCloseKey.Call(k)
	return regSet(k, "Path", value)
}

// ---- PATH helpers ----

func pathHasEntry(path string, entries []string) bool {
	for _, part := range strings.Split(path, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		for _, e := range entries {
			if strings.EqualFold(part, e) {
				return true
			}
		}
	}
	return false
}

func pathWithoutEntries(path string, entries []string) (string, bool) {
	var kept []string
	changed := false
	for _, part := range strings.Split(path, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		remove := false
		for _, e := range entries {
			if strings.EqualFold(part, e) {
				remove = true
				break
			}
		}
		if remove {
			changed = true
		} else {
			kept = append(kept, part)
		}
	}
	return strings.Join(kept, ";"), changed
}

// ---- install locations ----

func knownInstallDirs() []string {
	var dirs []string
	if pf := os.Getenv("ProgramFiles"); pf != "" {
		dirs = append(dirs, filepath.Join(pf, "SayLess"))
	}
	if la := os.Getenv("LOCALAPPDATA"); la != "" {
		dirs = append(dirs, filepath.Join(la, "Programs", "SayLess"))
	}
	return dirs
}

func isWithin(child, parent string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

// removeInstallDirExceptSelf deletes everything in dir except the currently
// running sale.exe (which is locked while in use) and its parent chain.
func removeInstallDirExceptSelf(dir, exePath string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.EqualFold(path, exePath) {
			return nil
		}
		os.Remove(path)
		return nil
	})
	exeDir := filepath.Dir(exePath)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		if strings.EqualFold(path, dir) || strings.EqualFold(path, exeDir) {
			return nil
		}
		os.Remove(path)
		return nil
	})
}

// scheduleSelfCleanup spawns a detached cmd that removes the directory a few
// seconds later, once this running sale.exe has exited.
func scheduleSelfCleanup(dir string) {
	cmdLine := fmt.Sprintf(`ping -n 3 127.0.0.1 >nul & rd /s /q "%s"`, dir)
	c := exec.Command("cmd", "/c", cmdLine)
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	c.Start()
}

func broadcastEnvChange() {
	var result uintptr
	procSendMessageTimeoutW.Call(
		0xFFFF, // HWND_BROADCAST
		wmSettingChange,
		0,
		uintptr(unsafe.Pointer(utf16Ptr("Environment"))),
		0x0002, // SMTO_ABORTIFHUNG
		5000,
		uintptr(unsafe.Pointer(&result)),
	)
}

// cmdUninstall removes Say Less from the system: install files, PATH entries
// (system and user), and requests elevation when system-level changes are
// needed. Invoked via `sale -g uninstall`.
func cmdUninstall() {
	installDirs := knownInstallDirs()
	binDirs := make([]string, 0, len(installDirs))
	for _, d := range installDirs {
		binDirs = append(binDirs, filepath.Join(d, "bin"))
	}

	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	selfDir := filepath.Dir(exePath)

	elevated := false
	for _, a := range os.Args {
		if a == "--elevated" {
			elevated = true
			break
		}
	}

	// System-level work (system PATH entry or Program Files install) needs admin.
	needsAdmin := false
	if sysPath, err := readPath(hKEYLocalMachine, environmentSystemKey); err == nil && pathHasEntry(sysPath, binDirs) {
		needsAdmin = true
	}
	if pf := os.Getenv("ProgramFiles"); pf != "" {
		if _, err := os.Stat(filepath.Join(pf, "SayLess")); err == nil {
			needsAdmin = true
		}
	}

	if !elevated && !isLocalElevated() && needsAdmin {
		fmt.Println("Removing Say Less needs administrator permission to update the system PATH and remove files from Program Files.")
		fmt.Println("Requesting administrator permission (UAC)...")
		if relaunchElevated() {
			fmt.Println("Say Less has been uninstalled.")
			fmt.Println("Open a new terminal to confirm the 'sale' command is gone.")
			os.Exit(0)
		}
		fmt.Println("Elevation declined - removing only the current user's files and PATH entries.")
	}

	var removedFiles []string
	scheduledCleanup := ""

	for _, dir := range installDirs {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}
		if isWithin(selfDir, dir) || strings.EqualFold(filepath.Clean(selfDir), filepath.Clean(dir)) {
			removeInstallDirExceptSelf(dir, exePath)
			removedFiles = append(removedFiles, dir+" (cleanup scheduled)")
			scheduledCleanup = dir
		} else {
			os.RemoveAll(dir)
			removedFiles = append(removedFiles, dir)
		}
	}

	sysChanged := false
	if sysPath, err := readPath(hKEYLocalMachine, environmentSystemKey); err == nil {
		if cleaned, ok := pathWithoutEntries(sysPath, binDirs); ok {
			if isLocalElevated() {
				if err := writePath(hKEYLocalMachine, environmentSystemKey, cleaned); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not update the system PATH: %v\n", err)
				} else {
					sysChanged = true
				}
			} else {
				fmt.Fprintln(os.Stderr, "Warning: the system PATH still contains Say Less. Re-run 'sale -g uninstall' from an administrator terminal.")
			}
		}
	}

	userChanged := false
	if userPath, err := readPath(hKEYCurrentUser, environmentUserKey); err == nil {
		if cleaned, ok := pathWithoutEntries(userPath, binDirs); ok {
			if err := writePath(hKEYCurrentUser, environmentUserKey, cleaned); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not update the user PATH: %v\n", err)
			} else {
				userChanged = true
			}
		}
	}

	if sysChanged || userChanged {
		broadcastEnvChange()
	}

	if len(removedFiles) == 0 && !sysChanged && !userChanged {
		fmt.Println("Say Less does not appear to be installed on this system.")
		os.Exit(0)
	}

	for _, f := range removedFiles {
		fmt.Printf("Removed %s\n", f)
	}
	if scheduledCleanup != "" {
		fmt.Println("The running copy will be removed once this process exits.")
	}
	fmt.Println("Say Less has been uninstalled.")
	if sysChanged || userChanged {
		fmt.Println("Open a new terminal to confirm the 'sale' command is gone.")
	}
	os.Exit(0)
}