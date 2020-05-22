package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/urfave/cli"
	"google.golang.org/grpc"
	pb "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

const (
	//the truncated length of containerID
	truncatedIDLen = 13
)

var runtimeClient pb.RuntimeServiceClient
var conn *grpc.ClientConn

type hcListOptions struct {
	// pid of container
	pid string
	// all containers
	all bool
	// state of the sandbox
	state string
	// Regular expression pattern to match pod or container
	nameRegexp string
	// out with truncating the id
	noTrunc bool
	// output format
	output string
}

type execOptions struct {
	// id of container
	id string
	// Command to exec
	cmd []string
}

func openFile(path string) (*os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config at %s not found", path)
		}
		return nil, err
	}
	return f, nil
}

func getRuntimeClient(context *cli.Context) error {
	// Set up a connection to the server.
	var err error
	conn, err = getRuntimeClientConnection(context)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	runtimeClient = pb.NewRuntimeServiceClient(conn)
	return nil
}

func closeConnection(context *cli.Context) error {
	if conn == nil {
		return nil
	}

	return conn.Close()
}

func getTruncatedID(id, prefix string) string {
	id = strings.TrimPrefix(id, prefix)
	if len(id) > truncatedIDLen {
		id = id[:truncatedIDLen]
	}
	return id
}

func matchesRegex(pattern, target string) bool {
	if pattern == "" {
		return true
	}
	matched, err := regexp.MatchString(pattern, target)
	if err != nil {
		return false
	}
	return matched
}

func convertContainerState(state pb.ContainerState) string {
	switch state {
	case pb.ContainerState_CONTAINER_CREATED:
		return "Created"
	case pb.ContainerState_CONTAINER_RUNNING:
		return "Running"
	case pb.ContainerState_CONTAINER_EXITED:
		return "Exited"
	case pb.ContainerState_CONTAINER_UNKNOWN:
		return "Unknown"
	default:
		log.Fatalf("Unknown container state %q", state)
		return ""
	}
}