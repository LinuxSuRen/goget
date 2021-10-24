package main

import (
	"context"
	"fmt"
	"github.com/linuxsuren/goget/pkg/ui"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func main() {
	cmd := CreateCLICommand()
	_ = cmd.ExecuteContext(context.TODO())
}

type option struct {
	server string

	os        string
	arch      string
	branch    string
	goPackage string
	upx       bool
}

func CreateCLICommand() (cmd *cobra.Command) {
	opt := &option{}
	cmd = &cobra.Command{
		Use:   "goget",
		Short: "The client of goget-server",
		Args:  cobra.MinimumNArgs(1),
		RunE:  opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.server, "server", "", "http://goget.surenpi.com", "The desired server address")
	flags.StringVarP(&opt.os, "os", "", runtime.GOOS, "The desired OS")
	flags.StringVarP(&opt.arch, "arch", "", runtime.GOARCH, "The desired Arch")
	flags.StringVarP(&opt.branch, "branch", "", "master", "The desired git branch name")
	flags.StringVarP(&opt.goPackage, "package", "", "", "The desired go packages for building")
	flags.BoolVarP(&opt.upx, "upx", "", true, "Indicate if you want to upx it")
	return
}

func (o *option) runE(cmd *cobra.Command, args []string) (err error) {
	var resp *http.Response

	binaryName := args[0][strings.LastIndex(args[0], "/")+1:]
	api := fmt.Sprintf("%s/%s?os=%s&arch=%s&upx=%v&branch=%s&package=%s",
		o.server, args[0], o.os, o.arch, o.upx, o.branch, o.goPackage)
	if resp, err = http.Get(api); err == nil {
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode == http.StatusOK {
			// Create the file
			out, err := os.OpenFile(binaryName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0744)
			if err != nil {
				return err
			}
			defer func() {
				_ = out.Close()
			}()

			// Write the body to file
			err = copyWithProcess(out, resp)
		} else {
			var message string
			if resp.Body != nil {
				if data, respReadErr := ioutil.ReadAll(resp.Body); respReadErr == nil {
					message = string(data)
				}
			}
			err = fmt.Errorf("unexpected response code: %d, error message: %s", resp.StatusCode, message)
		}
	}
	return
}

func copyWithProcess(out io.Writer, resp *http.Response) (err error) {
	progressIndicator := &ui.ProgressIndicator{}

	if total, ok := resp.Header["Content-Length"]; ok && len(total) > 0 {
		fileLength, err := strconv.ParseInt(total[0], 10, 64)
		if err == nil {
			progressIndicator.Total = float64(fileLength)
		}
	}
	progressIndicator.Writer = out
	progressIndicator.Init()
	_, err = io.Copy(progressIndicator, resp.Body)
	return
}
