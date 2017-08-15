package client

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SSHUserSpec is used for defining a user.
type SshUserSpec struct {
	AuthorizedKeys []string `json:"authorizedKeys"`
}

// SSHUser is our high level object being stored in Kubernetes.
type SshUser struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        metav1.ObjectMeta `json:"metadata"`

	Spec SshUserSpec `json:"spec"`
}

// SSHUserList is a list of our Kubernetes objects.
type SshUserList struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        metav1.ListMeta `json:"metadata"`

	Items []SshUser `json:"items"`
}

// GetObjectKind is required to satisfy Object interface
func (e *SshUser) GetObjectKind() schema.ObjectKind {
	return &e.TypeMeta
}

// GetObjectMeta is required to satisfy ObjectMetaAccessor interface
func (e *SshUser) GetObjectMeta() metav1.Object {
	return &e.Metadata
}

// GetObjectKind is required to satisfy Object interface
func (el *SshUserList) GetObjectKind() schema.ObjectKind {
	return &el.TypeMeta
}

// GetListMeta is required to satisfy ListMetaAccessor interface
func (el *SshUserList) GetListMeta() metav1.List {
	return &el.Metadata
}

// SshUserListCopy is used for unmarshalling data.
type SshUserListCopy SshUserList

// SshUserCopy is used for unmarshalling data.
type SshUserCopy SshUser

// UnmarshalJSON is used for unmarshalling objects.
func (e *SshUser) UnmarshalJSON(data []byte) error {
	tmp := SshUserCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmp2 := SshUser(tmp)
	*e = tmp2
	return nil
}

// UnmarshalJSON is used for unmarshalling objects.
func (el *SshUserList) UnmarshalJSON(data []byte) error {
	tmp := SshUserListCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmp2 := SshUserList(tmp)
	*el = tmp2
	return nil
}
