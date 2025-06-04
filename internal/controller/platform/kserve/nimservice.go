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

	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "github.com/NVIDIA/k8s-nim-operator/api/apps/v1alpha1"
	"github.com/NVIDIA/k8s-nim-operator/internal/conditions"
	"github.com/NVIDIA/k8s-nim-operator/internal/k8sutil"
	"github.com/NVIDIA/k8s-nim-operator/internal/render"
	"github.com/NVIDIA/k8s-nim-operator/internal/shared"
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
	// TODO: add cleanup logic specific to Kserve modelcache
	return nil
}

func (r *NIMServiceReconciler) reconcileNIMService(ctx context.Context, nimService *appsv1alpha1.NIMService) (ctrl.Result, error) {
	// TODO
	return ctrl.Result{}, nil
}
