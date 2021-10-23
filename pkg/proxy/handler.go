package proxy

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"net/http"
	"strings"
	"time"
)

// KCandidates is the config key of candidates
const KCandidates = "candidates"

// RedirectionHandler is the handler of proxy
func RedirectionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received a request", r.RequestURI)

	candidates := getCandidatesFromConfig()

	fmt.Println("found possible candidates", candidates.size())
	if candidate, ok := candidates.findAlive(); ok {
		fmt.Println("redirect to", candidate.address)
		http.Redirect(w, r, fmt.Sprintf("https://%s/%s", candidate.address, r.RequestURI), http.StatusMovedPermanently)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte("no candidates found"))
}

// RegistryHandler receive the proxy registry request
func RegistryHandler(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	address = strings.ReplaceAll(address, "http://", "")
	address = strings.ReplaceAll(address, "https://", "")

	var (
		candidates *candidateSlice
	)
	if candidatesRaw, ok := viper.Get(KCandidates).([]interface{}); !ok {
		if candidatesRaw, ok := viper.Get(KCandidates).([]map[interface{}]interface{}); !ok {
			candidates = newFromMap(candidatesRaw)
		} else {
			candidates = newFromMap(candidatesRaw)
		}
	} else {
		candidates = newArray(candidatesRaw)
	}

	candidates.addCandidate(address)

	if err := saveCandidates(candidates); err == nil {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
	}
}

func saveCandidates(candidates *candidateSlice) (err error) {
	viper.Set(KCandidates, candidates.getMap())
	return viper.WriteConfig()
}

// CandidatesGC removes the not alive candidates
func CandidatesGC(ctx context.Context, duration time.Duration) {
	go func(ctx context.Context) {
		ticker := time.NewTicker(duration)

		for {
			select {
			case <-ticker.C:
				candidates := getCandidatesFromConfig()
				candidates.markExpired(time.Minute * 2)
				// TODO avoid some unnecessary saving if there's no change
				_ = saveCandidates(candidates)
			case <-ctx.Done():
				ticker.Stop()
			}
		}
	}(ctx)
}

func getCandidatesFromConfig() (candidates *candidateSlice) {
	if candidatesRaw, ok := viper.Get(KCandidates).([]interface{}); !ok {
		if candidatesRaw, ok := viper.Get(KCandidates).([]map[interface{}]interface{}); !ok {
			candidates = newFromMap(candidatesRaw)
		} else {
			candidates = newFromMap(candidatesRaw)
		}
	} else {
		candidates = newArray(candidatesRaw)
	}
	return
}

func init() {
	viper.SetConfigName("goget-server")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			err = nil
		}
	}
}
