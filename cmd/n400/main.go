package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"n400/internal/flashcard"
	"n400/internal/store"
	"n400/internal/web"
)

func main() {
	port := flag.Int("port", 8400, "local HTTP port")
	noBrowser := flag.Bool("no-browser", false, "do not open the default browser")
	offline := flag.Bool("offline", false, "disable network lookups (this milestone has no outbound network code)")
	flag.Parse()
	if err := run(*port, *noBrowser, *offline); err != nil {
		log.Fatal(err)
	}
}

func run(port int, noBrowser, offline bool) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("finding user config directory: %w", err)
	}
	progress := store.NewFile(filepath.Join(configDir, "n400", "progress.json"))
	handler, err := web.NewWithStore(progress, flashcard.SystemClock{})
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fmt.Errorf("starting local server: %w", err)
	}
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	log.Printf("N400: %s (offline=%t; no outbound lookups implemented)", url, offline)
	if !noBrowser {
		if err := openBrowser(url); err != nil {
			log.Printf("Open %s in your browser: %v", url, err)
		}
	}
	select {
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			_ = server.Close()
			return fmt.Errorf("shutting down: %w", err)
		}
		return nil
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("browser launcher: %v", err)
		}
	}()
	return nil
}
