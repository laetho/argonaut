/*
Copyright 2021 The Argonaut authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ArgonautSpec defines the desired state of Argonaut
type ArgonautSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Argonaut. Edit argonaut_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// ArgonautStatus defines the observed state of Argonaut
type ArgonautStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Argonaut is the Schema for the argonauts API
type Argonaut struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArgonautSpec   `json:"spec,omitempty"`
	Status ArgonautStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ArgonautList contains a list of Argonaut
type ArgonautList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Argonaut `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Argonaut{}, &ArgonautList{})
}
