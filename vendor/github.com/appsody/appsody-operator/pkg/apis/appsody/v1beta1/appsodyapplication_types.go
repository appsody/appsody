package v1beta1

import (
	"github.com/appsody/appsody-operator/pkg/common"
	prometheusv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AppsodyApplicationSpec defines the desired state of AppsodyApplication
// +k8s:openapi-gen=true
type AppsodyApplicationSpec struct {
	Version          string                         `json:"version,omitempty"`
	ApplicationImage string                         `json:"applicationImage"`
	Replicas         *int32                         `json:"replicas,omitempty"`
	Autoscaling      *AppsodyApplicationAutoScaling `json:"autoscaling,omitempty"`
	PullPolicy       *corev1.PullPolicy             `json:"pullPolicy,omitempty"`
	PullSecret       *string                        `json:"pullSecret,omitempty"`
	// +listType=map
	// +listMapKey=name
	Volumes []corev1.Volume `json:"volumes,omitempty"`
	// +listType=atomic
	VolumeMounts        []corev1.VolumeMount         `json:"volumeMounts,omitempty"`
	ResourceConstraints *corev1.ResourceRequirements `json:"resourceConstraints,omitempty"`
	ReadinessProbe      *corev1.Probe                `json:"readinessProbe,omitempty"`
	LivenessProbe       *corev1.Probe                `json:"livenessProbe,omitempty"`
	Service             *AppsodyApplicationService   `json:"service,omitempty"`
	Expose              *bool                        `json:"expose,omitempty"`
	// +listType=atomic
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`
	// +listType=map
	// +listMapKey=name
	Env                []corev1.EnvVar `json:"env,omitempty"`
	ServiceAccountName *string         `json:"serviceAccountName,omitempty"`
	// +listType=set
	Architecture         []string                      `json:"architecture,omitempty"`
	Storage              *AppsodyApplicationStorage    `json:"storage,omitempty"`
	CreateKnativeService *bool                         `json:"createKnativeService,omitempty"`
	Stack                string                        `json:"stack,omitempty"`
	Monitoring           *AppsodyApplicationMonitoring `json:"monitoring,omitempty"`
	CreateAppDefinition  *bool                         `json:"createAppDefinition,omitempty"`
	// +listType=map
	// +listMapKey=name
	InitContainers []corev1.Container `json:"initContainers,omitempty"`
}

// AppsodyApplicationAutoScaling ...
// +k8s:openapi-gen=true
type AppsodyApplicationAutoScaling struct {
	TargetCPUUtilizationPercentage *int32 `json:"targetCPUUtilizationPercentage,omitempty"`
	MinReplicas                    *int32 `json:"minReplicas,omitempty"`

	// +kubebuilder:validation:Minimum=1
	MaxReplicas int32 `json:"maxReplicas,omitempty"`
}

// AppsodyApplicationService ...
// +k8s:openapi-gen=true
type AppsodyApplicationService struct {
	Type *corev1.ServiceType `json:"type,omitempty"`

	// +kubebuilder:validation:Maximum=65536
	// +kubebuilder:validation:Minimum=1
	Port int32 `json:"port,omitempty"`

	Annotations map[string]string `json:"annotations,omitempty"`
	// +listType=atomic
	Consumes []ServiceBindingConsumes `json:"consumes,omitempty"`
	Provides *ServiceBindingProvides  `json:"provides,omitempty"`
}

// ServiceBindingProvides represents information about
// +k8s:openapi-gen=true
type ServiceBindingProvides struct {
	Category common.ServiceBindingCategory `json:"category"`
	Context  string                        `json:"context,omitempty"`
	Protocol string                        `json:"protocol,omitempty"`
	Auth     *ServiceBindingAuth           `json:"auth,omitempty"`
}

// ServiceBindingConsumes represents a service to be consumed
// +k8s:openapi-gen=true
type ServiceBindingConsumes struct {
	Name      string                        `json:"name"`
	Namespace string                        `json:"namespace,omitempty"`
	Category  common.ServiceBindingCategory `json:"category"`
	MountPath string                        `json:"mountPath,omitempty"`
}

// AppsodyApplicationStorage ...
// +k8s:openapi-gen=false
type AppsodyApplicationStorage struct {
	// +kubebuilder:validation:Pattern=^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$
	Size                string                        `json:"size,omitempty"`
	MountPath           string                        `json:"mountPath,omitempty"`
	VolumeClaimTemplate *corev1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
}

// AppsodyApplicationMonitoring ...
type AppsodyApplicationMonitoring struct {
	Labels    map[string]string       `json:"labels,omitempty"`
	Endpoints []prometheusv1.Endpoint `json:"endpoints,omitempty"`
}

// ServiceBindingAuth allows a service to provide authentication information
type ServiceBindingAuth struct {
	// The secret that contains the username for authenticating
	Username corev1.SecretKeySelector `json:"username,omitempty"`
	// The secret that contains the password for authenticating
	Password corev1.SecretKeySelector `json:"password,omitempty"`
}

// AppsodyApplicationStatus defines the observed state of AppsodyApplication
// +k8s:openapi-gen=true
type AppsodyApplicationStatus struct {
	// +listType=atomic
	Conditions       []StatusCondition       `json:"conditions,omitempty"`
	ConsumedServices common.ConsumedServices `json:"consumedServices,omitempty"`
}

// StatusCondition ...
// +k8s:openapi-gen=true
type StatusCondition struct {
	LastTransitionTime *metav1.Time           `json:"lastTransitionTime,omitempty"`
	LastUpdateTime     metav1.Time            `json:"lastUpdateTime,omitempty"`
	Reason             string                 `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
	Status             corev1.ConditionStatus `json:"status,omitempty"`
	Type               StatusConditionType    `json:"type,omitempty"`
}

// StatusConditionType ...
type StatusConditionType string

const (
	// StatusConditionTypeReconciled ...
	StatusConditionTypeReconciled StatusConditionType = "Reconciled"

	// StatusConditionTypeDependenciesSatisfied ...
	StatusConditionTypeDependenciesSatisfied StatusConditionType = "DependenciesSatisfied"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppsodyApplication is the Schema for the appsodyapplications API
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=appsodyapplications,scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".spec.applicationImage",priority=0,description="Absolute name of the deployed image containing registry and tag"
// +kubebuilder:printcolumn:name="Exposed",type="boolean",JSONPath=".spec.expose",priority=0,description="Specifies whether deployment is exposed externally via default Route"
// +kubebuilder:printcolumn:name="Reconciled",type="string",JSONPath=".status.conditions[?(@.type=='Reconciled')].status",priority=0,description="Status of the reconcile condition"
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".status.conditions[?(@.type=='Reconciled')].reason",priority=1,description="Reason for the failure of reconcile condition"
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.conditions[?(@.type=='Reconciled')].message",priority=1,description="Failure message from reconcile condition"
// +kubebuilder:printcolumn:name="DependenciesSatisfied",type="string",JSONPath=".status.conditions[?(@.type=='DependenciesSatisfied')].status",priority=1,description="Status of the application dependencies"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",priority=0,description="Age of the resource"
type AppsodyApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppsodyApplicationSpec   `json:"spec,omitempty"`
	Status AppsodyApplicationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppsodyApplicationList contains a list of AppsodyApplication
type AppsodyApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppsodyApplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppsodyApplication{}, &AppsodyApplicationList{})
}

