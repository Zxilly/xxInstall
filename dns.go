//go:build !windows

package main

func setupDNS() error {
	return nil
}

func restoreDNS() error {
	return nil
}
