package client

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type SshUserSpec struct {
	AuthorizedKeys []string `json:"authorizedKeys"`
}

type SshUser struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        metav1.ObjectMeta `json:"metadata"`

	Spec SshUserSpec `json:"spec"`
}

type SshUserList struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        metav1.ListMeta `json:"metadata"`

	Items []SshUser `json:"items"`
}

// Required to satisfy Object interface
func (e *SshUser) GetObjectKind() schema.ObjectKind {
	return &e.TypeMeta
}

// Required to satisfy ObjectMetaAccessor interface
func (e *SshUser) GetObjectMeta() metav1.Object {
	return &e.Metadata
}

// Required to satisfy Object interface
func (el *SshUserList) GetObjectKind() schema.ObjectKind {
	return &el.TypeMeta
}

// Required to satisfy ListMetaAccessor interface
func (el *SshUserList) GetListMeta() metav1.List {
	return &el.Metadata
}

// The code below is used only to work around a known problem with third-party
// resources and ugorji. If/when these issues are resolved, the code below
// should no longer be required.

type SshUserListCopy SshUserList
type SshUserCopy SshUser

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
