/*
Copyright 2018 The Rook Authors. All rights reserved.

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
package fake

import (
	v1alpha1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeObjectStores implements ObjectStoreInterface
type FakeObjectStores struct {
	Fake *FakeCephV1alpha1
	ns   string
}

var objectstoresResource = schema.GroupVersionResource{Group: "ceph.rook.io", Version: "v1alpha1", Resource: "objectstores"}

var objectstoresKind = schema.GroupVersionKind{Group: "ceph.rook.io", Version: "v1alpha1", Kind: "ObjectStore"}

// Get takes name of the objectStore, and returns the corresponding objectStore object, and an error if there is any.
func (c *FakeObjectStores) Get(name string, options v1.GetOptions) (result *v1alpha1.ObjectStore, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(objectstoresResource, c.ns, name), &v1alpha1.ObjectStore{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ObjectStore), err
}

// List takes label and field selectors, and returns the list of ObjectStores that match those selectors.
func (c *FakeObjectStores) List(opts v1.ListOptions) (result *v1alpha1.ObjectStoreList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(objectstoresResource, objectstoresKind, c.ns, opts), &v1alpha1.ObjectStoreList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ObjectStoreList{}
	for _, item := range obj.(*v1alpha1.ObjectStoreList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested objectStores.
func (c *FakeObjectStores) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(objectstoresResource, c.ns, opts))

}

// Create takes the representation of a objectStore and creates it.  Returns the server's representation of the objectStore, and an error, if there is any.
func (c *FakeObjectStores) Create(objectStore *v1alpha1.ObjectStore) (result *v1alpha1.ObjectStore, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(objectstoresResource, c.ns, objectStore), &v1alpha1.ObjectStore{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ObjectStore), err
}

// Update takes the representation of a objectStore and updates it. Returns the server's representation of the objectStore, and an error, if there is any.
func (c *FakeObjectStores) Update(objectStore *v1alpha1.ObjectStore) (result *v1alpha1.ObjectStore, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(objectstoresResource, c.ns, objectStore), &v1alpha1.ObjectStore{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ObjectStore), err
}

// Delete takes name of the objectStore and deletes it. Returns an error if one occurs.
func (c *FakeObjectStores) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(objectstoresResource, c.ns, name), &v1alpha1.ObjectStore{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeObjectStores) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(objectstoresResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ObjectStoreList{})
	return err
}

// Patch applies the patch and returns the patched objectStore.
func (c *FakeObjectStores) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ObjectStore, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(objectstoresResource, c.ns, name, data, subresources...), &v1alpha1.ObjectStore{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ObjectStore), err
}
