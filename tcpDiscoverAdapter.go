package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

var TIME time.Duration = 10000000000

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Purple = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

var verbose = false
var warnings = false

func main() {

	if containsLower(os.Args, "help") || containsLower(os.Args, "--help") {

		println("Usage: exeName 80 [-v] [-w]")

		return
	}

	if containsLower(os.Args, "v") || containsLower(os.Args, "-v") {
		verbose = true
	}

	if containsLower(os.Args, "w") || containsLower(os.Args, "-w") {
		warnings = true
	}

	SetupCloseHandler()

	wg := sync.WaitGroup{}

	ifaces, err := net.Interfaces()
	if err != nil {
		println("Error ocurred on get adapters: " + err.Error())
		os.Exit(1)
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			println("Error ocurred on get adapter address: " + err.Error())
			continue
		}

		// handle err
		for _, addr := range addrs {

			ScanNetwork(addr.String(), os.Args[1], &wg)
		}
	}

	wg.Wait()
	os.Exit(0)
}

func ScanNetwork(network string, port string, wg *sync.WaitGroup) {

	if strings.HasPrefix(network, "169.254.") || strings.HasPrefix(network, "127.0") {
		return
	}

	_, ipv4Net, err := net.ParseCIDR(network)
	if err != nil {
		println("Error: ." + network + ". - " + err.Error())
		println(strings.Join(os.Args, ","))
		return
	}

	mask := binary.BigEndian.Uint32(ipv4Net.Mask)
	start := binary.BigEndian.Uint32(ipv4Net.IP) - 1

	finish := (start & mask) | (mask ^ 0xffffffff)

	for i := start; i < finish; i++ {

		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)
		wg.Add(1)
		go TestPort(ip.String(), os.Args[1], wg)
	}

	for i := 1; i < 255; i++ {
	}
}

func TestPort(addr string, port string, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr+":"+os.Args[1])
	if err != nil {
		if verbose {
			println(Gray + "[!] " + addr + " - " + Reset + " Is warning")
		}
		return
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		if verbose || warnings {
			show := verbose
			elapsed := time.Since(start)
			color := Red
			symbol := "-"
			if elapsed < TIME {
				color = Yellow
				symbol = "!"
				show = warnings
			}
			if show {
				println(color + "[" + symbol + "] " + port + " - Is closed " + Reset + addr)
			}
		}
		return
	} else {
		println(Green + "[+] " + port + " Is Open " + Reset + addr)
	}
	conn.Close()
}

func containsLower(s []string, e string) bool {
	for _, a := range s {
		if strings.ToLower(a) == e {
			return true
		}
	}
	return false
}

func SetupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\n\r- Ctrl+C pressed in Terminal")
		os.Exit(0)
	}()
}
