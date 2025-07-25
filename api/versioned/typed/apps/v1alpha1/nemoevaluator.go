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
// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"

	v1alpha1 "github.com/NVIDIA/k8s-nim-operator/api/apps/v1alpha1"
	scheme "github.com/NVIDIA/k8s-nim-operator/api/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// NemoEvaluatorsGetter has a method to return a NemoEvaluatorInterface.
// A group's client should implement this interface.
type NemoEvaluatorsGetter interface {
	NemoEvaluators(namespace string) NemoEvaluatorInterface
}

// NemoEvaluatorInterface has methods to work with NemoEvaluator resources.
type NemoEvaluatorInterface interface {
	Create(ctx context.Context, nemoEvaluator *v1alpha1.NemoEvaluator, opts v1.CreateOptions) (*v1alpha1.NemoEvaluator, error)
	Update(ctx context.Context, nemoEvaluator *v1alpha1.NemoEvaluator, opts v1.UpdateOptions) (*v1alpha1.NemoEvaluator, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, nemoEvaluator *v1alpha1.NemoEvaluator, opts v1.UpdateOptions) (*v1alpha1.NemoEvaluator, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.NemoEvaluator, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.NemoEvaluatorList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.NemoEvaluator, err error)
	NemoEvaluatorExpansion
}

// nemoEvaluators implements NemoEvaluatorInterface
type nemoEvaluators struct {
	*gentype.ClientWithList[*v1alpha1.NemoEvaluator, *v1alpha1.NemoEvaluatorList]
}

// newNemoEvaluators returns a NemoEvaluators
func newNemoEvaluators(c *AppsV1alpha1Client, namespace string) *nemoEvaluators {
	return &nemoEvaluators{
		gentype.NewClientWithList[*v1alpha1.NemoEvaluator, *v1alpha1.NemoEvaluatorList](
			"nemoevaluators",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1alpha1.NemoEvaluator { return &v1alpha1.NemoEvaluator{} },
			func() *v1alpha1.NemoEvaluatorList { return &v1alpha1.NemoEvaluatorList{} }),
	}
}
