package main

// Windows 10 local forward-proxy manager
// - Spins up local HTTP forward proxies on 127.0.0.1:<port>
// - Each local proxy forwards all traffic via its assigned upstream HTTP proxy (ip:port:user:pass)
// - Ports start at 10001 and increment
// - Web UI runs on 127.0.0.1:17890 for local-only management
// - When upstream health fails repeatedly, local listener is shut down (no leak).
// - Optional ADMIN_TOKEN (env) to protect management endpoints.
// - State saved to proxies.yaml in working directory.

import (
	"bufio"
	"context"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	goproxy "github.com/elazarl/goproxy"
	"gopkg.in/yaml.v3"
)

const (
	defaultUIAddr   = "127.0.0.1:17890"
	firstLocalPort  = 10001
	stateFile       = "proxies.yaml"
	healthURL       = "http://www.gstatic.com/generate_204" // lightweight 204
	healthInterval  = 10 * time.Second
	healthFailLimit = 3
)

type Upstream struct {
	ID        string `yaml:"id" json:"id"`
	Host      string `yaml:"host" json:"host"`
	Port      int    `yaml:"port" json:"port"`
	User      string `yaml:"user" json:"user"`
	Pass      string `yaml:"pass" json:"pass"`
	LocalPort int    `yaml:"local_port" json:"local_port"`

	Status    string `yaml:"status" json:"status"` // creating|live|dead|stopped
	LastError string `yaml:"last_error" json:"last_error"`
}

type State struct {
	Items []*Upstream `yaml:"items"`
	Next  int         `yaml:"next"`
}

type Manager struct {
	mu       sync.RWMutex
	items    map[string]*ProxyItem // id -> ProxyItem
	nextPort int

	adminToken string
}

type ProxyItem struct {
	cfg       *Upstream
	server    *http.Server
	listener  net.Listener
	stopFn    context.CancelFunc
	healthWg  sync.WaitGroup
	isRunning bool
}

func NewManager(adminToken string) *Manager {
	return &Manager{
		items:     make(map[string]*ProxyItem),
		nextPort:  firstLocalPort,
		adminToken: adminToken,
	}
}

func sanitizeID(host string, port int) string {
	s := strings.ReplaceAll(host, ".", "-")
	return fmt.Sprintf("%s-%d", s, port)
}

func parseProxyLine(line string) (*Upstream, error) {
	// Supports "ip:port:user:pass" or "ip:port" (no auth)
	parts := strings.Split(line, ":")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid line: %q", line)
	}
	host := strings.TrimSpace(parts[0])
	p, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid port in %q: %v", line, err)
	}
	up := &Upstream{
		ID:   sanitizeID(host, p),
		Host: host,
		Port: p,
	}
	if len(parts) >= 4 {
		up.User = strings.TrimSpace(parts[2])
		up.Pass = strings.TrimSpace(parts[3])
	}
	return up, nil
}

func (m *Manager) loadState() error {
	b, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var st State
	if err := yaml.Unmarshal(b, &st); err != nil {
		return err
	}
	if st.Next < firstLocalPort {
		st.Next = firstLocalPort
	}
	m.nextPort = st.Next
	for _, it := range st.Items {
		// reconstruct item but not running yet
		m.items[it.ID] = &ProxyItem{cfg: it}
	}
	return nil
}

func (m *Manager) saveState() error {
	st := State{Next: m.nextPort}
	for _, it := range m.items {
		st.Items = append(st.Items, it.cfg)
	}
	b, err := yaml.Marshal(&st)
	if err != nil {
		return err
	}
	tmp := stateFile + ".tmp"
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, stateFile)
}

func (m *Manager) allocPort() int {
	if m.nextPort < firstLocalPort {
		m.nextPort = firstLocalPort
	}
	p := m.nextPort
	m.nextPort++
	return p
}

func (m *Manager) addOrReplace(up *Upstream) (*Upstream, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if up.ID == "" {
		up.ID = sanitizeID(up.Host, up.Port)
	}
	if existing, ok := m.items[up.ID]; ok {
		// replace upstream credentials/host/port but keep local port
		up.LocalPort = existing.cfg.LocalPort
		m.items[up.ID].cfg = up
		return up, m.saveState()
	}
	up.LocalPort = m.allocPort()
	up.Status = "creating"
	m.items[up.ID] = &ProxyItem{cfg: up}
	return up, m.saveState()
}

