package proxy

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

type candidate struct {
	address   string
	heartBeat time.Time
	expired   bool
}

// NewCandidate creates a new candidate instance
func NewCandidate(address string) *candidate {
	return &candidate{address: address}
}

func (c *candidate) reachable() (ok bool) {
	if c.address == "" {
		return false
	}

	var address string
	if strings.HasPrefix(c.address, "https://") || strings.HasPrefix(c.address, "http://") {
		address = c.address
	} else {
		address = fmt.Sprintf("http://%s", c.address)
	}

	client := http.DefaultClient
	client.Timeout = time.Second * 10
	resp, err := client.Get(fmt.Sprintf("%s/health", address))
	ok = err == nil && resp.StatusCode == http.StatusOK
	if !ok {
		fmt.Printf("failed to check health: %s\n", err.Error())
	}
	return
}

func (c *candidate) getHost() (address string) {
	address = c.address
	address = strings.ReplaceAll(address, "http://", "")
	address = strings.ReplaceAll(address, "https://", "")
	return
}

var aliveDuration = time.Minute * 2

const timeFormat = time.RFC3339

type candidateSlice struct {
	candidates []candidate
}

func (c *candidateSlice) first() *candidate {
	if len(c.candidates) > 0 {
		return &c.candidates[0]
	}
	return nil
}

func (c *candidateSlice) findAlive() (candidate, bool) {
	sort.Sort(c)

	for i, _ := range c.candidates {
		candidateItem := c.candidates[i]
		if candidateItem.address != "" && candidateItem.heartBeat.Add(aliveDuration).After(time.Now()) {
			return candidateItem, true
		}
	}
	return candidate{}, false
}

func (c *candidateSlice) markExpired(toleration time.Duration) {
	for i, _ := range c.candidates {
		candidateItem := c.candidates[i]
		if candidateItem.address == "" ||
			candidateItem.heartBeat.Add(aliveDuration).Add(toleration).Before(time.Now()) {
			c.candidates[i].expired = true
		}
	}
}

func (c *candidateSlice) addCandidate(address string) {
	for i, _ := range c.candidates {
		if c.candidates[i].address == address {
			// update the existing candidate heartBeat
			c.candidates[i].heartBeat = time.Now()
			return
		}
	}

	// add a new candidate
	c.candidates = append(c.candidates, candidate{
		address:   address,
		heartBeat: time.Now(),
	})
}

func (c *candidateSlice) size() int {
	return len(c.candidates)
}

func (c *candidateSlice) Len() int {
	return c.size()
}

func (c *candidateSlice) Less(i, j int) bool {
	left := c.candidates[i].heartBeat
	right := c.candidates[j].heartBeat
	return left.After(right)
}

func (c *candidateSlice) Swap(i, j int) {
	c.candidates[i], c.candidates[j] = c.candidates[j], c.candidates[i]
}

func (c *candidateSlice) getMap() (result []map[interface{}]interface{}) {
	result = make([]map[interface{}]interface{}, 0)
	for i, _ := range c.candidates {
		can := c.candidates[i]
		// don't persistent the expired candidates
		if can.expired {
			fmt.Println("skip expired candidate:", can.address)
			continue
		}

		result = append(result, map[interface{}]interface{}{
			"address":   can.address,
			"heartBeat": can.heartBeat.Format(timeFormat),
		})
	}
	return
}

func newFromArray(candidates []interface{}) *candidateSlice {
	targetCandidates := make([]candidate, 0)

	for i, _ := range candidates {
		can := candidates[i]

		if canMap, ok := can.(map[interface{}]interface{}); ok {
			heartBeat, _ := time.Parse(time.RFC3339, fmt.Sprintf("%v", canMap["heartBeat"]))
			targetCandidates = append(targetCandidates, candidate{
				address:   fmt.Sprintf("%v", canMap["address"]),
				heartBeat: heartBeat,
			})
		}
	}
	return &candidateSlice{
		candidates: targetCandidates,
	}
}

func newFromMap(candidates []map[interface{}]interface{}) *candidateSlice {
	targetCandidates := make([]candidate, 0)

	for i, _ := range candidates {
		can := candidates[i]

		targetCandidate := candidate{
			address: fmt.Sprintf("%v", can["address"]),
		}

		switch v := can["heartBeat"].(type) {
		case time.Time:
			targetCandidate.heartBeat = v
		case string:
			targetCandidate.heartBeat, _ = time.Parse(timeFormat, v)
		}
		targetCandidates = append(targetCandidates, targetCandidate)
	}
	return &candidateSlice{
		candidates: targetCandidates,
	}
}
