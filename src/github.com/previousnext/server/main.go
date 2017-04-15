package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/gliderlabs/ssh"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned/remotecommand"
	remotecommandserver "k8s.io/kubernetes/pkg/kubelet/server/remotecommand"

	sshclient "github.com/previousnext/client"
	"github.com/previousnext/log"
)

const separator = "~"

var cliPort = kingpin.Flag("port", "Port to receive SSH requests").Default("22").OverrideDefaultFromEnvar("PORT").String()

func main() {
	kingpin.Parse()

	fmt.Println("Starting SSH Server")

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	sshClient, err := sshclient.NewClient(config)
	if err != nil {
		panic(err)
	}

	ssh.Handle(func(s ssh.Session) {
		// Generate a unique ID for this request.
		// This will be used for logging connections.
		logger := log.New()

		_, namespace, pod, container, err := splitUser(s.User())
		if err != nil {
			logger.Print(fmt.Sprintf("Failed to get namespace, pod and container from user: %s", s.User()))

			// Return the error code output to the end user so they can see why the request failed.
			io.WriteString(s, err.Error())

			// This will send an error code back to the SSH client.
			s.Exit(1)

			return
		}

		logger.Print(fmt.Sprintf("Starting connection for user: %s", s.User()))

		// These are default options which will be sent to the Kubernetes API.
		cmd := &api.PodExecOptions{
			Container: container,
			Stdout:    true,
			Stderr:    true,
			Command:   s.Command(),
		}
		opts := remotecommand.StreamOptions{
			SupportedProtocols: remotecommandserver.SupportedStreamingProtocols,
			Stdout:             s,
			Stderr:             s,
		}

		// This will handle "shell" calls.
		if len(cmd.Command) == 0 {
			logger.Print(fmt.Sprintf("Detected SHELL for:", s.User()))

			// Provide a fully featured shell from the remote environment.
			// @todo, This may need to change if we cut down our images.
			cmd.Command = []string{
				"/bin/bash",
			}

			// Allow TTY.
			cmd.TTY = true
			opts.Tty = true

			// Allow Input from the users CLI.
			cmd.Stdin = true
			opts.Stdin = s
		}

		// This will handle rsync support eg. stdin for syncing.
		if cmd.Command[0] == "rsync" {
			logger.Print(fmt.Sprintf("Detected rsync mode for:", s.User()))
			cmd.Stdin = true
			opts.Stdin = s
		}

		exec, err := remotecommand.NewExecutor(config, "POST", sshClient.Url(namespace, pod, container, cmd))
		if err != nil {
			logger.Print(fmt.Sprintf("Failed to run command - %s - '%s' - %s", s.User(), strings.Join(cmd.Command, " "), err.Error()))

			// Return the error code output to the end user so they can see why the request failed.
			io.WriteString(s, err.Error())

			// This will send an error code back to the SSH client.
			s.Exit(1)

			return
		}

		fmt.Println("Executing command:", s.User(), ":", strings.Join(cmd.Command, " "))

		err = exec.Stream(opts)
		if err != nil {
			logger.Print(fmt.Sprintf("Failed to stream command - %s - '%s' - %s", s.User(), strings.Join(cmd.Command, " "), err.Error()))

			// Return the error code output to the end user so they can see why the request failed.
			io.WriteString(s, err.Error())

			// This will send an error code back to the SSH client.
			s.Exit(1)

			return
		}
	})

	publicKeyHandler := ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		namespace, _, _, user, err := splitUser(ctx.User())
		if err != nil {
			fmt.Println("Failed to get namespace, pod and container from user:", err)
			return false
		}

		sshUser, err := sshClient.Get(namespace, user)
		if err != nil {
			fmt.Println("Failed to load the user objects:", err)
			return false
		}

		allowed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(sshUser.Spec.AuthorizedKey))
		if err != nil {
			fmt.Println("Failed to parse key:", err)
			return false
		}

		return ssh.KeysEqual(key, allowed)
	})

	err = ssh.ListenAndServe(":"+*cliPort, nil, publicKeyHandler)
	if err != nil {
		panic(err)
	}
}
