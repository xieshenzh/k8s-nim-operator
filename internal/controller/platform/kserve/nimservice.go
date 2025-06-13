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

package kserve

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	kservev1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	apiResource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appsv1alpha1 "github.com/NVIDIA/k8s-nim-operator/api/apps/v1alpha1"
	"github.com/NVIDIA/k8s-nim-operator/internal/conditions"
	"github.com/NVIDIA/k8s-nim-operator/internal/k8sutil"
	"github.com/NVIDIA/k8s-nim-operator/internal/render"
	rendertypes "github.com/NVIDIA/k8s-nim-operator/internal/render/types"
	"github.com/NVIDIA/k8s-nim-operator/internal/shared"
	"github.com/NVIDIA/k8s-nim-operator/internal/utils"
)

const (
	// ManifestsDir is the directory to render k8s resource manifests.
	ManifestsDir = "/manifests"
)

// NIMServiceReconciler represents the NIMService reconciler instance for KServe platform.
type NIMServiceReconciler struct {
	client.Client
	scheme *runtime.Scheme
	log    logr.Logger

	updater          conditions.Updater
	renderer         render.Renderer
	recorder         record.EventRecorder
	orchestratorType k8sutil.OrchestratorType
}

// NewNIMServiceReconciler returns NIMServiceReconciler for KServe platform.
func NewNIMServiceReconciler(r shared.Reconciler) *NIMServiceReconciler {
	return &NIMServiceReconciler{
		Client:   r.GetClient(),
		scheme:   r.GetScheme(),
		log:      r.GetLogger(),
		updater:  r.GetUpdater(),
		renderer: render.NewRenderer(ManifestsDir),
		recorder: r.GetEventRecorder(),
	}
}

func (r *NIMServiceReconciler) cleanupNIMService(ctx context.Context, nimService *appsv1alpha1.NIMService) error {
	// All dependent (owned) objects will be automatically garbage collected.
	return nil
}

