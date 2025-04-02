// Licensed under the Apache License, Version 2.0 . See LICENSE in the repository root for license information.
// Code generated by Microsoft (R) AutoRest Code Generator. DO NOT EDIT.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

package v20231001preview

import (
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"reflect"
)

// MarshalJSON implements the json.Marshaller interface for type AzureContainerInstanceCompute.
func (a AzureContainerInstanceCompute) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "identity", a.Identity)
	objectMap["kind"] = "aci"
	populate(objectMap, "resourceGroup", a.ResourceGroup)
	populate(objectMap, "resourceId", a.ResourceID)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type AzureContainerInstanceCompute.
func (a *AzureContainerInstanceCompute) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", a, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "identity":
				err = unpopulate(val, "Identity", &a.Identity)
			delete(rawMsg, key)
		case "kind":
				err = unpopulate(val, "Kind", &a.Kind)
			delete(rawMsg, key)
		case "resourceGroup":
				err = unpopulate(val, "ResourceGroup", &a.ResourceGroup)
			delete(rawMsg, key)
		case "resourceId":
				err = unpopulate(val, "ResourceID", &a.ResourceID)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", a, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type AzureResourceManagerCommonTypesTrackedResourceUpdate.
func (a AzureResourceManagerCommonTypesTrackedResourceUpdate) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "id", a.ID)
	populate(objectMap, "name", a.Name)
	populate(objectMap, "systemData", a.SystemData)
	populate(objectMap, "tags", a.Tags)
	populate(objectMap, "type", a.Type)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type AzureResourceManagerCommonTypesTrackedResourceUpdate.
func (a *AzureResourceManagerCommonTypesTrackedResourceUpdate) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", a, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "id":
				err = unpopulate(val, "ID", &a.ID)
			delete(rawMsg, key)
		case "name":
				err = unpopulate(val, "Name", &a.Name)
			delete(rawMsg, key)
		case "systemData":
				err = unpopulate(val, "SystemData", &a.SystemData)
			delete(rawMsg, key)
		case "tags":
				err = unpopulate(val, "Tags", &a.Tags)
			delete(rawMsg, key)
		case "type":
				err = unpopulate(val, "Type", &a.Type)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", a, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type EnvironmentCompute.
func (e EnvironmentCompute) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "identity", e.Identity)
	objectMap["kind"] = e.Kind
	populate(objectMap, "resourceId", e.ResourceID)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type EnvironmentCompute.
func (e *EnvironmentCompute) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", e, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "identity":
				err = unpopulate(val, "Identity", &e.Identity)
			delete(rawMsg, key)
		case "kind":
				err = unpopulate(val, "Kind", &e.Kind)
			delete(rawMsg, key)
		case "resourceId":
				err = unpopulate(val, "ResourceID", &e.ResourceID)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", e, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type ErrorAdditionalInfo.
func (e ErrorAdditionalInfo) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "info", e.Info)
	populate(objectMap, "type", e.Type)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type ErrorAdditionalInfo.
func (e *ErrorAdditionalInfo) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", e, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "info":
				err = unpopulate(val, "Info", &e.Info)
			delete(rawMsg, key)
		case "type":
				err = unpopulate(val, "Type", &e.Type)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", e, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type ErrorDetail.
func (e ErrorDetail) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "additionalInfo", e.AdditionalInfo)
	populate(objectMap, "code", e.Code)
	populate(objectMap, "details", e.Details)
	populate(objectMap, "message", e.Message)
	populate(objectMap, "target", e.Target)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type ErrorDetail.
func (e *ErrorDetail) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", e, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "additionalInfo":
				err = unpopulate(val, "AdditionalInfo", &e.AdditionalInfo)
			delete(rawMsg, key)
		case "code":
				err = unpopulate(val, "Code", &e.Code)
			delete(rawMsg, key)
		case "details":
				err = unpopulate(val, "Details", &e.Details)
			delete(rawMsg, key)
		case "message":
				err = unpopulate(val, "Message", &e.Message)
			delete(rawMsg, key)
		case "target":
				err = unpopulate(val, "Target", &e.Target)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", e, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type ErrorResponse.
