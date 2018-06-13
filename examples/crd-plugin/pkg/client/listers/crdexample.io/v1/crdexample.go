// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/ligato/cn-infra/examples/crd-plugin/pkg/apis/crdexample.io/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// CrdExampleLister helps list CrdExamples.
type CrdExampleLister interface {
	// List lists all CrdExamples in the indexer.
	List(selector labels.Selector) (ret []*v1.CrdExample, err error)
	// CrdExamples returns an object that can list and get CrdExamples.
	CrdExamples(namespace string) CrdExampleNamespaceLister
	CrdExampleListerExpansion
}

// crdExampleLister implements the CrdExampleLister interface.
type crdExampleLister struct {
	indexer cache.Indexer
}

// NewCrdExampleLister returns a new CrdExampleLister.
func NewCrdExampleLister(indexer cache.Indexer) CrdExampleLister {
	return &crdExampleLister{indexer: indexer}
}

// List lists all CrdExamples in the indexer.
func (s *crdExampleLister) List(selector labels.Selector) (ret []*v1.CrdExample, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.CrdExample))
	})
	return ret, err
}

// CrdExamples returns an object that can list and get CrdExamples.
func (s *crdExampleLister) CrdExamples(namespace string) CrdExampleNamespaceLister {
	return crdExampleNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// CrdExampleNamespaceLister helps list and get CrdExamples.
type CrdExampleNamespaceLister interface {
	// List lists all CrdExamples in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1.CrdExample, err error)
	// Get retrieves the CrdExample from the indexer for a given namespace and name.
	Get(name string) (*v1.CrdExample, error)
	CrdExampleNamespaceListerExpansion
}

// crdExampleNamespaceLister implements the CrdExampleNamespaceLister
// interface.
type crdExampleNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all CrdExamples in the indexer for a given namespace.
func (s crdExampleNamespaceLister) List(selector labels.Selector) (ret []*v1.CrdExample, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.CrdExample))
	})
	return ret, err
}

// Get retrieves the CrdExample from the indexer for a given namespace and name.
func (s crdExampleNamespaceLister) Get(name string) (*v1.CrdExample, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("crdexample"), name)
	}
	return obj.(*v1.CrdExample), nil
}
