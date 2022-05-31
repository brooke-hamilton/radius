//go:build go1.16
// +build go1.16

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

package v20220315privatepreview

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"net/http"
	"net/url"
	"strings"
)

// RabbitMQMessageQueuesClient contains the methods for the RabbitMQMessageQueues group.
// Don't use this type directly, use NewRabbitMQMessageQueuesClient() instead.
type RabbitMQMessageQueuesClient struct {
	con *connection
	subscriptionID string
}

// NewRabbitMQMessageQueuesClient creates a new instance of RabbitMQMessageQueuesClient with the specified values.
func NewRabbitMQMessageQueuesClient(con *connection, subscriptionID string) *RabbitMQMessageQueuesClient {
	return &RabbitMQMessageQueuesClient{con: con, subscriptionID: subscriptionID}
}

// CreateOrUpdate - Creates or updates a RabbitMQMessageQueue resource
// If the operation fails it returns the *ErrorResponse error type.
func (client *RabbitMQMessageQueuesClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, rabbitMQMessageQueueName string, rabbitMQMessageQueueParameters RabbitMQMessageQueueResource, options *RabbitMQMessageQueuesCreateOrUpdateOptions) (RabbitMQMessageQueuesCreateOrUpdateResponse, error) {
	req, err := client.createOrUpdateCreateRequest(ctx, resourceGroupName, rabbitMQMessageQueueName, rabbitMQMessageQueueParameters, options)
	if err != nil {
		return RabbitMQMessageQueuesCreateOrUpdateResponse{}, err
	}
	resp, err := 	client.con.Pipeline().Do(req)
	if err != nil {
		return RabbitMQMessageQueuesCreateOrUpdateResponse{}, err
	}
	if !runtime.HasStatusCode(resp, http.StatusOK, http.StatusCreated) {
		return RabbitMQMessageQueuesCreateOrUpdateResponse{}, client.createOrUpdateHandleError(resp)
	}
	return client.createOrUpdateHandleResponse(resp)
}

