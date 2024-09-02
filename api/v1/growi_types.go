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

// GrowiSpec defines the desired state of Growi
type GrowiSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//+kubebuilder:default:="growi-app"
	//+kubebuilder:validation:Optional
	Growi_app_namespace string `json:"growi_app_namespace"`

	//+kubebuilder:default:="7"
	//+kubebuilder:validation:Optional
	Growi_version string `json:"growi_version"`

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:default:=1
	//+kubebuilder:validation:Optional
	Growi_replicas int32 `json:"growi_replicas"`

	//+kubebuilder:default:="standard"
	//+kubebuilder:validation:Required
	Growi_storageClass string `json:"growi_storageclass"`

	//+kubebuilder:default:="10Gi"
	//+kubebuilder:validation:Optional
	Growi_storageRequest string `json:"growi_storagerequest"`

	//+kubebuilder:default:="growi-mongodb"
	//+kubebuilder:validation:Optional
	Mongo_db_namespace string `json:"mongo_db_namespace"`

	//+kubebuilder:default:="6.0"
	//+kubebuilder:validation:Optional
	Mongo_db_version string `json:"mongo_db_version"`

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:default:=1
	//+kubebuilder:validation:Optional
	Mongo_db_replicas int32 `json:"mongo_db_replicas"`

	//+kubebuilder:default:="10Gi"
	//+kubebuilder:validation:Optional
	Mongo_db_storageRequest string `json:"mongo_db_storagerequest"`

	//+kubebuilder:default:="standard"
	//+kubebuilder:validation:Required
	Mongo_db_storageClass string `json:"mongo_db_storageclass"`

	//+kubebuilder:default:="growi-es"
	//+kubebuilder:validation:Optional
	Elasticsearch_namespace string `json:"elasticsearch_namespace"`

	//+kubebuilder:default:="8.7.0"
	//+kubebuilder:validation:Optional
	Elasticsearch_version string `json:"elasticsearch_version"`

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:default:=1
	//+kubebuilder:validation:Optional
	Elasticsearch_replicas int32 `json:"elasticsearch_replicas"`

	//+kubebuilder:default:="standard"
	//+kubebuilder:validation:Required
	Elasticsearch_storageClass string `json:"elasticsearch_storageclass"`

	//+kubebuilder:default:="10Gi"
	//+kubebuilder:validation:Optional
	Elasticsearch_storageRequest string `json:"elasticsearch_storagerequest"`
}

// GrowiStatus defines the observed state of Growi
type GrowiStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Growi is the Schema for the growis API
type Growi struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrowiSpec   `json:"spec,omitempty"`
	Status GrowiStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GrowiList contains a list of Growi
type GrowiList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Growi `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Growi{}, &GrowiList{})
}
