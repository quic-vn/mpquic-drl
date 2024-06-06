package main

import (
	"bytes"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	quic "github.com/lucas-clemente/quic-go"

	"github.com/lucas-clemente/quic-go/h2quic"
	"github.com/lucas-clemente/quic-go/internal/utils"
)

func sendTrainSignal(wg *sync.WaitGroup) {
	defer wg.Done()
	data := map[string]interface{}{
		"train_flag": true,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	_, err = http.Post("http://10.0.0.20:8080/flag_training", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending POST request:", err)
	}
}

func main() {
	verbose := flag.Bool("v", false, "verbose")
	sleeptime := flag.Int("t", 0, "sleep time for request in second")
	num := flag.Int("n", 1, "number of request")
	clt := flag.Int("clt", 1, "number of client")
	multipath := flag.Bool("m", false, "multipath")
	output := flag.String("o", "", "logging output")
	cache := flag.Bool("c", false, "cache handshake information")
	flag.Parse()
	urls := flag.Args()

	filePath := "./logs/result" + fmt.Sprint(*clt) + ".csv"
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	csvwriter := csv.NewWriter(f)

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

	quicConfig := &quic.Config{
		CreatePaths:    *multipath,
		CacheHandshake: *cache,
	}

	hclient := &http.Client{
		Transport: &h2quic.RoundTripper{QuicConfig: quicConfig, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	var wg sync.WaitGroup
	for i := 0; i < *num; i++ {
		wg.Add(len(urls))
		for _, addr := range urls {
			utils.Infof("GET %s", addr)
			go func(addr string) {
				defer wg.Done()

				start := time.Now()
				rsp, err := hclient.Get(addr)
				if err != nil {
					log.Println("Error getting response:", err)
					return
				}
				defer rsp.Body.Close()

				body := &bytes.Buffer{}
				_, err = io.Copy(body, rsp.Body)
				if err != nil {
					log.Println("Error copying response body:", err)
					utils.Infof("%f", float64(30000))
					csvwriter.Write([]string{fmt.Sprint(float64(30000))})
				} else {
					elapsed := time.Since(start)
					utils.Infof("%f", float64(elapsed.Nanoseconds())/1000000)
					csvwriter.Write([]string{fmt.Sprint(float64(elapsed.Nanoseconds()) / 1000000)})
					csvwriter.Flush() // Gọi Flush() để đảm bảo dữ liệu được ghi ra file
				}
			}(addr)
		}
		wg.Wait()
		time.Sleep(time.Duration(*sleeptime) * time.Second)
	}
	// Thêm waitgroup cho sendTrainSignal
	wg.Add(1)
	go sendTrainSignal(&wg)
	wg.Wait()
}
