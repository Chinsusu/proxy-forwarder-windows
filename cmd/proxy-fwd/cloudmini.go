package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// detectProxyType detects proxy type based on hostname pattern
// - residential: hostname starts with ipv4- or ipv6- (e.g., ipv4-vt-01.resvn.net)
// - privatev4: hostname contains "isp" or premium static IPs
// - static: raw IP address format (e.g., 103.x.x.x)
// - datacenter: datacenter-related keywords
// - unknown: cannot determine type
func detectProxyType(host string) string {
	// Convert to lowercase for case-insensitive matching
	lowerHost := strings.ToLower(host)
	
	// Check for residential proxies (ipv4-*, ipv6-*)
	if len(lowerHost) >= 5 && (strings.HasPrefix(lowerHost, "ipv4-") || strings.HasPrefix(lowerHost, "ipv6-")) {
		return "residential"
	}
	
	// Check for PrivateV4 proxies
	if strings.Contains(lowerHost, "isp") || strings.Contains(lowerHost, "privatev4") {
		return "privatev4"
	}
	
	// Check for datacenter keywords
	if strings.Contains(lowerHost, "datacenter") || strings.Contains(lowerHost, "dc") || strings.Contains(lowerHost, "cloud") {
		return "datacenter"
	}
	
	// Check if it's a raw IP address (simple check for numbers and dots)
	// Format: xxx.xxx.xxx.xxx
	if isIPAddress(host) {
		return "static"
	}
	
	// Unknown type
	return "unknown"
}

// detectProxyTypeWithPrice detects proxy type using both hostname and price
// Price-based detection (CloudMini pricing):
// - >= 100000: Residential (premium rotating proxies)
// - 10000-99999: PrivateV4 (mid-tier premium static)
// - < 10000: Static/Datacenter (cheap/old proxies)
func detectProxyTypeWithPrice(host string, price int) string {
	// Convert to lowercase for case-insensitive matching
	lowerHost := strings.ToLower(host)
	
	// First check hostname patterns (most reliable)
	if len(lowerHost) >= 5 && (strings.HasPrefix(lowerHost, "ipv4-") || strings.HasPrefix(lowerHost, "ipv6-")) {
		return "residential"
	}
	
	if strings.Contains(lowerHost, "isp") || strings.Contains(lowerHost, "privatev4") {
		return "privatev4"
	}
	
	if strings.Contains(lowerHost, "datacenter") || strings.Contains(lowerHost, "dc") {
		return "datacenter"
	}
	
	// Use price to help determine type for ambiguous cases
	if price >= 100000 {
		// High price = Residential
		return "residential"
	} else if price >= 10000 && price < 100000 {
		// Medium price = PrivateV4 premium static
		if isIPAddress(host) {
			return "privatev4" // Premium static IP
		}
		return "privatev4"
	} else if price > 0 && price < 10000 {
		// Low price = Static/Datacenter
		if isIPAddress(host) {
			return "static" // Cheap static IP
		}
		return "datacenter"
	}
	
	// Fallback to basic detection
	if isIPAddress(host) {
		return "static"
	}
	
	return "unknown"
}

// isIPAddress checks if a string is an IP address
func isIPAddress(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if num, err := strconv.Atoi(part); err != nil || num < 0 || num > 255 {
			return false
		}
	}
	return true
}

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
		ID:        sanitizeID(hostname, port),
		Host:      hostname,
		Port:      port,
		User:      item.User,
		Pass:      item.Password,
		ProxyType: detectProxyType(hostname),
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
