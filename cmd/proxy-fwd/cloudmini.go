package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// parseCloudMiniProxy parses CloudMini proxy item to Upstream
// CloudMini ip field format: "hostname:internal_id" (e.g. "ipv4-vt-01.resvn.net:123")
// We need to extract hostname and use https field as port
func parseCloudMiniProxy(item CloudMiniProxyItem) (*Upstream, error) {
	// Parse IP field - format: "hostname:internal_id"
	// Strip the :internal_id suffix to get just the hostname
	hostname := item.IP
	if idx := strings.Index(item.IP, ":"); idx != -1 {
		hostname = item.IP[:idx]
	}
	fmt.Printf("[CloudMini Parse] Original IP: %s, Extracted hostname: %s, Port: %s\n", item.IP, hostname, item.HTTPS)

	// Parse HTTPS port
	port, err := strconv.Atoi(item.HTTPS)
	if err != nil {
		return nil, fmt.Errorf("invalid https port %q: %v", item.HTTPS, err)
	}

	up := &Upstream{
		ID:   sanitizeID(hostname, port),
		Host: hostname,
		Port: port,
		User: item.User,
		Pass: item.Password,
	}
	return up, nil
}

// checkProxyExitIP checks the exit IP of a proxy by making a request through it
func checkProxyExitIP(localPort int, token string) (string, error) {
	// Use a reliable IP check service
	checkURL := "https://api.ipify.org?format=text"

	proxyURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", localPort))
	if err != nil {
		return "", err
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", checkURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("got status %d", resp.StatusCode)
	}

	var buf [256]byte
	n, _ := resp.Body.Read(buf[:])
	ip := strings.TrimSpace(string(buf[:n]))

	return ip, nil
}