func (m *Manager) remove(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.items[id]
	if !ok {
		return os.ErrNotExist
	}
	if it.isRunning {
		_ = m.stopLocked(it)
	}
	delete(m.items, id)
	return m.saveState()
}

func (m *Manager) start(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.items[id]
	if !ok {
		return os.ErrNotExist
	}
	if it.isRunning {
		return nil
	}
	return m.startLocked(it)
}

func (m *Manager) stop(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.items[id]
	if !ok {
		return os.ErrNotExist
	}
	return m.stopLocked(it)
}

func (m *Manager) startLocked(it *ProxyItem) error {
	up := it.cfg
	upstreamURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", up.Host, up.Port),
	}
	if up.User != "" {
		upstreamURL.User = url.UserPassword(up.User, up.Pass)
	}

	tr := &http.Transport{
		Proxy: http.ProxyURL(upstreamURL),
		// Reasonable timeouts
		ProxyConnectHeader: http.Header{},
		MaxConnsPerHost:    0,
		MaxIdleConns:       128,
		IdleConnTimeout:    30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	px := goproxy.NewProxyHttpServer()
	px.Verbose = false
	px.Tr = tr
	// Suppress goproxy's verbose logging
	px.Logger = log.New(io.Discard, "", 0)
	
	// Force all CONNECT (HTTPS) requests through upstream proxy
	px.ConnectDial = func(network, addr string) (net.Conn, error) {
		// Connect to upstream proxy first
		proxyConn, err := net.DialTimeout("tcp", upstreamURL.Host, 10*time.Second)
		if err != nil {
			return nil, fmt.Errorf("dial upstream proxy: %w", err)
		}
		
		// Send CONNECT request to upstream proxy
		connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", addr, addr)
		if upstreamURL.User != nil {
			username := upstreamURL.User.Username()
			password, _ := upstreamURL.User.Password()
			auth := username + ":" + password
			encoded := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
			connectReq += "Proxy-Authorization: " + encoded + "\r\n"
		}
		connectReq += "\r\n"
		
		if _, err := proxyConn.Write([]byte(connectReq)); err != nil {
			proxyConn.Close()
			return nil, fmt.Errorf("write CONNECT: %w", err)
		}
		
		// Read response from upstream proxy
		br := bufio.NewReader(proxyConn)
		resp, err := http.ReadResponse(br, &http.Request{Method: "CONNECT"})
		if err != nil {
			proxyConn.Close()
			return nil, fmt.Errorf("read CONNECT response: %w", err)
		}
		resp.Body.Close()
		
		if resp.StatusCode != 200 {
			proxyConn.Close()
			return nil, fmt.Errorf("upstream proxy returned %d", resp.StatusCode)
		}
		
		return proxyConn, nil
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", up.LocalPort),
		Handler: px,
		// harden timeouts
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		up.Status = "dead"
		up.LastError = "listen failed: " + err.Error()
		_ = m.saveState()
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	it.server = srv
	it.listener = ln
	it.stopFn = cancel
	it.isRunning = true
	up.Status = "live"
	up.LastError = ""
	_ = m.saveState()

	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("[proxy %s] serve error: %v", up.ID, err)
		}
	}()

	// health watcher
	it.healthWg.Add(1)
	go func() {
		defer it.healthWg.Done()
		fail := 0
		t := time.NewTicker(healthInterval)
		defer t.Stop()
		client := &http.Client{
			Transport: tr,
			Timeout:   8 * time.Second,
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				req, _ := http.NewRequest("GET", healthURL, nil)
				// avoid cache
				req.Header.Set("Cache-Control", "no-cache")
				resp, err := client.Do(req)
				ok := err == nil && resp != nil && resp.StatusCode < 500
				if resp != nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}
				if ok {
					fail = 0
					continue
				}
				fail++
				if fail >= healthFailLimit {
					log.Printf("[proxy %s] upstream unhealthy (%d fails), shutting down local listener", up.ID, fail)
					m.mu.Lock()
					_ = m.stopLocked(it)
					up.Status = "dead"
					up.LastError = "upstream unhealthy (auto stop)"
					_ = m.saveState()
					m.mu.Unlock()
					return
				}
			}
		}
	}()

	log.Printf("[proxy %s] started at http://127.0.0.1:%d -> upstream %s", up.ID, up.LocalPort, upstreamURL.Redacted())
	return nil
}

