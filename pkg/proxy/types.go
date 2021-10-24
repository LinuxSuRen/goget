package proxy

import (
	"fmt"
	"time"
)

type candidate struct {
	address   string
	heartBeat time.Time
	expired   bool
}

var aliveDuration = time.Minute * 2

const timeFormat = time.RFC3339

type candidateSlice struct {
	candidates []candidate
}

func (c *candidateSlice) findAlive() (candidate, bool) {
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
