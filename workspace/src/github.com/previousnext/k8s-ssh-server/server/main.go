package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned/remotecommand"
	remotecommandserver "k8s.io/kubernetes/pkg/kubelet/server/remotecommand"

	sshclient "github.com/previousnext/k8s-ssh-server/client"
	"github.com/previousnext/k8s-ssh-server/log"
)

const separator = "~"

var (
	cliListen = kingpin.Flag("listen", "Port to receive SSH requests").Default(":22").OverrideDefaultFromEnvar("LISTEN").String()
	cliSigner = kingpin.Flag("signer", "Path to signer certificate").OverrideDefaultFromEnvar("SIGNER").String()
)

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

	srv := &ssh.Server{
		Addr: *cliListen,
	}

	ssh.Handle(func(sess ssh.Session) {
		// Generate a unique ID for this request.
		// This will be used for logging connections.
		logger := log.New()

		namespace, pod, container, user, err := splitUser(sess.User())
		if err != nil {
			logger.Print(fmt.Sprintf("Failed to get namespace, pod and container from user: %s", user))

			// Return the error code output to the end user so they can see why the request failed.
			io.WriteString(sess, err.Error())

			// This will send an error code back to the SSH client.
			sess.Exit(1)

			return
		}

		logger.Print(fmt.Sprintf("Starting connection for user: %s", user))

		// These are default options which will be sent to the Kubernetes API.
		cmd := &api.PodExecOptions{
			Container: container,
			Stdout:    true,
			Stderr:    true,
			Command:   sess.Command(),
		}
		opts := remotecommand.StreamOptions{
			SupportedProtocols: remotecommandserver.SupportedStreamingProtocols,
			Stdout:             sess,
			Stderr:             sess.Stderr(),
		}

		// This will handle "shell" calls.
		if isShell(cmd.Command) {
			logger.Print(fmt.Sprintf("Detected SHELL for: %s", user))

			// Provide a fully featured shell from the remote environment.
			// @todo, This may need to change if we cut down our images.
			cmd.Command = []string{
				"/bin/bash",
			}

			cmd.Stdin = true
			cmd.TTY = true
			opts.Stdin = sess
			opts.Tty = true
		}

		// This will handle rsync support eg. stdin for syncing.
		if isRsync(cmd.Command) {
			logger.Print(fmt.Sprintf("Detected rsync mode for: %s", user))
			cmd.Stdin = true
			cmd.TTY = false
			opts.Stdin = sess
			opts.Tty = false
		}

		if cmd.TTY {
			sizeQueue := NewResizeQueue(sess)
			opts.TerminalSizeQueue = sizeQueue
		}

		exec, err := remotecommand.NewExecutor(config, "POST", sshClient.Url(namespace, pod, container, cmd))
		if err != nil {
			logger.Print(fmt.Sprintf("Failed to run command '%s' as %s: %s", strings.Join(cmd.Command, " "), user, err.Error()))

			// Return the error code output to the end user so they can see why the request failed.
			io.WriteString(sess, err.Error())

			// This will send an error code back to the SSH client.
			sess.Exit(1)

			return
		}

		logger.Print(fmt.Sprintf("Executing command '%s'", strings.Join(cmd.Command, " ")))

		err = exec.Stream(opts)
		if err != nil {
			logger.Print(fmt.Sprintf("Failed to stream command '%s' as %s: %s", strings.Join(cmd.Command, " "), user, err.Error()))

			// Return the error code output to the end user so they can see why the request failed.
			io.WriteString(sess, err.Error())

			// This will send an error code back to the SSH client.
			sess.Exit(1)

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

		for _, authorizedKey := range sshUser.Spec.AuthorizedKeys {
			allowed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(authorizedKey))
			if err != nil {
				fmt.Println("Failed to parse key:", err)
				return false
			}

			if ssh.KeysEqual(key, allowed) {
				return true
			}
		}

		return false
	})
	srv.SetOption(publicKeyHandler)

	// Check if a signer was provided, if one was, load it and add to the server.
	if cliSigner != nil {
		file, err := ioutil.ReadFile(*cliSigner)
		if err != nil {
			panic(err)
		}

		signer, err := gossh.ParsePrivateKey(file)
		if err != nil {
			panic(err)
		}

		srv.HostSigners = append(srv.HostSigners, signer)
	}

	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
