package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func init() {
	// Set stateFile to executable directory + proxies.yaml
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("cannot get executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	stateFile = filepath.Join(exeDir, "proxies.yaml")
	log.Printf("state file: %s", stateFile)
}

func main() {
	// Force local-only binding for UI
	uiAddr := getenv("UI_ADDR", defaultUIAddr)
	adminToken := os.Getenv("ADMIN_TOKEN")
	if !strings.HasPrefix(uiAddr, "127.0.0.1:") && !strings.HasPrefix(uiAddr, "localhost:") {
		log.Fatalf("UI_ADDR must bind to 127.0.0.1, got %s", uiAddr)
	}

	// optional initial list
	initialList := os.Getenv("INITIAL_PROXIES") // "ip:port:user:pass,ip:port:..."

	m := NewManager(adminToken)

	// load state if exists
	if err := m.loadState(); err != nil {
		log.Printf("load state: %v", err)
	} else {
		log.Printf("loaded %d proxies from state", len(m.list()))
	}

	// Note: proxies are NOT auto-started on boot
	// User must manually start them from UI

	// optionally add initial proxies
	if initialList != "" {
		for _, line := range strings.Split(initialList, ",") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if _, err := m.addStartLine(line); err != nil {
				log.Printf("initial add failed: %v", err)
			}
		}
	}

	// UI server
	s := &http.Server{
		Addr:              uiAddr,
		Handler:           m.ui(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.Printf("UI listening at http://%s (local only)", uiAddr)
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ui server: %v", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Printf("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = s.Shutdown(ctx)

	// stop all proxies
	for _, it := range m.list() {
		_ = m.stop(it.ID)
	}
	log.Printf("bye")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
