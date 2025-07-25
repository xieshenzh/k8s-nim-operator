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
// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	appsv1alpha1 "github.com/NVIDIA/k8s-nim-operator/api/apps/v1alpha1"
	internalinterfaces "github.com/NVIDIA/k8s-nim-operator/api/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/NVIDIA/k8s-nim-operator/api/listers/apps/v1alpha1"
	versioned "github.com/NVIDIA/k8s-nim-operator/api/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// NIMBuildInformer provides access to a shared informer and lister for
// NIMBuilds.
type NIMBuildInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.NIMBuildLister
}

type nIMBuildInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewNIMBuildInformer constructs a new informer for NIMBuild type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewNIMBuildInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredNIMBuildInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredNIMBuildInformer constructs a new informer for NIMBuild type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredNIMBuildInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AppsV1alpha1().NIMBuilds(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AppsV1alpha1().NIMBuilds(namespace).Watch(context.TODO(), options)
			},
		},
		&appsv1alpha1.NIMBuild{},
		resyncPeriod,
		indexers,
	)
}

func (f *nIMBuildInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredNIMBuildInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *nIMBuildInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&appsv1alpha1.NIMBuild{}, f.defaultInformer)
}

func (f *nIMBuildInformer) Lister() v1alpha1.NIMBuildLister {
	return v1alpha1.NewNIMBuildLister(f.Informer().GetIndexer())
}