// GetApplicationImage returns application image
func (cr *AppsodyApplication) GetApplicationImage() string {
	return cr.Spec.ApplicationImage
}

// GetPullPolicy returns image pull policy
func (cr *AppsodyApplication) GetPullPolicy() *corev1.PullPolicy {
	return cr.Spec.PullPolicy
}

// GetPullSecret returns secret name for docker registry credentials
func (cr *AppsodyApplication) GetPullSecret() *string {
	return cr.Spec.PullSecret
}

// GetServiceAccountName returns service account name
func (cr *AppsodyApplication) GetServiceAccountName() *string {
	return cr.Spec.ServiceAccountName
}

// GetReplicas returns number of replicas
func (cr *AppsodyApplication) GetReplicas() *int32 {
	return cr.Spec.Replicas
}

// GetLivenessProbe returns liveness probe
func (cr *AppsodyApplication) GetLivenessProbe() *corev1.Probe {
	return cr.Spec.LivenessProbe
}

// GetReadinessProbe returns readiness probe
func (cr *AppsodyApplication) GetReadinessProbe() *corev1.Probe {
	return cr.Spec.ReadinessProbe
}

// GetVolumes returns volumes slice
func (cr *AppsodyApplication) GetVolumes() []corev1.Volume {
	return cr.Spec.Volumes
}

