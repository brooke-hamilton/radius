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
package aws

import (
	"context"
	"net/http"

	v1 "github.com/radius-project/radius/pkg/armrpc/api/v1"
	armrpc_controller "github.com/radius-project/radius/pkg/armrpc/frontend/controller"
	armrpc_rest "github.com/radius-project/radius/pkg/armrpc/rest"
	"github.com/radius-project/radius/pkg/components/secret"
	"github.com/radius-project/radius/pkg/ucp/datamodel"
	"github.com/radius-project/radius/pkg/ucp/datamodel/converter"
	"github.com/radius-project/radius/pkg/ucp/frontend/controller/credentials"
)

var _ armrpc_controller.Controller = (*CreateOrUpdateAWSCredential)(nil)

// CreateOrUpdateAWSCredential is the controller implementation to create/update a UCP AWS credential.
type CreateOrUpdateAWSCredential struct {
	armrpc_controller.Operation[*datamodel.AWSCredential, datamodel.AWSCredential]
	secretClient secret.Client
}

// NewCreateOrUpdateAWSCredential creates a new CreateOrUpdateAWSCredential controller which is used to create or update
// AWS credentials in the secret store.
func NewCreateOrUpdateAWSCredential(opts armrpc_controller.Options, secretClient secret.Client) (armrpc_controller.Controller, error) {
	return &CreateOrUpdateAWSCredential{
		Operation: armrpc_controller.NewOperation(opts,
			armrpc_controller.ResourceOptions[datamodel.AWSCredential]{
				RequestConverter:  converter.AWSCredentialDataModelFromVersioned,
				ResponseConverter: converter.AWSCredentialDataModelToVersioned,
			},
		),
		secretClient: secretClient,
	}, nil
}

// CreateOrUpdateAWSCredential validates the request, saves the AWS credential secret, and saves the resource in the
// metadata store. If an error occurs, it returns an error response.
func (c *CreateOrUpdateAWSCredential) Run(ctx context.Context, w http.ResponseWriter, req *http.Request) (armrpc_rest.Response, error) {
	serviceCtx := v1.ARMRequestContextFromContext(ctx)
	newResource, err := c.GetResourceFromRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if newResource.Properties.Kind != datamodel.AWSAccessKeyCredentialKind && newResource.Properties.Kind != datamodel.AWSIRSACredentialKind {
		return armrpc_rest.NewBadRequestResponse("Invalid Credential Kind"), nil
	}

	old, etag, err := c.GetResource(ctx, serviceCtx.ResourceID)
	if err != nil {
		return nil, err
	}

	if r, err := c.PrepareResource(ctx, req, newResource, old, etag); r != nil || err != nil {
		return r, err
	}

	secretName := credentials.GetSecretName(serviceCtx.ResourceID)
	if newResource.Properties.Storage.Kind == datamodel.InternalStorageKind {
		newResource.Properties.Storage.InternalCredential.SecretName = secretName
	}

	// Save the credential secret
	err = secret.SaveSecret(ctx, c.secretClient, secretName, newResource.Properties.AWSCredential)
	if err != nil {
		return nil, err
	}

	// Do not save the secret in metadata store.
	if newResource.Properties.AWSCredential.Kind == datamodel.AWSAccessKeyCredentialKind {
		newResource.Properties.AWSCredential.AccessKeyCredential.SecretAccessKey = ""
	}

	newResource.SetProvisioningState(v1.ProvisioningStateSucceeded)
	newEtag, err := c.SaveResource(ctx, serviceCtx.ResourceID.String(), newResource, etag)
	if err != nil {
		return nil, err
	}

	return c.ConstructSyncResponse(ctx, req.Method, newEtag, newResource)
}
