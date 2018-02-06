package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	apex "github.com/apex/go-apex"
)

func main() {
	apex.HandleFunc(func(event json.RawMessage, ctx *apex.Context) (interface{}, error) {
		tr := &http.Transport{}
		tr.Dial = func(network, addr string) (net.Conn, error) {
			host := strings.Split(addr, ":")[0]
			_, results, err := net.LookupSRV("", "", host)
			if err != nil {
				return nil, err
			}

			// range over DNS response array and return first succeed connection
			for i := range results {
				res := results[i]
				ip, port := strings.TrimRight(res.Target, "."), res.Port
				address := fmt.Sprintf("%v:%v", ip, port)
				conn, err := net.DialTimeout(network, address, 50*time.Millisecond)
				if err == nil {
					return conn, nil
				}
				if i == len(results)-1 {
					return nil, err
				}
			}

			return nil, nil
		}
		c1 := http.Client{
			Transport: tr,
		}
		req, _ := http.NewRequest("GET", "http://test.mmg.cloud", nil)
		start := time.Now()
		r1, err := c1.Do(req)
		if err != nil {
			return err, nil
		}
		log.Printf("CUSTOM CLIENT TIME: %v\n", time.Since(start))
		defer r1.Body.Close()
		buf1 := new(bytes.Buffer)
		buf1.ReadFrom(r1.Body)

		c2 := http.Client{}
		start = time.Now()
		r2, err := c2.Do(req)
		if err != nil {
			return err, nil
		}
		log.Printf("DEFAULT CLIENT TIME: %v\n", time.Since(start))
		defer r2.Body.Close()

		buf2 := new(bytes.Buffer)
		buf2.ReadFrom(r2.Body)
		return buf1.String() + " " + buf2.String(), nil
	})
}
