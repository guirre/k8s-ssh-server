package client

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SSHUserSpec is used for defining a user.
type SSHUserSpec struct {
	AuthorizedKeys []string `json:"authorizedKeys"`
}

// SSHUser is our high level object being stored in Kubernetes.
type SSHUser struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        metav1.ObjectMeta `json:"metadata"`

	Spec SSHUserSpec `json:"spec"`
}

// SSHUserList is a list of our Kubernetes objects.
type SSHUserList struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        metav1.ListMeta `json:"metadata"`

	Items []SSHUser `json:"items"`
}

// GetObjectKind is required to satisfy Object interface
func (e *SSHUser) GetObjectKind() schema.ObjectKind {
	return &e.TypeMeta
}

// GetObjectMeta is required to satisfy ObjectMetaAccessor interface
func (e *SSHUser) GetObjectMeta() metav1.Object {
	return &e.Metadata
}

// GetObjectKind is required to satisfy Object interface
func (el *SSHUserList) GetObjectKind() schema.ObjectKind {
	return &el.TypeMeta
}

// GetListMeta is required to satisfy ListMetaAccessor interface
func (el *SSHUserList) GetListMeta() metav1.List {
	return &el.Metadata
}

// SSHUserListCopy is used for unmarshalling data.
type SSHUserListCopy SSHUserList

// SSHUserCopy is used for unmarshalling data.
type SSHUserCopy SSHUser

// UnmarshalJSON is used for unmarshalling objects.
func (e *SSHUser) UnmarshalJSON(data []byte) error {
	tmp := SSHUserCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmp2 := SSHUser(tmp)
	*e = tmp2
	return nil
}

// UnmarshalJSON is used for unmarshalling objects.
func (el *SSHUserList) UnmarshalJSON(data []byte) error {
	tmp := SSHUserListCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmp2 := SSHUserList(tmp)
	*el = tmp2
	return nil
}