// createOrUpdateCreateRequest creates the CreateOrUpdate request.
func (client *RabbitMQMessageQueuesClient) createOrUpdateCreateRequest(ctx context.Context, resourceGroupName string, rabbitMQMessageQueueName string, rabbitMQMessageQueueParameters RabbitMQMessageQueueResource, options *RabbitMQMessageQueuesCreateOrUpdateOptions) (*policy.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Applications.Connector/rabbitMQMessageQueues/{rabbitMQMessageQueueName}"
	if client.subscriptionID == "" {
		return nil, errors.New("parameter client.subscriptionID cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionId}", url.PathEscape(client.subscriptionID))
	if resourceGroupName == "" {
		return nil, errors.New("parameter resourceGroupName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{resourceGroupName}", url.PathEscape(resourceGroupName))
	if rabbitMQMessageQueueName == "" {
		return nil, errors.New("parameter rabbitMQMessageQueueName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{rabbitMQMessageQueueName}", url.PathEscape(rabbitMQMessageQueueName))
	req, err := runtime.NewRequest(ctx, http.MethodPut, runtime.JoinPaths(client.con.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2022-03-15-privatepreview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")
	return req, runtime.MarshalAsJSON(req, rabbitMQMessageQueueParameters)
}

// createOrUpdateHandleResponse handles the CreateOrUpdate response.
func (client *RabbitMQMessageQueuesClient) createOrUpdateHandleResponse(resp *http.Response) (RabbitMQMessageQueuesCreateOrUpdateResponse, error) {
	result := RabbitMQMessageQueuesCreateOrUpdateResponse{RawResponse: resp}
	if err := runtime.UnmarshalAsJSON(resp, &result.RabbitMQMessageQueueResource); err != nil {
		return RabbitMQMessageQueuesCreateOrUpdateResponse{}, err
	}
	return result, nil
}

// createOrUpdateHandleError handles the CreateOrUpdate error response.
func (client *RabbitMQMessageQueuesClient) createOrUpdateHandleError(resp *http.Response) error {
	body, err := runtime.Payload(resp)
	if err != nil {
		return runtime.NewResponseError(err, resp)
	}
		errType := ErrorResponse{raw: string(body)}
	if err := runtime.UnmarshalAsJSON(resp, &errType); err != nil {
		return runtime.NewResponseError(fmt.Errorf("%s\n%s", string(body), err), resp)
	}
	return runtime.NewResponseError(&errType, resp)
}

// Delete - Deletes an existing rabbitMQMessageQueue resource
// If the operation fails it returns the *ErrorResponse error type.
func (client *RabbitMQMessageQueuesClient) Delete(ctx context.Context, resourceGroupName string, rabbitMQMessageQueueName string, options *RabbitMQMessageQueuesDeleteOptions) (RabbitMQMessageQueuesDeleteResponse, error) {
	req, err := client.deleteCreateRequest(ctx, resourceGroupName, rabbitMQMessageQueueName, options)
	if err != nil {
		return RabbitMQMessageQueuesDeleteResponse{}, err
	}
	resp, err := 	client.con.Pipeline().Do(req)
	if err != nil {
		return RabbitMQMessageQueuesDeleteResponse{}, err
	}
	if !runtime.HasStatusCode(resp, http.StatusOK, http.StatusAccepted, http.StatusNoContent) {
		return RabbitMQMessageQueuesDeleteResponse{}, client.deleteHandleError(resp)
	}
	return RabbitMQMessageQueuesDeleteResponse{RawResponse: resp}, nil
}

// deleteCreateRequest creates the Delete request.
func (client *RabbitMQMessageQueuesClient) deleteCreateRequest(ctx context.Context, resourceGroupName string, rabbitMQMessageQueueName string, options *RabbitMQMessageQueuesDeleteOptions) (*policy.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Applications.Connector/rabbitMQMessageQueues/{rabbitMQMessageQueueName}"
	if client.subscriptionID == "" {
		return nil, errors.New("parameter client.subscriptionID cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionId}", url.PathEscape(client.subscriptionID))
	if resourceGroupName == "" {
		return nil, errors.New("parameter resourceGroupName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{resourceGroupName}", url.PathEscape(resourceGroupName))
	if rabbitMQMessageQueueName == "" {
		return nil, errors.New("parameter rabbitMQMessageQueueName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{rabbitMQMessageQueueName}", url.PathEscape(rabbitMQMessageQueueName))
	req, err := runtime.NewRequest(ctx, http.MethodDelete, runtime.JoinPaths(client.con.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2022-03-15-privatepreview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")
	return req, nil
}

// deleteHandleError handles the Delete error response.
func (client *RabbitMQMessageQueuesClient) deleteHandleError(resp *http.Response) error {
	body, err := runtime.Payload(resp)
	if err != nil {
		return runtime.NewResponseError(err, resp)
	}
		errType := ErrorResponse{raw: string(body)}
	if err := runtime.UnmarshalAsJSON(resp, &errType); err != nil {
		return runtime.NewResponseError(fmt.Errorf("%s\n%s", string(body), err), resp)
	}
	return runtime.NewResponseError(&errType, resp)
}

// Get - Retrieves information about a rabbitMQMessageQueue resource
// If the operation fails it returns the *ErrorResponse error type.
func (client *RabbitMQMessageQueuesClient) Get(ctx context.Context, resourceGroupName string, rabbitMQMessageQueueName string, options *RabbitMQMessageQueuesGetOptions) (RabbitMQMessageQueuesGetResponse, error) {
	req, err := client.getCreateRequest(ctx, resourceGroupName, rabbitMQMessageQueueName, options)
	if err != nil {
		return RabbitMQMessageQueuesGetResponse{}, err
	}
	resp, err := 	client.con.Pipeline().Do(req)
	if err != nil {
		return RabbitMQMessageQueuesGetResponse{}, err
	}
	if !runtime.HasStatusCode(resp, http.StatusOK) {
		return RabbitMQMessageQueuesGetResponse{}, client.getHandleError(resp)
	}
	return client.getHandleResponse(resp)
}

// getCreateRequest creates the Get request.
func (client *RabbitMQMessageQueuesClient) getCreateRequest(ctx context.Context, resourceGroupName string, rabbitMQMessageQueueName string, options *RabbitMQMessageQueuesGetOptions) (*policy.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Applications.Connector/rabbitMQMessageQueues/{rabbitMQMessageQueueName}"
	if client.subscriptionID == "" {
		return nil, errors.New("parameter client.subscriptionID cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionId}", url.PathEscape(client.subscriptionID))
	if resourceGroupName == "" {
		return nil, errors.New("parameter resourceGroupName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{resourceGroupName}", url.PathEscape(resourceGroupName))
	if rabbitMQMessageQueueName == "" {
		return nil, errors.New("parameter rabbitMQMessageQueueName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{rabbitMQMessageQueueName}", url.PathEscape(rabbitMQMessageQueueName))
	req, err := runtime.NewRequest(ctx, http.MethodGet, runtime.JoinPaths(client.con.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2022-03-15-privatepreview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")
	return req, nil
}

// getHandleResponse handles the Get response.
func (client *RabbitMQMessageQueuesClient) getHandleResponse(resp *http.Response) (RabbitMQMessageQueuesGetResponse, error) {
	result := RabbitMQMessageQueuesGetResponse{RawResponse: resp}
	if err := runtime.UnmarshalAsJSON(resp, &result.RabbitMQMessageQueueResource); err != nil {
		return RabbitMQMessageQueuesGetResponse{}, err
	}
	return result, nil
}

// getHandleError handles the Get error response.
func (client *RabbitMQMessageQueuesClient) getHandleError(resp *http.Response) error {
	body, err := runtime.Payload(resp)
	if err != nil {
		return runtime.NewResponseError(err, resp)
	}
		errType := ErrorResponse{raw: string(body)}
	if err := runtime.UnmarshalAsJSON(resp, &errType); err != nil {
		return runtime.NewResponseError(fmt.Errorf("%s\n%s", string(body), err), resp)
	}
	return runtime.NewResponseError(&errType, resp)
}

// List - Lists information about all rabbitMQMessageQueue resources in the given subscription and resource group
// If the operation fails it returns the *ErrorResponse error type.
func (client *RabbitMQMessageQueuesClient) List(resourceGroupName string, options *RabbitMQMessageQueuesListOptions) (*RabbitMQMessageQueuesListPager) {
	return &RabbitMQMessageQueuesListPager{
		client: client,
		requester: func(ctx context.Context) (*policy.Request, error) {
			return client.listCreateRequest(ctx, resourceGroupName, options)
		},
		advancer: func(ctx context.Context, resp RabbitMQMessageQueuesListResponse) (*policy.Request, error) {
			return runtime.NewRequest(ctx, http.MethodGet, *resp.RabbitMQMessageQueueList.NextLink)
		},
	}
}

// listCreateRequest creates the List request.
func (client *RabbitMQMessageQueuesClient) listCreateRequest(ctx context.Context, resourceGroupName string, options *RabbitMQMessageQueuesListOptions) (*policy.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Applications.Connector/rabbitMQMessageQueues"
	if client.subscriptionID == "" {
		return nil, errors.New("parameter client.subscriptionID cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionId}", url.PathEscape(client.subscriptionID))
	if resourceGroupName == "" {
		return nil, errors.New("parameter resourceGroupName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{resourceGroupName}", url.PathEscape(resourceGroupName))
	req, err := runtime.NewRequest(ctx, http.MethodGet, runtime.JoinPaths(client.con.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2022-03-15-privatepreview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")
	return req, nil
}

// listHandleResponse handles the List response.
func (client *RabbitMQMessageQueuesClient) listHandleResponse(resp *http.Response) (RabbitMQMessageQueuesListResponse, error) {
	result := RabbitMQMessageQueuesListResponse{RawResponse: resp}
	if err := runtime.UnmarshalAsJSON(resp, &result.RabbitMQMessageQueueList); err != nil {
		return RabbitMQMessageQueuesListResponse{}, err
	}
	return result, nil
}

// listHandleError handles the List error response.
func (client *RabbitMQMessageQueuesClient) listHandleError(resp *http.Response) error {
	body, err := runtime.Payload(resp)
	if err != nil {
		return runtime.NewResponseError(err, resp)
	}
		errType := ErrorResponse{raw: string(body)}
	if err := runtime.UnmarshalAsJSON(resp, &errType); err != nil {
		return runtime.NewResponseError(fmt.Errorf("%s\n%s", string(body), err), resp)
	}
	return runtime.NewResponseError(&errType, resp)
}

// ListBySubscription - Lists information about all rabbitMQMessageQueue resources in the given subscription
// If the operation fails it returns the *ErrorResponse error type.
func (client *RabbitMQMessageQueuesClient) ListBySubscription(options *RabbitMQMessageQueuesListBySubscriptionOptions) (*RabbitMQMessageQueuesListBySubscriptionPager) {
	return &RabbitMQMessageQueuesListBySubscriptionPager{
		client: client,
		requester: func(ctx context.Context) (*policy.Request, error) {
			return client.listBySubscriptionCreateRequest(ctx, options)
		},
		advancer: func(ctx context.Context, resp RabbitMQMessageQueuesListBySubscriptionResponse) (*policy.Request, error) {
			return runtime.NewRequest(ctx, http.MethodGet, *resp.RabbitMQMessageQueueList.NextLink)
		},
	}
}

// listBySubscriptionCreateRequest creates the ListBySubscription request.
func (client *RabbitMQMessageQueuesClient) listBySubscriptionCreateRequest(ctx context.Context, options *RabbitMQMessageQueuesListBySubscriptionOptions) (*policy.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/providers/Applications.Connector/rabbitMQMessageQueues"
	if client.subscriptionID == "" {
		return nil, errors.New("parameter client.subscriptionID cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionId}", url.PathEscape(client.subscriptionID))
	req, err := runtime.NewRequest(ctx, http.MethodGet, runtime.JoinPaths(client.con.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2022-03-15-privatepreview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")
	return req, nil
}

// listBySubscriptionHandleResponse handles the ListBySubscription response.
func (client *RabbitMQMessageQueuesClient) listBySubscriptionHandleResponse(resp *http.Response) (RabbitMQMessageQueuesListBySubscriptionResponse, error) {
	result := RabbitMQMessageQueuesListBySubscriptionResponse{RawResponse: resp}
	if err := runtime.UnmarshalAsJSON(resp, &result.RabbitMQMessageQueueList); err != nil {
		return RabbitMQMessageQueuesListBySubscriptionResponse{}, err
	}
	return result, nil
}

// listBySubscriptionHandleError handles the ListBySubscription error response.
func (client *RabbitMQMessageQueuesClient) listBySubscriptionHandleError(resp *http.Response) error {
	body, err := runtime.Payload(resp)
	if err != nil {
		return runtime.NewResponseError(err, resp)
	}
		errType := ErrorResponse{raw: string(body)}
	if err := runtime.UnmarshalAsJSON(resp, &errType); err != nil {
		return runtime.NewResponseError(fmt.Errorf("%s\n%s", string(body), err), resp)
	}
	return runtime.NewResponseError(&errType, resp)
}