func (m *Manager) stopLocked(it *ProxyItem) error {
	if !it.isRunning {
		return nil
	}
	it.stopFn()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = it.server.Shutdown(ctx)
	_ = it.listener.Close()
	it.isRunning = false
	it.cfg.Status = "stopped"
	return nil
}

func (m *Manager) list() []*Upstream {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var res []*Upstream
	for _, it := range m.items {
		res = append(res, it.cfg)
	}
	// stable-ish order by LocalPort
	for i := 0; i < len(res); i++ {
		for j := i+1; j < len(res); j++ {
			if res[j].LocalPort < res[i].LocalPort {
				res[i], res[j] = res[j], res[i]
			}
		}
	}
	return res
}

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

// --- Web UI + API ---

func (m *Manager) ui() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !m.handleAuth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, indexHTML)
	})

	mux.HandleFunc("/api/list", func(w http.ResponseWriter, r *http.Request) {
		if !m.handleAuth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(struct{
			Items []*Upstream `json:"items"`
		}{Items: m.list()})
	})

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

	mux.HandleFunc("/api/sync", func(w http.ResponseWriter, r *http.Request) {
		if !m.handleAuth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		apiURL := r.URL.Query().Get("url")
		if apiURL == "" {
			http.Error(w, "url param required", 400)
			return
		}
		added, errs := m.syncFromAPI(apiURL)
		json.NewEncoder(w).Encode(map[string]any{
			"added": added,
			"errors": errs,
		})
	})

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

	return mux
}

func (m *Manager) syncFromAPI(urlStr string) (int, []string) {
	client := &http.Client{ Timeout: 10 * time.Second }
	resp, err := client.Get(urlStr)
	if err != nil {
		return 0, []string{err.Error()}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	text := strings.TrimSpace(string(body))
	added := 0
	var errs []string

	// Try JSON array of strings first
	var arr []string
	if strings.HasPrefix(text, "[") {
		if err := json.Unmarshal([]byte(text), &arr); err == nil {
			for _, line := range arr {
				if strings.TrimSpace(line) == "" { continue }
				if _, e := m.addStartLine(line); e != nil {
					errs = append(errs, e.Error())
				} else {
					added++
				}
			}
			return added, errs
		}
	}

	// Fallback: line-delimited
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" { continue }
		if _, e := m.addStartLine(line); e != nil {
			errs = append(errs, e.Error())
		} else {
			added++
		}
	}
	return added, errs
}

func (m *Manager) addStartLine(line string) (*Upstream, error) {
	up, err := parseProxyLine(line)
	if err != nil {
		return nil, err
	}
	up, err = m.addOrReplace(up)
	if err != nil {
		return nil, err
	}
	if err := m.start(up.ID); err != nil {
		return nil, err
	}
	return up, nil
}

func mustAtoi(s string, def int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return i
}

