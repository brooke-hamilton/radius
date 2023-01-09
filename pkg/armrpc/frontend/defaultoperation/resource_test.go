// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package defaultoperation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/golang/mock/gomock"
	"github.com/project-radius/radius/pkg/armrpc/api/conv"
	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
	"github.com/project-radius/radius/pkg/armrpc/asyncoperation/statusmanager"
	"github.com/project-radius/radius/pkg/armrpc/frontend/controller"
	"github.com/project-radius/radius/pkg/armrpc/rest"
	radiustesting "github.com/project-radius/radius/pkg/corerp/testing"
	"github.com/project-radius/radius/pkg/ucp/store"
)

const (
	resourceTestHeaderFile        = "resource_requestheaders.json"
	operationStatusTestHeaderFile = "operationstatus_requestheaders.json"
	testAPIVersion                = "2022-03-15-privatepreview"
)

// TestResourceDataModel represents test resource.
type TestResourceDataModel struct {
	v1.BaseResource

	// Properties is the properties of the resource.
	Properties *TestResourceDataModelProperties `json:"properties"`
}

// ResourceTypeName returns the qualified name of the resource
func (r *TestResourceDataModel) ResourceTypeName() string {
	return "Applications.Core/resources"
}

// TestResourceDataModelProperties represents the properties of TestResourceDataModel.
type TestResourceDataModelProperties struct {
	Application string `json:"application"`
	Environment string `json:"environment"`
	PropertyA   string `json:"propertyA,omitempty"`
	PropertyB   string `json:"propertyB,omitempty"`
}

// TestResource represents test resource for api version.
type TestResource struct {
	ID         *string                 `json:"id,omitempty"`
	Name       *string                 `json:"name,omitempty"`
	SystemData *v1.SystemData          `json:"systemData,omitempty"`
	Type       *string                 `json:"type,omitempty"`
	Location   *string                 `json:"location,omitempty"`
	Properties *TestResourceProperties `json:"properties,omitempty"`
	Tags       map[string]*string      `json:"tags,omitempty"`
}

// TestResourceProperties - HTTP Route properties
type TestResourceProperties struct {
	ProvisioningState *v1.ProvisioningState `json:"provisioningState,omitempty"`
	Environment       *string               `json:"environment,omitempty"`
	Application       *string               `json:"application,omitempty"`
	PropertyA         *string               `json:"propertyA,omitempty"`
	PropertyB         *string               `json:"propertyB,omitempty"`
}

func (src *TestResource) ConvertTo() (conv.DataModelInterface, error) {
	converted := &TestResourceDataModel{
		BaseResource: v1.BaseResource{
			TrackedResource: v1.TrackedResource{
				ID:       to.String(src.ID),
				Name:     to.String(src.Name),
				Type:     to.String(src.Type),
				Location: to.String(src.Location),
				Tags:     to.StringMap(src.Tags),
			},
			InternalMetadata: v1.InternalMetadata{
				UpdatedAPIVersion:      testAPIVersion,
				AsyncProvisioningState: toProvisioningStateDataModel(src.Properties.ProvisioningState),
			},
		},
		Properties: &TestResourceDataModelProperties{
			Application: to.String(src.Properties.Application),
			Environment: to.String(src.Properties.Environment),
			PropertyA:   to.String(src.Properties.PropertyA),
			PropertyB:   to.String(src.Properties.PropertyB),
		},
	}
	return converted, nil
}

func (dst *TestResource) ConvertFrom(src conv.DataModelInterface) error {
	dm, ok := src.(*TestResourceDataModel)
	if !ok {
		return conv.ErrInvalidModelConversion
	}

	dst.ID = to.StringPtr(dm.ID)
	dst.Name = to.StringPtr(dm.Name)
	dst.Type = to.StringPtr(dm.Type)
	dst.SystemData = &dm.SystemData
	dst.Location = to.StringPtr(dm.Location)
	dst.Tags = *to.StringMapPtr(dm.Tags)
	dst.Properties = &TestResourceProperties{
		ProvisioningState: fromProvisioningStateDataModel(dm.InternalMetadata.AsyncProvisioningState),
		Environment:       to.StringPtr(dm.Properties.Environment),
		Application:       to.StringPtr(dm.Properties.Application),
		PropertyA:         to.StringPtr(dm.Properties.PropertyA),
		PropertyB:         to.StringPtr(dm.Properties.PropertyB),
	}

	return nil
}

