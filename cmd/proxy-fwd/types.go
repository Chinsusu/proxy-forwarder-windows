package main

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	defaultUIAddr   = "127.0.0.1:17890"
	firstLocalPort  = 10001
	stateFile       = "proxies.yaml"
	healthURL       = "http://www.gstatic.com/generate_204" // lightweight 204
	healthInterval  = 10 * time.Second
	healthFailLimit = 3
)

// Upstream represents a single upstream proxy configuration
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

// State represents the persisted state
type State struct {
	Items []*Upstream `yaml:"items"`
	Next  int         `yaml:"next"`
}

// Manager manages all proxy items
type Manager struct {
	mu       sync.RWMutex
	items    map[string]*ProxyItem // id -> ProxyItem
	nextPort int

	adminToken string
}

// ProxyItem holds runtime data for a single proxy
type ProxyItem struct {
	cfg       *Upstream
	server    *http.Server
	listener  net.Listener
	stopFn    context.CancelFunc
	healthWg  sync.WaitGroup
	isRunning bool
}

// CloudMiniProxyItem represents a proxy from CloudMini API
type CloudMiniProxyItem struct {
	IP       string `json:"ip"`        // format: "hostname:internal_id"
	HTTPS    string `json:"https"`     // the actual proxy port
	Username string `json:"username"`
	Password string `json:"password"`
}

// CloudMiniOrderResponse represents the response from CloudMini order API
type CloudMiniOrderResponse struct {
	Code int                    `json:"code"`
	Msg  string                 `json:"msg"`
	Data []CloudMiniProxyItem   `json:"data"`
}
