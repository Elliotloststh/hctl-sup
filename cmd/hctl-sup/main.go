package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	internalapi "k8s.io/cri-api/pkg/apis"
	"k8s.io/kubernetes/pkg/kubelet/remote"
	"k8s.io/kubernetes/pkg/kubelet/util"
)

const (
	//defaultConfigPath      = "/etc/hctl-sup.yaml"
	defaultRuntimeEndpoint = "unix:///var/run/crio/crio.sock"
	defaultTimeout = 2 * time.Second
)

var (
	// RuntimeEndpoint is CRI server runtime endpoint
	RuntimeEndpoint string
	// Timeout  of connecting to server (default: 2s)
	Timeout time.Duration
)

func getRuntimeClientConnection(context *cli.Context) (*grpc.ClientConn, error) {
	if RuntimeEndpoint == "" {
		return nil, fmt.Errorf("--runtime-endpoint is not set")
	}

	addr, dialer, err := util.GetAddressAndDialer(RuntimeEndpoint)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(Timeout), grpc.WithDialer(dialer))
	if err != nil {
		return nil, fmt.Errorf("failed to connect, make sure you are running as root and the runtime has been started: %v", err)
	}
	return conn, nil
}

func getRuntimeService(context *cli.Context) (internalapi.RuntimeService, error) {
	return remote.NewRemoteRuntimeService(RuntimeEndpoint, Timeout)
}

func main() {
	app := cli.NewApp()
	app.Name = "hctl-sup"
	app.Usage = "support for hctl"

	app.Commands = []cli.Command{
		runtimeExecCommand,
		hcListCommand,
	}

	app.Flags = []cli.Flag{
		cli.DurationFlag{
			Name:  "timeout, t",
			Value: defaultTimeout,
			Usage: "Timeout of connecting to the server",
		},
	}

	app.Before = func(context *cli.Context) error {
		//TODO make it configurable
		RuntimeEndpoint = defaultRuntimeEndpoint
		Timeout = context.GlobalDuration("timeout")
		return nil
	}

	for _, cmd := range app.Commands {
		sort.Sort(cli.FlagsByName(cmd.Flags))
	}
	sort.Sort(cli.FlagsByName(app.Flags))

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
