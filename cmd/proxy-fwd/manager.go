package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// NewManager creates a new Manager instance
func NewManager(adminToken string) *Manager {
	return &Manager{
		items:      make(map[string]*ProxyItem),
		nextPort:   firstLocalPort,
		adminToken: adminToken,
	}
}

// sanitizeID creates a safe ID from host and port
func sanitizeID(host string, port int) string {
	s := strings.ReplaceAll(host, ".", "-")
	return fmt.Sprintf("%s-%d", s, port)
}

// parseProxyLine parses "ip:port:user:pass" or "ip:port"
func parseProxyLine(line string) (*Upstream, error) {
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

// loadState loads state from yaml file
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

// saveState saves state to yaml file
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

// allocPort allocates next available port
func (m *Manager) allocPort() int {
	if m.nextPort < firstLocalPort {
		m.nextPort = firstLocalPort
	}
	p := m.nextPort
	m.nextPort++
	return p
}

// addOrReplace adds or updates an upstream proxy
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

// addToPool adds proxy to pool without starting (no local port assigned yet)
func (m *Manager) addToPool(up *Upstream) (*Upstream, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if up.ID == "" {
		up.ID = sanitizeID(up.Host, up.Port)
	}
	if existing, ok := m.items[up.ID]; ok {
		// already exists, just update credentials
		existing.cfg.Host = up.Host
		existing.cfg.Port = up.Port
		existing.cfg.User = up.User
		existing.cfg.Pass = up.Pass
		return existing.cfg, m.saveState()
	}
	// Add to pool without local port (will be assigned on start)
	up.LocalPort = 0
	up.Status = "stopped"
	m.items[up.ID] = &ProxyItem{cfg: up}
	return up, m.saveState()
}

// remove removes a proxy by ID
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

// start starts a proxy by ID
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
	// Assign local port if not yet assigned (from pool)
	if it.cfg.LocalPort == 0 {
		it.cfg.LocalPort = m.allocPort()
		_ = m.saveState()
	}
	return m.startLocked(it)
}

// stop stops a proxy by ID
func (m *Manager) stop(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.items[id]
	if !ok {
		return os.ErrNotExist
	}
	if err := m.stopLocked(it); err != nil {
		return err
	}
	// Save state after stopping (port released, moved to pool)
	return m.saveState()
}

// list returns all upstream configs
func (m *Manager) list() []*Upstream {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]*Upstream, 0)
	for _, it := range m.items {
		res = append(res, it.cfg)
	}
	// stable-ish order by LocalPort
	for i := 0; i < len(res); i++ {
		for j := i + 1; j < len(res); j++ {
			if res[j].LocalPort < res[i].LocalPort {
				res[i], res[j] = res[j], res[i]
			}
		}
	}
	return res
}

// syncFromAPI syncs proxies from external API
func (m *Manager) syncFromAPI(urlStr string) (int, []string) {
	client := &http.Client{Timeout: 10 * time.Second}
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
				if strings.TrimSpace(line) == "" {
					continue
				}
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
		if line == "" {
			continue
		}
		if _, e := m.addStartLine(line); e != nil {
			errs = append(errs, e.Error())
		} else {
			added++
		}
	}
	return added, errs
}

// addStartLine parses and starts a proxy from a line
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
