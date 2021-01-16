/*
Copyright 2021.

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

// DeltaSpec defines the desired state of Delta
type DeltaSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Specifies the image and pull repository.
	Image string `json:"image,omitempty"`
	// Config contains the core pipeline configs.
	Config string `json:"config,omitempty"`
	// PodAnnotations are applied to the generated pods.
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`
	// ConfigMaps gives users the chance to mount a config map into the pods.
	ConfigMaps map[string]string `json:"configMaps,omitempty"`
	// Secrets gives users the chance to mount a secret into the pods.
	Secrets map[string]string `json:"secrets,omitempty"`
}
// DeltaStatus defines the observed state of Delta
type DeltaStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Delta is the Schema for the delta API
type Delta struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeltaSpec   `json:"spec,omitempty"`
	Status DeltaStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DeltaList contains a list of Delta
type DeltaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Delta `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Delta{}, &DeltaList{})
}
