[//]: # (Generated by grpc-framework using protoc-gen-doc)
[//]: # (Do not edit)

# gRPC API Reference

## Contents


- Services
    - [TenantService](#servicepublicportworxplatformtenantv1tenantservice)
  


- Messages
    - [CreateTenantRequest](#createtenantrequest)
    - [DeleteTenantRequest](#deletetenantrequest)
    - [GetTenantRequest](#gettenantrequest)
    - [ListTenantsRequest](#listtenantsrequest)
    - [ListTenantsResponse](#listtenantsresponse)
    - [Phase](#phase)
    - [Status](#status)
    - [Tenant](#tenant)
    - [UpdateTenantRequest](#updatetenantrequest)
  



- [Scalar Value Types](#scalar-value-types)




## TenantService {#servicepublicportworxplatformtenantv1tenantservice}
Tenant service provides APIs to interact with the Tenant entity.

### GetTenant {#methodpublicportworxplatformtenantv1tenantservicegettenant}

> **rpc** GetTenant([GetTenantRequest](#gettenantrequest))
    [Tenant](#tenant)

GetTenant API returns the info about  for given tenant id
### CreateTenant {#methodpublicportworxplatformtenantv1tenantservicecreatetenant}

> **rpc** CreateTenant([CreateTenantRequest](#createtenantrequest))
    [Tenant](#tenant)

CreateTenant API creates a new Tenant
### UpdateTenant {#methodpublicportworxplatformtenantv1tenantserviceupdatetenant}

> **rpc** UpdateTenant([UpdateTenantRequest](#updatetenantrequest))
    [Tenant](#tenant)

UpdateTenant API updates tenant.
### DeleteTenant {#methodpublicportworxplatformtenantv1tenantservicedeletetenant}

> **rpc** DeleteTenant([DeleteTenantRequest](#deletetenantrequest))
    [.google.protobuf.Empty](#googleprotobufempty)

Delete tenant removes a tenant record.
### ListTenants {#methodpublicportworxplatformtenantv1tenantservicelisttenants}

> **rpc** ListTenants([ListTenantsRequest](#listtenantsrequest))
    [ListTenantsResponse](#listtenantsresponse)

ListTenants API lists the tenants visible to the caller for the current account.
 <!-- end methods -->
 <!-- end services -->

## Messages


### CreateTenantRequest {#createtenantrequest}
Request for creating a tenant.


| Field | Type | Description |
| ----- | ---- | ----------- |
| tenant | [ Tenant](#tenant) | tenant to be created |
 <!-- end Fields -->
 <!-- end HasFields -->


### DeleteTenantRequest {#deletetenantrequest}
Request for deleting a tenant.


| Field | Type | Description |
| ----- | ---- | ----------- |
| tenant_id | [ string](#string) | ID of the tenant which needs to be deleted. |
 <!-- end Fields -->
 <!-- end HasFields -->


### GetTenantRequest {#gettenantrequest}
Request for getting  a tenant.


| Field | Type | Description |
| ----- | ---- | ----------- |
| tenant_id | [ string](#string) | ID of the tenant which needs to get info. |
 <!-- end Fields -->
 <!-- end HasFields -->


### ListTenantsRequest {#listtenantsrequest}
ListTenantsRequest  is the request message to the ListTenants API


| Field | Type | Description |
| ----- | ---- | ----------- |
| pagination | [ public.portworx.common.v1.PageBasedPaginationRequest](#publicportworxcommonv1pagebasedpaginationrequest) | Pagination parameters for listing tenants. |
 <!-- end Fields -->
 <!-- end HasFields -->


### ListTenantsResponse {#listtenantsresponse}
ListTenantsResponse is the response message to the ListTenants API and contains
the list of tenants visible to the caller


| Field | Type | Description |
| ----- | ---- | ----------- |
| tenants | [repeated Tenant](#tenant) | list of tenant response |
| pagination | [ public.portworx.common.v1.PageBasedPaginationResponse](#publicportworxcommonv1pagebasedpaginationresponse) | Pagination metadata for this response. (-- api-linter: core::0132::response-unknown-fields=disabled aip.dev/not-precedent: We need this field for pagination. --) |
 <!-- end Fields -->
 <!-- end HasFields -->


### Phase {#phase}
Phase represents the current status of the tenant.

 <!-- end HasFields -->


### Status {#status}
Status represents the current state of the tenant.


| Field | Type | Description |
| ----- | ---- | ----------- |
| reason | [ string](#string) | Textual information for the current state of the tenant. |
| phase | [ Phase.Type](#phasetype) | Current phase of the project. |
 <!-- end Fields -->
 <!-- end HasFields -->


### Tenant {#tenant}
Tenant is an organizational subunit of an account that represents an org or a unit of a large company.
A tenant comprises multiple projects.


| Field | Type | Description |
| ----- | ---- | ----------- |
| meta | [ public.portworx.common.v1.Meta](#publicportworxcommonv1meta) | Metadata of the tenant. |
| status | [ Status](#status) | status of the tenant |
 <!-- end Fields -->
 <!-- end HasFields -->


### UpdateTenantRequest {#updatetenantrequest}
Request for updating a tenant.


| Field | Type | Description |
| ----- | ---- | ----------- |
| tenant | [ Tenant](#tenant) | tenant which needs to be updated |
 <!-- end Fields -->
 <!-- end HasFields -->
 <!-- end messages -->

## Enums


### Phase.Type {#phasetype}
Type of phase the tenant is in currently should be one of the below.

| Name | Number | Description |
| ---- | ------ | ----------- |
| TYPE_UNSPECIFIED | 0 | Unspecified, do not use. |
| ACTIVE | 1 | The tenant is in use and active. |
| DELETE_PENDING | 2 | Deletion of tenant has not started. |
| DELETE_IN_PROGRESS | 3 | Deletion of the tenant is scheduled and in progress. |


 <!-- end Enums -->
 <!-- end Files -->

## Scalar Value Types

| .proto Type | Notes | C++ Type | Java Type | Python Type |
| ----------- | ----- | -------- | --------- | ----------- |
| <div><h4 id="double" /></div><a name="double" /> double |  | double | double | float |
| <div><h4 id="float" /></div><a name="float" /> float |  | float | float | float |
| <div><h4 id="int32" /></div><a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int |
| <div><h4 id="int64" /></div><a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long |
| <div><h4 id="uint32" /></div><a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long |
| <div><h4 id="uint64" /></div><a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long |
| <div><h4 id="sint32" /></div><a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int |
| <div><h4 id="sint64" /></div><a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long |
| <div><h4 id="fixed32" /></div><a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int |
| <div><h4 id="fixed64" /></div><a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long |
| <div><h4 id="sfixed32" /></div><a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int |
| <div><h4 id="sfixed64" /></div><a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long |
| <div><h4 id="bool" /></div><a name="bool" /> bool |  | bool | boolean | boolean |
| <div><h4 id="string" /></div><a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode |
| <div><h4 id="bytes" /></div><a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str |
