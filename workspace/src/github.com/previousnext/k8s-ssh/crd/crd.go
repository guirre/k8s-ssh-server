package crd

import (
	"reflect"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

const (
	Plural      string = "sshusers"
	Group       string = "skpr.io"
	Version     string = "v1"
	FullCRDName string = Plural + "." + Group
)

// Create the CRD resource, ignore error if it already exists
func Create(clientset apiextcs.Interface) error {
	definition := &v1beta1.CustomResourceDefinition{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: FullCRDName,
		},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group:   Group,
			Version: Version,
			Scope:   v1beta1.NamespaceScoped,
			Names: v1beta1.CustomResourceDefinitionNames{
				Plural: Plural,
				Kind:   reflect.TypeOf(SshUser{}).Name(),
			},
		},
	}

	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(definition)
	if err != nil && apierrors.IsAlreadyExists(err) {
		return nil
	}

	return err
}

// Definition of our CRD Example class
type SshUser struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               SshUserSpec `json:"spec"`
}
type SshUserSpec struct {
	Groups         []string `json:"groups"`
	AuthorizedKeys []string `json:"authorizedKeys"`
}

type SshUserList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []SshUser `json:"items"`
}

var SchemeGroupVersion = schema.GroupVersion{
	Group:   Group,
	Version: Version,
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&SshUser{},
		&SshUserList{},
	)

	meta_v1.AddToGroupVersion(scheme, SchemeGroupVersion)

	return nil
}

func NewClient(cfg *rest.Config) (*rest.RESTClient, *runtime.Scheme, error) {
	var (
		scheme        = runtime.NewScheme()
		SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	)

	err := SchemeBuilder.AddToScheme(scheme)
	if err != nil {
		return nil, nil, err
	}

	config := *cfg
	config.GroupVersion = &SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{
		CodecFactory: serializer.NewCodecFactory(scheme),
	}

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, nil, err
	}

	return client, scheme, nil
}
