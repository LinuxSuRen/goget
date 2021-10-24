package main

import (
	"context"
	"fmt"
	"github.com/linuxsuren/goget/pkg/proxy"
	"github.com/linuxsuren/goget/pkg/server"
	"github.com/spf13/cobra"
	"net/http"
	"time"
)

func main() {
	cmd := createServerCommand()
	if err := cmd.ExecuteContext(context.TODO()); err != nil {
		panic(err)
	}
}

type serverOption struct {
	port            int
	bind            string
	mode            string
	externalAddress string
	gcDuration      string
	proxyCenter     string
}

var defaultGCDuration = time.Minute * 4

func createServerCommand() (cmd *cobra.Command) {
	opt := &serverOption{}

	cmd = &cobra.Command{
		Use:   "goget-server",
		Short: "This is a server to help build a Golang application",
		RunE:  opt.runE,
	}

	flags := cmd.Flags()
	flags.IntVarP(&opt.port, "port", "p", 7878, "The port of server")
	flags.StringVarP(&opt.bind, "bind", "b", "0.0.0.0", "The binding address of server")
	flags.StringVarP(&opt.mode, "mode", "m", "server", "This could be a normal server or a proxy")
	flags.StringVarP(&opt.externalAddress, "externalAddress", "", "",
		"The external address which used to registry to the center proxy")
	flags.StringVarP(&opt.proxyCenter, "proxyCenter", "", "http://goget.surenpi.com",
		"The address of the center proxy")
	flags.StringVarP(&opt.gcDuration, "gc-duration", "", defaultGCDuration.String(),
		"The duration of not alive candidates gc")
	return
}

func (o *serverOption) runE(cmd *cobra.Command, args []string) (err error) {
	switch o.mode {
	case "server":
		http.HandleFunc("/", server.GogetHandler)
		if err = server.IntervalSelfRegistry(o.proxyCenter, o.externalAddress, time.Minute*1); err != nil {
			err = fmt.Errorf("failed to self registry to the center proxy, error: %v", err)
			return
		}
		fmt.Println("self registry success")
	case "proxy":
		http.HandleFunc("/registry", proxy.RegistryHandler)
		http.HandleFunc("/", proxy.RedirectionHandler)

		var duration time.Duration
		var durationErr error
		if duration, durationErr = time.ParseDuration(o.gcDuration); durationErr != nil {
			duration = defaultGCDuration
		}
		proxy.CandidatesGC(context.TODO(), duration)
	}
	err = http.ListenAndServe(fmt.Sprintf("%s:%d", o.bind, o.port), nil)
	return
}
