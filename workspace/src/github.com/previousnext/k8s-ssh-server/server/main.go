package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
	"k8s.io/api/core/v1"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/previousnext/k8s-ssh-server/client"
	"github.com/previousnext/k8s-ssh-server/crd"
	"github.com/previousnext/k8s-ssh-server/log"
)

var (
	cliListen = kingpin.Flag("listen", "Port to receive SSH requests").Default(":22").OverrideDefaultFromEnvar("LISTEN").String()
	cliSigner = kingpin.Flag("signer", "Path to signer certificate").OverrideDefaultFromEnvar("SIGNER").String()
)

func main() {
	kingpin.Parse()

	fmt.Println("Installing CRD:", crd.FullCRDName)

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	clientset, err := apiextcs.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	err = crd.Create(clientset)
	if err != nil {
		panic(err)
	}

	crdcs, scheme, err := crd.NewClient(config)
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting SSH Server")

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
		cmd := &v1.PodExecOptions{
			Container: container,
			Stdout:    true,
			Stderr:    true,
			Command:   sess.Command(),
		}
		opts := remotecommand.StreamOptions{
			Stdout: sess,
			Stderr: sess.Stderr(),
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
			opts.Stdin = sess

			cmd.TTY = true
			opts.Tty = true
		}

		// This will handle rsync support eg. stdin for syncing.
		if isRsync(cmd.Command) {
			logger.Print(fmt.Sprintf("Detected rsync mode for: %s", user))
			cmd.Stdin = true
			opts.Stdin = sess

			opts.Tty = false
			cmd.TTY = false
		}

		if cmd.TTY {
			sizeQueue := NewResizeQueue(sess)
			opts.TerminalSizeQueue = sizeQueue
		}

		crdclient := client.Client(crdcs, scheme, namespace)

		exec, err := remotecommand.NewExecutor(config, "POST", crdclient.URL(pod, container, cmd))
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

		crdclient := client.Client(crdcs, scheme, namespace)

		sshUser, err := crdclient.Get(user)
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
