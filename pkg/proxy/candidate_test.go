package proxy

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCandidate(t *testing.T) {
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
	emptyCandidates.addCandidate("fake")
	_, ok = emptyCandidates.findAlive()
	assert.True(t, ok, "should be able to find the alive candidate")
	// check the size
	emptyCandidates.addCandidate("fake")
	assert.Equal(t, emptyCandidates.size(), 1)

	expiredCandidates := candidateSlice{
		candidates: []candidate{{
			address: "fake",
			heartBeat: time.Now().Add(time.Minute * -10),
		}},
	}
	expiredCandidates.markExpired(0)
	_, ok = expiredCandidates.findAlive()
	assert.False(t, ok)
}