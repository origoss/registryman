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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeScanners implements ScannerInterface
type FakeScanners struct {
	Fake *FakeRegistrymanV1alpha1
	ns   string
}

var scannersResource = schema.GroupVersionResource{Group: "registryman.kubermatic.com", Version: "v1alpha1", Resource: "scanners"}

var scannersKind = schema.GroupVersionKind{Group: "registryman.kubermatic.com", Version: "v1alpha1", Kind: "Scanner"}

// Get takes name of the scanner, and returns the corresponding scanner object, and an error if there is any.
func (c *FakeScanners) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Scanner, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(scannersResource, c.ns, name), &v1alpha1.Scanner{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Scanner), err
}

// List takes label and field selectors, and returns the list of Scanners that match those selectors.
func (c *FakeScanners) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ScannerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(scannersResource, scannersKind, c.ns, opts), &v1alpha1.ScannerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ScannerList{ListMeta: obj.(*v1alpha1.ScannerList).ListMeta}
	for _, item := range obj.(*v1alpha1.ScannerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested scanners.
func (c *FakeScanners) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(scannersResource, c.ns, opts))

}

// Create takes the representation of a scanner and creates it.  Returns the server's representation of the scanner, and an error, if there is any.
func (c *FakeScanners) Create(ctx context.Context, scanner *v1alpha1.Scanner, opts v1.CreateOptions) (result *v1alpha1.Scanner, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(scannersResource, c.ns, scanner), &v1alpha1.Scanner{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Scanner), err
}

// Update takes the representation of a scanner and updates it. Returns the server's representation of the scanner, and an error, if there is any.
func (c *FakeScanners) Update(ctx context.Context, scanner *v1alpha1.Scanner, opts v1.UpdateOptions) (result *v1alpha1.Scanner, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(scannersResource, c.ns, scanner), &v1alpha1.Scanner{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Scanner), err
}

// Delete takes name of the scanner and deletes it. Returns an error if one occurs.
func (c *FakeScanners) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(scannersResource, c.ns, name, opts), &v1alpha1.Scanner{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeScanners) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(scannersResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.ScannerList{})
	return err
}

// Patch applies the patch and returns the patched scanner.
func (c *FakeScanners) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Scanner, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(scannersResource, c.ns, name, pt, data, subresources...), &v1alpha1.Scanner{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Scanner), err
}