func (r *NIMServiceReconciler) reconcileNIMService(ctx context.Context, nimService *appsv1alpha1.NIMService) (ctrl.Result, error) {
	logger := r.log

	var err error
	defer func() {
		if err != nil {
			r.recorder.Eventf(nimService, corev1.EventTypeWarning, conditions.Failed,
				"NIMService %s failed, msg: %s", nimService.Name, err.Error())
		}
	}()

	// Generate annotation for the current operator-version and apply to all resources
	// Get generic name for all resources
	namespacedName := types.NamespacedName{Name: nimService.GetName(), Namespace: nimService.GetNamespace()}

	// Sync serviceaccount
	err = r.renderAndSyncResource(ctx, nimService, &corev1.ServiceAccount{}, func() (client.Object, error) {
		return r.renderer.ServiceAccount(nimService.GetServiceAccountParams())
	}, "serviceaccount", conditions.ReasonServiceAccountFailed)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Sync role
	err = r.renderAndSyncResource(ctx, nimService, &rbacv1.Role{}, func() (client.Object, error) {
		return r.renderer.Role(nimService.GetRoleParams())
	}, "role", conditions.ReasonRoleFailed)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Sync rolebinding
	err = r.renderAndSyncResource(ctx, nimService, &rbacv1.RoleBinding{}, func() (client.Object, error) {
		return r.renderer.RoleBinding(nimService.GetRoleBindingParams())
	}, "rolebinding", conditions.ReasonRoleBindingFailed)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Sync Service Monitor
	if nimService.IsServiceMonitorEnabled() {
		err = r.renderAndSyncResource(ctx, nimService, &monitoringv1.ServiceMonitor{}, func() (client.Object, error) {
			return r.renderer.ServiceMonitor(nimService.GetServiceMonitorParams())
		}, "servicemonitor", conditions.ReasonServiceMonitorFailed)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	modelPVC, modelProfile, errCache := r.renderAndSyncCache(ctx, nimService)
	if errCache != nil {
		return ctrl.Result{}, errCache
	} else if modelPVC == nil {
		return ctrl.Result{}, nil
	}

	err = r.renderAndSyncInferenceService(ctx, nimService, modelPVC, modelProfile)
	if err != nil {

	}

	return ctrl.Result{}, nil
}

func (r *NIMServiceReconciler) renderAndSyncResource(ctx context.Context, nimService *appsv1alpha1.NIMService,
	obj client.Object, renderFunc func() (client.Object, error), conditionType string, reason string) error {
	logger := r.log

	namespacedName := types.NamespacedName{Name: nimService.GetName(), Namespace: nimService.GetNamespace()}

	err := r.Get(ctx, namespacedName, obj)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, fmt.Sprintf("Error is not NotFound for %s: %v", obj.GetObjectKind(), err))
		return err
	}
	// Don't do anything if CR is unchanged.
	if err == nil && !utils.IsParentSpecChanged(obj, utils.DeepHashObject(nimService.Spec)) {
		return nil
	}

	resource, err := renderFunc()
	if err != nil {
		logger.Error(err, "failed to render", conditionType, namespacedName)
		statusError := r.updater.SetConditionsFailed(ctx, nimService, reason, err.Error())
		if statusError != nil {
			logger.Error(statusError, "failed to update status", "nimservice", nimService.Name)
		}
		return err
	}

	// Check if the resource is nil
	if resource == nil {
		logger.V(2).Info("rendered nil resource")
		return nil
	}

	metaAccessor, ok := resource.(metav1.Object)
	if !ok {
		logger.V(2).Info("rendered un-initialized resource")
		return nil
	}

	if metaAccessor == nil || metaAccessor.GetName() == "" || metaAccessor.GetNamespace() == "" {
		logger.V(2).Info("rendered un-initialized resource")
		return nil
	}

	if err = controllerutil.SetControllerReference(nimService, resource, r.scheme); err != nil {
		logger.Error(err, "failed to set owner", conditionType, namespacedName)
		statusError := r.updater.SetConditionsFailed(ctx, nimService, reason, err.Error())
		if statusError != nil {
			logger.Error(statusError, "failed to update status", "nimservice", nimService.Name)
		}
		return err
	}

	err = k8sutil.SyncResource(ctx, r.Client, obj, resource)
	if err != nil {
		logger.Error(err, "failed to sync", conditionType, namespacedName)
		statusError := r.updater.SetConditionsFailed(ctx, nimService, reason, err.Error())
		if statusError != nil {
			logger.Error(statusError, "failed to update status", "nimservice", nimService.Name)
		}
		return err
	}
	return nil
}