// GetVolumeMounts returns volume mounts slice
func (cr *AppsodyApplication) GetVolumeMounts() []corev1.VolumeMount {
	return cr.Spec.VolumeMounts
}

// GetResourceConstraints returns resource constraints
func (cr *AppsodyApplication) GetResourceConstraints() *corev1.ResourceRequirements {
	return cr.Spec.ResourceConstraints
}

// GetExpose returns expose flag
func (cr *AppsodyApplication) GetExpose() *bool {
	return cr.Spec.Expose
}

// GetEnv returns slice of environment variables
func (cr *AppsodyApplication) GetEnv() []corev1.EnvVar {
	return cr.Spec.Env
}

// GetEnvFrom returns slice of environment variables from source
func (cr *AppsodyApplication) GetEnvFrom() []corev1.EnvFromSource {
	return cr.Spec.EnvFrom
}

// GetCreateKnativeService returns flag that toggles Knative service
func (cr *AppsodyApplication) GetCreateKnativeService() *bool {
	return cr.Spec.CreateKnativeService
}

// GetArchitecture returns slice of architectures
func (cr *AppsodyApplication) GetArchitecture() []string {
	return cr.Spec.Architecture
}

// GetAutoscaling returns autoscaling settings
func (cr *AppsodyApplication) GetAutoscaling() common.BaseApplicationAutoscaling {
	if cr.Spec.Autoscaling == nil {
		return nil
	}
	return cr.Spec.Autoscaling
}

// GetStorage returns storage settings
func (cr *AppsodyApplication) GetStorage() common.BaseApplicationStorage {
	if cr.Spec.Storage == nil {
		return nil
	}
	return cr.Spec.Storage
}

// GetService returns service settings
func (cr *AppsodyApplication) GetService() common.BaseApplicationService {
	if cr.Spec.Service == nil {
		return nil
	}
	return cr.Spec.Service
}

// GetVersion returns application version
func (cr *AppsodyApplication) GetVersion() string {
	return cr.Spec.Version
}

// GetCreateAppDefinition returns a toggle for integration with kAppNav
func (cr *AppsodyApplication) GetCreateAppDefinition() *bool {
	return cr.Spec.CreateAppDefinition
}

// GetMonitoring returns monitoring settings
func (cr *AppsodyApplication) GetMonitoring() common.BaseApplicationMonitoring {
	if cr.Spec.Monitoring == nil {
		return nil
	}
	return cr.Spec.Monitoring
}

// GetStatus returns AppsodyApplication status
func (cr *AppsodyApplication) GetStatus() common.BaseApplicationStatus {
	return &cr.Status
}

// GetInitContainers returns list of init containers
func (cr *AppsodyApplication) GetInitContainers() []corev1.Container {
	return cr.Spec.InitContainers
}

// GetGroupName returns group name to be used in labels and annotation
func (cr *AppsodyApplication) GetGroupName() string {
	return "appsody.dev"
}

// GetConsumedServices returns a map of all the service names to be consumed by the application
func (s *AppsodyApplicationStatus) GetConsumedServices() common.ConsumedServices {
	if s.ConsumedServices == nil {
		return nil
	}
	return s.ConsumedServices
}

// SetConsumedServices sets ConsumedServices
func (s *AppsodyApplicationStatus) SetConsumedServices(c common.ConsumedServices) {
	s.ConsumedServices = c
}