func (e ErrorResponse) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "error", e.Error)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type ErrorResponse.
func (e *ErrorResponse) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", e, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "error":
				err = unpopulate(val, "Error", &e.Error)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", e, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type IdentitySettings.
func (i IdentitySettings) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "kind", i.Kind)
	populate(objectMap, "oidcIssuer", i.OidcIssuer)
	populate(objectMap, "resource", i.Resource)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type IdentitySettings.
func (i *IdentitySettings) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", i, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "kind":
				err = unpopulate(val, "Kind", &i.Kind)
			delete(rawMsg, key)
		case "oidcIssuer":
				err = unpopulate(val, "OidcIssuer", &i.OidcIssuer)
			delete(rawMsg, key)
		case "resource":
				err = unpopulate(val, "Resource", &i.Resource)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", i, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type KubernetesCompute.
func (k KubernetesCompute) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "identity", k.Identity)
	objectMap["kind"] = "kubernetes"
	populate(objectMap, "namespace", k.Namespace)
	populate(objectMap, "resourceId", k.ResourceID)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type KubernetesCompute.
func (k *KubernetesCompute) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", k, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "identity":
				err = unpopulate(val, "Identity", &k.Identity)
			delete(rawMsg, key)
		case "kind":
				err = unpopulate(val, "Kind", &k.Kind)
			delete(rawMsg, key)
		case "namespace":
				err = unpopulate(val, "Namespace", &k.Namespace)
			delete(rawMsg, key)
		case "resourceId":
				err = unpopulate(val, "ResourceID", &k.ResourceID)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", k, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type Operation.
func (o Operation) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "actionType", o.ActionType)
	populate(objectMap, "display", o.Display)
	populate(objectMap, "isDataAction", o.IsDataAction)
	populate(objectMap, "name", o.Name)
	populate(objectMap, "origin", o.Origin)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type Operation.
func (o *Operation) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", o, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "actionType":
				err = unpopulate(val, "ActionType", &o.ActionType)
			delete(rawMsg, key)
		case "display":
				err = unpopulate(val, "Display", &o.Display)
			delete(rawMsg, key)
		case "isDataAction":
				err = unpopulate(val, "IsDataAction", &o.IsDataAction)
			delete(rawMsg, key)
		case "name":
				err = unpopulate(val, "Name", &o.Name)
			delete(rawMsg, key)
		case "origin":
				err = unpopulate(val, "Origin", &o.Origin)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", o, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type OperationDisplay.
func (o OperationDisplay) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "description", o.Description)
	populate(objectMap, "operation", o.Operation)
	populate(objectMap, "provider", o.Provider)
	populate(objectMap, "resource", o.Resource)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type OperationDisplay.
func (o *OperationDisplay) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", o, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "description":
				err = unpopulate(val, "Description", &o.Description)
			delete(rawMsg, key)
		case "operation":
				err = unpopulate(val, "Operation", &o.Operation)
			delete(rawMsg, key)
		case "provider":
				err = unpopulate(val, "Provider", &o.Provider)
			delete(rawMsg, key)
		case "resource":
				err = unpopulate(val, "Resource", &o.Resource)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", o, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type OperationListResult.
func (o OperationListResult) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "nextLink", o.NextLink)
	populate(objectMap, "value", o.Value)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type OperationListResult.
func (o *OperationListResult) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", o, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "nextLink":
				err = unpopulate(val, "NextLink", &o.NextLink)
			delete(rawMsg, key)
		case "value":
				err = unpopulate(val, "Value", &o.Value)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", o, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type OutputResource.
func (o OutputResource) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "id", o.ID)
	populate(objectMap, "localId", o.LocalID)
	populate(objectMap, "radiusManaged", o.RadiusManaged)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type OutputResource.
