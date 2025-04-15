/*
Copyright 2025.

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

// GrowiAppSpec difines the desired state of GrowiApp.
type GrowiAppSpec struct {
	// Version is the version of GrowiApp
	// +kubebuilder:default="latest"
	// +optional
	Version string `json:"version,omitempty"`

	// Repicas is the number of GrowiApp.
	// +kubebuilder:default=1
	// +kubebuilder:validation:MinPropates=1
	// +optional
	Repicas int `json:"replicas,omitempty"`
}

// MongoDBSpec difines the desired state of MongoDB.
type MongoDBSpec struct {
	// Version is the version of MongoDB.
	// +kubebuilder:default="6.0"
	// +optional
	Version string `json:"version,omitempty"`

	// Repicas is the number of MongoDB.
	// +kubebuilder:default=1
	// +kubebuilder:validation:MinPropates=1
	// +optional
	Repicas int `json:"replicas,omitempty"`
}

// ElasticSearchSpec difines the desired state of ElasticSearch.
type ElasticSearchSpec struct {
	// Version is the version of ElasticSearch,
	// +kubebuilder:default="8.7.0"
	// +optional
	Version string `json:"version,omitempty"`

	// Repicas is the number of ElasticSearch.
	// +kubebuilder:default=1
	// +kubebuilder:validation:MinPropates=1
	// +optional
	Repicas int `json:"replicas,omitempty"`
}

// GrowiSpec defines the desired state of Growi.
type GrowiSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// StorageClass for store data.
	// +kubebuilder:validation:Required
	StorageClass string `json:"storageclass"`

	// GrowiAppSpec is the desired state of GrowiApp.
	// +kubebuilder:validation:Required
	GrowiAppSpec GrowiAppSpec `json:"growiappspec"`

	// MongoDBSpec is the desired state of MongoDB.
	// +kubebuilder:validation:Required
	MongoDBSpec MongoDBSpec `json:"mongodbspec"`

	// ElasticSearchSpec is the desired state of ElasticSearch.
	// +kubebuilder:validation:Required
	ElasticSearchSpec ElasticSearchSpec `json:"elasticsearchspec"`
}

// GrowiStatus defines the observed state of Growi.
type GrowiStatus struct {
	// Conditions represent the latest available observations of an object's state
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// CurrentCondition stores the type of the condition whose status is currently True (for display purposes).
	// +optional
	CurrentCondition string `json:"currentcondition,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=.status.currentcondition

// Growi is the Schema for the growis API.
type Growi struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrowiSpec   `json:"spec,omitempty"`
	Status GrowiStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GrowiList contains a list of Growi.
type GrowiList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Growi `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Growi{}, &GrowiList{})
}
