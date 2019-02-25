package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GitOpsSpec defines the desired state of GitOps
type GitOpsSpec struct {
	//Repostory is the Git repository to operate on
	Repository string `json:"repository"`
	//Branch is the git branch to operate on
	Branch string `json:"branch",omitempty`

	//RootFolder is the root folder to start parsing yaml templates from in Repository
	RootFolder string `json:"rootFolder"`

	//User to pull from the Repository
	User string `json:"user"`

	//Password to pull from the repository
	Password string `json:"password"`

	//Templating defines if templating shall be done pre deploy
	*Templating `json:"templating,omitempty"`
}

type Templating struct {
	//Enabled
	Enabled bool `json:"enabled"`

	//SourceFolder is the folder containing source data for templating.
	//The folder is skipped from kubernetes yaml parsing
	SourceFolder string `json:"templateDataFolder,omitempty"`
	//Source
	Source *TemplateDataSource `json:"templateDataSource"`

	//Executor used to generate the template data source file
	*Executor `json:"preExecutor"`
}

// TemplateDataSource contains information about the source for template data
type TemplateDataSource struct {
	//TemplateDataFile relative to the root dir of repository
	TemplateDataFile []string `json:"templateDataFile"`
}

type Executor struct {
	Exec string   `json:"exec"`
	Args []string `json:"args"`
}

// GitOpsStatus defines the observed state of GitOps
type GitOpsStatus struct {
	RootFolder string `json:"rootFolder"`
	Updated    string `json:"lastUpdate"`
	Hash       string `json:"gitHash"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitOps is the Schema for the gitops API
// +k8s:openapi-gen=true
type GitOps struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitOpsSpec   `json:"spec,omitempty"`
	Status GitOpsStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitOpsList contains a list of GitOps
type GitOpsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitOps `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitOps{}, &GitOpsList{})
}
