/*


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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WioSpec defines the desired state of Wio
type WioSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// BaseUrl is the server host URL for querying wio nodes
	BaseUrl string `json:"base_url"`

	// SensorID is the grove sensor type
	SensorID string `json:"sensor_id"`

	// SensorPath is the API endpoint to query for data
	SensorPath string `json:"sensor_path"`

	// ResponsePath is the API endpoint to parse the response from
	ResponsePath string `json:"response_path"`

	// Token is the API Token for the Wio server
	Token string `json:"token"`
}

// WioStatus defines the observed state of Wio
type WioStatus struct {
	LastScrapeValue int          `json:"last_scrape_value,omitempty"`
	LastScrapeTime  *metav1.Time `json:"lastScrapeTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Wio is the Schema for the wios API
type Wio struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WioSpec   `json:"spec,omitempty"`
	Status WioStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WioList contains a list of Wio
type WioList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Wio `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Wio{}, &WioList{})
}
