package main

import (
	"fmt"
	"os/exec"
	"sync"
)

var mu sync.Mutex

// setupDNS 将所有网卡 DNS 设为静态 0.0.0.0
func setupDNS() error {
	mu.Lock()
	defer mu.Unlock()

	// 设置所有启用网卡 IPv4 DNS 为 0.0.0.0
	psSet := `Get-NetAdapter | Where-Object { $_.Status -eq 'Up' } | ForEach-Object {
    Set-DnsClientServerAddress -InterfaceIndex $_.InterfaceIndex -ServerAddresses ('0.0.0.0', '::')
}`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psSet)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("设置 IPv4 静态 DNS 失败: %v\n%s", err, string(out))
	}

	return nil
}

// restoreDNS 将所有网卡 DNS 恢复为自动(DHCP)
func restoreDNS() error {
	mu.Lock()
	defer mu.Unlock()

	// 重置所有网卡 DNS
	psReset := `Get-DnsClient | ForEach-Object { Set-DnsClientServerAddress -InterfaceIndex $_.InterfaceIndex -ResetServerAddresses }`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psReset)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("恢复自动 DNS 失败: %v\n%s", err, string(out))
	}

	return nil
}
