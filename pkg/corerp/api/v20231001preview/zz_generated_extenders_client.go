//go:build go1.18
// +build go1.18

// Licensed under the Apache License, Version 2.0 . See LICENSE in the repository root for license information.
// Code generated by Microsoft (R) AutoRest Code Generator. DO NOT EDIT.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

package v20231001preview

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"net/http"
	"net/url"
	"strings"
)

// ExtendersClient contains the methods for the Extenders group.
// Don't use this type directly, use NewExtendersClient() instead.
type ExtendersClient struct {
	internal *arm.Client
	rootScope string
}

// NewExtendersClient creates a new instance of ExtendersClient with the specified values.
//   - rootScope - The scope in which the resource is present. UCP Scope is /planes/{planeType}/{planeName}/resourceGroup/{resourcegroupID}
//     and Azure resource scope is
//     /subscriptions/{subscriptionID}/resourceGroup/{resourcegroupID}
//   - credential - used to authorize requests. Usually a credential from azidentity.
//   - options - pass nil to accept the default values.
func NewExtendersClient(rootScope string, credential azcore.TokenCredential, options *arm.ClientOptions) (*ExtendersClient, error) {
	cl, err := arm.NewClient(moduleName+".ExtendersClient", moduleVersion, credential, options)
	if err != nil {
		return nil, err
	}
	client := &ExtendersClient{
		rootScope: rootScope,
	internal: cl,
	}
	return client, nil
}

// BeginCreateOrUpdate - Create a ExtenderResource
// If the operation fails it returns an *azcore.ResponseError type.
//
// Generated from API version 2023-10-01-preview
//   - extenderName - The name of the ExtenderResource portable resource
//   - resource - Resource create parameters.
//   - options - ExtendersClientBeginCreateOrUpdateOptions contains the optional parameters for the ExtendersClient.BeginCreateOrUpdate
//     method.
func (client *ExtendersClient) BeginCreateOrUpdate(ctx context.Context, extenderName string, resource ExtenderResource, options *ExtendersClientBeginCreateOrUpdateOptions) (*runtime.Poller[ExtendersClientCreateOrUpdateResponse], error) {
	if options == nil || options.ResumeToken == "" {
		resp, err := client.createOrUpdate(ctx, extenderName, resource, options)
		if err != nil {
			return nil, err
		}
		poller, err := runtime.NewPoller(resp, client.internal.Pipeline(), &runtime.NewPollerOptions[ExtendersClientCreateOrUpdateResponse]{
			FinalStateVia: runtime.FinalStateViaAzureAsyncOp,
		})
		return poller, err
	} else {
		return runtime.NewPollerFromResumeToken[ExtendersClientCreateOrUpdateResponse](options.ResumeToken, client.internal.Pipeline(), nil)
	}
}

// CreateOrUpdate - Create a ExtenderResource
// If the operation fails it returns an *azcore.ResponseError type.
//
// Generated from API version 2023-10-01-preview
func (client *ExtendersClient) createOrUpdate(ctx context.Context, extenderName string, resource ExtenderResource, options *ExtendersClientBeginCreateOrUpdateOptions) (*http.Response, error) {
	var err error
	req, err := client.createOrUpdateCreateRequest(ctx, extenderName, resource, options)
	if err != nil {
		return nil, err
	}
	httpResp, err := client.internal.Pipeline().Do(req)
	if err != nil {
		return nil, err
	}
	if !runtime.HasStatusCode(httpResp, http.StatusOK, http.StatusCreated) {
		err = runtime.NewResponseError(httpResp)
		return nil, err
	}
	return httpResp, nil
}

