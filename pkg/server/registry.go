package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// selfRegistry registries myself to the center proxy
func selfRegistry(center, address string) (err error) {
	if address == "" {
		err = fmt.Errorf("the external address is empty")
	}

	var resp *http.Response
	if resp, err = http.Post(fmt.Sprintf("%s/registry?address=%s", center, address), "", nil); err == nil {
		if resp.StatusCode != http.StatusOK {
			data, _ := ioutil.ReadAll(resp.Body)
			err = fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(data))
		}
	}
	return
}

func IntervalSelfRegistry(center, address string, duration time.Duration) {
	go func() {
		ticker := time.NewTicker(duration)
		for range ticker.C {
			if err := selfRegistry(center, address); err != nil {
				fmt.Println("self registry failed", err)
			}
		}
	}()
}
