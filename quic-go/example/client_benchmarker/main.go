package main

import (
	"bytes"
	"crypto/tls"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	quic "github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/h2quic"
	"github.com/lucas-clemente/quic-go/internal/utils"
)

func main() {
	// Flag cấu hình
	verbose := flag.Bool("v", false, "verbose")
	sleeptime := flag.Int("t", 0, "sleep time (seconds) between each request")
	num := flag.Int("n", 10, "number of requests per thread")
	multipath := flag.Bool("m", false, "multipath")
	output := flag.String("o", "", "logging output file")
	cache := flag.Bool("c", false, "cache handshake information")
	flag.Int("clt", 1, "number of clients")

	flag.Parse()

	alphaValue := "1234"

	// Danh sách URLs
	urls := []string{
		"https://10.0.0.20:6121/files/2MB-3",
		"https://10.0.3.20:6121/files/2MB-3",
		"https://10.0.4.20:6121/files/2MB-3",
	}

	//log
	if *verbose {
		utils.SetLogLevel(utils.LogLevelDebug)
	} else {
		utils.SetLogLevel(utils.LogLevelInfo)
	}

	if *output != "" {
		logfile, err := os.Create(*output)
		if err != nil {
			panic(err)
		}
		defer logfile.Close()
		log.SetOutput(logfile)
	}

	// Cấu hình QUIC
	quicConfig := &quic.Config{
		CreatePaths:    *multipath,
		CacheHandshake: *cache,
	}

	// Tạo http.Client sử dụng h2quic
	hclient := &http.Client{
		Transport: &h2quic.RoundTripper{
			QuicConfig:      quicConfig,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	var wg sync.WaitGroup

	// Số thread = số lượng URL
	for i, rawURL := range urls {
		wg.Add(1)
		go func(threadID int, rawURL string) {
			defer wg.Done()

			// Tạo file CSV riêng cho thread này
			filePath := fmt.Sprintf("./logs/result_%d.csv", threadID)
			f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			csvWriter := csv.NewWriter(f)
			defer csvWriter.Flush()

			// Parse URL 1 lần bên ngoài vòng for
			parsedURL, err := url.Parse(rawURL)
			if err != nil {
				log.Printf("Thread %d - error parse URL %s: %v\n", threadID, rawURL, err)
				return
			}

			// Lặp lại *num lần GET request cho URL này
			for j := 0; j < *num; j++ {
				// Thêm alpha vào query
				q := parsedURL.Query()
				q.Set("alpha", alphaValue)
				parsedURL.RawQuery = q.Encode()

				fullURL := parsedURL.String()
				utils.Infof("Thread %d - full URL: %s", threadID, fullURL)

				start := time.Now()
				rsp, err := hclient.Get(fullURL)
				if err != nil {
					log.Printf("Thread %d - error GET %s: %v\n", threadID, fullURL, err)
					// Ghi giá trị 30000 ms để biểu thị lỗi
					csvWriter.Write([]string{fmt.Sprintf("Request %d error: 30000", j+1)})
					csvWriter.Flush()
				} else {
					// Đọc body để đảm bảo request được hoàn tất
					body := &bytes.Buffer{}
					_, err = io.Copy(body, rsp.Body)
					rsp.Body.Close()
					if err != nil {
						log.Printf("Thread %d - error reading response body: %v\n", threadID, err)
						csvWriter.Write([]string{fmt.Sprintf("Request %d error: 30000", j+1)})
						csvWriter.Flush()
					} else {
						elapsed := time.Since(start)
						elapsedMs := float64(elapsed.Nanoseconds()) / 1e6
						utils.Infof("Thread %d - GET %s time: %.3f ms", threadID, fullURL, elapsedMs)

						// Ghi thời gian vào file CSV
						csvWriter.Write([]string{
							fmt.Sprintf("Request %d", j+1),
							fmt.Sprintf("%.3f", elapsedMs),
						})
						csvWriter.Flush()
					}
				}

				// Nghỉ *sleeptime giây (nếu > 0) giữa các lần request
				time.Sleep(time.Duration(*sleeptime) * time.Second)
			}
		}(i, rawURL)
	}

	// Chờ tất cả thread hoàn thành
	wg.Wait()
}
