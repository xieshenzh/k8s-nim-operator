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

package v1alpha1

import (
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/utils/ptr"
)

const (
	// DefaultAPIPort is the default api port.
	DefaultAPIPort = 8000
	// DefaultNamedPortAPI is the default name for api port.
	DefaultNamedPortAPI = "api"
	// DefaultNamedPortGRPC is the default name for grpc port.
	DefaultNamedPortGRPC = "grpc"
	// DefaultNamedPortMetrics is the default name for metrics port.
	DefaultNamedPortMetrics = "metrics"
)

// Expose defines attributes to expose the service.
type Expose struct {
	Service Service `json:"service,omitempty"`
	Ingress Ingress `json:"ingress,omitempty"`
}

// Service defines attributes to create a service.
type Service struct {
	Type corev1.ServiceType `json:"type,omitempty"`
	// Override the default service name
	Name string `json:"name,omitempty"`
	// Port is the main api serving port (default: 8000)
	//
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default:=8000
	Port *int32 `json:"port,omitempty"`
	// GRPCPort is the GRPC serving port
	// Note: This port is only applicable for NIMs that runs a Triton GRPC Inference Server.
	//
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	GRPCPort *int32 `json:"grpcPort,omitempty"`
	// MetricsPort is the port for metrics
	// Note: This port is only applicable for NIMs that runs a separate metrics endpoint on Triton Inference Server.
	//
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	MetricsPort *int32            `json:"metricsPort,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ExposeV1 defines attributes to expose the service.
type ExposeV1 struct {
	Service Service   `json:"service,omitempty"`
	Ingress IngressV1 `json:"ingress,omitempty"`
}

// Metrics defines attributes to setup metrics collection.
type Metrics struct {
	Enabled *bool `json:"enabled,omitempty"`
	// for use with the Prometheus Operator and the primary service object
	ServiceMonitor ServiceMonitor `json:"serviceMonitor,omitempty"`
}

// ServiceMonitor defines attributes to create a service monitor.
type ServiceMonitor struct {
	AdditionalLabels map[string]string `json:"additionalLabels,omitempty"`
	Annotations      map[string]string `json:"annotations,omitempty"`
	Interval         promv1.Duration   `json:"interval,omitempty"`
	ScrapeTimeout    promv1.Duration   `json:"scrapeTimeout,omitempty"`
}

// Autoscaling defines attributes to automatically scale the service based on metrics.
type Autoscaling struct {
	Enabled     *bool                       `json:"enabled,omitempty"`
	HPA         HorizontalPodAutoscalerSpec `json:"hpa,omitempty"`
	Annotations map[string]string           `json:"annotations,omitempty"`
}

// HorizontalPodAutoscalerSpec defines the parameters required to setup HPA.
type HorizontalPodAutoscalerSpec struct {
	MinReplicas *int32                                         `json:"minReplicas,omitempty"`
	MaxReplicas int32                                          `json:"maxReplicas"`
	Metrics     []autoscalingv2.MetricSpec                     `json:"metrics,omitempty"`
	Behavior    *autoscalingv2.HorizontalPodAutoscalerBehavior `json:"behavior,omitempty" `
}

// Image defines image attributes.
type Image struct {
	Repository  string   `json:"repository"`
	PullPolicy  string   `json:"pullPolicy,omitempty"`
	Tag         string   `json:"tag"`
	PullSecrets []string `json:"pullSecrets,omitempty"`
}

// Ingress defines attributes to enable ingress for the service.
type Ingress struct {
	// ingress, or virtualService - not both
	Enabled     *bool                    `json:"enabled,omitempty"`
	Annotations map[string]string        `json:"annotations,omitempty"`
	Spec        networkingv1.IngressSpec `json:"spec,omitempty"`
}

// IngressV1 defines attributes for ingress
//
// +kubebuilder:validation:XValidation:rule="(has(self.spec) && has(self.enabled) && self.enabled) || !has(self.enabled) || !self.enabled", message="spec cannot be nil when ingress is enabled"
type IngressV1 struct {
	Enabled     *bool             `json:"enabled,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Spec        *IngressSpec      `json:"spec,omitempty"`
}

