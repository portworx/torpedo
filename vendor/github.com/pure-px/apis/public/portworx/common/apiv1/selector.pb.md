[//]: # (Generated by grpc-framework using protoc-gen-doc)
[//]: # (Do not edit)


# selector

## Contents


- Messages
    - [ResourceSelector](#resourceselector)
    - [ResourceSelector.ResourceFilter](#resourceselectorresourcefilter)
    - [Selector](#selector)
    - [Selector.Filter](#selectorfilter)
  


- Enums
    - [RespData](#respdata)
    - [Selector.Operator](#selectoroperator)
  


- [Scalar Value Types](#scalar-value-types)



 <!-- end services -->

## Messages


### ResourceSelector {#resourceselector}
ResourceSelector is used to query resources using the associated infra resources.


| Field | Type | Description |
| ----- | ---- | ----------- |
| infra_resource_filters | [repeated ResourceSelector.ResourceFilter](#resourceselectorresourcefilter) | Infra_resource_filters is the list of all filters that should be applied to fetch data related to infra resource. Each filter will have AND relationship. |
 <!-- end Fields -->
 <!-- end HasFields -->


### ResourceSelector.ResourceFilter {#resourceselectorresourcefilter}
ResourceFilter is filter for a given resource type.


| Field | Type | Description |
| ----- | ---- | ----------- |
| resource_type | [ InfraResource.Type](#infraresourcetype) | Key of key,value pair against which filtering needs to be performs based on associated infra resource type. |
| op | [ Selector.Operator](#selectoroperator) | Op provides the relationship between the key,value pair in the resp element(s). |
| values | [repeated string](#string) | Value of key,value pair against which filtering needs to be performs. |
 <!-- end Fields -->
 <!-- end HasFields -->


### Selector {#selector}
Selector is used to query resources using the associated labels or field names.


| Field | Type | Description |
| ----- | ---- | ----------- |
| filters | [repeated Selector.Filter](#selectorfilter) | FilterList is the list of all filters that should be applied. |
 <!-- end Fields -->
 <!-- end HasFields -->


### Selector.Filter {#selectorfilter}
Filter for a given key.


| Field | Type | Description |
| ----- | ---- | ----------- |
| key | [ string](#string) | Key of key,value pair against which filtering needs to be performs. |
| op | [ Selector.Operator](#selectoroperator) | Op provides the relationship between the key,value pair in the resp element(s). |
| values | [repeated string](#string) | Value of key,value pair against which filtering needs to be performs if operator is EXIST, value should be an empty array. |
 <!-- end Fields -->
 <!-- end HasFields -->
 <!-- end messages -->

## Enums


### RespData {#respdata}
RespData provides flags which provides info about the fields that should be populated in the response.

| Name | Number | Description |
| ---- | ------ | ----------- |
| RESP_DATA_UNSPECIFIED | 0 | RespData Unspecified. complete resource will be populated. |
| INDEX | 1 | only uid, name, labels should be populated. |
| LITE | 2 | only meta data should be populated. |
| FULL | 3 | complete resource should be populated. |




### Selector.Operator {#selectoroperator}
Operator specifies the relationship between the provided (key,value) pairs in the response.

| Name | Number | Description |
| ---- | ------ | ----------- |
| OPERATOR_UNSPECIFIED | 0 | Unspecified, do not use. |
| IN | 1 | IN specifies that the key should be associated with atleast 1 of the element in value list. |
| NOT_IN | 2 | NOT_IN specifies that the key should not be associated with any of the element in value list. |


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