// GetMinReplicas returns minimum replicas
func (a *AppsodyApplicationAutoScaling) GetMinReplicas() *int32 {
	return a.MinReplicas
}

// GetMaxReplicas returns maximum replicas
func (a *AppsodyApplicationAutoScaling) GetMaxReplicas() int32 {
	return a.MaxReplicas
}

// GetTargetCPUUtilizationPercentage returns target cpu usage
func (a *AppsodyApplicationAutoScaling) GetTargetCPUUtilizationPercentage() *int32 {
	return a.TargetCPUUtilizationPercentage
}

// GetSize returns persistent volume size
func (s *AppsodyApplicationStorage) GetSize() string {
	return s.Size
}

// GetMountPath returns mount path for persistent volume
func (s *AppsodyApplicationStorage) GetMountPath() string {
	return s.MountPath
}

// GetVolumeClaimTemplate returns a template representing requested persitent volume
func (s *AppsodyApplicationStorage) GetVolumeClaimTemplate() *corev1.PersistentVolumeClaim {
	return s.VolumeClaimTemplate
}

// GetAnnotations returns a set of annotations to be added to the service
func (s *AppsodyApplicationService) GetAnnotations() map[string]string {
	return s.Annotations
}

// GetPort returns service port
func (s *AppsodyApplicationService) GetPort() int32 {
	return s.Port
}

// GetType returns service type
func (s *AppsodyApplicationService) GetType() *corev1.ServiceType {
	return s.Type
}

// GetProvides returns service provider configuration
func (s *AppsodyApplicationService) GetProvides() common.ServiceBindingProvides {
	if s.Provides == nil {
		return nil
	}
	return s.Provides
}

// GetCategory returns category of a service provider configuration
func (p *ServiceBindingProvides) GetCategory() common.ServiceBindingCategory {
	return p.Category
}

// GetContext returns context of a service provider configuration
func (p *ServiceBindingProvides) GetContext() string {
	return p.Context
}

// GetAuth returns secret of a service provider configuration
func (p *ServiceBindingProvides) GetAuth() common.ServiceBindingAuth {
	if p.Auth == nil {
		return nil
	}
	return p.Auth
}

// GetProtocol returns protocol of a service provider configuration
func (p *ServiceBindingProvides) GetProtocol() string {
	return p.Protocol
}

// GetConsumes returns a list of service consumers' configuration
func (s *AppsodyApplicationService) GetConsumes() []common.ServiceBindingConsumes {
	consumes := make([]common.ServiceBindingConsumes, len(s.Consumes))
	for i := range s.Consumes {
		consumes[i] = &s.Consumes[i]
	}
	return consumes
}

// GetName returns service name of a service consumer configuration
func (c *ServiceBindingConsumes) GetName() string {
	return c.Name
}

// GetNamespace returns namespace of a service consumer configuration
func (c *ServiceBindingConsumes) GetNamespace() string {
	return c.Namespace
}

// GetCategory returns category of a service consumer configuration
func (c *ServiceBindingConsumes) GetCategory() common.ServiceBindingCategory {
	return common.ServiceBindingCategoryOpenAPI
}

// GetMountPath returns mount path of a service consumer configuration
func (c *ServiceBindingConsumes) GetMountPath() string {
	return c.MountPath
}

// GetUsername returns username of a service binding auth object
func (a *ServiceBindingAuth) GetUsername() corev1.SecretKeySelector {
	return a.Username
}

// GetPassword returns password of a service binding auth object
func (a *ServiceBindingAuth) GetPassword() corev1.SecretKeySelector {
	return a.Password
}

// GetLabels returns labels to be added on ServiceMonitor
func (m *AppsodyApplicationMonitoring) GetLabels() map[string]string {
	return m.Labels
}

// GetEndpoints returns endpoints to be added to ServiceMonitor
func (m *AppsodyApplicationMonitoring) GetEndpoints() []prometheusv1.Endpoint {
	return m.Endpoints
}

