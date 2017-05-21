package main

import (
	"context"
	"fmt"

	"github.com/alecthomas/kingpin"
	"github.com/google/go-github/github"
	sshclient "github.com/previousnext/client"
	"golang.org/x/oauth2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

var (
	cliToken = kingpin.Flag("token", "Github token for authentication").OverrideDefaultFromEnvar("TOKEN").String()
	cliOrg   = kingpin.Flag("org", "Organisation members to sync").OverrideDefaultFromEnvar("ORG").String()
)

func main() {
	kingpin.Parse()

	// Load all the users we will be syncing to Kubernetes.
	users, err := getGithubKeys(*cliToken, *cliOrg)
	if err != nil {
		panic(err)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	sshClient, err := sshclient.NewClient(config)
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
		// Get all the users in this namespace, this will tell us if we need to update or create new.
		existingUsers, err := sshClient.List(namespace.ObjectMeta.Name)
		if err != nil {
			panic(err)
		}

		for _, user := range users {
			if userExists(user, existingUsers) {
				fmt.Printf("Updating user %s in namespace %s", user.Metadata.Name, namespace.ObjectMeta.Name)

				err := sshClient.Put(user)
				if err != nil {
					panic(err)
				}
			} else {
				fmt.Printf("Creating user %s in namespace %s", user.Metadata.Name, namespace.ObjectMeta.Name)

				err := sshClient.Post(user)
				if err != nil {
					panic(err)
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
				Name: *member.Login,
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

func userExists(user sshclient.SshUser, existingUsers sshclient.SshUserList) bool {
	for _, existingUser := range existingUsers.Items {
		if existingUser.Metadata.Name == user.Metadata.Name {
			return true
		}
	}

	return false
}
