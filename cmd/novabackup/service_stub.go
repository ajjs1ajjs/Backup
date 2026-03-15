//go:build !windows

package main

import "fmt"

func runAsServiceIfNeeded() bool {
	return false
}

func installService() {
	fmt.Println("Service installation is only supported on Windows.")
	fmt.Println("Please run novabackup server to start manually.")
}

func removeService() {
	fmt.Println("Service removal is only supported on Windows.")
}