// Initialize the AppsodyApplication instance with values from the default and constant ConfigMap
func (cr *AppsodyApplication) Initialize(defaults AppsodyApplicationSpec, constants *AppsodyApplicationSpec) {

	if cr.Spec.PullPolicy == nil {
		cr.Spec.PullPolicy = defaults.PullPolicy
		if cr.Spec.PullPolicy == nil {
			pp := corev1.PullIfNotPresent
			cr.Spec.PullPolicy = &pp
		}
	}

	if cr.Spec.PullSecret == nil {
		cr.Spec.PullSecret = defaults.PullSecret
	}

	if cr.Spec.ServiceAccountName == nil {
		cr.Spec.ServiceAccountName = defaults.ServiceAccountName
	}

	if cr.Spec.ReadinessProbe == nil {
		cr.Spec.ReadinessProbe = defaults.ReadinessProbe
	}
	if cr.Spec.LivenessProbe == nil {
		cr.Spec.LivenessProbe = defaults.LivenessProbe
	}
	if cr.Spec.Env == nil {
		cr.Spec.Env = defaults.Env
	}
	if cr.Spec.EnvFrom == nil {
		cr.Spec.EnvFrom = defaults.EnvFrom
	}

	if cr.Spec.Volumes == nil {
		cr.Spec.Volumes = defaults.Volumes
	}

	if cr.Spec.VolumeMounts == nil {
		cr.Spec.VolumeMounts = defaults.VolumeMounts
	}

	if cr.Spec.ResourceConstraints == nil {
		if defaults.ResourceConstraints != nil {
			cr.Spec.ResourceConstraints = defaults.ResourceConstraints
		} else {
			cr.Spec.ResourceConstraints = &corev1.ResourceRequirements{}
		}
	}

	if cr.Spec.Autoscaling == nil {
		cr.Spec.Autoscaling = defaults.Autoscaling
	}

	if cr.Spec.Expose == nil {
		cr.Spec.Expose = defaults.Expose
	}

	if cr.Spec.CreateKnativeService == nil {
		cr.Spec.CreateKnativeService = defaults.CreateKnativeService
	}

	if cr.Spec.Service == nil {
		cr.Spec.Service = defaults.Service
	}

	// This is to handle when there is no service in the CR nor defaults
	if cr.Spec.Service == nil {
		cr.Spec.Service = &AppsodyApplicationService{}
	}

	if cr.Spec.Service.Type == nil {
		st := corev1.ServiceTypeClusterIP
		cr.Spec.Service.Type = &st
	}
	if cr.Spec.Service.Port == 0 {
		if defaults.Service != nil && defaults.Service.Port != 0 {
			cr.Spec.Service.Port = defaults.Service.Port
		} else {
			cr.Spec.Service.Port = 8080
		}
	}

	if cr.Spec.CreateAppDefinition == nil {
		if defaults.CreateAppDefinition != nil {
			cr.Spec.CreateAppDefinition = defaults.CreateAppDefinition
		}
	}

	if cr.Spec.Monitoring == nil {
		if defaults.Monitoring != nil {
			cr.Spec.Monitoring = defaults.Monitoring
		}
	}

	if cr.Spec.InitContainers == nil {
		if defaults.InitContainers != nil {
			cr.Spec.InitContainers = defaults.InitContainers
		}
	}

	if cr.Spec.Service.Provides != nil && cr.Spec.Service.Provides.Protocol == "" {
		cr.Spec.Service.Provides.Protocol = "http"
	}

	for i := range cr.Spec.Service.Consumes {
		if cr.Spec.Service.Consumes[i].Category == common.ServiceBindingCategoryOpenAPI {
			if cr.Spec.Service.Consumes[i].Namespace == "" {
				cr.Spec.Service.Consumes[i].Namespace = cr.Namespace
			}
		}
	}

	if constants != nil {
		cr.applyConstants(defaults, constants)
	}
}

