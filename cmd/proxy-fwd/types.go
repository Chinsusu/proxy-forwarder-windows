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
	healthURL       = "http://www.gstatic.com/generate_204" // lightweight 204
	healthInterval  = 10 * time.Second
	healthFailLimit = 3
)

// stateFile will be set to executable_dir/proxies.yaml in init()
var stateFile string

// Upstream represents a single upstream proxy configuration
type Upstream struct {
	ID        string `yaml:"id" json:"id"`
	Host      string `yaml:"host" json:"host"`
	Port      int    `yaml:"port" json:"port"`
	User      string `yaml:"user" json:"user"`
	Pass      string `yaml:"pass" json:"pass"`
	LocalPort int    `yaml:"local_port" json:"local_port"`
	ProxyType string `yaml:"proxy_type" json:"proxy_type"` // residential|privatev4|datacenter|static|unknown
	Location  string `yaml:"location" json:"location"`     // Geographic location

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
	IP       string `json:"ip"`        // format: "hostname:port"
	HTTPS    string `json:"https"`     // the actual proxy port
	User     string `json:"user"`      // username for auth
	Password string `json:"password"`  // password for auth
}

// CloudMiniOrderResponse represents the response from CloudMini order API
type CloudMiniOrderResponse struct {
	Code int                    `json:"code"`
	Msg  string                 `json:"msg"`
	Data []CloudMiniProxyItem   `json:"data"`
}

// CloudMiniProxyFull represents a full proxy item from /proxy endpoint
type CloudMiniProxyFull struct {
	PK       int    `json:"pk"`
	IP       string `json:"ip"`
	HTTPS    string `json:"https"`
	Socks    string `json:"socks"`
	User     string `json:"user"`
	Password string `json:"password"`
	Location string `json:"location"` // Already exists
	Status   string `json:"status"`
	Price    int    `json:"price"`
}

// CloudMiniRegionResponse represents the region config response
type CloudMiniRegionResponse struct {
	Error bool `json:"error"`
	Msg   string `json:"msg"`
	Data  []struct {
		Type   string   `json:"type"`
		Region []string `json:"region"`
	} `json:"data"`
}
