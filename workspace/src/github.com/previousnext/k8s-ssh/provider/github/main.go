package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/previousnext/k8s-ssh/client"
	"github.com/previousnext/k8s-ssh/crd"
)

var (
	cliToken      = kingpin.Flag("token", "Github token for authentication").OverrideDefaultFromEnvar("TOKEN").String()
	cliOrg        = kingpin.Flag("org", "Organisation members to sync").OverrideDefaultFromEnvar("ORG").String()
	cliExclude    = kingpin.Flag("exclude", "A list of namespaces to skip").Default("kube-system,kube-public").OverrideDefaultFromEnvar("EXCLUDE").String()
	cliFrequency  = kingpin.Flag("frequency", "How often to sync Github users").Default("120s").OverrideDefaultFromEnvar("FREQUENCY").Duration()
	cliNamespaces = kingpin.Flag("namespaces", "Comma separated list of namespaces to sync keys to").Default("default").OverrideDefaultFromEnvar("NAMESPACES").String()
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

		for _, namespace := range strings.Split(*cliNamespaces, ",") {
			crdclient := client.Client(crdcs, scheme, namespace)

			// Check if we need to skip this namespace.
			if contains(strings.Split(*cliExclude, ","), namespace) {
				fmt.Println("Skipping namespace:", namespace)
				continue
			}

			// Get all the users in this namespace, this will tell us if we need to update or create new.
			existingUsers, err := crdclient.List(meta_v1.ListOptions{})
			if err != nil {
				panic(err)
			}

			// Delete our the old users.
			for _, existingUser := range existingUsers.Items {
				if !userExists(existingUser, users) {
					fmt.Printf("Deleting user %s in namespace %s\n", namespace, existingUser.Name)

					err := crdclient.Delete(existingUser.Name, &meta_v1.DeleteOptions{})
					if err != nil {
						panic(err)
					}
				}
			}

			// Add in the new ones.
			for _, user := range users {
				user.Namespace = namespace

				if userExists(user, existingUsers.Items) {
					fmt.Printf("Updating user %s in namespace %s\n", user, namespace)

					_, err := crdclient.Update(&user)
					if err != nil {
						panic(err)
					}
				} else {
					fmt.Printf("Creating user %s in namespace %s\n", user, namespace)

					_, err := crdclient.Create(&user)
					if err != nil {
						panic(err)
					}
				}
			}
		}
	}
}

func getGithubKeys(token, org string) ([]crd.SSH, error) {
	var users []crd.SSH

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
		user := crd.SSH{
			ObjectMeta: metav1.ObjectMeta{
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

func userExists(user crd.SSH, existingUsers []crd.SSH) bool {
	for _, existingUser := range existingUsers {
		if existingUser.Name == user.Name {
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
