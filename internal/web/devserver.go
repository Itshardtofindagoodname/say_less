package web

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DevServer serves the web build output with live reload and automatic
// rebuilds when .sl source files change. It uses only the Go standard library.
type DevServer struct {
	mu       sync.Mutex
	revision int

	buildDir string
	rootDir  string
	port     int
	build    func() error
}

// NewDevServer creates a dev server. It serves static files from buildDir,
// watches *.sl files under rootDir, and calls build() to rebuild on change.
func NewDevServer(buildDir, rootDir string, port int, build func() error) *DevServer {
	return &DevServer{
		buildDir: buildDir,
		rootDir:  rootDir,
		port:     port,
		build:    build,
	}
}

// Run starts the file watcher and HTTP server, blocking until the server stops.
func (s *DevServer) Run() error {
	go s.watchLoop()
	return s.serve()
}

func (s *DevServer) bump() {
	s.mu.Lock()
	s.revision++
	s.mu.Unlock()
}

func (s *DevServer) currentRevision() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.revision
}

func (s *DevServer) serve() error {
	mux := http.NewServeMux()
	mux.Handle("/__sale_reload", s.reloadHandler())
	mux.Handle("/", s.fileHandler())
	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("Dev server running at http://localhost:%d\n", s.port)
	fmt.Println("Press Ctrl+C to stop.")
	return http.ListenAndServe(addr, mux)
}

func (s *DevServer) reloadHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		fmt.Fprintf(w, "%d", s.currentRevision())
	})
}

// fileHandler serves build output files, injecting a live-reload script into
// HTML documents so the browser auto-refreshes after a rebuild.
func (s *DevServer) fileHandler() http.Handler {
	fs := http.FileServer(http.Dir(s.buildDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		if strings.HasSuffix(path, ".html") {
			full := filepath.Join(s.buildDir, filepath.FromSlash(path))
			if _, err := os.Stat(full); err == nil {
				s.serveHTML(w, r, full)
				return
			}
		}
		fs.ServeHTTP(w, r)
	})
}

func (s *DevServer) serveHTML(w http.ResponseWriter, r *http.Request, full string) {
	data, err := os.ReadFile(full)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	html := string(data)
	if strings.Contains(html, "</body>") {
		html = strings.Replace(html, "</body>", reloadScript+"\n</body>", 1)
	} else {
		html += "\n" + reloadScript + "\n"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	fmt.Fprint(w, html)
}

func (s *DevServer) watchLoop() {
	last := s.snapshot()
	for {
		time.Sleep(400 * time.Millisecond)
		cur := s.snapshot()
		if cur != last {
			last = cur
			s.rebuild()
		}
	}
}

func (s *DevServer) rebuild() {
	fmt.Println("[sale] Source changed, rebuilding...")
	if err := s.build(); err != nil {
		fmt.Fprintf(os.Stderr, "[sale] Build failed: %v\n", err)
		return
	}
	s.bump()
	fmt.Println("[sale] Build successful.")
}

// snapshot returns a signature of all .sl files under rootDir so the watcher
// can detect changes by comparing consecutive snapshots.
func (s *DevServer) snapshot() string {
	var sb strings.Builder
	_ = filepath.Walk(s.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			switch info.Name() {
			case "build", ".git", "node_modules", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".sl") {
			fmt.Fprintf(&sb, "%s|%d|%d\n", path, info.Size(), info.ModTime().UnixNano())
		}
		return nil
	})
	return sb.String()
}

// reloadScript is injected into served HTML and polls the dev server for a
// new build revision, reloading the page when one is detected.
const reloadScript = `<script>
(function() {
  var rev = 0;
  function check() {
    fetch('/__sale_reload', {cache: 'no-store'})
      .then(function(r) { return r.text(); })
      .then(function(t) {
        var n = parseInt(t, 10) || 0;
        if (rev === 0) { rev = n; return; }
        if (n !== rev) { window.location.reload(); }
      })
      .catch(function() {});
  }
  check();
  setInterval(check, 1000);
})();
</script>`