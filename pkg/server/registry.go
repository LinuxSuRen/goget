package server

import (
	"fmt"
	"net/http"
	"time"
)

// selfRegistry registries myself to the center proxy
func selfRegistry(address string) (err error) {
	if address == "" {
		err = fmt.Errorf("the external address is empty")
	}

	var resp *http.Response
	if resp, err = http.Post(fmt.Sprintf("http://goget.surenpi.com/registry?address=%s", address), "", nil); err == nil {
		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	}
	return
}

func IntervalSelfRegistry(address string, duration time.Duration) (err error) {
	err = selfRegistry(address)

	go func() {
		ticker := time.NewTicker(duration)
		for range ticker.C {
			selfRegistry(address)
		}
	}()
	return
}
