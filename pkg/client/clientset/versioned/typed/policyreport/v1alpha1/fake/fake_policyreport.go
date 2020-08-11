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
	"context"

	v1alpha1 "github.com/nirmata/kyverno/pkg/api/policyreport/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePolicyReports implements PolicyReportInterface
type FakePolicyReports struct {
	Fake *FakePolicyV1alpha1
	ns   string
}

var policyreportsResource = schema.GroupVersionResource{Group: "policy.kubernetes.io", Version: "v1alpha1", Resource: "policyreports"}

var policyreportsKind = schema.GroupVersionKind{Group: "policy.kubernetes.io", Version: "v1alpha1", Kind: "PolicyReport"}

// Get takes name of the policyReport, and returns the corresponding policyReport object, and an error if there is any.
func (c *FakePolicyReports) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.PolicyReport, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(policyreportsResource, c.ns, name), &v1alpha1.PolicyReport{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PolicyReport), err
}

// List takes label and field selectors, and returns the list of PolicyReports that match those selectors.
func (c *FakePolicyReports) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.PolicyReportList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(policyreportsResource, policyreportsKind, c.ns, opts), &v1alpha1.PolicyReportList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.PolicyReportList{ListMeta: obj.(*v1alpha1.PolicyReportList).ListMeta}
	for _, item := range obj.(*v1alpha1.PolicyReportList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested policyReports.
func (c *FakePolicyReports) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(policyreportsResource, c.ns, opts))

}

// Create takes the representation of a policyReport and creates it.  Returns the server's representation of the policyReport, and an error, if there is any.
func (c *FakePolicyReports) Create(ctx context.Context, policyReport *v1alpha1.PolicyReport, opts v1.CreateOptions) (result *v1alpha1.PolicyReport, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(policyreportsResource, c.ns, policyReport), &v1alpha1.PolicyReport{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PolicyReport), err
}

// Update takes the representation of a policyReport and updates it. Returns the server's representation of the policyReport, and an error, if there is any.
func (c *FakePolicyReports) Update(ctx context.Context, policyReport *v1alpha1.PolicyReport, opts v1.UpdateOptions) (result *v1alpha1.PolicyReport, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(policyreportsResource, c.ns, policyReport), &v1alpha1.PolicyReport{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PolicyReport), err
}

// Delete takes name of the policyReport and deletes it. Returns an error if one occurs.
func (c *FakePolicyReports) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(policyreportsResource, c.ns, name), &v1alpha1.PolicyReport{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePolicyReports) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(policyreportsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.PolicyReportList{})
	return err
}

// Patch applies the patch and returns the patched policyReport.
func (c *FakePolicyReports) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.PolicyReport, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(policyreportsResource, c.ns, name, pt, data, subresources...), &v1alpha1.PolicyReport{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PolicyReport), err
}