// createOrUpdateCreateRequest creates the CreateOrUpdate request.
func (client *ExtendersClient) createOrUpdateCreateRequest(ctx context.Context, extenderName string, resource ExtenderResource, options *ExtendersClientBeginCreateOrUpdateOptions) (*policy.Request, error) {
	urlPath := "/{rootScope}/providers/Applications.Core/extenders/{extenderName}"
	urlPath = strings.ReplaceAll(urlPath, "{rootScope}", client.rootScope)
	if extenderName == "" {
		return nil, errors.New("parameter extenderName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{extenderName}", url.PathEscape(extenderName))
	req, err := runtime.NewRequest(ctx, http.MethodPut, runtime.JoinPaths(client.internal.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2023-10-01-preview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header["Accept"] = []string{"application/json"}
	if err := runtime.MarshalAsJSON(req, resource); err != nil {
	return nil, err
}
	return req, nil
}

// BeginDelete - Delete a ExtenderResource
// If the operation fails it returns an *azcore.ResponseError type.
//
// Generated from API version 2023-10-01-preview
//   - extenderName - The name of the ExtenderResource portable resource
//   - options - ExtendersClientBeginDeleteOptions contains the optional parameters for the ExtendersClient.BeginDelete method.
func (client *ExtendersClient) BeginDelete(ctx context.Context, extenderName string, options *ExtendersClientBeginDeleteOptions) (*runtime.Poller[ExtendersClientDeleteResponse], error) {
	if options == nil || options.ResumeToken == "" {
		resp, err := client.deleteOperation(ctx, extenderName, options)
		if err != nil {
			return nil, err
		}
		poller, err := runtime.NewPoller(resp, client.internal.Pipeline(), &runtime.NewPollerOptions[ExtendersClientDeleteResponse]{
			FinalStateVia: runtime.FinalStateViaLocation,
		})
		return poller, err
	} else {
		return runtime.NewPollerFromResumeToken[ExtendersClientDeleteResponse](options.ResumeToken, client.internal.Pipeline(), nil)
	}
}

// Delete - Delete a ExtenderResource
// If the operation fails it returns an *azcore.ResponseError type.
//
// Generated from API version 2023-10-01-preview
func (client *ExtendersClient) deleteOperation(ctx context.Context, extenderName string, options *ExtendersClientBeginDeleteOptions) (*http.Response, error) {
	var err error
	req, err := client.deleteCreateRequest(ctx, extenderName, options)
	if err != nil {
		return nil, err
	}
	httpResp, err := client.internal.Pipeline().Do(req)
	if err != nil {
		return nil, err
	}
	if !runtime.HasStatusCode(httpResp, http.StatusOK, http.StatusAccepted, http.StatusNoContent) {
		err = runtime.NewResponseError(httpResp)
		return nil, err
	}
	return httpResp, nil
}

// deleteCreateRequest creates the Delete request.
func (client *ExtendersClient) deleteCreateRequest(ctx context.Context, extenderName string, options *ExtendersClientBeginDeleteOptions) (*policy.Request, error) {
	urlPath := "/{rootScope}/providers/Applications.Core/extenders/{extenderName}"
	urlPath = strings.ReplaceAll(urlPath, "{rootScope}", client.rootScope)
	if extenderName == "" {
		return nil, errors.New("parameter extenderName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{extenderName}", url.PathEscape(extenderName))
	req, err := runtime.NewRequest(ctx, http.MethodDelete, runtime.JoinPaths(client.internal.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2023-10-01-preview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header["Accept"] = []string{"application/json"}
	return req, nil
}

// Get - Get a ExtenderResource
// If the operation fails it returns an *azcore.ResponseError type.
//
// Generated from API version 2023-10-01-preview
//   - extenderName - The name of the ExtenderResource portable resource
//   - options - ExtendersClientGetOptions contains the optional parameters for the ExtendersClient.Get method.
func (client *ExtendersClient) Get(ctx context.Context, extenderName string, options *ExtendersClientGetOptions) (ExtendersClientGetResponse, error) {
	var err error
	req, err := client.getCreateRequest(ctx, extenderName, options)
	if err != nil {
		return ExtendersClientGetResponse{}, err
	}
	httpResp, err := client.internal.Pipeline().Do(req)
	if err != nil {
		return ExtendersClientGetResponse{}, err
	}
	if !runtime.HasStatusCode(httpResp, http.StatusOK) {
		err = runtime.NewResponseError(httpResp)
		return ExtendersClientGetResponse{}, err
	}
	resp, err := client.getHandleResponse(httpResp)
	return resp, err
}

// getCreateRequest creates the Get request.
func (client *ExtendersClient) getCreateRequest(ctx context.Context, extenderName string, options *ExtendersClientGetOptions) (*policy.Request, error) {
	urlPath := "/{rootScope}/providers/Applications.Core/extenders/{extenderName}"
	urlPath = strings.ReplaceAll(urlPath, "{rootScope}", client.rootScope)
	if extenderName == "" {
		return nil, errors.New("parameter extenderName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{extenderName}", url.PathEscape(extenderName))
	req, err := runtime.NewRequest(ctx, http.MethodGet, runtime.JoinPaths(client.internal.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2023-10-01-preview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header["Accept"] = []string{"application/json"}
	return req, nil
}

// getHandleResponse handles the Get response.
func (client *ExtendersClient) getHandleResponse(resp *http.Response) (ExtendersClientGetResponse, error) {
	result := ExtendersClientGetResponse{}
	if err := runtime.UnmarshalAsJSON(resp, &result.ExtenderResource); err != nil {
		return ExtendersClientGetResponse{}, err
	}
	return result, nil
}

// NewListByScopePager - List ExtenderResource resources by Scope
//
// Generated from API version 2023-10-01-preview
//   - options - ExtendersClientListByScopeOptions contains the optional parameters for the ExtendersClient.NewListByScopePager
//     method.
func (client *ExtendersClient) NewListByScopePager(options *ExtendersClientListByScopeOptions) (*runtime.Pager[ExtendersClientListByScopeResponse]) {
	return runtime.NewPager(runtime.PagingHandler[ExtendersClientListByScopeResponse]{
		More: func(page ExtendersClientListByScopeResponse) bool {
			return page.NextLink != nil && len(*page.NextLink) > 0
		},
		Fetcher: func(ctx context.Context, page *ExtendersClientListByScopeResponse) (ExtendersClientListByScopeResponse, error) {
			var req *policy.Request
			var err error
			if page == nil {
				req, err = client.listByScopeCreateRequest(ctx, options)
			} else {
				req, err = runtime.NewRequest(ctx, http.MethodGet, *page.NextLink)
			}
			if err != nil {
				return ExtendersClientListByScopeResponse{}, err
			}
			resp, err := client.internal.Pipeline().Do(req)
			if err != nil {
				return ExtendersClientListByScopeResponse{}, err
			}
			if !runtime.HasStatusCode(resp, http.StatusOK) {
				return ExtendersClientListByScopeResponse{}, runtime.NewResponseError(resp)
			}
			return client.listByScopeHandleResponse(resp)
		},
	})
}

// listByScopeCreateRequest creates the ListByScope request.
func (client *ExtendersClient) listByScopeCreateRequest(ctx context.Context, options *ExtendersClientListByScopeOptions) (*policy.Request, error) {
	urlPath := "/{rootScope}/providers/Applications.Core/extenders"
	urlPath = strings.ReplaceAll(urlPath, "{rootScope}", client.rootScope)
	req, err := runtime.NewRequest(ctx, http.MethodGet, runtime.JoinPaths(client.internal.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2023-10-01-preview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header["Accept"] = []string{"application/json"}
	return req, nil
}

// listByScopeHandleResponse handles the ListByScope response.
func (client *ExtendersClient) listByScopeHandleResponse(resp *http.Response) (ExtendersClientListByScopeResponse, error) {
	result := ExtendersClientListByScopeResponse{}
	if err := runtime.UnmarshalAsJSON(resp, &result.ExtenderResourceListResult); err != nil {
		return ExtendersClientListByScopeResponse{}, err
	}
	return result, nil
}

// ListSecrets - Lists secrets values for the specified Extender resource
// If the operation fails it returns an *azcore.ResponseError type.
//
// Generated from API version 2023-10-01-preview
//   - extenderName - The name of the ExtenderResource portable resource
//   - body - The content of the action request
//   - options - ExtendersClientListSecretsOptions contains the optional parameters for the ExtendersClient.ListSecrets method.
func (client *ExtendersClient) ListSecrets(ctx context.Context, extenderName string, body map[string]any, options *ExtendersClientListSecretsOptions) (ExtendersClientListSecretsResponse, error) {
	var err error
	req, err := client.listSecretsCreateRequest(ctx, extenderName, body, options)
	if err != nil {
		return ExtendersClientListSecretsResponse{}, err
	}
	httpResp, err := client.internal.Pipeline().Do(req)
	if err != nil {
		return ExtendersClientListSecretsResponse{}, err
	}
	if !runtime.HasStatusCode(httpResp, http.StatusOK) {
		err = runtime.NewResponseError(httpResp)
		return ExtendersClientListSecretsResponse{}, err
	}
	resp, err := client.listSecretsHandleResponse(httpResp)
	return resp, err
}

// listSecretsCreateRequest creates the ListSecrets request.
func (client *ExtendersClient) listSecretsCreateRequest(ctx context.Context, extenderName string, body map[string]any, options *ExtendersClientListSecretsOptions) (*policy.Request, error) {
	urlPath := "/{rootScope}/providers/Applications.Core/extenders/{extenderName}/listSecrets"
	urlPath = strings.ReplaceAll(urlPath, "{rootScope}", client.rootScope)
	if extenderName == "" {
		return nil, errors.New("parameter extenderName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{extenderName}", url.PathEscape(extenderName))
	req, err := runtime.NewRequest(ctx, http.MethodPost, runtime.JoinPaths(client.internal.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2023-10-01-preview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header["Accept"] = []string{"application/json"}
	if err := runtime.MarshalAsJSON(req, body); err != nil {
	return nil, err
}
	return req, nil
}

// listSecretsHandleResponse handles the ListSecrets response.
func (client *ExtendersClient) listSecretsHandleResponse(resp *http.Response) (ExtendersClientListSecretsResponse, error) {
	result := ExtendersClientListSecretsResponse{}
	if err := runtime.UnmarshalAsJSON(resp, &result.Object); err != nil {
		return ExtendersClientListSecretsResponse{}, err
	}
	return result, nil
}

// BeginUpdate - Update a ExtenderResource
// If the operation fails it returns an *azcore.ResponseError type.
//
// Generated from API version 2023-10-01-preview
//   - extenderName - The name of the ExtenderResource portable resource
//   - properties - The resource properties to be updated.
//   - options - ExtendersClientBeginUpdateOptions contains the optional parameters for the ExtendersClient.BeginUpdate method.
func (client *ExtendersClient) BeginUpdate(ctx context.Context, extenderName string, properties ExtenderResourceUpdate, options *ExtendersClientBeginUpdateOptions) (*runtime.Poller[ExtendersClientUpdateResponse], error) {
	if options == nil || options.ResumeToken == "" {
		resp, err := client.update(ctx, extenderName, properties, options)
		if err != nil {
			return nil, err
		}
		poller, err := runtime.NewPoller(resp, client.internal.Pipeline(), &runtime.NewPollerOptions[ExtendersClientUpdateResponse]{
			FinalStateVia: runtime.FinalStateViaLocation,
		})
		return poller, err
	} else {
		return runtime.NewPollerFromResumeToken[ExtendersClientUpdateResponse](options.ResumeToken, client.internal.Pipeline(), nil)
	}
}

// Update - Update a ExtenderResource
// If the operation fails it returns an *azcore.ResponseError type.
//
// Generated from API version 2023-10-01-preview
func (client *ExtendersClient) update(ctx context.Context, extenderName string, properties ExtenderResourceUpdate, options *ExtendersClientBeginUpdateOptions) (*http.Response, error) {
	var err error
	req, err := client.updateCreateRequest(ctx, extenderName, properties, options)
	if err != nil {
		return nil, err
	}
	httpResp, err := client.internal.Pipeline().Do(req)
	if err != nil {
		return nil, err
	}
	if !runtime.HasStatusCode(httpResp, http.StatusOK, http.StatusAccepted) {
		err = runtime.NewResponseError(httpResp)
		return nil, err
	}
	return httpResp, nil
}

// updateCreateRequest creates the Update request.
func (client *ExtendersClient) updateCreateRequest(ctx context.Context, extenderName string, properties ExtenderResourceUpdate, options *ExtendersClientBeginUpdateOptions) (*policy.Request, error) {
	urlPath := "/{rootScope}/providers/Applications.Core/extenders/{extenderName}"
	urlPath = strings.ReplaceAll(urlPath, "{rootScope}", client.rootScope)
	if extenderName == "" {
		return nil, errors.New("parameter extenderName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{extenderName}", url.PathEscape(extenderName))
	req, err := runtime.NewRequest(ctx, http.MethodPatch, runtime.JoinPaths(client.internal.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2023-10-01-preview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header["Accept"] = []string{"application/json"}
	if err := runtime.MarshalAsJSON(req, properties); err != nil {
	return nil, err
}
	return req, nil
}
