/*
public/portworx/platform/serviceaccount/apiv1/serviceaccount.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package serviceaccount

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
)


// ServiceAccountServiceAPIService ServiceAccountServiceAPI service
type ServiceAccountServiceAPIService service

type ApiServiceAccountServiceCreateServiceAccountRequest struct {
	ctx context.Context
	ApiService *ServiceAccountServiceAPIService
	tenantId string
	v1ServiceAccount *V1ServiceAccount
}

// Service account details.
func (r ApiServiceAccountServiceCreateServiceAccountRequest) V1ServiceAccount(v1ServiceAccount V1ServiceAccount) ApiServiceAccountServiceCreateServiceAccountRequest {
	r.v1ServiceAccount = &v1ServiceAccount
	return r
}

func (r ApiServiceAccountServiceCreateServiceAccountRequest) Execute() (*V1ServiceAccount, *http.Response, error) {
	return r.ApiService.ServiceAccountServiceCreateServiceAccountExecute(r)
}

/*
ServiceAccountServiceCreateServiceAccount Create a requested service account.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param tenantId The parent tenant under which the service account will be created
 @return ApiServiceAccountServiceCreateServiceAccountRequest
*/
func (a *ServiceAccountServiceAPIService) ServiceAccountServiceCreateServiceAccount(ctx context.Context, tenantId string) ApiServiceAccountServiceCreateServiceAccountRequest {
	return ApiServiceAccountServiceCreateServiceAccountRequest{
		ApiService: a,
		ctx: ctx,
		tenantId: tenantId,
	}
}

