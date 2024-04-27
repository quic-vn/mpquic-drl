package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	// "encoding/csv"
	// "fmt"

	quic "github.com/lucas-clemente/quic-go"

	"github.com/lucas-clemente/quic-go/h2quic"
	"github.com/lucas-clemente/quic-go/internal/utils"
)

func main() {

	verbose := flag.Bool("v", false, "verbose")
	sleeptime := flag.Int("t", 0, "sleep time for request in second")
	num := flag.Int("n", 1, "number of request")
	multipath := flag.Bool("m", false, "multipath")
	output := flag.String("o", "", "logging output")
	// address := flag.String("a", "", "mac address")
	cache := flag.Bool("c", false, "cache handshake information")
	flag.Parse()
	urls := flag.Args()

	// f, err := os.OpenFile("/App/output/result"+ *address + ".csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	// if err != nil{
	// 	panic(err)
	// }
	// csvwriter := csv.NewWriter(f)

	// os.Remove("/notebooks/result.pdf")
    // destination, err := os.Create("/notebooks/result.pdf")
    // if err != nil {
	// 	panic(err)
    // }
    // defer destination.Close()

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
	for i:= 0; i < *num; i++ {
		wg.Add(len(urls))
		for _, addr := range urls {
			utils.Infof("GET %s", addr)
			go func(addr string) {
				start := time.Now()
				rsp, err := hclient.Get(addr)
				if err != nil {
					panic(err)
				}

				body := &bytes.Buffer{}
				_, err = io.Copy(body, rsp.Body)
				if err != nil {
					//panic(err)
					utils.Infof("%f", float64(30000))
					// csvwriter.Write([]string{fmt.Sprint(float64(30000))})
					wg.Done()
				}else {
					elapsed := time.Since(start)
					utils.Infof("%f", float64(elapsed.Nanoseconds())/1000000)
					// csvwriter.Write([]string{fmt.Sprint(float64(elapsed.Nanoseconds())/1000000)})
					// io.Copy(destination, body)
					wg.Done()
				}
				// csvwriter.Flush()
				
			}(addr)
		}
		wg.Wait()
		time.Sleep(time.Duration(*sleeptime)*time.Second)
	}
}