func main() {
	// Force local-only binding for UI
	uiAddr := getenv("UI_ADDR", defaultUIAddr)
	adminToken := os.Getenv("ADMIN_TOKEN")
	if !strings.HasPrefix(uiAddr, "127.0.0.1:") && !strings.HasPrefix(uiAddr, "localhost:") {
		log.Fatalf("UI_ADDR must bind to 127.0.0.1, got %s", uiAddr)
	}

	// optional initial API to sync, or initial list
	initialAPI := os.Getenv("INITIAL_API")
	initialList := os.Getenv("INITIAL_PROXIES") // "ip:port:user:pass,ip:port:..."
	statePath := getenv("STATE_FILE", stateFile)
	if statePath != stateFile {
		// move working dir state filename
	}

	m := NewManager(adminToken)

	// load state if exists
	if err := m.loadState(); err != nil {
		log.Printf("load state: %v", err)
	}

	// start any previously known proxies (best-effort)
	for _, it := range m.list() {
		_ = m.start(it.ID)
	}

	// optionally sync
	if initialAPI != "" {
		added, errs := m.syncFromAPI(initialAPI)
		log.Printf("initial sync added=%d errs=%v", added, errs)
	}
	if initialList != "" {
		for _, line := range strings.Split(initialList, ",") {
			line = strings.TrimSpace(line)
			if line == "" { continue }
			if _, err := m.addStartLine(line); err != nil {
				log.Printf("initial add failed: %v", err)
			}
		}
	}

	// UI server
	s := &http.Server{
		Addr:    uiAddr,
		Handler: m.ui(),
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

var indexHTML = `<!doctype html>
<html lang="vi">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Proxy Forward Grid</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
  <script src="https://cdn.tailwindcss.com"></script>
  <style>
    body{background:#f8f9fa;font-family:'Inter',system-ui,-apple-system,sans-serif}
    input,textarea,select,button{font-family:'Inter',system-ui,-apple-system,sans-serif}
    .sidebar{background:linear-gradient(180deg,#2d1b4e 0%,#1a0f2e 100%)}
    .sidebar-item{transition:all .2s;border-radius:.5rem;margin:.25rem 0}
    .sidebar-item:hover{background:rgba(255,255,255,.1)}
    .sidebar-item.active{background:rgba(147,51,234,.3);border-left:3px solid #a855f7}
    .table-header{background:#2d1b4e;color:#fff}
    .status-active{background:#d1fae5;color:#065f46;padding:.25rem .75rem;border-radius:9999px;font-size:.75rem;font-weight:600}
    .status-inactive{background:#fee2e2;color:#991b1b;padding:.25rem .75rem;border-radius:9999px;font-size:.75rem;font-weight:600}
    .action-btn{padding:.4rem;border-radius:.375rem;transition:all .2s}
    .action-btn:hover{background:#f3f4f6}
    .avatar{width:2rem;height:2rem;border-radius:9999px}
  </style>
</head>
<body class="flex h-screen">
  <aside class="sidebar w-64 p-4 flex flex-col">
    <div class="flex items-center gap-2 mb-8">
      <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-pink-500 to-purple-600 grid place-items-center">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-white" viewBox="0 0 24 24" fill="currentColor"><path d="M12 1a5 5 0 0 0-5 5v2H6a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V10a2 2 0 0 0-2-2h-1V6a5 5 0 0 0-5-5m-3 7V6a3 3 0 1 1 6 0v2z"/></svg>
      </div>
      <div class="text-white font-bold text-lg">Proxy Forward</div>
    </div>
    <nav class="flex-1 space-y-1">
      <a href="#" class="sidebar-item active flex items-center gap-3 px-3 py-2 text-white">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor"><path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2M9 17H7v-7h2v7m4 0h-2V7h2v10m4 0h-2v-4h2v4z"/></svg>
        <span>Proxies</span>
      </a>
    </nav>
    <div class="pt-4 border-t border-slate-700">
      <div class="p-2 bg-slate-800/50 rounded-lg mb-2">
        <label class="text-xs text-slate-400 block mb-1">Admin Token</label>
        <input id="tokenInput" type="password" placeholder="Enter token..." class="w-full px-2 py-1 text-xs bg-slate-900/70 border border-slate-700 rounded text-white">
      </div>
      <div class="flex items-center gap-3 px-3 py-2">
        <img src="https://ui-avatars.com/api/?name=Admin&background=8b5cf6&color=fff" class="avatar">
        <div class="flex-1">
          <div class="text-white text-sm font-medium">Admin</div>
          <div class="text-slate-400 text-xs">127.0.0.1 Only</div>
        </div>
      </div>
    </div>
  </aside>
  <main class="flex-1 overflow-auto">
    <header class="bg-white border-b px-6 py-4 flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-800">Proxies Management</h1>
        <p class="text-sm text-gray-500">Proxy Forward Dashboard</p>
      </div>
      <div class="flex items-center gap-2">
<button onclick="handleExport" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">
          <span>📦 Export Local</span>
        </button>
<button onclick="openBulkModal()" class="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700">
          <span>📥 Bulk Add</span>
        </button>
      </div>
    </header>
    
    <div class="m-6 space-y-4">
      <div class="bg-white p-4 rounded-xl shadow-sm">
        <h3 class="font-bold mb-3">Add Single Proxy</h3>
        <div class="flex gap-2">
          <input id="singleProxyInput" type="text" placeholder="ip:port:user:pass hoặc ip:port" class="flex-1 px-3 py-2 border rounded-lg">
<button onclick="handleAddSingle()" class="px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700">
            ➕ Add Proxy
          </button>
        </div>
      </div>

      <div class="bg-white p-4 rounded-xl shadow-sm">
        <h3 class="font-bold mb-3">Sync from API</h3>
        <div class="flex gap-2">
          <input id="apiUrlInput" type="text" placeholder="API URL trả về danh sách proxy (text hoặc JSON array)" class="flex-1 px-3 py-2 border rounded-lg">
<button onclick="handleSyncAPI()" class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
            🔄 Sync API
          </button>
        </div>
      </div>

      <div class="bg-white p-4 rounded-xl shadow-sm">
        <div class="flex items-center justify-between mb-3">
          <h3 class="font-bold">Proxy List</h3>
          <div class="flex items-center gap-2">
            <label class="flex items-center gap-2">
              <input id="autoRefreshCheck" type="checkbox" class="accent-purple-600" onchange="toggleAutoRefresh()">
              <span class="text-sm">Auto 5s</span>
            </label>
            <input id="searchInput" type="text" placeholder="Search..." class="px-3 py-2 border rounded-lg">
<button onclick="handleSearch()" class="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700">
              🔍 Search
            </button>
          </div>
        </div>
        <div class="overflow-hidden rounded-lg border">
          <table class="min-w-full">
            <thead class="table-header">
              <tr>
                <th class="py-3 px-4 text-left">#</th>
                <th class="py-3 px-4 text-left">Proxy Address</th>
                <th class="py-3 px-4 text-left">Local Port</th>
                <th class="py-3 px-4 text-left">Status</th>
                <th class="py-3 px-4 text-left">Action</th>
              </tr>
            </thead>
            <tbody id="rows" class="divide-y divide-gray-200"></tbody>
          </table>
        </div>
        <div class="mt-3 flex items-center justify-between">
          <div id="countText" class="text-sm text-gray-600">Loading...</div>
          <div class="text-xs text-gray-500">Local ports start at 127.0.0.1:10001+</div>
        </div>
      </div>
    </div>
  </main>

  <div id="bulkModal" class="hidden fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
    <div class="bg-white rounded-xl shadow-xl max-w-2xl w-full">
      <div class="p-4 border-b flex items-center justify-between">
        <h3 class="font-bold text-lg">Bulk Add Proxies</h3>
<button onclick="closeBulkModal()" class="text-gray-500 hover:text-gray-700">✖</button>
      </div>
      <div class="p-4">
        <p class="text-sm text-gray-600 mb-2">Mỗi dòng 1 proxy: <code class="bg-gray-100 px-2 py-1 rounded">ip:port:user:pass</code> hoặc <code class="bg-gray-100 px-2 py-1 rounded">ip:port</code></p>
        <textarea id="bulkTextarea" class="w-full h-64 px-3 py-2 border rounded-lg font-mono text-sm" placeholder="1.2.3.4:8080:user:pass
5.6.7.8:3128
9.10.11.12:1080:admin:secret"></textarea>
      </div>
      <div class="p-4 border-t flex justify-end gap-2">
        <button onclick="closeBulkModal()" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">Cancel</button>
        <button onclick="handleBulkAdd()" class="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700">Add All</button>
      </div>
    </div>
  </div>

  <div id="toast" class="hidden fixed bottom-4 right-4 bg-gray-900 text-white px-4 py-3 rounded-lg shadow-lg"></div>

  <script>
    var rowsEl = document.getElementById('rows');
    var countEl = document.getElementById('countText');
    var toastEl = document.getElementById('toast');
    var tokenInput = document.getElementById('tokenInput');
    var DATA = { items: [] };
    var autoRefreshTimer = null;

    tokenInput.value = localStorage.getItem('admintoken') || '';
    tokenInput.addEventListener('change', function(){
      localStorage.setItem('admintoken', tokenInput.value.trim());
      showToast('Token saved');
    });

    function hdr(){ 
      var t = localStorage.getItem('admintoken') || ''; 
      var h = {}; 
      if(t){ h['X-Admin-Token'] = t; } 
      return h; 
    }
    
    function GET(u){ 
      return fetch(u, {headers: hdr()}).then(function(r){ 
        if(!r.ok) return r.text().then(function(t){ throw new Error(t); }); 
        return r.json(); 
      }); 
    }

    function POST(u, body){ 
      var opt = {method:'POST', headers: hdr()}; 
      if(body) opt.body = body;
      return fetch(u, opt).then(function(r){ 
        if(!r.ok) return r.text().then(function(t){ throw new Error(t); }); 
        return r.text().then(function(t){ try{ return JSON.parse(t); }catch(e){ return {}; } }); 
      }); 
    }

    function showToast(msg){
      toastEl.textContent = msg;
      toastEl.classList.remove('hidden');
      setTimeout(function(){ toastEl.classList.add('hidden'); }, 2000);
    }

    function getInitial(name){ 
      return name ? name.charAt(0).toUpperCase() : 'P'; 
    }
    
    function getRandomColor(){ 
      var colors = ['ef4444','f59e0b','10b981','3b82f6','8b5cf6','ec4899']; 
      return colors[Math.floor(Math.random() * colors.length)]; 
    }

    function rowItem(it, idx){
      var tr = document.createElement('tr');
      tr.className = 'hover:bg-gray-50';
      
      var tdIdx = document.createElement('td'); 
      tdIdx.className = 'py-3 px-4 text-gray-600'; 
      tdIdx.textContent = String(idx+1); 
      tr.appendChild(tdIdx);

      var up = (it.user ? it.user + ':' + it.pass + '@' : '') + it.host + ':' + it.port;
      var color = getRandomColor();
      var initial = getInitial(it.host);
      
      var tdUp = document.createElement('td'); 
      tdUp.className = 'py-3 px-4';
      var divFlex = document.createElement('div'); 
      divFlex.className = 'flex items-center gap-3';
      var avatar = document.createElement('div');
      avatar.className = 'w-8 h-8 rounded-full grid place-items-center text-white font-bold text-sm';
      avatar.style.background = '#' + color;
      avatar.textContent = initial;
      var divText = document.createElement('div');
      var divMain = document.createElement('div');
      divMain.className = 'font-medium text-gray-800'; 
      divMain.textContent = up;
      var divId = document.createElement('div');
      divId.className = 'text-xs text-gray-500';
      divId.textContent = 'ID: ' + it.id;
      divText.appendChild(divMain);
      divText.appendChild(divId);
      divFlex.appendChild(avatar); 
      divFlex.appendChild(divText);
      tdUp.appendChild(divFlex); 
      tr.appendChild(tdUp);

      var local = '127.0.0.1:' + it.local_port;
      var tdLocal = document.createElement('td'); 
      tdLocal.className = 'py-3 px-4';
      var localDiv = document.createElement('div');
      localDiv.className = 'font-mono text-sm text-gray-800';
      localDiv.textContent = local;
      var copyBtn = document.createElement('button');
      copyBtn.className = 'text-xs text-blue-600 hover:underline mt-1';
      copyBtn.textContent = 'Copy';
      copyBtn.onclick = function(){ 
        navigator.clipboard.writeText(local);
        showToast('Copied: ' + local);
      };
      tdLocal.appendChild(localDiv);
      tdLocal.appendChild(copyBtn);
      tr.appendChild(tdLocal);

      var tdStatus = document.createElement('td'); 
      tdStatus.className = 'py-3 px-4';
      var badge = document.createElement('span');
      badge.className = it.status === 'live' ? 'status-active' : 'status-inactive';
      badge.textContent = it.status === 'live' ? 'Active' : 'Inactive';
      tdStatus.appendChild(badge); 
      tr.appendChild(tdStatus);

      var tdAction = document.createElement('td'); 
      tdAction.className = 'py-3 px-4';
      var btnStart = document.createElement('button'); 
      btnStart.className = 'action-btn mr-2 text-green-600'; 
btnStart.textContent = '▶ Start';
      btnStart.onclick = function(){ startProxy(it.id); };
      var btnStop = document.createElement('button'); 
      btnStop.className = 'action-btn mr-2 text-orange-600'; 
btnStop.textContent = '⏸ Stop';
      btnStop.onclick = function(){ stopProxy(it.id); };
      var btnDel = document.createElement('button'); 
      btnDel.className = 'action-btn text-red-600'; 
btnDel.textContent = '🗑 Del';
      btnDel.onclick = function(){ deleteProxy(it.id); };
      tdAction.appendChild(btnStart); 
      tdAction.appendChild(btnStop); 
      tdAction.appendChild(btnDel);
      tr.appendChild(tdAction);

      return tr;
    }

    function render(list){
      rowsEl.innerHTML = '';
      countEl.textContent = 'Total: ' + list.length + ' proxies';
      if(list.length === 0){
        var tr = document.createElement('tr');
        var td = document.createElement('td');
        td.colSpan = 5;
        td.className = 'py-8 text-center text-gray-500';
        td.textContent = 'No proxies yet. Add one above!';
        tr.appendChild(td);
        rowsEl.appendChild(tr);
        return;
      }
      for(var i=0; i<list.length; i++){ 
        rowsEl.appendChild(rowItem(list[i], i)); 
      }
    }

    function reload(){
      GET('/api/list').then(function(data){ 
        DATA = data; 
        render(DATA.items); 
      }).catch(function(e){ 
        console.error(e); 
        showToast('Error loading');
      });
    }

    function handleSearch(){
      var query = document.getElementById('searchInput').value.toLowerCase();
      if(!query){ render(DATA.items); return; }
      var filtered = DATA.items.filter(function(it){
        var up = (it.user ? it.user + ':' + it.pass + '@' : '') + it.host + ':' + it.port;
        var local = '127.0.0.1:' + it.local_port;
        return up.toLowerCase().indexOf(query) !== -1 || local.indexOf(query) !== -1;
      });
      render(filtered);
    }

    function handleAddSingle(){
      var line = document.getElementById('singleProxyInput').value.trim();
      if(!line){ showToast('Enter proxy address'); return; }
      POST('/api/add', line).then(function(){
        document.getElementById('singleProxyInput').value = '';
        showToast('Added successfully');
        reload();
      }).catch(function(e){
        showToast('Error: ' + e.message);
      });
    }

    function handleSyncAPI(){
      var url = document.getElementById('apiUrlInput').value.trim();
      if(!url){ showToast('Enter API URL'); return; }
      GET('/api/sync?url=' + encodeURIComponent(url)).then(function(result){
        showToast('Synced: ' + (result.added || 0) + ' added');
        reload();
      }).catch(function(e){
        showToast('Error: ' + e.message);
      });
    }

    function handleExport(){
      window.open('/api/export-local', '_blank');
    }

    function openBulkModal(){
      document.getElementById('bulkModal').classList.remove('hidden');
    }

    function closeBulkModal(){
      document.getElementById('bulkModal').classList.add('hidden');
    }

    function handleBulkAdd(){
      var text = document.getElementById('bulkTextarea').value;
      var lines = text.split('\n').map(function(s){ return s.trim(); }).filter(Boolean);
      if(!lines.length){ showToast('No lines'); return; }
      var ok = 0, fail = 0;
      function addLine(i){
        if(i >= lines.length){
          closeBulkModal();
          document.getElementById('bulkTextarea').value = '';
          showToast('Done: ' + ok + ' ok, ' + fail + ' fail');
          reload();
          return;
        }
        POST('/api/add', lines[i]).then(function(){
          ok++;
          addLine(i+1);
        }).catch(function(){
          fail++;
          addLine(i+1);
        });
      }
      addLine(0);
    }

    function startProxy(id){
      POST('/api/start?id=' + encodeURIComponent(id)).then(function(){ 
        showToast('Started');
        reload(); 
      }).catch(function(e){
        showToast('Error: ' + e.message);
      });
    }

    function stopProxy(id){
      POST('/api/stop?id=' + encodeURIComponent(id)).then(function(){ 
        showToast('Stopped');
        reload(); 
      }).catch(function(e){
        showToast('Error: ' + e.message);
      });
    }

    function deleteProxy(id){
      if(!confirm('Delete this proxy?')) return;
      POST('/api/remove?id=' + encodeURIComponent(id)).then(function(){ 
        showToast('Deleted');
        reload(); 
      }).catch(function(e){
        showToast('Error: ' + e.message);
      });
    }

    function toggleAutoRefresh(){
      var checked = document.getElementById('autoRefreshCheck').checked;
      if(autoRefreshTimer){ clearInterval(autoRefreshTimer); autoRefreshTimer = null; }
      if(checked){
        autoRefreshTimer = setInterval(reload, 5000);
        showToast('Auto refresh ON');
      } else {
        showToast('Auto refresh OFF');
      }
    }

    reload();
  </script>
</body>
</html>`