/*
public/portworx/platform/template/apiv1/template.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package template

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
)


// TemplateServiceAPIService TemplateServiceAPI service
type TemplateServiceAPIService service

type ApiTemplateServiceCreateTemplateRequest struct {
	ctx context.Context
	ApiService *TemplateServiceAPIService
	tenantId string
	v1Template *V1Template
}

// Information about the template instance that needs to be created.
func (r ApiTemplateServiceCreateTemplateRequest) V1Template(v1Template V1Template) ApiTemplateServiceCreateTemplateRequest {
	r.v1Template = &v1Template
	return r
}

func (r ApiTemplateServiceCreateTemplateRequest) Execute() (*V1Template, *http.Response, error) {
	return r.ApiService.TemplateServiceCreateTemplateExecute(r)
}

/*
TemplateServiceCreateTemplate Create API creates a set of templates for a tenant.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param tenantId The parent tenant id under which template will be created.
 @return ApiTemplateServiceCreateTemplateRequest
*/
func (a *TemplateServiceAPIService) TemplateServiceCreateTemplate(ctx context.Context, tenantId string) ApiTemplateServiceCreateTemplateRequest {
	return ApiTemplateServiceCreateTemplateRequest{
		ApiService: a,
		ctx: ctx,
		tenantId: tenantId,
	}
}

