package editor

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var (
	port        int
	host        string
	openBrowser bool
	dev         bool
)

var Cmd = &cobra.Command{
	Use:   "editor",
	Short: "Start Flow visual editor",
	Long: `Start Web UI for visual editing of team-flow workflows.

Features:
  - Visual flow diagram editor with React Flow
  - Drag-and-drop node creation
  - Real-time property editing
  - Flow validation and preview`,
	RunE: runEditor,
}

func init() {
	Cmd.Flags().IntVarP(&port, "port", "p", 5174, "Port for API server")
	Cmd.Flags().StringVar(&host, "host", "localhost", "Listen address")
	Cmd.Flags().BoolVar(&openBrowser, "open", true, "Auto-open browser")
	Cmd.Flags().BoolVar(&dev, "dev", true, "Start in dev mode (frontend dev server + API server)")
}

func findEditorDir() string {
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		candidates := []string{
			filepath.Join(dir, "..", "..", "projects", "team-flow", "editor"),
			filepath.Join(dir, "editor"),
		}
		for _, c := range candidates {
			abs, _ := filepath.Abs(c)
			if _, err := os.Stat(filepath.Join(abs, "package.json")); err == nil {
				return abs
			}
		}
	}

	wd, _ := os.Getwd()
	cwdCandidates := []string{
		filepath.Join(wd, "editor"),
		filepath.Join(wd, "projects", "team-flow", "editor"),
		filepath.Join(wd, "..", "editor"),
	}
	for _, c := range cwdCandidates {
		abs, _ := filepath.Abs(c)
		if _, err := os.Stat(filepath.Join(abs, "package.json")); err == nil {
			return abs
		}
	}

	return ""
}

func findBun() string {
	paths := []string{
		filepath.Join(os.Getenv("USERPROFILE"), ".bun", "bin", "bun-windows-x64", "bun.exe"),
		filepath.Join(os.Getenv("HOME"), ".bun", "bin", "bun"),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	if p, err := exec.LookPath("bun"); err == nil {
		return p
	}

	return ""
}

func findNpx() string {
	if p, err := exec.LookPath("npx"); err == nil {
		return p
	}
	return ""
}

func runEditor(cmd *cobra.Command, args []string) error {
	rootDir, _ := os.Getwd()
	api := newAPIServer(rootDir)

	if dev {
		go func() {
			apiAddr := fmt.Sprintf("%s:%d", host, port)
			fmt.Printf("API server: http://%s/api/\n", apiAddr)
			if err := http.ListenAndServe(apiAddr, api); err != nil {
				fmt.Printf("API server error: %v\n", err)
			}
		}()

		editorDir := findEditorDir()
		if editorDir != "" {
			return runDevServer(editorDir, rootDir)
		}
		fmt.Println("Editor directory not found, API-only mode.")
		select {}
	}

	return runProductionServer(api)
}

func runDevServer(editorDir string, rootDir string) error {
	bunPath := findBun()

	var cmdStr string
	var cmdArgs []string

	if bunPath != "" {
		cmdStr = bunPath
		cmdArgs = []string{"run", "dev", "--", "--open"}
	} else {
		npxPath := findNpx()
		if npxPath == "" {
			fmt.Println("Neither bun nor npx found. Starting API-only server.")
			addr := fmt.Sprintf("%s:%d", host, port)
			fmt.Printf("API server: http://%s/api/\n", addr)
			return http.ListenAndServe(addr, newAPIServer(rootDir))
		}
		cmdStr = npxPath
		cmdArgs = []string{"rsbuild", "dev", "--open"}
	}

	addr := fmt.Sprintf("http://localhost:%d", port)
	fmt.Printf("Starting Flow Editor dev server...\n")
	fmt.Printf("Editor dir: %s\n", editorDir)
	fmt.Printf("Runner: %s %v\n", cmdStr, cmdArgs)
	fmt.Printf("URL: %s\n", addr)
	fmt.Printf("Press Ctrl+C to stop\n\n")

	if openBrowser {
		go func() {
			<-time.After(3 * time.Second)
			openURL(addr)
		}()
	}

	c := exec.Command(cmdStr, cmdArgs...)
	c.Dir = editorDir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		fmt.Printf("Dev server error: %v\n", err)
		fmt.Println("Falling back to production server.")
		return runProductionServer(newAPIServer(rootDir))
	}

	return nil
}

func openURL(url string) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command(os.Getenv("SystemRoot")+"\\System32\\cmd.exe", "/c", "start", "", url)
	} else if runtime.GOOS == "darwin" {
		cmd = exec.Command("open", url)
	} else {
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
		fmt.Printf("Open manually: %s\n", url)
	}
}

func runProductionServer(api *apiServer) error {
	addr := fmt.Sprintf("%s:%d", host, port)

	mux := http.NewServeMux()
	mux.Handle("/api/", api)

	editorDir := findEditorDir()
	distDir := ""
	if editorDir != "" {
		distDir = filepath.Join(editorDir, "dist")
		if _, err := os.Stat(distDir); err != nil {
			distDir = ""
		}
	}

	if distDir != "" {
		fs := http.FileServer(http.Dir(distDir))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fsPath := filepath.Join(distDir, r.URL.Path)
			if _, err := os.Stat(fsPath); err != nil {
				http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
			} else {
				fs.ServeHTTP(w, r)
			}
		})
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(fallbackHTML))
		})
	}

	fmt.Printf("Flow Editor starting...\n")
	fmt.Printf("Address: http://%s\n", addr)
	fmt.Printf("API: http://%s/api/\n", addr)
	fmt.Printf("Press Ctrl+C to stop\n\n")

	if openBrowser {
		go func() {
			<-time.After(1 * time.Second)
			openURL(fmt.Sprintf("http://%s", addr))
		}()
	}

	return http.ListenAndServe(addr, mux)
}

const fallbackHTML = `<!DOCTYPE html>
<html>
<head>
    <title>team-flow Editor</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: system-ui, sans-serif; background: #1a1a2e; color: #eee; height: 100vh; display: flex; align-items: center; justify-content: center; }
        .msg { text-align: center; }
        .msg h1 { color: #e94560; margin-bottom: 16px; }
        .msg p { color: #888; margin-bottom: 8px; }
        .msg code { background: #16213e; padding: 8px 16px; border-radius: 6px; color: #10B981; display: inline-block; margin-top: 16px; }
    </style>
</head>
<body>
    <div class="msg">
        <h1>Flow Editor</h1>
        <p>The React editor is not available.</p>
        <p>Run the following command to start the dev server:</p>
        <code>cd editor && bun run dev</code>
    </div>
</body>
</html>`
