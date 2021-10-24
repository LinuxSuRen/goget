package server

import (
	"fmt"
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
			err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	}
	return
}

func IntervalSelfRegistry(center, address string, duration time.Duration) (err error) {
	err = selfRegistry(center, address)

	go func() {
		ticker := time.NewTicker(duration)
		for range ticker.C {
			if err = selfRegistry(center, address); err != nil {
				fmt.Println("self registry failed", err)
			}
		}
	}()
	return
}
