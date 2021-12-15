/*
Copyright 2021 The Kubermatic Kubernetes Platform contributors.

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

// Code generated by main. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ScannerLister helps list Scanners.
// All objects returned here must be treated as read-only.
type ScannerLister interface {
	// List lists all Scanners in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Scanner, err error)
	// Scanners returns an object that can list and get Scanners.
	Scanners(namespace string) ScannerNamespaceLister
	ScannerListerExpansion
}

// scannerLister implements the ScannerLister interface.
type scannerLister struct {
	indexer cache.Indexer
}

// NewScannerLister returns a new ScannerLister.
func NewScannerLister(indexer cache.Indexer) ScannerLister {
	return &scannerLister{indexer: indexer}
}

// List lists all Scanners in the indexer.
func (s *scannerLister) List(selector labels.Selector) (ret []*v1alpha1.Scanner, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Scanner))
	})
	return ret, err
}

// Scanners returns an object that can list and get Scanners.
func (s *scannerLister) Scanners(namespace string) ScannerNamespaceLister {
	return scannerNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ScannerNamespaceLister helps list and get Scanners.
// All objects returned here must be treated as read-only.
type ScannerNamespaceLister interface {
	// List lists all Scanners in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Scanner, err error)
	// Get retrieves the Scanner from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.Scanner, error)
	ScannerNamespaceListerExpansion
}

// scannerNamespaceLister implements the ScannerNamespaceLister
// interface.
type scannerNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Scanners in the indexer for a given namespace.
func (s scannerNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Scanner, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Scanner))
	})
	return ret, err
}

// Get retrieves the Scanner from the indexer for a given namespace and name.
func (s scannerNamespaceLister) Get(name string) (*v1alpha1.Scanner, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("scanner"), name)
	}
	return obj.(*v1alpha1.Scanner), nil
}