func (r *NIMServiceReconciler) renderAndSyncCache(ctx context.Context,
	nimService *appsv1alpha1.NIMService) (*appsv1alpha1.PersistentVolumeClaim, string, error) {
	logger := r.log

	var modelPVC *appsv1alpha1.PersistentVolumeClaim
	modelProfile := ""

	// Select PVC for model store
	nimCacheName := nimService.GetNIMCacheName()
	if nimCacheName != "" {
		nimCache := appsv1alpha1.NIMCache{}
		if err := r.Get(ctx, types.NamespacedName{Name: nimCacheName, Namespace: nimService.GetNamespace()}, &nimCache); err != nil {
			// Fail the NIMService if the NIMCache is not found
			if errors.IsNotFound(err) {
				msg := fmt.Sprintf("NIMCache %s not found", nimCacheName)
				statusUpdateErr := r.updater.SetConditionsFailed(ctx, nimService, conditions.ReasonNIMCacheNotFound, msg)
				r.recorder.Eventf(nimService, corev1.EventTypeWarning, conditions.Failed, msg)
				logger.Info(msg, "nimcache", nimCacheName, "nimservice", nimService.Name)
				if statusUpdateErr != nil {
					logger.Error(statusUpdateErr, "failed to update status", "nimservice", nimService.Name)
					return nil, "", statusUpdateErr
				}
				return nil, "", nil
			}
			return nil, "", err
		}

		switch nimCache.Status.State {
		case appsv1alpha1.NimCacheStatusReady:
			logger.V(4).Info("NIMCache is ready", "nimcache", nimCacheName, "nimservice", nimService.Name)
		case appsv1alpha1.NimCacheStatusFailed:
			var msg string
			cond := meta.FindStatusCondition(nimCache.Status.Conditions, conditions.Failed)
			if cond != nil && cond.Status == metav1.ConditionTrue {
				msg = cond.Message
			} else {
				msg = ""
			}
			err := r.updater.SetConditionsFailed(ctx, nimService, conditions.ReasonNIMCacheFailed, msg)
			r.recorder.Eventf(nimService, corev1.EventTypeWarning, conditions.Failed, msg)
			logger.Info(msg, "nimcache", nimCacheName, "nimservice", nimService.Name)
			if err != nil {
				logger.Error(err, "failed to update status", "nimservice", nimService.Name)
			}
			return nil, "", err
		default:
			msg := fmt.Sprintf("NIMCache %s not ready", nimCacheName)
			err := r.updater.SetConditionsNotReady(ctx, nimService, conditions.ReasonNIMCacheNotReady, msg)
			r.recorder.Eventf(nimService, corev1.EventTypeNormal, conditions.NotReady,
				"NIMService %s not ready yet, msg: %s", nimService.Name, msg)
			logger.V(4).Info(msg, "nimservice", nimService.Name)
			if err != nil {
				logger.Error(err, "failed to update status", "nimservice", nimService.Name)
			}
			return nil, "", err
		}

		// Fetch PVC for the associated NIMCache instance and mount it
		if nimCache.Status.PVC == "" {
			err := fmt.Errorf("missing PVC for the nimcache instance %s", nimCache.GetName())
			logger.Error(err, "unable to obtain pvc backing the nimcache instance")
			return nil, "", err
		}
		if nimCache.Spec.Storage.PVC.Name == "" {
			nimCache.Spec.Storage.PVC.Name = nimCache.Status.PVC
		}
		// Get the underlying PVC for the NIMCache instance
		modelPVC = &nimCache.Spec.Storage.PVC
		logger.V(2).Info("obtained the backing pvc for nimcache instance", "pvc", modelPVC)

		if profile := nimService.GetNIMCacheProfile(); profile != "" {
			logger.Info("overriding model profile", "profile", profile)
			modelProfile = profile
		}
	} else if nimService.Spec.Storage.PVC.Create != nil && *nimService.Spec.Storage.PVC.Create {
		// Create a new PVC
		var err error
		modelPVC, err = r.reconcilePVC(ctx, nimService)
		if err != nil {
			logger.Error(err, "unable to create pvc")
			return nil, "", err
		}
	} else if nimService.Spec.Storage.PVC.Name != "" {
		// Use an existing PVC
		modelPVC = &nimService.Spec.Storage.PVC
	} else {
		err := fmt.Errorf("neither external PVC name or NIMCache volume is provided")
		logger.Error(err, "failed to determine PVC for model-store")
		return nil, "", err
	}

	return modelPVC, modelProfile, nil
}

func (r *NIMServiceReconciler) reconcilePVC(ctx context.Context, nimService *appsv1alpha1.NIMService) (*appsv1alpha1.PersistentVolumeClaim, error) {
	logger := r.log

	pvcName := nimService.GetPVCName(nimService.Spec.Storage.PVC)
	pvcNamespacedName := types.NamespacedName{Name: pvcName, Namespace: nimService.GetNamespace()}
	pvc := &corev1.PersistentVolumeClaim{}
	err := r.Get(ctx, pvcNamespacedName, pvc)
	if err != nil && client.IgnoreNotFound(err) != nil {
		return nil, err
	}

	// If PVC does not exist, create a new one if creation flag is enabled
	if err != nil {
		if nimService.Spec.Storage.PVC.Create != nil && *nimService.Spec.Storage.PVC.Create {
			pvc, err = shared.ConstructPVC(nimService.Spec.Storage.PVC, metav1.ObjectMeta{Name: pvcName, Namespace: nimService.GetNamespace()})
			if err != nil {
				logger.Error(err, "Failed to construct pvc", "name", pvcName)
				return nil, err
			}
			if err := controllerutil.SetControllerReference(nimService, pvc, r.scheme); err != nil {
				return nil, err
			}
			err = r.Create(ctx, pvc)
			if err != nil {
				logger.Error(err, "Failed to create pvc", "name", pvc.Name)
				return nil, err
			}
			logger.Info("Created PVC for NIM Service", "pvc", pvcName)

			conditions.UpdateCondition(&nimService.Status.Conditions, appsv1alpha1.NimCacheConditionPVCCreated, metav1.ConditionTrue, "PVCCreated", "The PVC has been created for storing NIM")
			nimService.Status.State = appsv1alpha1.NimCacheStatusPVCCreated
			if err := r.Status().Update(ctx, nimService); err != nil {
				logger.Error(err, "Failed to update status", "NIMService", nimService.Name)
				return nil, err
			}
		} else {
			logger.Error(err, "PVC doesn't exist and auto-creation is not enabled", "name", pvcNamespacedName)
			return nil, err
		}
	}

	// If explicit name is not provided in the spec, update it with the one created
	if nimService.Spec.Storage.PVC.Name == "" {
		nimService.Spec.Storage.PVC.Name = pvc.Name
	}

	return &nimService.Spec.Storage.PVC, nil
}