func (o *OutputResource) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", o, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "id":
				err = unpopulate(val, "ID", &o.ID)
			delete(rawMsg, key)
		case "localId":
				err = unpopulate(val, "LocalID", &o.LocalID)
			delete(rawMsg, key)
		case "radiusManaged":
				err = unpopulate(val, "RadiusManaged", &o.RadiusManaged)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", o, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type RabbitMQListSecretsResult.
func (r RabbitMQListSecretsResult) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "password", r.Password)
	populate(objectMap, "uri", r.URI)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type RabbitMQListSecretsResult.
func (r *RabbitMQListSecretsResult) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", r, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "password":
				err = unpopulate(val, "Password", &r.Password)
			delete(rawMsg, key)
		case "uri":
				err = unpopulate(val, "URI", &r.URI)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", r, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type RabbitMQQueueProperties.
func (r RabbitMQQueueProperties) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "application", r.Application)
	populate(objectMap, "environment", r.Environment)
	populate(objectMap, "host", r.Host)
	populate(objectMap, "port", r.Port)
	populate(objectMap, "provisioningState", r.ProvisioningState)
	populate(objectMap, "queue", r.Queue)
	populate(objectMap, "recipe", r.Recipe)
	populate(objectMap, "resourceProvisioning", r.ResourceProvisioning)
	populate(objectMap, "resources", r.Resources)
	populate(objectMap, "secrets", r.Secrets)
	populate(objectMap, "status", r.Status)
	populate(objectMap, "tls", r.TLS)
	populate(objectMap, "username", r.Username)
	populate(objectMap, "vHost", r.VHost)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type RabbitMQQueueProperties.
func (r *RabbitMQQueueProperties) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", r, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "application":
				err = unpopulate(val, "Application", &r.Application)
			delete(rawMsg, key)
		case "environment":
				err = unpopulate(val, "Environment", &r.Environment)
			delete(rawMsg, key)
		case "host":
				err = unpopulate(val, "Host", &r.Host)
			delete(rawMsg, key)
		case "port":
				err = unpopulate(val, "Port", &r.Port)
			delete(rawMsg, key)
		case "provisioningState":
				err = unpopulate(val, "ProvisioningState", &r.ProvisioningState)
			delete(rawMsg, key)
		case "queue":
				err = unpopulate(val, "Queue", &r.Queue)
			delete(rawMsg, key)
		case "recipe":
				err = unpopulate(val, "Recipe", &r.Recipe)
			delete(rawMsg, key)
		case "resourceProvisioning":
				err = unpopulate(val, "ResourceProvisioning", &r.ResourceProvisioning)
			delete(rawMsg, key)
		case "resources":
				err = unpopulate(val, "Resources", &r.Resources)
			delete(rawMsg, key)
		case "secrets":
				err = unpopulate(val, "Secrets", &r.Secrets)
			delete(rawMsg, key)
		case "status":
				err = unpopulate(val, "Status", &r.Status)
			delete(rawMsg, key)
		case "tls":
				err = unpopulate(val, "TLS", &r.TLS)
			delete(rawMsg, key)
		case "username":
				err = unpopulate(val, "Username", &r.Username)
			delete(rawMsg, key)
		case "vHost":
				err = unpopulate(val, "VHost", &r.VHost)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", r, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type RabbitMQQueueResource.
func (r RabbitMQQueueResource) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "id", r.ID)
	populate(objectMap, "location", r.Location)
	populate(objectMap, "name", r.Name)
	populate(objectMap, "properties", r.Properties)
	populate(objectMap, "systemData", r.SystemData)
	populate(objectMap, "tags", r.Tags)
	populate(objectMap, "type", r.Type)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type RabbitMQQueueResource.