func (cr *AppsodyApplication) applyConstants(defaults AppsodyApplicationSpec, constants *AppsodyApplicationSpec) {

	if constants.Replicas != nil {
		cr.Spec.Replicas = constants.Replicas
	}

	if constants.Stack != "" {
		cr.Spec.Stack = constants.Stack
	}

	if constants.ApplicationImage != "" {
		cr.Spec.ApplicationImage = constants.ApplicationImage
	}

	if constants.PullPolicy != nil {
		cr.Spec.PullPolicy = constants.PullPolicy
	}

	if constants.PullSecret != nil {
		cr.Spec.PullSecret = constants.PullSecret
	}

	if constants.Expose != nil {
		cr.Spec.Expose = constants.Expose
	}

	if constants.CreateKnativeService != nil {
		cr.Spec.CreateKnativeService = constants.CreateKnativeService
	}

	if constants.ServiceAccountName != nil {
		cr.Spec.ServiceAccountName = constants.ServiceAccountName
	}

	if constants.Architecture != nil {
		cr.Spec.Architecture = constants.Architecture
	}

	if constants.ReadinessProbe != nil {
		cr.Spec.ReadinessProbe = constants.ReadinessProbe
	}

	if constants.LivenessProbe != nil {
		cr.Spec.LivenessProbe = constants.LivenessProbe
	}

	if constants.EnvFrom != nil {
		for _, v := range constants.EnvFrom {

			found := false
			for _, v2 := range cr.Spec.EnvFrom {
				if v2 == v {
					found = true
				}
			}
			if !found {
				cr.Spec.EnvFrom = append(cr.Spec.EnvFrom, v)
			}
		}
	}

	if constants.Env != nil {
		for _, v := range constants.Env {
			found := false
			for _, v2 := range cr.Spec.Env {
				if v2.Name == v.Name {
					found = true
				}
			}
			if !found {
				cr.Spec.Env = append(cr.Spec.Env, v)
			}
		}
	}

	if constants.Volumes != nil {
		for _, v := range constants.Volumes {
			found := false
			for _, v2 := range cr.Spec.Volumes {
				if v2.Name == v.Name {
					found = true
				}
			}
			if !found {
				cr.Spec.Volumes = append(cr.Spec.Volumes, v)
			}
		}
	}

	if constants.VolumeMounts != nil {
		for _, v := range constants.VolumeMounts {
			found := false
			for _, v2 := range cr.Spec.VolumeMounts {
				if v2.Name == v.Name && v2.SubPath == v.SubPath {
					found = true
				}
			}
			if !found {
				cr.Spec.VolumeMounts = append(cr.Spec.VolumeMounts, v)
			}
		}
	}

	if constants.ResourceConstraints != nil {
		cr.Spec.ResourceConstraints = constants.ResourceConstraints
	}

	if constants.Service != nil {
		if constants.Service.Type != nil {
			cr.Spec.Service.Type = constants.Service.Type
		}
		if constants.Service.Port != 0 {
			cr.Spec.Service.Port = constants.Service.Port
		}
	}

	if constants.Autoscaling != nil {
		cr.Spec.Autoscaling = constants.Autoscaling
	}

	if constants.InitContainers != nil {
		cr.Spec.InitContainers = constants.InitContainers
	}

	if constants.Monitoring != nil {
		cr.Spec.Monitoring = constants.Monitoring
	}

	if constants.CreateAppDefinition != nil {
		cr.Spec.CreateAppDefinition = constants.CreateAppDefinition
	}

}

// GetLabels returns set of labels to be added to all resources
func (cr *AppsodyApplication) GetLabels() map[string]string {

	labels := map[string]string{
		"app.kubernetes.io/instance":   cr.Name,
		"app.kubernetes.io/name":       cr.Name,
		"app.kubernetes.io/managed-by": "appsody-operator",
	}

	if cr.Spec.Stack != "" {
		labels["stack.appsody.dev/id"] = cr.Spec.Stack
	}

	if cr.Spec.Version != "" {
		labels["app.kubernetes.io/version"] = cr.Spec.Version
	}

	for key, value := range cr.Labels {
		if key != "app.kubernetes.io/instance" {
			labels[key] = value
		}
	}

	return labels
}

