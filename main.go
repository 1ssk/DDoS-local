package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	startIP       = "192.168.1.1"   // Замените на ваш начальный IP
	endIP         = "192.168.1.255" // Замените на ваш конечный IP
	startPort     = 1
	endPort       = 65535
	numPackets    = 1000
	packetSize    = 1024
	numGoroutines = 100
)

func loadNetwork(wg *sync.WaitGroup, ip, port string) {
	defer wg.Done()

	conn, err := net.Dial("udp", fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		return
	}
	defer conn.Close()

	packet := make([]byte, packetSize)
	for i := 0; i < numPackets; i++ {
		conn.Write(packet)
	}
}

func main() {
	var wg sync.WaitGroup

	startTime := time.Now()
	for ipNum := ipToInt(startIP); ipNum <= ipToInt(endIP); ipNum++ {
		for port := startPort; port <= endPort; port++ {
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go loadNetwork(&wg, intToIP(ipNum).String(), fmt.Sprintf("%d", port))
			}
		}
	}

	wg.Wait()
	fmt.Println("Нагрузка завершена. Время выполнения:", time.Since(startTime))
}

// Функции для работы с IP
func ipToInt(ip string) int {
	ret := 0
	for _, b := range net.ParseIP(ip).To4() {
		ret <<= 8
		ret |= int(b)
	}
	return ret
}

func intToIP(nn int) net.IP {
	ip := make(net.IP, 4)
	for i := 0; i < 4; i++ {
		ip[3-i] = byte(nn & 0xFF)
		nn >>= 8
	}
	return ip
}