func (r *RabbitMQQueueResource) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", r, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "id":
				err = unpopulate(val, "ID", &r.ID)
			delete(rawMsg, key)
		case "location":
				err = unpopulate(val, "Location", &r.Location)
			delete(rawMsg, key)
		case "name":
				err = unpopulate(val, "Name", &r.Name)
			delete(rawMsg, key)
		case "properties":
				err = unpopulate(val, "Properties", &r.Properties)
			delete(rawMsg, key)
		case "systemData":
				err = unpopulate(val, "SystemData", &r.SystemData)
			delete(rawMsg, key)
		case "tags":
				err = unpopulate(val, "Tags", &r.Tags)
			delete(rawMsg, key)
		case "type":
				err = unpopulate(val, "Type", &r.Type)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", r, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type RabbitMQQueueResourceListResult.
func (r RabbitMQQueueResourceListResult) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "nextLink", r.NextLink)
	populate(objectMap, "value", r.Value)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type RabbitMQQueueResourceListResult.
func (r *RabbitMQQueueResourceListResult) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", r, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "nextLink":
				err = unpopulate(val, "NextLink", &r.NextLink)
			delete(rawMsg, key)
		case "value":
				err = unpopulate(val, "Value", &r.Value)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", r, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type RabbitMQQueueResourceUpdate.
func (r RabbitMQQueueResourceUpdate) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "id", r.ID)
	populate(objectMap, "name", r.Name)
	populate(objectMap, "systemData", r.SystemData)
	populate(objectMap, "tags", r.Tags)
	populate(objectMap, "type", r.Type)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type RabbitMQQueueResourceUpdate.
func (r *RabbitMQQueueResourceUpdate) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", r, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "id":
				err = unpopulate(val, "ID", &r.ID)
			delete(rawMsg, key)
		case "name":
				err = unpopulate(val, "Name", &r.Name)
			delete(rawMsg, key)
		case "systemData":
				err = unpopulate(val, "SystemData", &r.SystemData)
			delete(rawMsg, key)
		case "tags":
				err = unpopulate(val, "Tags", &r.Tags)
			delete(rawMsg, key)
		case "type":
				err = unpopulate(val, "Type", &r.Type)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", r, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type RabbitMQSecrets.
func (r RabbitMQSecrets) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "password", r.Password)
	populate(objectMap, "uri", r.URI)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type RabbitMQSecrets.
func (r *RabbitMQSecrets) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", r, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "password":
				err = unpopulate(val, "Password", &r.Password)
			delete(rawMsg, key)
		case "uri":
				err = unpopulate(val, "URI", &r.URI)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", r, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type Recipe.
func (r Recipe) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "name", r.Name)
	populate(objectMap, "parameters", r.Parameters)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type Recipe.
func (r *Recipe) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", r, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "name":
				err = unpopulate(val, "Name", &r.Name)
			delete(rawMsg, key)
		case "parameters":
				err = unpopulate(val, "Parameters", &r.Parameters)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", r, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type RecipeStatus.
func (r RecipeStatus) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "templateKind", r.TemplateKind)
	populate(objectMap, "templatePath", r.TemplatePath)
	populate(objectMap, "templateVersion", r.TemplateVersion)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type RecipeStatus.
func (r *RecipeStatus) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", r, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "templateKind":
				err = unpopulate(val, "TemplateKind", &r.TemplateKind)
			delete(rawMsg, key)
		case "templatePath":
				err = unpopulate(val, "TemplatePath", &r.TemplatePath)
			delete(rawMsg, key)
		case "templateVersion":
				err = unpopulate(val, "TemplateVersion", &r.TemplateVersion)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", r, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type Resource.
func (r Resource) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "id", r.ID)
	populate(objectMap, "name", r.Name)
	populate(objectMap, "systemData", r.SystemData)
	populate(objectMap, "type", r.Type)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type Resource.
func (r *Resource) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", r, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "id":
				err = unpopulate(val, "ID", &r.ID)
			delete(rawMsg, key)
		case "name":
				err = unpopulate(val, "Name", &r.Name)
			delete(rawMsg, key)
		case "systemData":
				err = unpopulate(val, "SystemData", &r.SystemData)
			delete(rawMsg, key)
		case "type":
				err = unpopulate(val, "Type", &r.Type)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", r, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type ResourceReference.
