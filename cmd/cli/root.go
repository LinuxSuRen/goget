package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
)

func main() {
	cmd := CreateCLICommand()
	if err := cmd.ExecuteContext(context.TODO()); err != nil {
		panic(err)
	}
}

type option struct {
	server string

	os   string
	arch string
	upx  bool
}

func CreateCLICommand() (cmd *cobra.Command) {
	opt := &option{}
	cmd = &cobra.Command{
		Use:   "goget",
		Short: "The client of goget-server",
		RunE:  opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.server, "server", "", "localhost:7878", "The desired server address")
	flags.StringVarP(&opt.os, "os", "", runtime.GOOS, "The desired OS")
	flags.StringVarP(&opt.arch, "arch", "", runtime.GOARCH, "The desired Arch")
	flags.BoolVarP(&opt.upx, "upx", "", true, "Indicate if you want to upx it")
	return
}

func (o *option) runE(cmd *cobra.Command, args []string) (err error) {
	var resp *http.Response

	binaryName := args[0][strings.LastIndex(args[0], "/")+1:]

	api := fmt.Sprintf("http://%s/%s?os=%s&arch=%s&upx=%s", o.server, args[0], o.os, o.arch, o.upx)
	if resp, err = http.Get(api); err == nil && resp.StatusCode == http.StatusOK {
		var data []byte
		if data, err = ioutil.ReadAll(resp.Body); err == nil {
			err = ioutil.WriteFile(binaryName, data, 0544)
		}
	}
	return
}