// Execute executes the request
//  @return V1Template
func (a *TemplateServiceAPIService) TemplateServiceCreateTemplateExecute(r ApiTemplateServiceCreateTemplateRequest) (*V1Template, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *V1Template
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "TemplateServiceAPIService.TemplateServiceCreateTemplate")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/core/v1/tenants/{tenantId}/templates"
	localVarPath = strings.Replace(localVarPath, "{"+"tenantId"+"}", url.PathEscape(parameterValueToString(r.tenantId, "tenantId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.v1Template == nil {
		return localVarReturnValue, nil, reportError("v1Template is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.v1Template
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
			var v GooglerpcStatus
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
					newErr.error = formatErrorMessage(localVarHTTPResponse.Status, &v)
					newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiTemplateServiceDeleteTemplateRequest struct {
	ctx context.Context
	ApiService *TemplateServiceAPIService
	id string
}

func (r ApiTemplateServiceDeleteTemplateRequest) Execute() (map[string]interface{}, *http.Response, error) {
	return r.ApiService.TemplateServiceDeleteTemplateExecute(r)
}

/*
TemplateServiceDeleteTemplate Delete API deletes the templates.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id ID of the template instance.
 @return ApiTemplateServiceDeleteTemplateRequest
*/
func (a *TemplateServiceAPIService) TemplateServiceDeleteTemplate(ctx context.Context, id string) ApiTemplateServiceDeleteTemplateRequest {
	return ApiTemplateServiceDeleteTemplateRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
//  @return map[string]interface{}
func (a *TemplateServiceAPIService) TemplateServiceDeleteTemplateExecute(r ApiTemplateServiceDeleteTemplateRequest) (map[string]interface{}, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodDelete
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  map[string]interface{}
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "TemplateServiceAPIService.TemplateServiceDeleteTemplate")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/core/v1/templates/{id}"
	localVarPath = strings.Replace(localVarPath, "{"+"id"+"}", url.PathEscape(parameterValueToString(r.id, "id")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
			var v GooglerpcStatus
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
					newErr.error = formatErrorMessage(localVarHTTPResponse.Status, &v)
					newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiTemplateServiceGetTemplateRequest struct {
	ctx context.Context
	ApiService *TemplateServiceAPIService
	id string
}

func (r ApiTemplateServiceGetTemplateRequest) Execute() (*V1Template, *http.Response, error) {
	return r.ApiService.TemplateServiceGetTemplateExecute(r)
}

/*
TemplateServiceGetTemplate Get API returns the template details sans the actual credentials.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id ID of the template instance.
 @return ApiTemplateServiceGetTemplateRequest
*/
func (a *TemplateServiceAPIService) TemplateServiceGetTemplate(ctx context.Context, id string) ApiTemplateServiceGetTemplateRequest {
	return ApiTemplateServiceGetTemplateRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
//  @return V1Template
func (a *TemplateServiceAPIService) TemplateServiceGetTemplateExecute(r ApiTemplateServiceGetTemplateRequest) (*V1Template, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *V1Template
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "TemplateServiceAPIService.TemplateServiceGetTemplate")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/core/v1/templates/{id}"
	localVarPath = strings.Replace(localVarPath, "{"+"id"+"}", url.PathEscape(parameterValueToString(r.id, "id")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
			var v GooglerpcStatus
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
					newErr.error = formatErrorMessage(localVarHTTPResponse.Status, &v)
					newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiTemplateServiceListTemplatesRequest struct {
	ctx context.Context
	ApiService *TemplateServiceAPIService
	tenantId *string
	paginationPageNumber *string
	paginationPageSize *string
	respData *string
	sortSortBy *string
	sortSortOrder *string
}

// Tenant ID for which the credentials will be listed.
func (r ApiTemplateServiceListTemplatesRequest) TenantId(tenantId string) ApiTemplateServiceListTemplatesRequest {
	r.tenantId = &tenantId
	return r
}

// Page number is the page number to return based on the size.
func (r ApiTemplateServiceListTemplatesRequest) PaginationPageNumber(paginationPageNumber string) ApiTemplateServiceListTemplatesRequest {
	r.paginationPageNumber = &paginationPageNumber
	return r
}

// Page size is the maximum number of records to include per page.
func (r ApiTemplateServiceListTemplatesRequest) PaginationPageSize(paginationPageSize string) ApiTemplateServiceListTemplatesRequest {
	r.paginationPageSize = &paginationPageSize
	return r
}

// Response data flags for listing templates.   - RESP_DATA_UNSPECIFIED: RespData Unspecified. complete resource will be populated.  - INDEX: only uid, name, labels should be populated.  - LITE: only meta data should be populated.  - FULL: complete resource should be populated.
func (r ApiTemplateServiceListTemplatesRequest) RespData(respData string) ApiTemplateServiceListTemplatesRequest {
	r.respData = &respData
	return r
}

// Name of the attribute to sort results by.   - FIELD_UNSPECIFIED: Unspecified, do not use.  - NAME: Sorting based on the name of the resource.  - CREATED_AT: Sorting on create time of the resource.  - UPDATED_AT: Sorting on update time of the resource.  - PHASE: Sorting on phase of the resource.
func (r ApiTemplateServiceListTemplatesRequest) SortSortBy(sortSortBy string) ApiTemplateServiceListTemplatesRequest {
	r.sortSortBy = &sortSortBy
	return r
}

// Order of sorting to be applied on requested list. If sort_by having some value and sort_order is not provided, by default ascending order will be used to sort the list.   - VALUE_UNSPECIFIED: Unspecified, do not use.  - ASC: Sort order ascending.  - DESC: Sort order descending.
func (r ApiTemplateServiceListTemplatesRequest) SortSortOrder(sortSortOrder string) ApiTemplateServiceListTemplatesRequest {
	r.sortSortOrder = &sortSortOrder
	return r
}

func (r ApiTemplateServiceListTemplatesRequest) Execute() (*V1ListTemplatesResponse, *http.Response, error) {
	return r.ApiService.TemplateServiceListTemplatesExecute(r)
}

/*
TemplateServiceListTemplates List API lists all the templates for a tenant

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiTemplateServiceListTemplatesRequest
*/
func (a *TemplateServiceAPIService) TemplateServiceListTemplates(ctx context.Context) ApiTemplateServiceListTemplatesRequest {
	return ApiTemplateServiceListTemplatesRequest{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
//  @return V1ListTemplatesResponse
func (a *TemplateServiceAPIService) TemplateServiceListTemplatesExecute(r ApiTemplateServiceListTemplatesRequest) (*V1ListTemplatesResponse, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *V1ListTemplatesResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "TemplateServiceAPIService.TemplateServiceListTemplates")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/core/v1/templates"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.tenantId != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "tenantId", r.tenantId, "")
	}
	if r.paginationPageNumber != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "pagination.pageNumber", r.paginationPageNumber, "")
	}
	if r.paginationPageSize != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "pagination.pageSize", r.paginationPageSize, "")
	}
	if r.respData != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "respData", r.respData, "")
	} else {
		var defaultValue string = "RESP_DATA_UNSPECIFIED"
		r.respData = &defaultValue
	}
	if r.sortSortBy != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "sort.sortBy", r.sortSortBy, "")
	} else {
		var defaultValue string = "FIELD_UNSPECIFIED"
		r.sortSortBy = &defaultValue
	}
	if r.sortSortOrder != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "sort.sortOrder", r.sortSortOrder, "")
	} else {
		var defaultValue string = "VALUE_UNSPECIFIED"
		r.sortSortOrder = &defaultValue
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
			var v GooglerpcStatus
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
					newErr.error = formatErrorMessage(localVarHTTPResponse.Status, &v)
					newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiTemplateServiceListTemplates2Request struct {
	ctx context.Context
	ApiService *TemplateServiceAPIService
	v1ListTemplatesRequest *V1ListTemplatesRequest
}

// Request to list the templates for a tenant.
func (r ApiTemplateServiceListTemplates2Request) V1ListTemplatesRequest(v1ListTemplatesRequest V1ListTemplatesRequest) ApiTemplateServiceListTemplates2Request {
	r.v1ListTemplatesRequest = &v1ListTemplatesRequest
	return r
}

func (r ApiTemplateServiceListTemplates2Request) Execute() (*V1ListTemplatesResponse, *http.Response, error) {
	return r.ApiService.TemplateServiceListTemplates2Execute(r)
}

/*
TemplateServiceListTemplates2 List API lists all the templates for a tenant

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiTemplateServiceListTemplates2Request
*/
func (a *TemplateServiceAPIService) TemplateServiceListTemplates2(ctx context.Context) ApiTemplateServiceListTemplates2Request {
	return ApiTemplateServiceListTemplates2Request{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
//  @return V1ListTemplatesResponse
func (a *TemplateServiceAPIService) TemplateServiceListTemplates2Execute(r ApiTemplateServiceListTemplates2Request) (*V1ListTemplatesResponse, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *V1ListTemplatesResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "TemplateServiceAPIService.TemplateServiceListTemplates2")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/core/v1/templates:search"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.v1ListTemplatesRequest == nil {
		return localVarReturnValue, nil, reportError("v1ListTemplatesRequest is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.v1ListTemplatesRequest
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
			var v GooglerpcStatus
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
					newErr.error = formatErrorMessage(localVarHTTPResponse.Status, &v)
					newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiTemplateServiceUpdateTemplateRequest struct {
	ctx context.Context
	ApiService *TemplateServiceAPIService
	id string
	v1Template *V1Template
}

// Desired template configuration.
func (r ApiTemplateServiceUpdateTemplateRequest) V1Template(v1Template V1Template) ApiTemplateServiceUpdateTemplateRequest {
	r.v1Template = &v1Template
	return r
}

func (r ApiTemplateServiceUpdateTemplateRequest) Execute() (*V1Template, *http.Response, error) {
	return r.ApiService.TemplateServiceUpdateTemplateExecute(r)
}

/*
TemplateServiceUpdateTemplate Update API updates a template.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id ID of the template to be updated.
 @return ApiTemplateServiceUpdateTemplateRequest
*/
func (a *TemplateServiceAPIService) TemplateServiceUpdateTemplate(ctx context.Context, id string) ApiTemplateServiceUpdateTemplateRequest {
	return ApiTemplateServiceUpdateTemplateRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
//  @return V1Template
func (a *TemplateServiceAPIService) TemplateServiceUpdateTemplateExecute(r ApiTemplateServiceUpdateTemplateRequest) (*V1Template, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPut
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *V1Template
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "TemplateServiceAPIService.TemplateServiceUpdateTemplate")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/core/v1/templates/{id}"
	localVarPath = strings.Replace(localVarPath, "{"+"id"+"}", url.PathEscape(parameterValueToString(r.id, "id")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.v1Template == nil {
		return localVarReturnValue, nil, reportError("v1Template is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.v1Template
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
			var v GooglerpcStatus
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
					newErr.error = formatErrorMessage(localVarHTTPResponse.Status, &v)
					newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}