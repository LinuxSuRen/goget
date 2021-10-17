package main

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
)

func main() {
	cmd := CreateServerCommand()
	if err := cmd.ExecuteContext(context.TODO()); err != nil {
		panic(err)
	}
}

func CreateServerCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "goget-server",
		Short: "This is a server to help build a Golang application",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Println(r.RequestURI)

				dir := path.Join("tmp", strings.ReplaceAll(r.RequestURI, "/github.com", ""))
				gitRepo := fmt.Sprintf("https:/%s", r.RequestURI)

				if ok, _ := PathExists(dir); ok {
					var repo *git.Repository
					if repo, err = git.PlainOpen(dir); err == nil {
						var wd *git.Worktree

						if wd, err = repo.Worktree(); err == nil {
							if err = wd.Pull(&git.PullOptions{
								Progress: cmd.OutOrStdout(),
								Force:    true,
							}); err != nil && err != git.NoErrAlreadyUpToDate {
								err = fmt.Errorf("failed to pull git repository '%s', error: %v", repo, err)
								return
							}
							err = nil
						}
					} else {
						err = fmt.Errorf("failed to open git local repository, error: %v", err)
					}
				} else if _, err = git.PlainClone(dir, false, &git.CloneOptions{
					URL:      gitRepo,
					Progress: cmd.OutOrStdout(),
				}); err != nil {
					fmt.Println(err)
					return
				}

				binaryName := r.RequestURI[strings.LastIndex(r.RequestURI, "/")+1:]

				args := []string{"build"}
				env := []string{
					pairFromQuery(r.URL, "os", "GOOS"),
					pairFromQuery(r.URL, "arch", "GOARCH"),
					"CGO_ENABLE=0",
				}
				if err := RunCommandInDir("go", dir, env, args...); err == nil {
					fmt.Println("success", binaryName)

					if data, err := ioutil.ReadFile(path.Join(dir, binaryName)); err == nil {
						_, _ = w.Write(data)
					}
				}
			})
			err = http.ListenAndServe("0.0.0.0:7878", nil)
			return nil
		},
	}
	return
}

func pairFromQuery(httpURL *url.URL, key, pairKey string) string {
	if val := httpURL.Query().Get(key); val != "" {
		return pair(pairKey, val)
	}
	return ""
}

func pair(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

// RunCommandInDir runs a command
func RunCommandInDir(name, dir string, env []string, args ...string) error {
	return RunCommandWithIO(name, dir, os.Stdout, os.Stderr, env, args...)
}

func RunCommandWithIO(name, dir string, stdout, stderr io.Writer, env []string, args ...string) (err error) {
	command := exec.Command(name, args...)
	if dir != "" {
		command.Dir = dir
	}
	env = append(env, os.Environ()...)

	//var stdout []byte
	//var errStdout error
	stdoutIn, _ := command.StdoutPipe()
	stderrIn, _ := command.StderrPipe()
	command.Env = env
	err = command.Start()
	if err != nil {
		return
	}

	// cmd.Wait() should be called only after we finish reading
	// from stdoutIn and stderrIn.
	// wg ensures that we finish
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_, _ = copyAndCapture(stdout, stdoutIn)
		wg.Done()
	}()

	_, _ = copyAndCapture(stderr, stderrIn)

	wg.Wait()

	err = command.Wait()
	return
}

func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}

// PathExists checks if the target path exist or not
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
