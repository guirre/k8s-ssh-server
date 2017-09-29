package main

import (
	"fmt"
	"strings"

	"github.com/gosuri/uitable"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/previousnext/k8s-ssh-server/client"
	"github.com/previousnext/k8s-ssh-server/crd"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	crdcs, scheme, err := crd.NewClient(config)
	if err != nil {
		panic(err)
	}

	list, err := client.Client(crdcs, scheme, "all").List(meta_v1.ListOptions{})
	if err != nil {
		panic(err)
	}

	table := uitable.New()
	table.MaxColWidth = 80

	table.AddRow("NAMESPACE", "NAME", "KEYS")
	for _, key := range list.Items {
		table.AddRow(key.Namespace, key.Name, strings.Join(key.Spec.AuthorizedKeys, "\n"))
	}
	fmt.Println(table)
}
