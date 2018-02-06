package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

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
				conn, err := net.Dial(network, address)
				if err == nil {
					return conn, nil
				}
				if i == len(results)-1 {
					return nil, err
				}
			}

			return nil, nil
		}
		c := http.Client{
			Transport: tr,
		}
		req, _ := http.NewRequest("GET", "http://test.some.cloud", nil)
		r, err := c.Do(req)
		if err != nil {
			return err, nil
		}
		defer r.Body.Close()

		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		return buf.String(), nil
	})
}
