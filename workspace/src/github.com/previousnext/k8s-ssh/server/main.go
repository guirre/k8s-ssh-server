package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/gliderlabs/ssh"
	promlog "github.com/prometheus/common/log"
	gossh "golang.org/x/crypto/ssh"
	"k8s.io/api/core/v1"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/previousnext/k8s-ssh/client"
	"github.com/previousnext/k8s-ssh/crd"
	"github.com/previousnext/log"
)

var (
	cliListen = kingpin.Flag("listen", "Port to receive SSH requests").Default(":22").OverrideDefaultFromEnvar("SSH_LISTEN").String()
	cliSigner = kingpin.Flag("signer", "Path to signer certificate").OverrideDefaultFromEnvar("SSH_SIGNER").String()
	cliShell  = kingpin.Flag("shell", "Shell type to use if the user requests a Shell session").Default("/bin/bash").OverrideDefaultFromEnvar("SSH_SHELL").String()
	cliK8s    = kingpin.Flag("k8s", "K8s endpoint. If left blank we assume this is 'in cluster' and using K8s native auth").String()
)

func main() {
	kingpin.Parse()

	promlog.Info("Installing CRD:", crd.FullCRDName)

	var (
		config *rest.Config
		err    error
	)

	if *cliK8s != "" {
		config = &rest.Config{
			Host: *cliK8s,
		}
	} else {
		// This must be an deployed to the cluster.
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
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

	promlog.Info("Starting SSH Server")

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
			cmd.Command = []string{
				*cliShell,
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
			promlog.Info("Failed to get namespace, pod and container from user:", err)
			return false
		}

		crdclient := client.Client(crdcs, scheme, namespace)

		sshUser, err := crdclient.Get(user)
		if err != nil {
			promlog.Info("Failed to load the user objects:", err)
			return false
		}

		for _, authorizedKey := range sshUser.Spec.AuthorizedKeys {
			allowed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(authorizedKey))
			if err != nil {
				promlog.Info("Failed to parse key:", err)
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
	if *cliSigner != "" {
		signer, err := getSigner(*cliSigner)
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

// Helper function to generate a signer certificate if one does not exist.
func getSigner(path string) (ssh.Signer, error) {
	var signer ssh.Signer

	if _, err := os.Stat(path); os.IsNotExist(err) {
		key, err := rsa.GenerateKey(rand.Reader, 768)
		if err != nil {
			return signer, err
		}

		priv_der := x509.MarshalPKCS1PrivateKey(key)

		priv_blk := pem.Block{
			Type:    "RSA PRIVATE KEY",
			Headers: nil,
			Bytes:   priv_der,
		}

		err = ioutil.WriteFile(path, pem.EncodeToMemory(&priv_blk), 0644)
		if err != nil {
			return signer, err
		}
	}

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return signer, err
	}

	return gossh.ParsePrivateKey(file)
}
