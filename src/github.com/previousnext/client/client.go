package client

import (
	"net/url"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/api"
)

const (
	API      = "ssh-user"
	Group    = "skpr.io"
	Resource = "sshusers"
)

type Client struct {
	rc *rest.RESTClient
}

// Create a client for SSH User interactions.
func NewClient(config *rest.Config) (Client, error) {
	var c Client

	rc, err := rest.RESTClientFor(sshUserConfig(config))
	if err != nil {
		return c, err
	}

	// Store the rest client for future queries.
	// We will use this client for Get() and List() operations.
	c.rc = rc

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return c, err
	}

	// Create a K8s Third Party Resource object.
	// This allows us to start creating Solr objects under a bespoke, K8s backed, API.
	_, err = clientset.Extensions().ThirdPartyResources().Create(&v1beta1.ThirdPartyResource{
		ObjectMeta: metav1.ObjectMeta{
			Name: API + "." + Group,
		},
		Versions: []v1beta1.APIVersion{
			{Name: "v1"},
		},
		Description: "A SSH User ThirdPartyResource",
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		return c, err
	}

	return c, nil
}

// Returns a single SSH User for a Namespace.
func (c *Client) Get(namespace, name string) (SshUser, error) {
	var s SshUser
	err := c.rc.Get().Resource(Resource).Namespace(namespace).Name(name).Do().Into(&s)
	return s, err
}

// Sets the entire SSH User (Spec + Status) if it exists.
func (c *Client) Put(user SshUser) error {
	return c.rc.Put().Resource(Resource).Namespace(user.Metadata.Namespace).Name(user.Metadata.Name).Body(&user).Do().Error()
}

// Sets the entire SSH User (Spec + Status).
func (c *Client) Post(user SshUser) error {
	return c.rc.Post().Resource(Resource).Namespace(user.Metadata.Namespace).Body(&user).Do().Error()
}

// Returns a list of SSH Users from all namespaces.
func (c *Client) List(namespace string) (SshUserList, error) {
	s := SshUserList{}
	err := c.rc.Get().Resource(Resource).Namespace(namespace).Do().Into(&s)
	if err != nil {
		return s, err
	}
	return s, nil
}

// Returns a list of SSH Users from all namespaces.
func (c *Client) ListAll() (SshUserList, error) {
	s := SshUserList{}
	err := c.rc.Get().Resource(Resource).Do().Into(&s)
	if err != nil {
		return s, err
	}
	return s, nil
}

func (c *Client) Url(namespace, pod, container string, cmd *api.PodExecOptions) *url.URL {
	return c.rc.Post().Resource("pods").Name(pod).Namespace(namespace).SubResource("exec").Param("container", container).VersionedParams(cmd, api.ParameterCodec).URL()
}