func toProvisioningStateDataModel(state *v1.ProvisioningState) v1.ProvisioningState {
	if state == nil {
		return v1.ProvisioningStateAccepted
	}
	return *state
}

func fromProvisioningStateDataModel(state v1.ProvisioningState) *v1.ProvisioningState {
	converted := v1.ProvisioningStateAccepted
	if state != "" {
		converted = state
	}

	return &converted
}

func testResourceDataModelToVersioned(model *TestResourceDataModel, version string) (conv.VersionedModelInterface, error) {
	switch version {
	case testAPIVersion:
		versioned := &TestResource{}
		err := versioned.ConvertFrom(model)
		return versioned, err

	default:
		return nil, v1.ErrUnsupportedAPIVersion
	}
}

func testResourceDataModelFromVersioned(content []byte, version string) (*TestResourceDataModel, error) {
	switch version {
	case testAPIVersion:
		am := &TestResource{}
		if err := json.Unmarshal(content, am); err != nil {
			return nil, err
		}
		dm, err := am.ConvertTo()
		return dm.(*TestResourceDataModel), err

	default:
		return nil, v1.ErrUnsupportedAPIVersion
	}
}

// testValidateRequest is an example resource filter.
//
// In this case we're validating that the application of an existing resource can't change. This is one of our scenarios
// for the corerp and linkrp. However we're avoiding calling into that code directly from here to avoid coupling.
func testValidateRequest(ctx context.Context, newResource *TestResourceDataModel, oldResource *TestResourceDataModel, options *controller.Options) (rest.Response, error) {
	if oldResource == nil {
		return nil, nil
	}

	if newResource.Properties.Application != oldResource.Properties.Application {
		return rest.NewBadRequestResponse("Oh no!"), nil
	}

	return nil, nil
}

func loadTestResurce() (*TestResource, *TestResourceDataModel, *TestResource) {
	reqBody := radiustesting.ReadFixture("resource-request.json")
	reqModel := &TestResource{}
	_ = json.Unmarshal(reqBody, reqModel)

	rawDataModel := radiustesting.ReadFixture("resource-datamodel.json")
	datamodel := &TestResourceDataModel{}
	_ = json.Unmarshal(rawDataModel, datamodel)

	respBody := radiustesting.ReadFixture("resource-response.json")
	respModel := &TestResource{}
	_ = json.Unmarshal(respBody, respModel)

	return reqModel, datamodel, respModel
}

func setupTest(tb testing.TB) (func(testing.TB), *store.MockStorageClient, *statusmanager.MockStatusManager) {
	mctrl := gomock.NewController(tb)
	mds := store.NewMockStorageClient(mctrl)
	msm := statusmanager.NewMockStatusManager(mctrl)

	return func(tb testing.TB) {
		mctrl.Finish()
	}, mds, msm
}

// TODO: Use Referer header instead of X-Forwarded-Proto by following ARM RPC spec - https://github.com/project-radius/radius/issues/3068
func getAsyncLocationPath(sCtx *v1.ARMRequestContext, location string, resourceType string, req *http.Request) string {
	dest := url.URL{
		Host:   req.Host,
		Scheme: req.URL.Scheme,
		Path: fmt.Sprintf("%s/providers/%s/locations/%s/%s/%s", sCtx.ResourceID.PlaneScope(),
			sCtx.ResourceID.ProviderNamespace(), location, resourceType, sCtx.OperationID.String()),
	}

	query := url.Values{}
	query.Add("api-version", sCtx.APIVersion)
	dest.RawQuery = query.Encode()

	protocol := req.Header.Get("X-Forwarded-Proto")
	if protocol != "" {
		dest.Scheme = protocol
	}

	if dest.Scheme == "" {
		dest.Scheme = "http"
	}

	return dest.String()
}