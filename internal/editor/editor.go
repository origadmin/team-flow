package editor

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/spf13/cobra"
)

var (
	port        int
	host        string
	openBrowser bool
)

var Cmd = &cobra.Command{
	Use:   "editor",
	Short: "Start Flow visual editor",
	Long: `Start Web UI for visual editing of team-flow workflows.

Features:
  - View role flow diagrams
  - Edit workflow configuration
  - Real-time preview`,
	RunE: runEditor,
}

func init() {
	Cmd.Flags().IntVarP(&port, "port", "p", 3000, "Port")
	Cmd.Flags().StringVar(&host, "host", "localhost", "Listen address")
	Cmd.Flags().BoolVar(&openBrowser, "open", true, "Auto-open browser")
}

func runEditor(cmd *cobra.Command, args []string) error {
	addr := fmt.Sprintf("%s:%d", host, port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("port %d in use: %w", port, err)
	}
	ln.Close()

	fmt.Printf("Flow Editor starting...\n")
	fmt.Printf("Address: http://%s\n", addr)
	fmt.Printf("Press Ctrl+C to stop\n\n")

	html := `<!DOCTYPE html>
<html>
<head>
    <title>team-flow Editor</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: system-ui, sans-serif; background: #1a1a2e; color: #eee; height: 100vh; }
        .header { background: #16213e; padding: 16px 24px; border-bottom: 1px solid #0f3460; }
        .header h1 { font-size: 18px; color: #e94560; }
        .main { display: flex; height: calc(100vh - 60px); }
        .sidebar { width: 200px; background: #16213e; padding: 16px; border-right: 1px solid #0f3460; }
        .sidebar h3 { font-size: 12px; color: #888; text-transform: uppercase; margin-bottom: 12px; }
        .role-item { padding: 8px 12px; margin: 4px 0; background: #1a1a2e; border-radius: 6px; cursor: pointer; }
        .role-item:hover { background: #0f3460; }
        .canvas { flex: 1; background: #0f0f23; display: flex; align-items: center; justify-content: center; }
        .canvas-placeholder { color: #666; text-align: center; }
        .canvas-placeholder .icon { font-size: 64px; margin-bottom: 16px; }
        .properties { width: 280px; background: #16213e; padding: 16px; border-left: 1px solid #0f3460; }
        .properties h3 { font-size: 12px; color: #888; text-transform: uppercase; margin-bottom: 12px; }
        .prop-group { margin-bottom: 16px; }
        .prop-label { font-size: 12px; color: #aaa; margin-bottom: 4px; }
        .prop-value { background: #1a1a2e; padding: 8px; border-radius: 4px; font-size: 14px; }
        .btn { background: #e94560; color: white; border: none; padding: 8px 16px; border-radius: 6px; cursor: pointer; }
        .btn:hover { background: #d63850; }
    </style>
</head>
<body>
    <div class="header">
        <h1>team-flow Editor</h1>
    </div>
    <div class="main">
        <div class="sidebar">
            <h3>Roles</h3>
            <div class="role-item">Triage</div>
            <div class="role-item">Tech Lead</div>
            <div class="role-item">Dev</div>
            <div class="role-item">QA</div>
            <div class="role-item">PM</div>
            <h3 style="margin-top: 24px;">Workflows</h3>
            <div class="role-item">Three Gates</div>
            <div class="role-item">Task Lifecycle</div>
            <div class="role-item">Role Handoff</div>
        </div>
        <div class="canvas">
            <div class="canvas-placeholder">
                <div class="icon">Flow</div>
                <p>Drag nodes to canvas</p>
                <p style="font-size: 12px; color: #666;">Visual editor in development...</p>
            </div>
        </div>
        <div class="properties">
            <h3>Properties</h3>
            <div class="prop-group">
                <div class="prop-label">Node Name</div>
                <div class="prop-value">-</div>
            </div>
            <div class="prop-group">
                <div class="prop-label">Type</div>
                <div class="prop-value">-</div>
            </div>
            <div class="prop-group">
                <button class="btn">Save</button>
            </div>
        </div>
    </div>
</body>
</html>`

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	})

	log.Fatal(http.ListenAndServe(addr, nil))
	return nil
}