func (r *NIMServiceReconciler) renderAndSyncInferenceService(ctx context.Context,
	nimService *appsv1alpha1.NIMService, modelPVC *appsv1alpha1.PersistentVolumeClaim, modelProfile string) error {
	logger := r.log

	inferenceServiceParams := nimService.GetInferenceServiceParams()
	inferenceServiceParams.OrchestratorType = string(r.orchestratorType)

	// Setup volume mounts with model store
	inferenceServiceParams.Volumes = nimService.GetVolumes(*modelPVC)
	inferenceServiceParams.VolumeMounts = nimService.GetVolumeMounts(*modelPVC)

	// Setup env for explicit override profile is specified
	if modelProfile != "" {
		profileEnv := corev1.EnvVar{
			Name:  "NIM_MODEL_PROFILE",
			Value: modelProfile,
		}
		inferenceServiceParams.Env = append(inferenceServiceParams.Env, profileEnv)

		// Retrieve and set profile details from NIMCache
		var profile *appsv1alpha1.NIMProfile
		var err error
		profile, err = r.getNIMCacheProfile(ctx, nimService, modelProfile)
		if err != nil {
			logger.Error(err, "Failed to get cached NIM profile")
			return err
		}

		// Auto assign GPU resources in case of the optimized profile
		if profile != nil {
			if err = r.assignGPUResources(ctx, nimService, profile, inferenceServiceParams); err != nil {
				return err
			}
		}

		// TODO: assign GPU resources and node selector that is required for the selected profile
	}

	// Setup pod resource claims
	draResources := shared.GenerateNamedDRAResources(nimService)
	inferenceServiceParams.PodResourceClaims = shared.GetPodResourceClaims(draResources)

	// Sync InferenceService
	err := r.renderAndSyncResource(ctx, nimService, &kservev1beta1.InferenceService{}, func() (client.Object, error) {
		result, err := r.renderer.InferenceService(inferenceServiceParams)
		if err != nil {
			return nil, err
		}
		initContainers := nimService.GetInitContainers()
		if len(initContainers) > 0 {
			result.Spec.Predictor.InitContainers = initContainers
		}
		// Update Container resources with DRA resource claims.
		shared.UpdateContainerResourceClaims(result.Spec.Predictor.Containers, draResources)
		return result, err

	}, "inferenceservice", conditions.ReasonInferenceServiceFailed)
	if err != nil {
		return err
	}

	return nil
}

// getNIMCacheProfile returns model profile info from the NIM cache instance.
func (r *NIMServiceReconciler) getNIMCacheProfile(ctx context.Context, nimService *appsv1alpha1.NIMService, profile string) (*appsv1alpha1.NIMProfile, error) {
	logger := r.log

	if nimService.GetNIMCacheName() == "" {
		// NIM cache is not used
		return nil, nil
	}

	// Lookup NIMCache instance in the same namespace as the NIMService instance
	nimCache := &appsv1alpha1.NIMCache{}
	if err := r.Get(ctx, types.NamespacedName{Name: nimService.GetNIMCacheName(), Namespace: nimService.Namespace}, nimCache); err != nil {
		logger.Error(err, "unable to fetch nimcache", "nimcache", nimService.GetNIMCacheName(), "nimservice", nimService.Name)
		return nil, err
	}

	// Get the status of NIMCache
	if nimCache.Status.State != appsv1alpha1.NimCacheStatusReady {
		return nil, fmt.Errorf("nimcache %s is not ready, nimservice %s", nimCache.GetName(), nimService.GetName())
	}

	for _, cachedProfile := range nimCache.Status.Profiles {
		if cachedProfile.Name == profile {
			return &cachedProfile, nil
		}
	}

	// If the specified profile is not cached, return nil
	return nil, nil
}

