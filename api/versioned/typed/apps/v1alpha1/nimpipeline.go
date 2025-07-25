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

// NIMPipelinesGetter has a method to return a NIMPipelineInterface.
// A group's client should implement this interface.
type NIMPipelinesGetter interface {
	NIMPipelines(namespace string) NIMPipelineInterface
}

// NIMPipelineInterface has methods to work with NIMPipeline resources.
type NIMPipelineInterface interface {
	Create(ctx context.Context, nIMPipeline *v1alpha1.NIMPipeline, opts v1.CreateOptions) (*v1alpha1.NIMPipeline, error)
	Update(ctx context.Context, nIMPipeline *v1alpha1.NIMPipeline, opts v1.UpdateOptions) (*v1alpha1.NIMPipeline, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, nIMPipeline *v1alpha1.NIMPipeline, opts v1.UpdateOptions) (*v1alpha1.NIMPipeline, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.NIMPipeline, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.NIMPipelineList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.NIMPipeline, err error)
	NIMPipelineExpansion
}

// nIMPipelines implements NIMPipelineInterface
type nIMPipelines struct {
	*gentype.ClientWithList[*v1alpha1.NIMPipeline, *v1alpha1.NIMPipelineList]
}

// newNIMPipelines returns a NIMPipelines
func newNIMPipelines(c *AppsV1alpha1Client, namespace string) *nIMPipelines {
	return &nIMPipelines{
		gentype.NewClientWithList[*v1alpha1.NIMPipeline, *v1alpha1.NIMPipelineList](
			"nimpipelines",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1alpha1.NIMPipeline { return &v1alpha1.NIMPipeline{} },
			func() *v1alpha1.NIMPipelineList { return &v1alpha1.NIMPipelineList{} }),
	}
}
