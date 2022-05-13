// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package hostoptions

import (
	"github.com/project-radius/radius/pkg/corerp/dataprovider"
	"github.com/project-radius/radius/pkg/telemetry/metrics"
)

// ProviderConfig includes the resource provider configuration.
type ProviderConfig struct {
	CloudEnv        CloudEnvironmentOptions             `yaml:"cloudEnvironment"`
	Identity        IdentityOptions                     `yaml:"identity"`
	StorageProvider dataprovider.StorageProviderOptions `yaml:"storageProvider"`
	Server          ServerOptions                       `yaml:"server"`
	MetricsProvider metrics.MetricsOptions              `yaml:"metricsProvider"`

	// FeatureFlags includes the list of feature flags.
	FeatureFlags []string `yaml:"featureFlags"`
}

type CloudEnvironmentType string

const (
	AzureDogfood      CloudEnvironmentType = "Dogfood"
	AzureCloud        CloudEnvironmentType = "AzureCloud"
	AzureChinaCloud   CloudEnvironmentType = "AzureChinaCloud"
	AzureUSGovernment CloudEnvironmentType = "AzureUSGovernment"
	AzureGermanCloud  CloudEnvironmentType = "AzureGermanCloud"
)

type AuthentificationType string

const (
	ClientCertificateAuthType AuthentificationType = "ClientCertificate"
	AADPoPAuthType            AuthentificationType = "PoP"
)

// CloudEnvironmentOptions represents the cloud environment.
type CloudEnvironmentOptions struct {
	Name         CloudEnvironmentType `yaml:"name"`
	RoleLocation string               `yaml:"roleLocation"`
}

// IdentityOptions includes authentication options to issue JWT from Azure AD.
type IdentityOptions struct {
	ClientID    string `yaml:"clientId"`
	Instance    string `yaml:"instance"`
	TenantID    string `yaml:"tenantId"`
	ArmEndpoint string `yaml:"armEndpoint"`
	Audience    string `yaml:"audience"`
	PemCertPath string `yaml:"pemCertPath"`
}

// ServerOptions includes http server bootstrap options.
type ServerOptions struct {
	Host     string               `yaml:"host"`
	Port     int                  `yaml:"port"`
	PathBase string               `yaml:"pathBase,omitempty"`
	AuthType AuthentificationType `yaml:"authType,omitempty"`
	// ArmMetadataEndpoints provides the client certification to authenticate between ARM and RP.
	// https://armwiki.azurewebsites.net/authorization/AuthenticateBetweenARMandRP.html
	ArmMetadataEndpoint string `yaml:"armMetadataEndpoint,omitempty"`
	// EnableAuth when set the arm client authetication will be performed
	EnableArmAuth bool `yaml:"enableArmAuth,omitempty"`
}