// GetAnnotations returns set of annotations to be added to all resources
func (cr *AppsodyApplication) GetAnnotations() map[string]string {
	return cr.Annotations
}

// GetType returns status condition type
func (c *StatusCondition) GetType() common.StatusConditionType {
	return convertToCommonStatusConditionType(c.Type)
}

// SetType returns status condition type
func (c *StatusCondition) SetType(ct common.StatusConditionType) {
	c.Type = convertFromCommonStatusConditionType(ct)
}

// GetLastTransitionTime return time of last status change
func (c *StatusCondition) GetLastTransitionTime() *metav1.Time {
	return c.LastTransitionTime
}

// SetLastTransitionTime sets time of last status change
func (c *StatusCondition) SetLastTransitionTime(t *metav1.Time) {
	c.LastTransitionTime = t
}

// GetLastUpdateTime return time of last status update
func (c *StatusCondition) GetLastUpdateTime() metav1.Time {
	return c.LastUpdateTime
}

// SetLastUpdateTime sets time of last status update
func (c *StatusCondition) SetLastUpdateTime(t metav1.Time) {
	c.LastUpdateTime = t
}

// GetMessage return condition's message
func (c *StatusCondition) GetMessage() string {
	return c.Message
}

// SetMessage sets condition's message
func (c *StatusCondition) SetMessage(m string) {
	c.Message = m
}

// GetReason return condition's message
func (c *StatusCondition) GetReason() string {
	return c.Reason
}

// SetReason sets condition's reason
func (c *StatusCondition) SetReason(r string) {
	c.Reason = r
}

// GetStatus return condition's status
func (c *StatusCondition) GetStatus() corev1.ConditionStatus {
	return c.Status
}

// SetStatus sets condition's status
func (c *StatusCondition) SetStatus(s corev1.ConditionStatus) {
	c.Status = s
}

// NewCondition returns new condition
func (s *AppsodyApplicationStatus) NewCondition() common.StatusCondition {
	return &StatusCondition{}
}

// GetConditions returns slice of conditions
func (s *AppsodyApplicationStatus) GetConditions() []common.StatusCondition {
	var conditions = make([]common.StatusCondition, len(s.Conditions))
	for i := range s.Conditions {
		conditions[i] = &s.Conditions[i]
	}
	return conditions
}

// GetCondition ...
func (s *AppsodyApplicationStatus) GetCondition(t common.StatusConditionType) common.StatusCondition {
	for i := range s.Conditions {
		if s.Conditions[i].GetType() == t {
			return &s.Conditions[i]
		}
	}
	return nil
}

// SetCondition ...
func (s *AppsodyApplicationStatus) SetCondition(c common.StatusCondition) {
	condition := &StatusCondition{}
	found := false
	for i := range s.Conditions {
		if s.Conditions[i].GetType() == c.GetType() {
			condition = &s.Conditions[i]
			found = true
		}
	}

	condition.SetLastTransitionTime(c.GetLastTransitionTime())
	condition.SetLastUpdateTime(c.GetLastUpdateTime())
	condition.SetReason(c.GetReason())
	condition.SetMessage(c.GetMessage())
	condition.SetStatus(c.GetStatus())
	condition.SetType(c.GetType())
	if !found {
		s.Conditions = append(s.Conditions, *condition)
	}
}

func convertToCommonStatusConditionType(c StatusConditionType) common.StatusConditionType {
	switch c {
	case StatusConditionTypeReconciled:
		return common.StatusConditionTypeReconciled
	case StatusConditionTypeDependenciesSatisfied:
		return common.StatusConditionTypeDependenciesSatisfied
	default:
		panic(c)
	}
}

func convertFromCommonStatusConditionType(c common.StatusConditionType) StatusConditionType {
	switch c {
	case common.StatusConditionTypeReconciled:
		return StatusConditionTypeReconciled
	case common.StatusConditionTypeDependenciesSatisfied:
		return StatusConditionTypeDependenciesSatisfied
	default:
		panic(c)
	}
}
