package editor

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var (
	port     int
	host     string
	openBrowser bool
)

var editorCmd = &cobra.Command{
	Use:   "editor",
	Short: "启动 Flow 可视化编辑器",
	Long: `启动 Web UI，用于可视化编辑 _team 工作流。

功能:
  - 查看角色流转图
  - 编辑流程配置
  - 实时预览变化`,
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := fmt.Sprintf("%s:%d", host, port)

		// 检查端口是否可用
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("端口 %d 已被占用: %w", port, err)
		}
		ln.Close()

		fmt.Printf("🎨 Flow Editor 启动中...\n")
		fmt.Printf("📍 地址: http://%s\n", addr)
		fmt.Printf("按 Ctrl+C 停止\n\n")

		// 创建简单的 HTML 编辑器
		html := `<!DOCTYPE html>
<html>
<head>
    <title>_team Flow Editor</title>
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
        <h1>🎨 _team Flow Editor</h1>
    </div>
    <div class="main">
        <div class="sidebar">
            <h3>角色</h3>
            <div class="role-item">📋 Triage</div>
            <div class="role-item">👨‍💻 Tech Lead</div>
            <div class="role-item">🔧 Dev</div>
            <div class="role-item">🧪 QA</div>
            <div class="role-item">📊 PM</div>
            
            <h3 style="margin-top: 24px;">工作流</h3>
            <div class="role-item">三层门禁</div>
            <div class="role-item">任务生命周期</div>
            <div class="role-item">角色交接</div>
        </div>
        <div class="canvas">
            <div class="canvas-placeholder">
                <div class="icon">📊</div>
                <p>拖拽节点到画布</p>
                <p style="font-size: 12px; color: #666;">Flow 可视化编辑开发中...</p>
            </div>
        </div>
        <div class="properties">
            <h3>属性</h3>
            <div class="prop-group">
                <div class="prop-label">节点名称</div>
                <div class="prop-value">-</div>
            </div>
            <div class="prop-group">
                <div class="prop-label">类型</div>
                <div class="prop-value">-</div>
            </div>
            <div class="prop-group">
                <button class="btn">保存</button>
            </div>
        </div>
    </div>
</body>
</html>`

		// 简单的 Web 服务器
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(html))
		})

		if openBrowser {
			go func() {
				// 延迟打开浏览器
				os.WriteFile("/tmp/.open_browser", []byte(""), 0644)
			}()
		}

		log.Fatal(http.ListenAndServe(addr, nil))
		return nil
	},
}

func init() {
	editorCmd.Flags().IntVarP(&port, "port", "p", 3000, "端口")
	editorCmd.Flags().StringVar(&host, "host", "localhost", "监听地址")
	editorCmd.Flags().BoolVar(&openBrowser, "open", true, "自动打开浏览器")
}
