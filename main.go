package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	// Define command-line flags
	input := flag.String("in", "", "Input file name (csv)")
	output := flag.String("out", "/dev/stdout", "Output file name")
	domain := flag.String("domain", "", "Domain to lookup")
	maxParallelism := flag.Int("parall", 1, "Maximum parallelism")
	flag.Parse()

	if *input == "" {
		fmt.Println("Input filename is required")
		return
	}

	// Open the CSV file
	file, err := os.Open(*input)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Read the CSV records
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV:", err)
		return
	}

	// Results writer
	writer, err := os.Create(*output)
	if err != nil {
		fmt.Println("Error reading output file:", err)
		return
	}
	defer writer.Close()

	wg := &sync.WaitGroup{}
	done := make(chan struct{})
	ipCh := make(chan string, *maxParallelism*10)
	ipCount := int64(0)

	go func() {
		defer close(ipCh)
		for _, record := range records {
			ipCh <- record[0]
		}
	}()

	// Resolving dns records for ips
	for i := 0; i < *maxParallelism; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range ipCh {
				names, err := net.LookupAddr(ip)
				if err == nil && len(names) > 0 && strings.Contains(names[0], *domain) {
					fmt.Fprintf(writer, "%s: %s\n", ip, names[0])
				}

				// Increment the proceeded IPs count
				atomic.AddInt64(&ipCount, 1)
			}
		}()
	}

	// Checking progress
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fmt.Printf("Records processed: %d\n", atomic.LoadInt64(&ipCount))
			case <-done:
				return
			}
		}
	}()

	wg.Wait()
	close(done)

	// Print the final count
	fmt.Printf("Total IP addresses processed: %d\n", atomic.LoadInt64(&ipCount))
}
