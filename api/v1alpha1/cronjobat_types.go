/*
Copyright 2022 ByteGopher.

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
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConcurrencyPolicy describe how the job will be handled.
// Only one of the following concurrent policies may be specified.
// If none of the following policies is specified, the default one is
// AllowConcurrent.
// +kubebuilder:validation:Enum=Allow;Forbid;Replace
type ConcurrencyPolicy string

const (
	// AllowConcurrent allows CronJobs to run concurrently
	AllowConcurrent ConcurrencyPolicy = "Allow"

	// ForbidConcurrent forbids concurrent runs, skipping next run if previous
	// hasn't finished yet.
	ForbidConcurrent ConcurrencyPolicy = "Forbid"

	// ReplaceConcurrent cancels currently running job and replaces it with a new one.
	ReplaceConcurrent ConcurrencyPolicy = "Replace"
)

// CronJobAtSpec defines the desired state of CronJobAt
type CronJobAtSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of CronJobAt. Edit cronjobat_types.go to remove/update

	//+kubebuilder:validation:MinLength=0

	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron
	Schedule string `json:"schedule"`

	//+kubebuilder:validation:Minimum=0

	//Optional deadline in seconds for starting the job if it misses scheduled time
	// for any reason. Missed jobs executions will be counted as failed ones.
	//+optional
	StartingDeadlineSeconds *int64 `json:"startingDeadlineSeconds,omitempty"`

	//Specified how to treat concurrent executions of a Job.
	// Valid values are:
	// - "Allow" (default): allows CronJobs to run concurrently
	// - "Forbid": forbids concurrent runs, skipping next run if previous run hasn't finished yet;
	// - "Replace": cancels currently running job and replaces it with a new one
	//+optional
	ConcurrencyPolicy ConsurrencyPolicy `json:"concurrencyPolicy,omitempty"`

	// This flag tells the controller to suspend subsequent executions, it does not
	// apply to already started executions. Default to false.
	// +optional
	Suspend *bool `json:"suspend,omitempty`

	// Specifies the job that will be created when executing a CronJob
	JobTemplate batchv1beta1.JobTemplateSpec `json:"jobTemplate"`

	//+kubebuilder: validation:Minimum=0
	// The number of successful finished jobs to retain.
	// This is a pointer to distinguish between explicit zero and not specified.
	// +optional
	SuccessfulJobHistoryLimit *int32 `json:"successfulJobHistoryLimit,omitempty"`

	//+kubebuilder:validation:Minimum=0

	// The number of failed finished jobs to retain.
	// This is a pointer to distinguish between explicit zero and not specified.
	// +optional
	FailedJobHistoryLimit *int32 `json:"failedJobHistoryLimit,omitempty"`
}

// CronJobAtStatus defines the observed state of CronJobAt
type CronJobAtStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// A list of pointers to currently running jobs.
	// +optional
	Active []corev1.ObjectReference `json:"active,omitempty"`

	// Information when was the last time the job was successfully scheduled.
	// +optional
	LastScheduleTime *metav1.Time `json:"LastScheduleTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CronJobAt is the Schema for the cronjobats API
type CronJobAt struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronJobAtSpec   `json:"spec,omitempty"`
	Status CronJobAtStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CronJobAtList contains a list of CronJobAt
type CronJobAtList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CronJobAt `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CronJobAt{}, &CronJobAtList{})
}
