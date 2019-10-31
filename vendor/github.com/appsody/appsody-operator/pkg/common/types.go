package common

import (
	prometheusv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatusConditionType ...
type StatusConditionType string

// StatusCondition ...
type StatusCondition interface {
	GetLastTransitionTime() *metav1.Time
	SetLastTransitionTime(*metav1.Time)

	GetLastUpdateTime() metav1.Time
	SetLastUpdateTime(metav1.Time)

	GetReason() string
	SetReason(string)

	GetMessage() string
	SetMessage(string)

	GetStatus() corev1.ConditionStatus
	SetStatus(corev1.ConditionStatus)

	GetType() StatusConditionType
	SetType(StatusConditionType)
}

// BaseApplicationStatus returns base appplication status
type BaseApplicationStatus interface {
	GetConditions() []StatusCondition
	GetCondition(StatusConditionType) StatusCondition
	SetCondition(StatusCondition)
	NewCondition() StatusCondition
}

const (
	// StatusConditionTypeReconciled ...
	StatusConditionTypeReconciled StatusConditionType = "Reconciled"
)

// BaseApplicationAutoscaling represents basic HPA configuration
type BaseApplicationAutoscaling interface {
	GetMinReplicas() *int32
	GetMaxReplicas() int32
	GetTargetCPUUtilizationPercentage() *int32
}

// BaseApplicationStorage represents basic PVC configuration
type BaseApplicationStorage interface {
	GetSize() string
	GetMountPath() string
	GetVolumeClaimTemplate() *corev1.PersistentVolumeClaim
}

// BaseApplicationService epresents basic service configuration
type BaseApplicationService interface {
	GetPort() int32
	GetType() *corev1.ServiceType
	GetAnnotations() map[string]string
}

// BaseApplicationMonitoring epresents basic service configuration
type BaseApplicationMonitoring interface {
	GetLabels() map[string]string
	GetEndpoints() []prometheusv1.Endpoint
}

// BaseApplication represents basic kubernetes application
type BaseApplication interface {
	GetApplicationImage() string
	GetPullPolicy() *corev1.PullPolicy
	GetPullSecret() *string
	GetServiceAccountName() *string
	GetReplicas() *int32
	GetLivenessProbe() *corev1.Probe
	GetReadinessProbe() *corev1.Probe
	GetVolumes() []corev1.Volume
	GetVolumeMounts() []corev1.VolumeMount
	GetResourceConstraints() *corev1.ResourceRequirements
	GetExpose() *bool
	GetEnv() []corev1.EnvVar
	GetEnvFrom() []corev1.EnvFromSource
	GetCreateKnativeService() *bool
	GetArchitecture() []string
	GetAutoscaling() BaseApplicationAutoscaling
	GetStorage() BaseApplicationStorage
	GetService() BaseApplicationService
	GetVersion() string
	GetCreateAppDefinition() *bool
	GetMonitoring() BaseApplicationMonitoring
	GetLabels() map[string]string
	GetStatus() BaseApplicationStatus
}
