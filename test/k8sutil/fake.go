/*
Copyright 2023 The Radius Authors.

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

package k8sutil

import (
	"context"

	openapi_v2 "github.com/google/gnostic-models/openapiv2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/openapi"
	"k8s.io/client-go/rest"
	clienttesting "k8s.io/client-go/testing"
)

// NewFakeKubeClient create new fake kube dynamic client with the given scheme and initial objects.
func NewFakeKubeClient(scheme *runtime.Scheme, initObjs ...client.Object) client.WithWatch {
	builder := fake.NewClientBuilder()
	if scheme != nil {
		builder = builder.WithScheme(scheme)
	}
	return &testClient{builder.WithObjects(initObjs...).Build()}
}

// PrependPatchReactor prepends patch reactor to fake client. This is workaround because clientset
// fake doesn't support patch verb. https://github.com/kubernetes/client-go/issues/1184
func PrependPatchReactor(f *k8sfake.Clientset, resource string, objFunc func(clienttesting.PatchAction) runtime.Object) {
	f.PrependReactor(
		"patch",
		resource,
		func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
			pa := action.(clienttesting.PatchAction)
			if pa.GetPatchType() == types.ApplyPatchType {
				rfunc := clienttesting.ObjectReaction(f.Tracker())
				_, obj, err := rfunc(
					clienttesting.NewGetAction(pa.GetResource(), pa.GetNamespace(), pa.GetName()),
				)
				if apierrors.IsNotFound(err) || obj == nil {
					_, _, _ = rfunc(
						clienttesting.NewCreateAction(
							pa.GetResource(),
							pa.GetNamespace(),
							objFunc(pa),
						),
					)
				}
				return rfunc(clienttesting.NewPatchAction(
					pa.GetResource(),
					pa.GetNamespace(),
					pa.GetName(),
					types.StrategicMergePatchType,
					pa.GetPatch()))
			}
			return false, nil, nil
		},
	)
}

// testClient is a fake client that implements the Patch method.
type testClient struct {
	client.WithWatch
}

// Patch implements client.Patch for apply patches. It checks if the patch type is Apply, then attempts to get
// the object, create it if it doesn't exist, or update it if it does. If an error is encountered, it is returned.
func (c *testClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	if patch.Type() != client.Apply.Type() {
		return c.WithWatch.Patch(ctx, obj, patch, opts...)
	}

	// This is not exactly the same as the real implementation, but it's good enough for our tests.
	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())
	err := c.WithWatch.Get(ctx, client.ObjectKey{Name: obj.GetName(), Namespace: obj.GetNamespace()}, existing)
	if client.IgnoreNotFound(err) != nil {
		return err
	} else if err != nil {
		return c.WithWatch.Create(ctx, obj)
	} else {
		return c.WithWatch.Update(ctx, obj)
	}
}

type DiscoveryClient struct {
	Groups    *metav1.APIGroupList
	Resources []*metav1.APIResourceList
	APIGroup  []*metav1.APIGroup
}

// OpenAPISchema implements discovery.DiscoveryInterface.
func (d *DiscoveryClient) OpenAPISchema() (*openapi_v2.Document, error) {
	panic("unimplemented")
}

// OpenAPIV3 implements discovery.DiscoveryInterface.
func (d *DiscoveryClient) OpenAPIV3() openapi.Client {
	panic("unimplemented")
}

// RESTClient implements discovery.DiscoveryInterface.
func (d *DiscoveryClient) RESTClient() rest.Interface {
	panic("unimplemented")
}

// ServerVersion implements discovery.DiscoveryInterface.
func (d *DiscoveryClient) ServerVersion() (*version.Info, error) {
	panic("unimplemented")
}

// WithLegacy implements discovery.DiscoveryInterface.
func (d *DiscoveryClient) WithLegacy() discovery.DiscoveryInterface {
	panic("unimplemented")
}

// ServerGroups returns a list of API groups supported by the server.
func (d *DiscoveryClient) ServerGroups() (*metav1.APIGroupList, error) {
	return d.Groups, nil
}

// This function returns a slice of API resource lists.
func (d *DiscoveryClient) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	return d.Resources, nil
}

// This function returns a slice of API resource lists.
func (d *DiscoveryClient) ServerPreferredNamespacedResources() ([]*metav1.APIResourceList, error) {
	return d.Resources, nil
}

// ServerGroupsAndResources returns a list of API groups and resources associated with the discovery client.
func (d *DiscoveryClient) ServerGroupsAndResources() ([]*metav1.APIGroup, []*metav1.APIResourceList, error) {
	return d.APIGroup, d.Resources, nil
}

// ServerResourcesForGroupVersion returns nil for the API resource list.
func (d *DiscoveryClient) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	return nil, nil
}
