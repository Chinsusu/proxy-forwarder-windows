package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const cloudMiniBaseURL = "https://client.cloudmini.net/api/v2"

// handleCloudMiniRegions proxies request to CloudMini API to get regions
func (m *Manager) handleCloudMiniRegions(w http.ResponseWriter, r *http.Request) {
	if !m.handleAuth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token := r.URL.Query().Get("token")
	proxyType := r.URL.Query().Get("type")
	if token == "" {
		http.Error(w, "token required", http.StatusBadRequest)
		return
	}
	if proxyType == "" {
		proxyType = "proxy-res"
	}

	url := fmt.Sprintf("%s/order_config?type=%s", cloudMiniBaseURL, proxyType)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Token "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "CloudMini API error: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("CloudMini API returned %d: %s", resp.StatusCode, string(body)), resp.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}

// handleCloudMiniOrder proxies order request to CloudMini API
func (m *Manager) handleCloudMiniOrder(w http.ResponseWriter, r *http.Request) {
	if !m.handleAuth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("X-CloudMini-Token")
	if token == "" {
		http.Error(w, "X-CloudMini-Token header required", http.StatusBadRequest)
		return
	}

	// Parse JSON body from frontend
	var reqData struct {
		Type     string `json:"type"`
		Region   string `json:"region"`
		Quantity int    `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("[CloudMini Order] Type: %s, Region: %s, Quantity: %d\n", reqData.Type, reqData.Region, reqData.Quantity)

	// CloudMini API expects form-urlencoded
	formData := fmt.Sprintf("type=%s&region=%s", reqData.Type, reqData.Region)
	url := fmt.Sprintf("%s/order", cloudMiniBaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(formData))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Token "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	fmt.Printf("[CloudMini Order] Calling: %s with form: %s\n", url, formData)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "CloudMini API error: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("[CloudMini Order] API Error %d: %s\n", resp.StatusCode, string(bodyBytes))
		http.Error(w, fmt.Sprintf("CloudMini API returned %d: %s", resp.StatusCode, string(bodyBytes)), resp.StatusCode)
		return
	}

	// Parse order response to get order_id
	var orderResult struct {
		Error bool   `json:"error"`
		Msg   string `json:"msg"`
		Data  []struct {
			OrderID interface{} `json:"order_id"` // Can be string or number
		} `json:"data"`
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Printf("[CloudMini Order] Response: %s\n", string(bodyBytes))
	if err := json.Unmarshal(bodyBytes, &orderResult); err != nil {
		http.Error(w, "Failed to parse response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if orderResult.Error || len(orderResult.Data) == 0 || orderResult.Data[0].OrderID == nil {
		http.Error(w, fmt.Sprintf("Order failed: %s", orderResult.Msg), http.StatusBadRequest)
		return
	}

	// Convert order_id to string (can be int or string)
	orderID := fmt.Sprintf("%v", orderResult.Data[0].OrderID)
	fmt.Printf("[CloudMini Order] Created order_id: %s\n", orderID)

	// Poll order status until success (max 60 attempts, 5s each = 5 min)
	for i := 0; i < 60; i++ {
		time.Sleep(5 * time.Second)
		statusURL := fmt.Sprintf("%s/order?id=%s", cloudMiniBaseURL, orderID)
		statusReq, _ := http.NewRequest("GET", statusURL, nil)
		statusReq.Header.Set("Authorization", "Token "+token)

		statusResp, err := client.Do(statusReq)
		if err != nil {
			continue
		}

		var statusResult struct {
			Data []struct {
				OrderStatus string `json:"order_status"`
			} `json:"data"`
		}
		json.NewDecoder(statusResp.Body).Decode(&statusResult)
		statusResp.Body.Close()

		if len(statusResult.Data) > 0 {
			status := statusResult.Data[0].OrderStatus
			fmt.Printf("[CloudMini Order] Status check #%d: %s\n", i+1, status)
			if status == "success" {
				break
			}
		}
	}

	// Fetch proxies from order
	proxyURL := fmt.Sprintf("%s/proxy?order_id=%s", cloudMiniBaseURL, orderID)
	proxyReq, _ := http.NewRequest("GET", proxyURL, nil)
	proxyReq.Header.Set("Authorization", "Token "+token)

	proxyResp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Failed to fetch proxies: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer proxyResp.Body.Close()

	var proxyResult struct {
		Error bool                  `json:"error"`
		Msg   string                `json:"msg"`
		Data  []CloudMiniProxyItem `json:"data"`
	}

	proxyBytes, _ := io.ReadAll(proxyResp.Body)
	fmt.Printf("[CloudMini Order] Proxies response: %s\n", string(proxyBytes))
	if err := json.Unmarshal(proxyBytes, &proxyResult); err != nil {
		http.Error(w, "Failed to parse proxies: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if proxyResult.Error || len(proxyResult.Data) == 0 {
		http.Error(w, "No proxies returned: "+proxyResult.Msg, http.StatusBadRequest)
		return
	}

	// Return proxies in same format as original order endpoint expected
	w.Header().Set("Content-Type", "application/json")
	w.Write(proxyBytes)
}

// handleCloudMiniSync syncs all existing proxy-res proxies from CloudMini
func (m *Manager) handleCloudMiniSync(w http.ResponseWriter, r *http.Request) {
	if !m.handleAuth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token required", http.StatusBadRequest)
		return
	}

	client := &http.Client{Timeout: 30 * time.Second}

	// Fetch all proxies from CloudMini
	proxyURL := fmt.Sprintf("%s/proxy", cloudMiniBaseURL)
	proxyReq, _ := http.NewRequest("GET", proxyURL, nil)
	proxyReq.Header.Set("Authorization", "Token "+token)

	proxyResp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Failed to fetch proxies: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer proxyResp.Body.Close()

	var proxyResult struct {
		Error bool                   `json:"error"`
		Msg   string                 `json:"msg"`
		Data  []CloudMiniProxyFull `json:"data"`
	}

	if err := json.NewDecoder(proxyResp.Body).Decode(&proxyResult); err != nil {
		http.Error(w, "Failed to parse proxies: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if proxyResult.Error {
		http.Error(w, "CloudMini error: "+proxyResult.Msg, http.StatusBadRequest)
		return
	}

	// Sync all proxies without filtering by type
	fmt.Printf("[CloudMini Sync] Syncing all %d proxies (no filtering)\n", len(proxyResult.Data))
	
	// Log first 5 proxies for debugging
	for i := 0; i < len(proxyResult.Data) && i < 5; i++ {
		proxy := proxyResult.Data[i]
		hostname := proxy.IP
		if idx := len(proxy.IP); idx > 0 {
			for j := len(proxy.IP) - 1; j >= 0; j-- {
				if proxy.IP[j] == ':' {
					hostname = proxy.IP[:j]
					break
				}
			}
		}
		fmt.Printf("  [%d] ip=%q hostname=%q port=%s\n", proxy.PK, proxy.IP, hostname, proxy.HTTPS)
	}
	
	filtered := proxyResult.Data

	// 4) Add filtered proxies to pool
	added := 0
	var errors []string
	m.mu.Lock()
	for _, proxy := range filtered {
		// Parse hostname (strip :ID suffix)
		hostname := proxy.IP
		if idx := len(proxy.IP); idx > 0 {
			for i := len(proxy.IP) - 1; i >= 0; i-- {
				if proxy.IP[i] == ':' {
					hostname = proxy.IP[:i]
					break
				}
			}
		}

		// Parse port
		var port int
		fmt.Sscanf(proxy.HTTPS, "%d", &port)
		if port == 0 {
			errors = append(errors, fmt.Sprintf("Invalid port for %s", proxy.IP))
			continue
		}

		up := &Upstream{
			ID:        sanitizeID(hostname, port),
			Host:      hostname,
			Port:      port,
			User:      proxy.User,
			Pass:      proxy.Password,
			LocalPort: 0, // Add to pool
			ProxyType: detectProxyTypeWithPrice(hostname, proxy.Price),
			Location:  proxy.Location,
			Status:    "stopped",
		}

		// Check if already exists
		if existing, ok := m.items[up.ID]; ok {
			// Update credentials
			existing.cfg.User = up.User
			existing.cfg.Pass = up.Pass
		} else {
			m.items[up.ID] = &ProxyItem{cfg: up}
			added++
		}
	}
	_ = m.saveState()
	m.mu.Unlock()

	fmt.Printf("[CloudMini Sync] Added %d new proxies to pool (total: %d)\n", added, len(filtered))

	// Return result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total":    len(filtered),
		"added":    added,
		"existing": len(filtered) - added,
		"errors":   errors,
	})
}
