package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/google/go-github/github"
	sshclient "github.com/previousnext/k8s-ssh-server/client"
	"golang.org/x/oauth2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

var (
	cliToken     = kingpin.Flag("token", "Github token for authentication").OverrideDefaultFromEnvar("TOKEN").String()
	cliOrg       = kingpin.Flag("org", "Organisation members to sync").OverrideDefaultFromEnvar("ORG").String()
	cliExclude   = kingpin.Flag("exclude", "A list of namespaces to skip").Default("kube-system,kube-public").OverrideDefaultFromEnvar("EXCLUDE").String()
	cliFrequency = kingpin.Flag("frequency", "How often to sync Github users").Default("120s").OverrideDefaultFromEnvar("FREQUENCY").Duration()
)

func main() {
	kingpin.Parse()

	limiter := time.Tick(*cliFrequency)

	for {
		<-limiter

		// Load all the users we will be syncing to Kubernetes.
		users, err := getGithubKeys(*cliToken, *cliOrg)
		if err != nil {
			// If Github is down, wait until the next loop.
			fmt.Println("Failed to lookup Github keys:", err)
			continue
		}

		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err)
		}

		sshClient, err := sshclient.New(config)
		if err != nil {
			panic(err)
		}

		clientset, err := client.NewForConfig(config)
		if err != nil {
			panic(err)
		}

		namespaces, err := clientset.Namespaces().List(metav1.ListOptions{})
		if err != nil {
			panic(err)
		}

		for _, namespace := range namespaces.Items {
			// Check if we need to skip this namespace.
			if contains(strings.Split(*cliExclude, ","), namespace.ObjectMeta.Name) {
				fmt.Println("Skipping namespace:", namespace.ObjectMeta.Name)
				continue
			}

			// Get all the users in this namespace, this will tell us if we need to update or create new.
			existingUsers, err := sshClient.List(namespace.ObjectMeta.Name)
			if err != nil {
				panic(err)
			}

			// Delete our the old users.
			for _, existingUser := range existingUsers.Items {
				if !userExists(existingUser, users) {
					fmt.Printf("Deleting user %s in namespace %s\n", namespace.ObjectMeta.Name, existingUser.Metadata.Name)

					err := sshClient.Delete(namespace.ObjectMeta.Name, existingUser.Metadata.Name)
					if err != nil {
						panic(err)
					}
				}
			}

			// Add in the new ones.
			for _, user := range users {
				user.Metadata.Namespace = namespace.ObjectMeta.Name

				if userExists(user, existingUsers.Items) {
					fmt.Printf("Updating user %s in namespace %s\n", user.Metadata.Name, namespace.ObjectMeta.Name)

					err := sshClient.Put(user)
					if err != nil {
						panic(err)
					}
				} else {
					fmt.Printf("Creating user %s in namespace %s\n", user.Metadata.Name, namespace.ObjectMeta.Name)

					err := sshClient.Post(user)
					if err != nil {
						panic(err)
					}
				}
			}
		}
	}
}

func getGithubKeys(token, org string) ([]sshclient.SshUser, error) {
	var users []sshclient.SshUser

	gh := github.NewClient(oauth2.NewClient(oauth2.NoContext, oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token,
		},
	)))

	members, _, err := gh.Organizations.ListMembers(context.Background(), org, &github.ListMembersOptions{})
	if err != nil {
		return users, err
	}

	// Loop over the members, look up their ssh keys and add to all namespaces.
	for _, member := range members {
		user := sshclient.SshUser{
			Metadata: metav1.ObjectMeta{
				Name: strings.ToLower(*member.Login),
			},
		}

		keys, _, err := gh.Users.ListKeys(context.Background(), *member.Login, &github.ListOptions{})
		if err != nil {
			return users, err
		}

		for _, key := range keys {
			user.Spec.AuthorizedKeys = append(user.Spec.AuthorizedKeys, *key.Key)
		}

		users = append(users, user)
	}

	return users, nil
}

func userExists(user sshclient.SshUser, existingUsers []sshclient.SshUser) bool {
	for _, existingUser := range existingUsers {
		if existingUser.Metadata.Name == user.Metadata.Name {
			return true
		}
	}

	return false
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}