// assignGPUResources automatically assigns GPU resources to the NIMService based on the provided profile,
// but retains any user-specified GPU resources if they are explicitly provided.
//
// This function retrieves the tensor parallelism (TP) value from the provided profile config to determine
// the number of GPUs to be allocated. If the TP value is defined and no GPU resources have been
// explicitly provided by the user, the function allocates GPUs according to the TP value.
// If the TP value is not present, the function defaults to allocating 1 GPU.
func (r *NIMServiceReconciler) assignGPUResources(ctx context.Context, nimService *appsv1alpha1.NIMService, profile *appsv1alpha1.NIMProfile, inferenceServiceParams *rendertypes.InferenceServiceParams) error {
	logger := r.log

	// TODO: Refine this to detect GPU claims and only assign GPU resources if no GPU claims are present.
	if len(nimService.Spec.DRAResources) > 0 {
		logger.Info("DRAResources found, skipping GPU resource assignment", "DRAResources", nimService.Spec.DRAResources)
		return nil
	}

	// TODO: Make the resource name configurable
	const gpuResourceName = corev1.ResourceName("nvidia.com/gpu")

	// Check if the user has already provided a GPU resource quantity in the requests or limits
	if inferenceServiceParams.Resources != nil {
		if _, gpuRequested := inferenceServiceParams.Resources.Requests[gpuResourceName]; gpuRequested {
			logger.V(2).Info("User has provided GPU resource requests, skipping auto-assignment", "gpuResource", gpuResourceName)
			return nil
		}
		if _, gpuLimit := inferenceServiceParams.Resources.Limits[gpuResourceName]; gpuLimit {
			logger.V(2).Info("User has provided GPU resource limits, skipping auto-assignment", "gpuResource", gpuResourceName)
			return nil
		}
	}

	// If no user-provided GPU resource is found, proceed with auto-assignment
	// Get tensorParallelism from the profile
	tensorParallelism, err := r.getTensorParallelismByProfile(profile)
	if err != nil {
		logger.Error(err, "Failed to retrieve tensorParallelism")
		return err
	}

	// Initialize the Resources field if not already initialized
	if inferenceServiceParams.Resources == nil {
		inferenceServiceParams.Resources = &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{},
			Limits:   corev1.ResourceList{},
		}
	}

	// Assign GPU resources based on tensorParallelism, or default to 1 GPU if tensorParallelism is not available
	gpuQuantity := apiResource.MustParse("1") // Default to 1 GPU

	if tensorParallelism != "" {
		gpuQuantity, err = apiResource.ParseQuantity(tensorParallelism)
		if err != nil {
			return fmt.Errorf("failed to parse tensorParallelism: %w", err)
		}

		logger.V(2).Info("Auto-assigning GPU resources based on tensorParallelism", "tensorParallelism", tensorParallelism, "gpuQuantity", gpuQuantity.String())
	} else {
		logger.V(2).Info("tensorParallelism not found, assigning 1 GPU by default", "Profile", profile.Name)
	}

	// Assign the GPU quantity for both requests and limits
	inferenceServiceParams.Resources.Requests[gpuResourceName] = gpuQuantity
	inferenceServiceParams.Resources.Limits[gpuResourceName] = gpuQuantity

	return nil
}

// getTensorParallelismByProfile returns the value of tensor parallelism parameter in the given NIM profile.
func (r *NIMServiceReconciler) getTensorParallelismByProfile(profile *appsv1alpha1.NIMProfile) (string, error) {
	// List of possible keys for tensor parallelism
	possibleKeys := []string{"tensorParallelism", "tp"}

	tensorParallelism := ""

	// Iterate through possible keys and return the first valid value
	for _, key := range possibleKeys {
		if value, exists := profile.Config[key]; exists {
			tensorParallelism = value
			break
		}
	}

	return tensorParallelism, nil
}

func (r *NIMServiceReconciler) isInferenceServiceReady(ctx context.Context, namespacedName *types.NamespacedName) {
}
