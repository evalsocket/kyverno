/*
Copyright The Kubernetes Authors.

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
	kyvernov1 "github.com/nirmata/kyverno/pkg/api/kyverno/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeGenerateRequests implements GenerateRequestInterface
type FakeGenerateRequests struct {
	Fake *FakeKyvernoV1
	ns   string
}

var generaterequestsResource = schema.GroupVersionResource{Group: "kyverno.io", Version: "v1", Resource: "generaterequests"}

var generaterequestsKind = schema.GroupVersionKind{Group: "kyverno.io", Version: "v1", Kind: "GenerateRequest"}

// Get takes name of the generateRequest, and returns the corresponding generateRequest object, and an error if there is any.
func (c *FakeGenerateRequests) Get(name string, options v1.GetOptions) (result *kyvernov1.GenerateRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(generaterequestsResource, c.ns, name), &kyvernov1.GenerateRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kyvernov1.GenerateRequest), err
}

// List takes label and field selectors, and returns the list of GenerateRequests that match those selectors.
func (c *FakeGenerateRequests) List(opts v1.ListOptions) (result *kyvernov1.GenerateRequestList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(generaterequestsResource, generaterequestsKind, c.ns, opts), &kyvernov1.GenerateRequestList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &kyvernov1.GenerateRequestList{ListMeta: obj.(*kyvernov1.GenerateRequestList).ListMeta}
	for _, item := range obj.(*kyvernov1.GenerateRequestList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested generateRequests.
func (c *FakeGenerateRequests) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(generaterequestsResource, c.ns, opts))

}

// Create takes the representation of a generateRequest and creates it.  Returns the server's representation of the generateRequest, and an error, if there is any.
func (c *FakeGenerateRequests) Create(generateRequest *kyvernov1.GenerateRequest) (result *kyvernov1.GenerateRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(generaterequestsResource, c.ns, generateRequest), &kyvernov1.GenerateRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kyvernov1.GenerateRequest), err
}

// Update takes the representation of a generateRequest and updates it. Returns the server's representation of the generateRequest, and an error, if there is any.
func (c *FakeGenerateRequests) Update(generateRequest *kyvernov1.GenerateRequest) (result *kyvernov1.GenerateRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(generaterequestsResource, c.ns, generateRequest), &kyvernov1.GenerateRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kyvernov1.GenerateRequest), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeGenerateRequests) UpdateStatus(generateRequest *kyvernov1.GenerateRequest) (*kyvernov1.GenerateRequest, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(generaterequestsResource, "status", c.ns, generateRequest), &kyvernov1.GenerateRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kyvernov1.GenerateRequest), err
}

// Delete takes name of the generateRequest and deletes it. Returns an error if one occurs.
func (c *FakeGenerateRequests) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(generaterequestsResource, c.ns, name), &kyvernov1.GenerateRequest{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeGenerateRequests) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(generaterequestsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &kyvernov1.GenerateRequestList{})
	return err
}

// Patch applies the patch and returns the patched generateRequest.
func (c *FakeGenerateRequests) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kyvernov1.GenerateRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(generaterequestsResource, c.ns, name, pt, data, subresources...), &kyvernov1.GenerateRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kyvernov1.GenerateRequest), err
}