// ResourceRequirements defines the resources required for a container.
type ResourceRequirements struct {
	// Limits describes the maximum amount of compute resources allowed.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Limits corev1.ResourceList `json:"limits,omitempty" protobuf:"bytes,1,rep,name=limits,casttype=ResourceList,castkey=ResourceName"`
	// Requests describes the minimum amount of compute resources required.
	// If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
	// otherwise to an implementation-defined value. Requests cannot exceed Limits.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Requests corev1.ResourceList `json:"requests,omitempty" protobuf:"bytes,2,rep,name=requests,casttype=ResourceList,castkey=ResourceName"`
}

func (i *IngressV1) GenerateNetworkingV1IngressSpec(name string) networkingv1.IngressSpec {
	if i.Spec == nil {
		return networkingv1.IngressSpec{}
	}

	ingressSpec := networkingv1.IngressSpec{
		IngressClassName: &i.Spec.IngressClassName,
		Rules: []networkingv1.IngressRule{
			{
				Host: i.Spec.Host,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{},
				},
			},
		},
	}

	svcBackend := networkingv1.IngressBackend{
		Service: &networkingv1.IngressServiceBackend{
			Name: name,
			Port: networkingv1.ServiceBackendPort{
				Name: DefaultNamedPortAPI,
			},
		},
	}
	if len(i.Spec.Paths) == 0 {
		ingressSpec.Rules[0].HTTP.Paths = append(ingressSpec.Rules[0].HTTP.Paths, networkingv1.HTTPIngressPath{
			Path:     "/",
			PathType: ptr.To(networkingv1.PathTypePrefix),
			Backend:  svcBackend,
		})
	}
	for _, path := range i.Spec.Paths {
		ingressSpec.Rules[0].HTTP.Paths = append(ingressSpec.Rules[0].HTTP.Paths, networkingv1.HTTPIngressPath{
			Path:     path.Path,
			PathType: path.PathType,
			Backend:  svcBackend,
		})
	}
	return ingressSpec
}

type IngressSpec struct {
	// +kubebuilder:validation:Pattern=`[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`
	IngressClassName string        `json:"ingressClassName"`
	Host             string        `json:"host,omitempty"`
	Paths            []IngressPath `json:"paths,omitempty"`
}

// IngressPath defines attributes for ingress paths.
type IngressPath struct {
	// +kubebuilder:default="/"
	Path string `json:"path,omitempty"`
	// +kubebuilder:default=Prefix
	PathType *networkingv1.PathType `json:"pathType,omitempty"`
}

// Probe defines attributes for startup/liveness/readiness probes.
type Probe struct {
	Enabled *bool         `json:"enabled,omitempty"`
	Probe   *corev1.Probe `json:"probe,omitempty"`
}

// CertConfig defines the configuration for custom certificates.
type CertConfig struct {
	// Name of the ConfigMap containing the certificate data.
	Name string `json:"name"`
	// MountPath is the path where the certificates should be mounted in the container.
	MountPath string `json:"mountPath"`
}

// ProxySpec defines the proxy configuration for NIMService.
type ProxySpec struct {
	HttpProxy     string `json:"httpProxy,omitempty"`
	HttpsProxy    string `json:"httpsProxy,omitempty"`
	NoProxy       string `json:"noProxy,omitempty"`
	CertConfigMap string `json:"certConfigMap,omitempty"`
}

// NGCSecret represents the secret and key details for NGC.
type NGCSecret struct {
	// Name of the Kubernetes secret containing NGC API key
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Key in the key containing the actual API key value
	// +kubebuilder:default:="NGC_API_KEY"
	Key string `json:"key"`
}

// HFSecret represents the secret and key details for HuggingFace.
type HFSecret struct {
	// Name of the Kubernetes secret containing HF_TOKEN key
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Key in the key containing the actual token value
	// +kubebuilder:default:="HF_TOKEN"
	Key string `json:"key"`
}

// PersistentVolumeClaim defines the attributes of PVC.
type PersistentVolumeClaim struct {
	// Create specifies whether to create a new PersistentVolumeClaim (PVC).
	// If set to false, an existing PVC must be referenced via the `Name` field.
	Create *bool `json:"create,omitempty"`
	// Name of the PVC to use. Required if `Create` is false (i.e., using an existing PVC).
	Name string `json:"name,omitempty"`
	// StorageClass to be used for PVC creation. Leave it as empty if the PVC is already created or
	// a default storage class is set in the cluster.
	StorageClass string `json:"storageClass,omitempty"`
	// Size of the NIM cache in Gi, used during PVC creation
	Size string `json:"size,omitempty"`
	// VolumeAccessMode is the volume access mode of the PVC
	VolumeAccessMode corev1.PersistentVolumeAccessMode `json:"volumeAccessMode,omitempty"`
	// SubPath is the path inside the PVC that should be mounted
	SubPath string `json:"subPath,omitempty"`
	// Annotations for the PVC
	Annotations map[string]string `json:"annotations,omitempty"`
}