func (r ResourceReference) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "id", r.ID)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type ResourceReference.
func (r *ResourceReference) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", r, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "id":
				err = unpopulate(val, "ID", &r.ID)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", r, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type ResourceStatus.
func (r ResourceStatus) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "compute", r.Compute)
	populate(objectMap, "outputResources", r.OutputResources)
	populate(objectMap, "recipe", r.Recipe)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type ResourceStatus.
func (r *ResourceStatus) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", r, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "compute":
			r.Compute, err = unmarshalEnvironmentComputeClassification(val)
			delete(rawMsg, key)
		case "outputResources":
				err = unpopulate(val, "OutputResources", &r.OutputResources)
			delete(rawMsg, key)
		case "recipe":
				err = unpopulate(val, "Recipe", &r.Recipe)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", r, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type SystemData.
func (s SystemData) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populateDateTimeRFC3339(objectMap, "createdAt", s.CreatedAt)
	populate(objectMap, "createdBy", s.CreatedBy)
	populate(objectMap, "createdByType", s.CreatedByType)
	populateDateTimeRFC3339(objectMap, "lastModifiedAt", s.LastModifiedAt)
	populate(objectMap, "lastModifiedBy", s.LastModifiedBy)
	populate(objectMap, "lastModifiedByType", s.LastModifiedByType)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type SystemData.
func (s *SystemData) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", s, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "createdAt":
				err = unpopulateDateTimeRFC3339(val, "CreatedAt", &s.CreatedAt)
			delete(rawMsg, key)
		case "createdBy":
				err = unpopulate(val, "CreatedBy", &s.CreatedBy)
			delete(rawMsg, key)
		case "createdByType":
				err = unpopulate(val, "CreatedByType", &s.CreatedByType)
			delete(rawMsg, key)
		case "lastModifiedAt":
				err = unpopulateDateTimeRFC3339(val, "LastModifiedAt", &s.LastModifiedAt)
			delete(rawMsg, key)
		case "lastModifiedBy":
				err = unpopulate(val, "LastModifiedBy", &s.LastModifiedBy)
			delete(rawMsg, key)
		case "lastModifiedByType":
				err = unpopulate(val, "LastModifiedByType", &s.LastModifiedByType)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", s, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type TrackedResource.
func (t TrackedResource) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "id", t.ID)
	populate(objectMap, "location", t.Location)
	populate(objectMap, "name", t.Name)
	populate(objectMap, "systemData", t.SystemData)
	populate(objectMap, "tags", t.Tags)
	populate(objectMap, "type", t.Type)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type TrackedResource.
func (t *TrackedResource) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", t, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "id":
				err = unpopulate(val, "ID", &t.ID)
			delete(rawMsg, key)
		case "location":
				err = unpopulate(val, "Location", &t.Location)
			delete(rawMsg, key)
		case "name":
				err = unpopulate(val, "Name", &t.Name)
			delete(rawMsg, key)
		case "systemData":
				err = unpopulate(val, "SystemData", &t.SystemData)
			delete(rawMsg, key)
		case "tags":
				err = unpopulate(val, "Tags", &t.Tags)
			delete(rawMsg, key)
		case "type":
				err = unpopulate(val, "Type", &t.Type)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", t, err)
		}
	}
	return nil
}

func populate(m map[string]any, k string, v any) {
	if v == nil {
		return
	} else if azcore.IsNullValue(v) {
		m[k] = nil
	} else if !reflect.ValueOf(v).IsNil() {
		m[k] = v
	}
}

func unpopulate(data json.RawMessage, fn string, v any) error {
	if data == nil || string(data) == "null" {
		return nil
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("struct field %s: %v", fn, err)
	}
	return nil
}

