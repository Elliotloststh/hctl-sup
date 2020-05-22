package main

import (
	"fmt"
	"net/url"

	dockerterm "github.com/docker/docker/pkg/term"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
	restclient "k8s.io/client-go/rest"
	remoteclient "k8s.io/client-go/tools/remotecommand"
	pb "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"k8s.io/kubernetes/pkg/kubectl/util/term"
)

const (
	kubeletURLSchema = "http"
	kubeletURLHost   = "http://127.0.0.1:10250"
)

var runtimeExecCommand = cli.Command{
	Name:                   "exec",
	Usage:                  "Run a command in a running container",
	ArgsUsage:              "CONTAINER-ID COMMAND [ARG...]",

	Action: func(context *cli.Context) error {
		if len(context.Args()) < 2 {
			return cli.ShowSubcommandHelp(context)
		}

		if err := getRuntimeClient(context); err != nil {
			return err
		}

		var opts = execOptions{
			id:      context.Args().First(),
			cmd:     context.Args()[1:],
		}

		err := Exec(runtimeClient, opts)
		if err != nil {
			return fmt.Errorf("execing command in container failed: %v", err)
		}
		return nil
	},
	After: closeConnection,
}

// Exec sends an ExecRequest to server, and parses the returned ExecResponse
func Exec(client pb.RuntimeServiceClient, opts execOptions) error {
	request := &pb.ExecRequest{
		ContainerId: opts.id,
		Cmd:         opts.cmd,
		Tty:         true,
		Stdin:       true,
		Stdout:      true,
		Stderr:      false,
	}

	r, err := client.Exec(context.Background(), request)
	if err != nil {
		return err
	}
	execURL := r.Url

	URL, err := url.Parse(execURL)
	if err != nil {
		return err
	}

	if URL.Host == "" {
		URL.Host = kubeletURLHost
	}

	if URL.Scheme == "" {
		URL.Scheme = kubeletURLSchema
	}

	return stream(URL)
}

func stream(url *url.URL) error {
	executor, err := remoteclient.NewSPDYExecutor(&restclient.Config{TLSClientConfig: restclient.TLSClientConfig{Insecure: true}}, "POST", url)
	if err != nil {
		return err
	}

	stdin, stdout, stderr := dockerterm.StdStreams()
	streamOptions := remoteclient.StreamOptions{
		Stdout: stdout,
		Stderr: stderr,
		Tty:    true,
	}
	streamOptions.Stdin = stdin
	t := term.TTY{
		In:  stdin,
		Out: stdout,
		Raw: true,
	}
	if !t.IsTerminalIn() {
		return fmt.Errorf("input is not a terminal")
	}
	streamOptions.TerminalSizeQueue = t.MonitorSize(t.GetSize())
	return t.Safe(func() error { return executor.Stream(streamOptions) })
}