// DRAResource references exactly one ResourceClaim, either directly
// or by naming a ResourceClaimTemplate which is then turned into a ResourceClaim.
//
// When creating the NIMService pods, it adds a name (`DNS_LABEL` format) to it
// that uniquely identifies the DRA resource.
// +kubebuilder:validation:XValidation:rule="has(self.resourceClaimName) != has(self.resourceClaimTemplateName)",message="exactly one of spec.resourceClaimName and spec.resourceClaimTemplateName must be set."
type DRAResource struct {
	// ResourceClaimName is the name of a ResourceClaim object in the same
	// namespace as the NIMService.
	//
	// Exactly one of ResourceClaimName and ResourceClaimTemplateName must
	// be set.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`
	ResourceClaimName *string `json:"resourceClaimName,omitempty"`

	// ResourceClaimTemplateName is the name of a ResourceClaimTemplate
	// object in the same namespace as the pods for this NIMService.
	//
	// The template will be used to create a new ResourceClaim, which will
	// be bound to the pods created for this NIMService.
	//
	// Exactly one of ResourceClaimName and ResourceClaimTemplateName must
	// be set.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`
	ResourceClaimTemplateName *string `json:"resourceClaimTemplateName,omitempty"`

	// Requests is the list of requests in the referenced ResourceClaim/ResourceClaimTemplate
	// to be made available to the model container of the NIMService pods.
	//
	// If empty, everything from the claim is made available, otherwise
	// only the result of this subset of requests.
	//
	// +kubebuilder:validation:items:MinLength=1
	Requests []string `json:"requests,omitempty"`
}

// DRAResourceStatus defines the status of the DRAResource.
// +kubebuilder:validation:XValidation:rule="has(self.resourceClaimStatus) != has(self.resourceClaimTemplateStatus)",message="exactly one of resourceClaimStatus and resourceClaimTemplateStatus must be set."
type DRAResourceStatus struct {
	// Name is the pod claim name referenced in the pod spec as `spec.resourceClaims[].name` for this DRA resource.
	Name string `json:"name"`
	// ResourceClaimStatus is the status of the resource claim in this DRA resource.
	//
	// Exactly one of resourceClaimStatus and resourceClaimTemplateStatus will be set.
	ResourceClaimStatus *DRAResourceClaimStatusInfo `json:"resourceClaimStatus,omitempty"`
	// ResourceClaimTemplateStatus is the status of the resource claim template in this DRA resource.
	//
	// Exactly one of resourceClaimStatus and resourceClaimTemplateStatus will be set.
	ResourceClaimTemplateStatus *DRAResourceClaimTemplateStatusInfo `json:"resourceClaimTemplateStatus,omitempty"`
}

// DRAResourceClaimStatusInfo defines the status of a ResourceClaim referenced in the DRAResource.
type DRAResourceClaimStatusInfo struct {
	// Name is the name of the ResourceClaim.
	Name string `json:"name"`
	// State is the state of the ResourceClaim.
	// * pending: the resource claim is pending allocation.
	// * deleted: the resource claim has a deletion timestamp set but is not yet finalized.
	// * allocated: the resource claim is allocated to a pod.
	// * reserved: the resource claim is consumed by a pod.
	// This field will have one or more of the above values depending on the status of the resource claim.
	//
	// +kubebuilder:validation:default=pending
	State string `json:"state"`
}

// DRAResourceClaimTemplateStatusInfo defines the status of a ResourceClaimTemplate referenced in the DRAResource.
type DRAResourceClaimTemplateStatusInfo struct {
	// Name is the name of the resource claim template.
	Name string `json:"name"`
	// ResourceClaimStatuses is the statuses of the generated resource claims from this resource claim template.
	ResourceClaimStatuses []DRAResourceClaimStatusInfo `json:"resourceClaimStatuses,omitempty"`
}
