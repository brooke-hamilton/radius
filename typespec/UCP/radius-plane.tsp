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
import "../radius/v1/trackedresource.tsp";
import "./common.tsp";
import "./planes.tsp";
import "./ucp-operations.tsp";

using TypeSpec.Http;
using TypeSpec.Rest;
using TypeSpec.Versioning;
using Autorest;
using Azure.Core;
using Azure.ResourceManager;
using Azure.ResourceManager.Foundations;
using OpenAPI;

namespace Ucp;

@doc("The Radius plane resource.")
model RadiusPlaneResource
  is TrackedResourceRequired<
    RadiusPlaneResourceProperties,
    "System.Radius/planes"
  > {
  @doc("The plane name.")
  @segment("radius")
  @path
  @key("planeName")
  name: ResourceNameString;
}

@doc("The Plane properties.")
model RadiusPlaneResourceProperties {
  @doc("The status of the asynchronous operation.")
  @visibility("read")
  provisioningState?: ProvisioningState;

  @doc("Resource Providers for UCP Native Plane")
  resourceProviders: Record<string>;
}

@route("/planes")
@armResourceOperations
interface RadiusPlanes {
  @doc("List Radius planes")
  @get
  @route("/radius")
  @armResourceList(RadiusPlaneResource)
  list(
    ...ApiVersionParameter,
  ): ArmResponse<ResourceListResult<RadiusPlaneResource>> | ErrorResponse;

  @doc("Get a plane by name")
  get is UcpResourceRead<
    RadiusPlaneResource,
    PlaneBaseParameters<RadiusPlaneResource>
  >;

  @doc("Create or update a plane")
  createOrUpdate is UcpResourceCreateOrUpdateAsync<
    RadiusPlaneResource,
    PlaneBaseParameters<RadiusPlaneResource>
  >;

  @doc("Update a plane")
  update is UcpCustomPatchAsync<
    RadiusPlaneResource,
    PlaneBaseParameters<RadiusPlaneResource>
  >;

  @doc("Delete a plane")
  delete is UcpResourceDeleteAsync<
    RadiusPlaneResource,
    PlaneBaseParameters<RadiusPlaneResource>
  >;
}
