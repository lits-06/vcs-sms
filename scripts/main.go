package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/xuri/excelize/v2"
)

const (
	startPort = 20000
	endPort   = 29999
	outFile   = "fake_servers.xlsx"
	onServers = 8000
)

func main() {
	total := endPort - startPort + 1

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)

	f.SetCellValue(sheet, "A1", "name")
	f.SetCellValue(sheet, "B1", "ipv4")
	f.SetCellValue(sheet, "C1", "port")

	var wg sync.WaitGroup
	wg.Add(total)

	concurrency := 200
	sem := make(chan struct{}, concurrency)

	row := 2
	for p := startPort; p <= endPort; p++ {
		ip := fmt.Sprintf("%d.%d.%d.%d",
			rand.Intn(256),
			rand.Intn(256),
			rand.Intn(256),
			rand.Intn(256),
		)

		name := fmt.Sprintf("server-%05d", p-startPort+1)
		f.SetCellValue(sheet, "A"+strconv.Itoa(row), name)
		f.SetCellValue(sheet, "B"+strconv.Itoa(row), ip)
		f.SetCellValue(sheet, "C"+strconv.Itoa(row), p)
		row++

		if p-startPort < onServers {
			go func(port int) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				addr := fmt.Sprintf(":%d", port)
				// try to listen first to detect failure early
				ln, err := net.Listen("tcp", addr)
				if err != nil {
					log.Printf("[ERROR] %s - listen %s failed: %v", name, addr, err)
					return
				}

				mux := http.NewServeMux()
				mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintln(w, "OK")
				})
				server := &http.Server{
					Handler: mux,
				}

				go func() {
					if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
						log.Printf("[ERROR] %s - serve %s failed: %v", name, addr, err)
					}
				}()
			}(p)
		}
	}

	// Xóa file cũ nếu tồn tại
	if _, err := os.Stat(outFile); err == nil {
		if err := os.Remove(outFile); err != nil {
			log.Printf("Can not remove old file: %v", err)
		}
	}

	if err := f.SaveAs(outFile); err != nil {
		log.Fatalf("failed to save excel: %v", err)
	}

	wg.Wait()
	select {}
}
