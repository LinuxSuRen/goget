package proxy

import (
	"fmt"
	"github.com/spf13/viper"
	"net/http"
)

// KCandidates is the config key of candidates
const KCandidates = "candidates"

// RedirectionHandler is the handler of proxy
func RedirectionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received a request", r.RequestURI)

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

	viper.Set(KCandidates, candidates.getMap())
	if err := viper.WriteConfig(); err == nil {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
	}
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
