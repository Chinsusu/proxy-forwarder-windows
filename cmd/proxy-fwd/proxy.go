package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	goproxy "github.com/elazarl/goproxy"
)

// startLocked starts a proxy (must be called with Manager lock held)
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
		ProxyConnectHeader:    http.Header{},
		MaxConnsPerHost:       0,
		MaxIdleConns:          128,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
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

// stopLocked stops a proxy (must be called with Manager lock held)
// Releases the local port so proxy moves to pool
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
	
	// Release the port by setting LocalPort to 0
	// This moves the proxy to pool and allows port reuse
	oldPort := it.cfg.LocalPort
	it.cfg.LocalPort = 0
	log.Printf("[proxy %s] stopped and released port %d (moved to pool)", it.cfg.ID, oldPort)
	
	return nil
}
