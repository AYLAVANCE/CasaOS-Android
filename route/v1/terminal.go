//go:build android

package v1

import (
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var termUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// HandleWebTerminal provides a WebSocket-based terminal using a direct PTY
func HandleWebTerminal(ctx echo.Context) error {
	conn, err := termUpgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Determine the shell path
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/data/data/com.termux/files/usr/bin/bash"
	}

	cmd := exec.Command(shell, "-l")
	cmd.Env = os.Environ()

	ptmx, err := pty.Start(cmd)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Failed to start terminal: "+err.Error()))
		return err
	}
	defer ptmx.Close()

	// PTY -> WebSocket
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				if err != io.EOF {
					conn.WriteMessage(websocket.TextMessage, []byte("Terminal closed"))
				}
				conn.Close()
				return
			}
			conn.WriteMessage(websocket.BinaryMessage, buf[:n])
		}
	}()

	// WebSocket -> PTY
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		ptmx.Write(msg)
	}

	cmd.Process.Kill()
	cmd.Wait()
	return nil
}
