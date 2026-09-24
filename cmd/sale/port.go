package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// resolvePort returns a port to serve on. If the requested port is already in
// use it prints instructions for freeing it (including the owning PID via
// `kill <pid>` / `taskkill`) and falls back to the next available port so, for
// example, a frontend and backend can both run at the same time. The original
// port is reused as soon as it is free and the command is restarted.
func resolvePort(requested int) int {
	if canBind(requested) {
		return requested
	}

	pid := processOnPort(requested)
	if pid == "" {
		pid = "unknown"
	}

	fmt.Fprintf(os.Stderr, "\nPort %d is already in use (PID %s).\n", requested, pid)
	fmt.Fprintf(os.Stderr, "Free it first with one of:\n")
	if runtime.GOOS == "windows" {
		fmt.Fprintf(os.Stderr, "  taskkill /PID %s /F\n", pid)
		fmt.Fprintf(os.Stderr, "  netstat -ano | findstr :%d\n", requested)
	} else {
		fmt.Fprintf(os.Stderr, "  lsof -i :%d\n", requested)
	}
	fmt.Fprintf(os.Stderr, "  kill %s\n", pid)

	for candidate := requested + 1; candidate < requested+100; candidate++ {
		if canBind(candidate) {
			fmt.Fprintf(os.Stderr, "\nServing on http://localhost:%d instead.\n", candidate)
			fmt.Fprintf(os.Stderr, "Free port %d and run this command again to use it.\n", requested)
			fmt.Fprintf(os.Stderr, "\n")
			return candidate
		}
	}

	fmt.Fprintf(os.Stderr, "No fallback port available; trying %d anyway.\n", requested)
	return requested
}

// canBind reports whether the port can be bound right now.
func canBind(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// processOnPort returns the PID of the process listening on the port, or an
// empty string if it cannot be determined. It is best effort: missing tools are
// handled silently.
func processOnPort(port int) string {
	if runtime.GOOS == "windows" {
		out, err := exec.Command("netstat", "-ano").Output()
		if err != nil {
			return ""
		}
		suffix := fmt.Sprintf(":%d", port)
		for _, line := range strings.Split(string(out), "\n") {
			if !strings.Contains(strings.ToUpper(line), "LISTENING") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) >= 2 && strings.HasSuffix(fields[1], suffix) {
				return fields[len(fields)-1]
			}
		}
		return ""
	}

	out, err := exec.Command("lsof", "-t", "-i", fmt.Sprintf(":%d", port)).Output()
	if err == nil {
		pids := strings.Fields(string(out))
		if len(pids) > 0 {
			return pids[0]
		}
	}
	return ""
}
