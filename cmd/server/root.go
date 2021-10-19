package main

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
	cmd := createServerCommand()
	if err := cmd.ExecuteContext(context.TODO()); err != nil {
		panic(err)
	}
}

func createServerCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "goget-server",
		Short: "This is a server to help build a Golang application",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Println(r.RequestURI)

				requestPath := strings.Split(r.RequestURI, "?")[0]
				dir := path.Join("tmp", strings.ReplaceAll(requestPath, "/github.com", ""))
				gitRepo := fmt.Sprintf("https:/%s", requestPath)
				branch := getBranch(r.URL)

				if ok, _ := PathExists(dir); ok {
					var repo *git.Repository
					if repo, err = git.PlainOpen(dir); err == nil {
						var wd *git.Worktree

						if wd, err = repo.Worktree(); err == nil {
							if err = wd.Pull(&git.PullOptions{
								Progress: cmd.OutOrStdout(),
								ReferenceName: plumbing.NewBranchReferenceName(branch),
								Force:    true,
							}); err != nil && err != git.NoErrAlreadyUpToDate {
								err = fmt.Errorf("failed to pull git repository '%s', error: %v", repo, err)
							} else {
								err = nil
							}
						}
					} else {
						err = fmt.Errorf("failed to open git local repository, error: %v", err)
					}
				} else {
					_, err = git.PlainClone(dir, false, &git.CloneOptions{
						URL:      gitRepo,
						ReferenceName: plumbing.NewBranchReferenceName(branch),
						Progress: cmd.OutOrStdout(),
					})
				}

				if err == nil {
					fmt.Println("get the desired git repository", gitRepo)
				} else {
					w.WriteHeader(http.StatusBadRequest)
					fmt.Println(err.Error())
					return
				}

				binaryName := requestPath[strings.LastIndex(requestPath, "/")+1:]

				args := []string{"build"}
				env := []string{
					pairFromQuery(r.URL, "os", "GOOS"),
					pairFromQuery(r.URL, "arch", "GOARCH"),
					"CGO_ENABLE=0",
				}

				if strings.Contains(r.Header.Get("user-agent"), "Macintosh") {
					env = append(env, "GOOS=darwin")
				}

				fmt.Println("start to build", binaryName)
				if err := RunCommandInDir("go", dir, env, args...); err == nil {
					fmt.Println("success", binaryName)
					binaryFilePath := path.Join(dir, binaryName)

					if binaryFileInfo, err := os.Stat(binaryFilePath); err == nil {
						w.Header().Set("Content-Length", fmt.Sprintf("%d", binaryFileInfo.Size()))
						w.Header().Set("Content-Type", "application/octet-stream")
						w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", binaryName))
						w.Header().Set("Content-Transfer-Encoding", "binary")
						w.Header().Set("Expires", "0")
						w.WriteHeader(http.StatusOK)

						if data, err := ioutil.ReadFile(binaryFilePath); err == nil {
							_, _ = w.Write(data)
							return
						}
					} else {
						fmt.Println("cannot get info of file", binaryName)
					}
				} else {
					fmt.Printf("failed to build, error: %v\n", err)
					_, _ = w.Write([]byte(err.Error()))
				}
				w.WriteHeader(http.StatusBadRequest)
			})
			err = http.ListenAndServe("0.0.0.0:7878", nil)
			return nil
		},
	}
	return
}

func getBranch(httpURL *url.URL) (branch string) {
	if branch = httpURL.Query().Get("branch"); branch == "" {
		branch = "master"
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
