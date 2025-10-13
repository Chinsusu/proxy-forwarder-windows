package main

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// handleAuth checks admin token authorization
func (m *Manager) handleAuth(r *http.Request) bool {
	if m.adminToken == "" {
		return true
	}
	token := r.Header.Get("X-Admin-Token")
	if token == "" {
		token = r.URL.Query().Get("token")
	}
	if subtle.ConstantTimeCompare([]byte(token), []byte(m.adminToken)) == 1 {
		return true
	}
	return false
}

// ui returns the HTTP handler for web UI and API
func (m *Manager) ui() http.Handler {
	mux := http.NewServeMux()

	// Web UI
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !m.handleAuth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, indexHTML)
	})

	// API: List all proxies
	mux.HandleFunc("/api/list", func(w http.ResponseWriter, r *http.Request) {
		if !m.handleAuth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(struct {
			Items []*Upstream `json:"items"`
		}{Items: m.list()})
	})

	// API: Add proxy and auto-start
	mux.HandleFunc("/api/add", func(w http.ResponseWriter, r *http.Request) {
		if !m.handleAuth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		b, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		line := strings.TrimSpace(string(b))
		if line == "" {
			http.Error(w, "body must be 'ip:port:user:pass' or 'ip:port'", 400)
			return
		}
		up, err := parseProxyLine(line)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		up, err = m.addOrReplace(up)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		_ = m.start(up.ID)
		json.NewEncoder(w).Encode(up)
	})

	// API: Add proxy to pool (no auto-start)
	mux.HandleFunc("/api/add-pool", func(w http.ResponseWriter, r *http.Request) {
		if !m.handleAuth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		b, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		line := strings.TrimSpace(string(b))
		if line == "" {
			http.Error(w, "body must be 'ip:port:user:pass' or 'ip:port'", 400)
			return
		}
		up, err := parseProxyLine(line)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		up, err = m.addToPool(up)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(up)
	})

	// API: Remove proxy
	mux.HandleFunc("/api/remove", func(w http.ResponseWriter, r *http.Request) {
		if !m.handleAuth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "missing id", 400)
			return
		}
		if err := m.remove(id); err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		w.WriteHeader(204)
	})

	// API: Stop proxy
	mux.HandleFunc("/api/stop", func(w http.ResponseWriter, r *http.Request) {
		if !m.handleAuth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "missing id", 400)
			return
		}
		if err := m.stop(id); err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		w.WriteHeader(204)
	})

	// API: Start proxy
	mux.HandleFunc("/api/start", func(w http.ResponseWriter, r *http.Request) {
		if !m.handleAuth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "missing id", 400)
			return
		}
		if err := m.start(id); err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		w.WriteHeader(204)
	})

	// API: Export local proxy addresses
	mux.HandleFunc("/api/export-local", func(w http.ResponseWriter, r *http.Request) {
		if !m.handleAuth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var lines []string
		for _, it := range m.list() {
			lines = append(lines, fmt.Sprintf("127.0.0.1:%d", it.LocalPort))
		}
		io.WriteString(w, strings.Join(lines, "\n"))
	})

	// API: CloudMini regions proxy
	mux.HandleFunc("/api/cloudmini/regions", m.handleCloudMiniRegions)

	// API: CloudMini order proxy
	mux.HandleFunc("/api/cloudmini/order", m.handleCloudMiniOrder)

	// API: CloudMini sync all proxy-res to pool
	mux.HandleFunc("/api/cloudmini/sync", m.handleCloudMiniSync)

	// API: Firewall status
	mux.HandleFunc("/api/firewall/status", func(w http.ResponseWriter, r *http.Request) {
		if !m.handleAuth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		rules, _ := listFirewallRules()
		json.NewEncoder(w).Encode(map[string]any{
			"enabled":    len(rules) > 0,
			"is_admin":   isAdmin(),
			"rule_count": len(rules),
			"rules":      rules,
		})
	})

	// API: Check exit IP of a proxy
	mux.HandleFunc("/api/check-ip", func(w http.ResponseWriter, r *http.Request) {
		if !m.handleAuth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "missing id", 400)
			return
		}

		m.mu.RLock()
		it, ok := m.items[id]
		m.mu.RUnlock()

		if !ok {
			http.Error(w, "proxy not found", 404)
			return
		}

		if !it.isRunning {
			http.Error(w, "proxy not running", 400)
			return
		}

		ip, err := checkProxyExitIP(it.cfg.LocalPort, m.adminToken)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"ip": ip,
		})
	})

	return mux
}
