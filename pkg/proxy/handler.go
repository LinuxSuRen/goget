package proxy

import (
	"context"
	"fmt"
	"github.com/linuxsuren/goget/pkg/common"
	"github.com/spf13/viper"
	"net/http"
	"sync"
	"time"
)

// KCandidates is the config key of candidates
const KCandidates = "candidates"

// RedirectionHandler is the handler of proxy
func RedirectionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received a request", r.RequestURI)
	if !common.IsValid(r.RequestURI) {
		// TODO do the validation check
		w.WriteHeader(http.StatusBadRequest)
		_,_ = w.Write([]byte("invalid request, please check https://github.com/LinuxSuRen/goget"))
		return
	}

	candidates := getCandidatesFromConfig()

	fmt.Println("found possible candidates", candidates.size())
	if candidate, ok := candidates.findAlive(); ok {
		fmt.Println("redirect to", candidate.address)
		http.Redirect(w, r, fmt.Sprintf("http://%s/%s", candidate.address, r.RequestURI), http.StatusMovedPermanently)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte("no candidates found, please feel free to be a candidate with command 'goget-server --mode proxy --externalAddress your-ip:port'"))
}

// RegistryHandler receive the proxy registry request
func RegistryHandler(w http.ResponseWriter, r *http.Request) {
	candidate := NewCandidate(r.URL.Query().Get("address"))
	if !candidate.reachable() {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(fmt.Sprintf("%s is not reachable", candidate.address)))
		return
	}

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
		candidates = newFromArray(candidatesRaw)
	}

	candidates.addCandidate(candidate.getHost())

	fmt.Println("receive candidate server", candidate.getHost())
	if err := saveCandidates(candidates); err == nil {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
	}
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

var mutex = &sync.Mutex{}
func saveCandidates(candidates *candidateSlice) (err error) {
	mutex.Lock()
	defer func() {
		mutex.Unlock()
	}()
	viper.Set(KCandidates, candidates.getMap())
	return viper.WriteConfig()
}

func getCandidatesFromConfig() (candidates *candidateSlice) {
	mutex.Lock()
	defer func() {
		mutex.Unlock()
	}()
	switch val := viper.Get(KCandidates).(type) {
	case []interface{}:
		candidates = newFromArray(val)
	case []map[interface{}]interface{}:
		candidates = newFromMap(val)
	default:
		fmt.Println(val)
		candidates = &candidateSlice{}
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