// Execute executes the request
//  @return V1ServiceAccount
func (a *ServiceAccountServiceAPIService) ServiceAccountServiceCreateServiceAccountExecute(r ApiServiceAccountServiceCreateServiceAccountRequest) (*V1ServiceAccount, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *V1ServiceAccount
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ServiceAccountServiceAPIService.ServiceAccountServiceCreateServiceAccount")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v1/tenants/{tenantId}/serviceAccounts"
	localVarPath = strings.Replace(localVarPath, "{"+"tenantId"+"}", url.PathEscape(parameterValueToString(r.tenantId, "tenantId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.v1ServiceAccount == nil {
		return localVarReturnValue, nil, reportError("v1ServiceAccount is required and must be specified")
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
	localVarPostBody = r.v1ServiceAccount
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

type ApiServiceAccountServiceDeleteServiceAccountRequest struct {
	ctx context.Context
	ApiService *ServiceAccountServiceAPIService
	id string
}

func (r ApiServiceAccountServiceDeleteServiceAccountRequest) Execute() (map[string]interface{}, *http.Response, error) {
	return r.ApiService.ServiceAccountServiceDeleteServiceAccountExecute(r)
}

/*
ServiceAccountServiceDeleteServiceAccount Initiates deletion of a service account.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id Unique identifier for the service account to be deleted.
 @return ApiServiceAccountServiceDeleteServiceAccountRequest
*/
func (a *ServiceAccountServiceAPIService) ServiceAccountServiceDeleteServiceAccount(ctx context.Context, id string) ApiServiceAccountServiceDeleteServiceAccountRequest {
	return ApiServiceAccountServiceDeleteServiceAccountRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
//  @return map[string]interface{}
func (a *ServiceAccountServiceAPIService) ServiceAccountServiceDeleteServiceAccountExecute(r ApiServiceAccountServiceDeleteServiceAccountRequest) (map[string]interface{}, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodDelete
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  map[string]interface{}
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ServiceAccountServiceAPIService.ServiceAccountServiceDeleteServiceAccount")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v1/serviceAccounts/{id}"
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

type ApiServiceAccountServiceGetAccessTokenRequest struct {
	ctx context.Context
	ApiService *ServiceAccountServiceAPIService
	tenantId string
	serviceAccountServiceGetAccessTokenBody *ServiceAccountServiceGetAccessTokenBody
}

func (r ApiServiceAccountServiceGetAccessTokenRequest) ServiceAccountServiceGetAccessTokenBody(serviceAccountServiceGetAccessTokenBody ServiceAccountServiceGetAccessTokenBody) ApiServiceAccountServiceGetAccessTokenRequest {
	r.serviceAccountServiceGetAccessTokenBody = &serviceAccountServiceGetAccessTokenBody
	return r
}

func (r ApiServiceAccountServiceGetAccessTokenRequest) Execute() (*V1AccessToken, *http.Response, error) {
	return r.ApiService.ServiceAccountServiceGetAccessTokenExecute(r)
}

/*
ServiceAccountServiceGetAccessToken Get access token for a service account.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param tenantId ID of the tenant under which the service account was created.
 @return ApiServiceAccountServiceGetAccessTokenRequest
*/
func (a *ServiceAccountServiceAPIService) ServiceAccountServiceGetAccessToken(ctx context.Context, tenantId string) ApiServiceAccountServiceGetAccessTokenRequest {
	return ApiServiceAccountServiceGetAccessTokenRequest{
		ApiService: a,
		ctx: ctx,
		tenantId: tenantId,
	}
}

// Execute executes the request
//  @return V1AccessToken
func (a *ServiceAccountServiceAPIService) ServiceAccountServiceGetAccessTokenExecute(r ApiServiceAccountServiceGetAccessTokenRequest) (*V1AccessToken, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *V1AccessToken
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ServiceAccountServiceAPIService.ServiceAccountServiceGetAccessToken")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v1/tenants/{tenantId}:getToken"
	localVarPath = strings.Replace(localVarPath, "{"+"tenantId"+"}", url.PathEscape(parameterValueToString(r.tenantId, "tenantId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.serviceAccountServiceGetAccessTokenBody == nil {
		return localVarReturnValue, nil, reportError("serviceAccountServiceGetAccessTokenBody is required and must be specified")
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
	localVarPostBody = r.serviceAccountServiceGetAccessTokenBody
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

type ApiServiceAccountServiceGetServiceAccountRequest struct {
	ctx context.Context
	ApiService *ServiceAccountServiceAPIService
	id string
	tenantId *string
}

// Tenant id to which a service account is associated.
func (r ApiServiceAccountServiceGetServiceAccountRequest) TenantId(tenantId string) ApiServiceAccountServiceGetServiceAccountRequest {
	r.tenantId = &tenantId
	return r
}

func (r ApiServiceAccountServiceGetServiceAccountRequest) Execute() (*V1ServiceAccount, *http.Response, error) {
	return r.ApiService.ServiceAccountServiceGetServiceAccountExecute(r)
}

/*
ServiceAccountServiceGetServiceAccount Returns a requested service account.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id Unique identifier for the service account to be fetched.
 @return ApiServiceAccountServiceGetServiceAccountRequest
*/
func (a *ServiceAccountServiceAPIService) ServiceAccountServiceGetServiceAccount(ctx context.Context, id string) ApiServiceAccountServiceGetServiceAccountRequest {
	return ApiServiceAccountServiceGetServiceAccountRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
//  @return V1ServiceAccount
func (a *ServiceAccountServiceAPIService) ServiceAccountServiceGetServiceAccountExecute(r ApiServiceAccountServiceGetServiceAccountRequest) (*V1ServiceAccount, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *V1ServiceAccount
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ServiceAccountServiceAPIService.ServiceAccountServiceGetServiceAccount")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v1/serviceAccounts/{id}"
	localVarPath = strings.Replace(localVarPath, "{"+"id"+"}", url.PathEscape(parameterValueToString(r.id, "id")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.tenantId != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "tenantId", r.tenantId, "")
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

type ApiServiceAccountServiceListServiceAccountRequest struct {
	ctx context.Context
	ApiService *ServiceAccountServiceAPIService
	tenantId *string
	sortSortBy *string
	sortSortOrder *string
	paginationPageNumber *string
	paginationPageSize *string
}

// id of tenant on which service account should be listed. If not provided, then list will filtered on account id present in the context.
func (r ApiServiceAccountServiceListServiceAccountRequest) TenantId(tenantId string) ApiServiceAccountServiceListServiceAccountRequest {
	r.tenantId = &tenantId
	return r
}

// Name of the attribute to sort results by.   - FIELD_UNSPECIFIED: Unspecified, do not use.  - NAME: Sorting based on the name of the resource.  - CREATED_AT: Sorting on create time of the resource.  - UPDATED_AT: Sorting on update time of the resource.  - PHASE: Sorting on phase of the resource.
func (r ApiServiceAccountServiceListServiceAccountRequest) SortSortBy(sortSortBy string) ApiServiceAccountServiceListServiceAccountRequest {
	r.sortSortBy = &sortSortBy
	return r
}

// Order of sorting to be applied on requested list. If sort_by having some value and sort_order is not provided, by default ascending order will be used to sort the list.   - VALUE_UNSPECIFIED: Unspecified, do not use.  - ASC: Sort order ascending.  - DESC: Sort order descending.
func (r ApiServiceAccountServiceListServiceAccountRequest) SortSortOrder(sortSortOrder string) ApiServiceAccountServiceListServiceAccountRequest {
	r.sortSortOrder = &sortSortOrder
	return r
}

// Page number is the page number to return based on the size.
func (r ApiServiceAccountServiceListServiceAccountRequest) PaginationPageNumber(paginationPageNumber string) ApiServiceAccountServiceListServiceAccountRequest {
	r.paginationPageNumber = &paginationPageNumber
	return r
}

// Page size is the maximum number of records to include per page.
func (r ApiServiceAccountServiceListServiceAccountRequest) PaginationPageSize(paginationPageSize string) ApiServiceAccountServiceListServiceAccountRequest {
	r.paginationPageSize = &paginationPageSize
	return r
}

func (r ApiServiceAccountServiceListServiceAccountRequest) Execute() (*V1ListServiceAccountResponse, *http.Response, error) {
	return r.ApiService.ServiceAccountServiceListServiceAccountExecute(r)
}

/*
ServiceAccountServiceListServiceAccount Returns a requested list of service accounts.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiServiceAccountServiceListServiceAccountRequest
*/
func (a *ServiceAccountServiceAPIService) ServiceAccountServiceListServiceAccount(ctx context.Context) ApiServiceAccountServiceListServiceAccountRequest {
	return ApiServiceAccountServiceListServiceAccountRequest{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
//  @return V1ListServiceAccountResponse
func (a *ServiceAccountServiceAPIService) ServiceAccountServiceListServiceAccountExecute(r ApiServiceAccountServiceListServiceAccountRequest) (*V1ListServiceAccountResponse, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *V1ListServiceAccountResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ServiceAccountServiceAPIService.ServiceAccountServiceListServiceAccount")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v1/serviceAccounts"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.tenantId != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "tenantId", r.tenantId, "")
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
	if r.paginationPageNumber != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "pagination.pageNumber", r.paginationPageNumber, "")
	}
	if r.paginationPageSize != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "pagination.pageSize", r.paginationPageSize, "")
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

type ApiServiceAccountServiceRegenerateServiceAccountSecretRequest struct {
	ctx context.Context
	ApiService *ServiceAccountServiceAPIService
	id string
}

func (r ApiServiceAccountServiceRegenerateServiceAccountSecretRequest) Execute() (*V1ServiceAccount, *http.Response, error) {
	return r.ApiService.ServiceAccountServiceRegenerateServiceAccountSecretExecute(r)
}

/*
ServiceAccountServiceRegenerateServiceAccountSecret Regenerate access token for a service account.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id Unique identifier for the service account.
 @return ApiServiceAccountServiceRegenerateServiceAccountSecretRequest
*/
func (a *ServiceAccountServiceAPIService) ServiceAccountServiceRegenerateServiceAccountSecret(ctx context.Context, id string) ApiServiceAccountServiceRegenerateServiceAccountSecretRequest {
	return ApiServiceAccountServiceRegenerateServiceAccountSecretRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
//  @return V1ServiceAccount
func (a *ServiceAccountServiceAPIService) ServiceAccountServiceRegenerateServiceAccountSecretExecute(r ApiServiceAccountServiceRegenerateServiceAccountSecretRequest) (*V1ServiceAccount, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *V1ServiceAccount
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ServiceAccountServiceAPIService.ServiceAccountServiceRegenerateServiceAccountSecret")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v1/serviceAccounts/{id}:regenerate"
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

type ApiServiceAccountServiceUpdateServiceAccountRequest struct {
	ctx context.Context
	ApiService *ServiceAccountServiceAPIService
	id string
	v1ServiceAccount *V1ServiceAccount
}

// Service account to be updated.
func (r ApiServiceAccountServiceUpdateServiceAccountRequest) V1ServiceAccount(v1ServiceAccount V1ServiceAccount) ApiServiceAccountServiceUpdateServiceAccountRequest {
	r.v1ServiceAccount = &v1ServiceAccount
	return r
}

func (r ApiServiceAccountServiceUpdateServiceAccountRequest) Execute() (*V1ServiceAccount, *http.Response, error) {
	return r.ApiService.ServiceAccountServiceUpdateServiceAccountExecute(r)
}

/*
ServiceAccountServiceUpdateServiceAccount Updates a service account.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id Unique identifier for the service account.
 @return ApiServiceAccountServiceUpdateServiceAccountRequest
*/
func (a *ServiceAccountServiceAPIService) ServiceAccountServiceUpdateServiceAccount(ctx context.Context, id string) ApiServiceAccountServiceUpdateServiceAccountRequest {
	return ApiServiceAccountServiceUpdateServiceAccountRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
//  @return V1ServiceAccount
func (a *ServiceAccountServiceAPIService) ServiceAccountServiceUpdateServiceAccountExecute(r ApiServiceAccountServiceUpdateServiceAccountRequest) (*V1ServiceAccount, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPut
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *V1ServiceAccount
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ServiceAccountServiceAPIService.ServiceAccountServiceUpdateServiceAccount")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v1/serviceAccounts/{id}"
	localVarPath = strings.Replace(localVarPath, "{"+"id"+"}", url.PathEscape(parameterValueToString(r.id, "id")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.v1ServiceAccount == nil {
		return localVarReturnValue, nil, reportError("v1ServiceAccount is required and must be specified")
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
	localVarPostBody = r.v1ServiceAccount
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