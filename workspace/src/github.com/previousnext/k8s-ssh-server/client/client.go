package client

import (
	"net/url"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

const (
	apiName  = "ssh-user"
	apiGroup = "skpr.io"
	resource = "sshusers"
)

// Client used for interacting with the "k8s-ssh-server" third party resource.
type Client struct {
	rc *rest.RESTClient
	cs *client.Clientset
}

// New is used for setting up the Third Party Resource and returning a client.
func New(config *rest.Config) (Client, error) {
	var c Client

	rc, err := rest.RESTClientFor(sshUserConfig(config))
	if err != nil {
		return c, err
	}

	// Store the rest client for future queries.
	// We will use this client for Get() and List() operations.
	c.rc = rc

	clientset, err := client.NewForConfig(config)
	if err != nil {
		return c, err
	}

	// This client will allow us to perform Kubernetes operations.
	c.cs = clientset

	// Create a K8s Third Party Resource object.
	// This allows us to start creating Solr objects under a bespoke, K8s backed, API.
	_, err = clientset.Extensions().ThirdPartyResources().Create(&v1beta1.ThirdPartyResource{
		ObjectMeta: metav1.ObjectMeta{
			Name: apiName + "." + apiGroup,
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

// Get returns a single SSH User for a Namespace.
func (c *Client) Get(namespace, name string) (SSHUser, error) {
	var s SSHUser
	err := c.rc.Get().Resource(resource).Namespace(namespace).Name(name).Do().Into(&s)
	return s, err
}

// Put sets the entire SSH User (Spec + Status) if it exists.
func (c *Client) Put(user SSHUser) error {
	return c.rc.Put().Resource(resource).Namespace(user.Metadata.Namespace).Name(user.Metadata.Name).Body(&user).Do().Error()
}

// Post sets the entire SSH User (Spec + Status).
func (c *Client) Post(user SSHUser) error {
	return c.rc.Post().Resource(resource).Namespace(user.Metadata.Namespace).Body(&user).Do().Error()
}

// Delete the entire SSH User.
func (c *Client) Delete(namespace, name string) error {
	return c.rc.Delete().Resource(resource).Namespace(namespace).Name(name).Do().Error()
}

// List returns a list of SSH Users from all namespaces.
func (c *Client) List(namespace string) (SSHUserList, error) {
	s := SSHUserList{}
	err := c.rc.Get().Resource(resource).Namespace(namespace).Do().Into(&s)
	if err != nil {
		return s, err
	}
	return s, nil
}

// ListAll returns a list of SSH Users from all namespaces.
func (c *Client) ListAll() (SSHUserList, error) {
	s := SSHUserList{}
	err := c.rc.Get().Resource(resource).Do().Into(&s)
	if err != nil {
		return s, err
	}
	return s, nil
}

// URL returns a url to the resource.
func (c *Client) URL(namespace, pod, container string, cmd *api.PodExecOptions) *url.URL {
	return c.cs.Core().RESTClient().Post().Resource("pods").Name(pod).Namespace(namespace).SubResource("exec").Param("container", container).VersionedParams(cmd, api.ParameterCodec).URL()
}
