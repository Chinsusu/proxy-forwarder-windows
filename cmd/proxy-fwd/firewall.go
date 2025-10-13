package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	firewallGroup = "ProxyFwd Rules"
	portRange     = "10001-20000"
	loopback      = "127.0.0.1"
)

// isAdmin checks if the process is running with administrator privileges
func isAdmin() bool {
	cmd := exec.Command("net", "session")
	err := cmd.Run()
	return err == nil
}

// setupFirewall creates Windows firewall rules to prevent WebRTC and direct internet access
func setupFirewall() error {
	if !isAdmin() {
		log.Printf("[Firewall] Warning: Not running as Administrator. Firewall rules cannot be created.")
		log.Printf("[Firewall] To enable firewall protection, run as Administrator or manually execute: scripts\\firewall_rules.ps1")
		return nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot get executable path: %w", err)
	}

	log.Printf("[Firewall] Setting up firewall rules...")

	// 1. Allow proxy-fwd.exe to access internet
	if err := createAllowRule(exePath); err != nil {
		log.Printf("[Firewall] Warning: Failed to create allow rule for proxy-fwd: %v", err)
	} else {
		log.Printf("[Firewall] ✅ Allow rule created for proxy-fwd.exe")
	}

	// 2. Block browsers except localhost proxy ports
	browsers := []string{
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files\Mozilla Firefox\firefox.exe`,
		`C:\Program Files (x86)\Mozilla Firefox\firefox.exe`,
		`C:\Program Files\BraveSoftware\Brave-Browser\Application\brave.exe`,
		`C:\Program Files (x86)\BraveSoftware\Brave-Browser\Application\brave.exe`,
	}

	blockedCount := 0
	for _, browser := range browsers {
		if _, err := os.Stat(browser); err == nil {
			// Allow localhost:10001-20000
			if err := createLocalhostAllowRule(browser); err != nil {
				log.Printf("[Firewall] Warning: Failed to create localhost rule for %s: %v", filepath.Base(browser), err)
			}
			// Block all other internet access
			if err := createBlockRule(browser); err != nil {
				log.Printf("[Firewall] Warning: Failed to create block rule for %s: %v", filepath.Base(browser), err)
			} else {
				blockedCount++
			}
		}
	}

	if blockedCount > 0 {
		log.Printf("[Firewall] ✅ Firewall rules created for %d browser(s)", blockedCount)
		log.Printf("[Firewall] 🛡️  WebRTC leak protection active")
	} else {
		log.Printf("[Firewall] ℹ️  No browsers found for firewall rules")
	}

	return nil
}

// createAllowRule allows proxy-fwd.exe to access internet
func createAllowRule(exePath string) error {
	ruleName := "ProxyFwd Allow Out"
	
	// Check if rule already exists
	if ruleExists(ruleName) {
		return nil // Already exists
	}

	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		fmt.Sprintf(`New-NetFirewallRule -DisplayName "%s" -Direction Outbound -Program "%s" -Action Allow -Profile Any -Protocol TCP -RemoteAddress Any -EdgeTraversalPolicy Block -Group "%s" -ErrorAction SilentlyContinue`,
			ruleName, exePath, firewallGroup))
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create allow rule: %w, output: %s", err, output)
	}
	return nil
}

// createLocalhostAllowRule allows browser to access localhost proxy ports
func createLocalhostAllowRule(browserPath string) error {
	browserName := filepath.Base(browserPath)
	ruleName := fmt.Sprintf("ProxyFwd Allow Localhost for %s", browserName)
	
	// Check if rule already exists
	if ruleExists(ruleName) {
		return nil
	}

	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		fmt.Sprintf(`New-NetFirewallRule -DisplayName "%s" -Direction Outbound -Program "%s" -Action Allow -Profile Any -Protocol TCP -RemoteAddress %s -RemotePort %s -Group "%s" -ErrorAction SilentlyContinue`,
			ruleName, browserPath, loopback, portRange, firewallGroup))
	
	_, err := cmd.CombinedOutput()
	return err
}

// createBlockRule blocks browser from accessing internet directly
func createBlockRule(browserPath string) error {
	browserName := filepath.Base(browserPath)
	ruleName := fmt.Sprintf("ProxyFwd Block Internet for %s", browserName)
	
	// Check if rule already exists
	if ruleExists(ruleName) {
		return nil
	}

	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		fmt.Sprintf(`New-NetFirewallRule -DisplayName "%s" -Direction Outbound -Program "%s" -Action Block -Profile Any -Protocol TCP -RemoteAddress Any -Group "%s" -ErrorAction SilentlyContinue`,
			ruleName, browserPath, firewallGroup))
	
	_, err := cmd.CombinedOutput()
	return err
}

// ruleExists checks if a firewall rule already exists
func ruleExists(ruleName string) bool {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		fmt.Sprintf(`Get-NetFirewallRule -DisplayName "%s" -ErrorAction SilentlyContinue`, ruleName))
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != ""
}

// cleanupFirewall removes all ProxyFwd firewall rules
func cleanupFirewall() error {
	if !isAdmin() {
		log.Printf("[Firewall] Warning: Not running as Administrator. Cannot cleanup firewall rules.")
		return nil
	}

	log.Printf("[Firewall] Cleaning up firewall rules...")

	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		fmt.Sprintf(`Get-NetFirewallRule -Group "%s" -ErrorAction SilentlyContinue | Remove-NetFirewallRule -ErrorAction SilentlyContinue`, firewallGroup))
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Not an error if no rules exist
		if strings.Contains(string(output), "No MSFT_NetFirewallRule objects found") {
			log.Printf("[Firewall] No rules to cleanup")
			return nil
		}
		return fmt.Errorf("failed to cleanup firewall rules: %w, output: %s", err, output)
	}

	log.Printf("[Firewall] ✅ Firewall rules removed")
	return nil
}

// listFirewallRules lists all ProxyFwd firewall rules
func listFirewallRules() ([]string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		fmt.Sprintf(`Get-NetFirewallRule -Group "%s" -ErrorAction SilentlyContinue | Select-Object -ExpandProperty DisplayName`, firewallGroup))
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	rules := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			rules = append(rules, line)
		}
	}
	return rules, nil
}
