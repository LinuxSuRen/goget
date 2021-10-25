package proxy

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
	"time"
)

func TestCandidateSlice(t *testing.T) {
	candidates := candidateSlice{
		[]candidate{{
			address:   "fake",
			heartBeat: time.Now(),
		}},
	}
	_, ok := candidates.findAlive()
	assert.True(t, ok, "should be able to find the alive candidate")

	oldCandidates := candidateSlice{
		[]candidate{{
			address:   "fake",
			heartBeat: time.Now().Add(aliveDuration * -4),
		}},
	}
	_, ok = oldCandidates.findAlive()
	assert.False(t, ok, "should not to find the alive candidate")
	// check if the heartBeat will be updated
	oldCandidates.addCandidate("fake")
	_, ok = oldCandidates.findAlive()
	assert.True(t, ok, "should be able to find the alive candidate which was updated")

	emptyCandidates := candidateSlice{}
	assert.Nil(t, emptyCandidates.first())
	emptyCandidates.addCandidate("fake")
	_, ok = emptyCandidates.findAlive()
	assert.True(t, ok, "should be able to find the alive candidate")
	// check the size
	emptyCandidates.addCandidate("fake")
	assert.Equal(t, emptyCandidates.size(), 1)

	expiredCandidates := candidateSlice{
		candidates: []candidate{{
			address:   "fake",
			heartBeat: time.Now().Add(time.Minute * -10),
		}},
	}
	expiredCandidates.markExpired(0)
	_, ok = expiredCandidates.findAlive()
	assert.False(t, ok)
}

func TestCandidatesHelper(t *testing.T) {
	// invalid candidates array
	candidatesArray := []interface{}{
		struct {
		}{},
	}
	candidates := newFromArray(candidatesArray)
	assert.Equal(t, 0, candidates.size())

	// valid candidates array
	candidatesArray = []interface{}{
		map[interface{}]interface{}{
			"address":   "fake",
			"heartBeat": time.Now().Format(timeFormat),
		},
	}
	candidates = newFromArray(candidatesArray)
	assert.Equal(t, 1, candidates.size())
	aliveCandidate, ok := candidates.findAlive()
	assert.True(t, ok)
	assert.Equal(t, "fake", aliveCandidate.address)

	// from map
	candidatesMap := []map[interface{}]interface{}{
		{
			"address":   "fake",
			"heartBeat": time.Now(),
		},
	}
	candidates = newFromMap(candidatesMap)
	assert.Equal(t, 1, candidates.size())
	aliveCandidate, ok = candidates.findAlive()
	assert.True(t, ok)
	assert.Equal(t, "fake", aliveCandidate.address)
}

func TestCandidateSliceSort(t *testing.T) {
	candidates := &candidateSlice{candidates: []candidate{{
		address:   "one",
		heartBeat: time.Now().Add(time.Minute),
	}, {
		address:   "two",
		heartBeat: time.Now().Add(time.Minute * 2),
	}}}
	sort.Sort(candidates)
	firstCandidate := candidates.first()
	assert.NotNil(t, firstCandidate)
	assert.Equal(t, firstCandidate.address, "two")
}

func TestCandidate(t *testing.T) {
	candidate := NewCandidate("http://fake")
	assert.Equal(t, "fake", candidate.getHost())

	candidate = NewCandidate("https://fake")
	assert.Equal(t, "fake", candidate.getHost())
}
