/*
Copyright 2024.

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

// LayerSpec defines the desired state of Layer.
type LayerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Layers can be ordered into tree topology
	// Layers at the same node-level - are alternates
	// if unspecified, the layer is a child of the root layer
	Parent string `json:"parent,omitempty"`
}

// Important: Run "make" to regenerate code after modifying this file

// LayerStatus defines the observed state of Layer.
type LayerStatus struct {
	// Current state of the layer
	// TODO insert known FSM's,
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// Layer is the Schema for the layers API.
type Layer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LayerSpec   `json:"spec,omitempty"`
	Status LayerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LayerList contains a list of Layer.
type LayerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Layer `json:"items"`
}

// LayerServiceSpec defines the desired state of Layer.
type LayerServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Reference to the layer must be defined.
	// +kubebuilder:validation:Required
	Layer string `json:"layer"`
	// Host - is the name of the service to route on the basis.
	// +kubebuilder:validation:Required
	Host string `json:"host"`
	// Labels optional labels to defined
	// When defined they will be used to create an istio DestinationRule with subsets
	// -- see https://istio.io/latest/docs/reference/config/networking/virtual-service/#Destination
	Labels map[string]string `json:"labels,omitempty"`
	// Destination - optional destination (must be different from the host)
	// Either Destination or Labels must be specified.
	Destination string `json:"destination,omitempty"`
}

type LayerServiceStatus struct {
	// Current state of the layerservice
	// TODO insert known FSM's,
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced

// LayerService is the Schema for the layers API.
type LayerService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LayerSpec   `json:"spec,omitempty"`
	Status LayerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LayerServiceList contains a list of Layer.
type LayerServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Layer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Layer{}, &LayerList{})
	SchemeBuilder.Register(&LayerService{}, &LayerServiceList{})
}
