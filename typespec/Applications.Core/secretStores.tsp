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

import "@typespec/rest";
import "@typespec/versioning";
import "@typespec/openapi";
import "@azure-tools/typespec-autorest";
import "@azure-tools/typespec-azure-core";
import "@azure-tools/typespec-azure-resource-manager";

import "../radius/v1/ucprootscope.tsp";
import "../radius/v1/resources.tsp";
import "./common.tsp";
import "../radius/v1/trackedresource.tsp";

using TypeSpec.Http;
using TypeSpec.Rest;
using TypeSpec.Versioning;
using Autorest;
using Azure.Core;
using Azure.ResourceManager;
using OpenAPI;

namespace Applications.Core;

model SecretStoreResource
  is TrackedResourceRequired<SecretStoreProperties, "secretStores"> {
  @doc("SecretStore name")
  @key("secretStoreName")
  @path
  @segment("secretStores")
  name: ResourceNameString;
}

@doc("The properties of SecretStore")
model SecretStoreProperties {
  ...GlobalScopedResource;

  #suppress "@azure-tools/typespec-azure-resource-manager/arm-resource-duplicate-property"
  @doc("The type of secret store data")
  type?: SecretStoreDataType = SecretStoreDataType.generic;

  @doc("An object to represent key-value type secrets")
  data: Record<SecretValueProperties>;

  @doc("The resource id of external secret store.")
  resource?: string;
}

@doc("The type of SecretStore data")
enum SecretStoreDataType {
  @doc("Generic secret data type")
  generic,

  @doc("Certificate secret data type")
  certificate,

  @doc("basicAuthentication type is used to represent username and password based authentication and the secretstore resource is expected to have the keys 'username' and 'password'.")
  basicAuthentication,

  @doc("azureWorkloadIdentity type is used to represent authentication using azure federated identity and the secretstore resource is expected to have the keys 'clientId' and 'tenantId'.")
  azureWorkloadIdentity,

  @doc("awsIRSA type is used to represent authentication using AWS IRSA (IAM Roles for Service accounts) and the secretstore resource is expected to have the key 'roleARN'.")
  awsIRSA,
}

@doc("The type of SecretValue Encoding")
enum SecretValueEncoding {
  @doc("The raw secret value")
  raw,

  @doc("The base64-encoded secret value")
  base64,
}

@doc("The Secret value source properties")
model ValueFromProperties {
  @doc("The name of the referenced secret.")
  name: string;

  @doc("The version of the referenced secret.")
  version?: string;
}

@doc("The properties of SecretValue")
model SecretValueProperties {
  @doc("The encoding of value")
  encoding?: SecretValueEncoding = SecretValueEncoding.raw;

  @doc("The value of secret.")
  @secret
  value?: string;

  @doc("The referenced secret in properties.resource")
  valueFrom?: ValueFromProperties;
}

@doc("The list of secrets")
model SecretStoreListSecretsResult {
  @doc("The type of secret store data")
  type: SecretStoreDataType;

  @doc("An object to represent key-value type secrets")
  data: Record<SecretValueProperties>;
}

@armResourceOperations
interface SecretStores {
  get is ArmResourceRead<
    SecretStoreResource,
    UCPBaseParameters<SecretStoreResource>
  >;

  createOrUpdate is ArmResourceCreateOrReplaceAsync<
    SecretStoreResource,
    UCPBaseParameters<SecretStoreResource>
  >;

  update is ArmResourcePatchAsync<
    SecretStoreResource,
    SecretStoreProperties,
    UCPBaseParameters<SecretStoreResource>
  >;

  delete is ArmResourceDeleteWithoutOkAsync<
    SecretStoreResource,
    UCPBaseParameters<SecretStoreResource>
  >;

  listByScope is ArmResourceListByParent<
    SecretStoreResource,
    UCPBaseParameters<SecretStoreResource>,
    "Scope",
    "Scope"
  >;

  @doc("List the secrets of a secret stores.")
  @action("listSecrets")
  listSecrets is ArmResourceActionSync<
    SecretStoreResource,
    {},
    SecretStoreListSecretsResult,
    UCPBaseParameters<SecretStoreResource>
  >;
}
