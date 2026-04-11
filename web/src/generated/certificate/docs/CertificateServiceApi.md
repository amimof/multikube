# CertificateServiceApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**certificateServiceCreate**](CertificateServiceApi.md#certificateservicecreate) | **POST** /api/v1/certificates |  |
| [**certificateServiceDelete**](CertificateServiceApi.md#certificateservicedelete) | **DELETE** /api/v1/certificates/{uid} |  |
| [**certificateServiceDelete2**](CertificateServiceApi.md#certificateservicedelete2) | **DELETE** /api/v1/certificates/{name} |  |
| [**certificateServiceGet**](CertificateServiceApi.md#certificateserviceget) | **GET** /api/v1/certificates/{uid} |  |
| [**certificateServiceGet2**](CertificateServiceApi.md#certificateserviceget2) | **GET** /api/v1/certificates/{name} |  |
| [**certificateServiceList**](CertificateServiceApi.md#certificateservicelist) | **GET** /api/v1/certificates |  |
| [**certificateServicePatch**](CertificateServiceApi.md#certificateservicepatch) | **PATCH** /api/v1/certificates/{uid} |  |
| [**certificateServicePatch2**](CertificateServiceApi.md#certificateservicepatch2) | **PATCH** /api/v1/certificates/{name} |  |
| [**certificateServiceUpdate**](CertificateServiceApi.md#certificateserviceupdate) | **PUT** /api/v1/certificates/{uid} |  |
| [**certificateServiceUpdate2**](CertificateServiceApi.md#certificateserviceupdate2) | **PUT** /api/v1/certificates/{name} |  |



## certificateServiceCreate

> V1CreateResponse certificateServiceCreate(certificate)



### Example

```ts
import {
  Configuration,
  CertificateServiceApi,
} from '';
import type { CertificateServiceCreateRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateServiceApi();

  const body = {
    // V1Certificate
    certificate: ...,
  } satisfies CertificateServiceCreateRequest;

  try {
    const data = await api.certificateServiceCreate(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **certificate** | [V1Certificate](V1Certificate.md) |  | |

### Return type

[**V1CreateResponse**](V1CreateResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | A successful response. |  -  |
| **0** | An unexpected error response. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## certificateServiceDelete

> object certificateServiceDelete(uid, name, purge)



### Example

```ts
import {
  Configuration,
  CertificateServiceApi,
} from '';
import type { CertificateServiceDeleteRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateServiceApi();

  const body = {
    // string
    uid: uid_example,
    // string (optional)
    name: name_example,
    // boolean (optional)
    purge: true,
  } satisfies CertificateServiceDeleteRequest;

  try {
    const data = await api.certificateServiceDelete(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **uid** | `string` |  | [Defaults to `undefined`] |
| **name** | `string` |  | [Optional] [Defaults to `undefined`] |
| **purge** | `boolean` |  | [Optional] [Defaults to `undefined`] |

### Return type

**object**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | A successful response. |  -  |
| **0** | An unexpected error response. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## certificateServiceDelete2

> object certificateServiceDelete2(name, uid, purge)



### Example

```ts
import {
  Configuration,
  CertificateServiceApi,
} from '';
import type { CertificateServiceDelete2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateServiceApi();

  const body = {
    // string
    name: name_example,
    // string (optional)
    uid: uid_example,
    // boolean (optional)
    purge: true,
  } satisfies CertificateServiceDelete2Request;

  try {
    const data = await api.certificateServiceDelete2(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **name** | `string` |  | [Defaults to `undefined`] |
| **uid** | `string` |  | [Optional] [Defaults to `undefined`] |
| **purge** | `boolean` |  | [Optional] [Defaults to `undefined`] |

### Return type

**object**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | A successful response. |  -  |
| **0** | An unexpected error response. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## certificateServiceGet

> V1GetResponse certificateServiceGet(uid, name)



### Example

```ts
import {
  Configuration,
  CertificateServiceApi,
} from '';
import type { CertificateServiceGetRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateServiceApi();

  const body = {
    // string
    uid: uid_example,
    // string (optional)
    name: name_example,
  } satisfies CertificateServiceGetRequest;

  try {
    const data = await api.certificateServiceGet(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **uid** | `string` |  | [Defaults to `undefined`] |
| **name** | `string` |  | [Optional] [Defaults to `undefined`] |

### Return type

[**V1GetResponse**](V1GetResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | A successful response. |  -  |
| **0** | An unexpected error response. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## certificateServiceGet2

> V1GetResponse certificateServiceGet2(name, uid)



### Example

```ts
import {
  Configuration,
  CertificateServiceApi,
} from '';
import type { CertificateServiceGet2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateServiceApi();

  const body = {
    // string
    name: name_example,
    // string (optional)
    uid: uid_example,
  } satisfies CertificateServiceGet2Request;

  try {
    const data = await api.certificateServiceGet2(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **name** | `string` |  | [Defaults to `undefined`] |
| **uid** | `string` |  | [Optional] [Defaults to `undefined`] |

### Return type

[**V1GetResponse**](V1GetResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | A successful response. |  -  |
| **0** | An unexpected error response. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## certificateServiceList

> V1ListResponse certificateServiceList(limit, selector)



### Example

```ts
import {
  Configuration,
  CertificateServiceApi,
} from '';
import type { CertificateServiceListRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateServiceApi();

  const body = {
    // number (optional)
    limit: 56,
    // string | This is a request variable of the map type. The query format is \"map_name[key]=value\", e.g. If the map name is Age, the key type is string, and the value type is integer, the query parameter is expressed as Age[\"bob\"]=18 (optional)
    selector: selector_example,
  } satisfies CertificateServiceListRequest;

  try {
    const data = await api.certificateServiceList(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **limit** | `number` |  | [Optional] [Defaults to `undefined`] |
| **selector** | `string` | This is a request variable of the map type. The query format is \&quot;map_name[key]&#x3D;value\&quot;, e.g. If the map name is Age, the key type is string, and the value type is integer, the query parameter is expressed as Age[\&quot;bob\&quot;]&#x3D;18 | [Optional] [Defaults to `undefined`] |

### Return type

[**V1ListResponse**](V1ListResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | A successful response. |  -  |
| **0** | An unexpected error response. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## certificateServicePatch

> V1PatchResponse certificateServicePatch(uid, certificate, name)



### Example

```ts
import {
  Configuration,
  CertificateServiceApi,
} from '';
import type { CertificateServicePatchRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateServiceApi();

  const body = {
    // string
    uid: uid_example,
    // V1Certificate
    certificate: ...,
    // string (optional)
    name: name_example,
  } satisfies CertificateServicePatchRequest;

  try {
    const data = await api.certificateServicePatch(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **uid** | `string` |  | [Defaults to `undefined`] |
| **certificate** | [V1Certificate](V1Certificate.md) |  | |
| **name** | `string` |  | [Optional] [Defaults to `undefined`] |

### Return type

[**V1PatchResponse**](V1PatchResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | A successful response. |  -  |
| **0** | An unexpected error response. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## certificateServicePatch2

> V1PatchResponse certificateServicePatch2(name, certificate, uid)



### Example

```ts
import {
  Configuration,
  CertificateServiceApi,
} from '';
import type { CertificateServicePatch2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateServiceApi();

  const body = {
    // string
    name: name_example,
    // V1Certificate
    certificate: ...,
    // string (optional)
    uid: uid_example,
  } satisfies CertificateServicePatch2Request;

  try {
    const data = await api.certificateServicePatch2(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **name** | `string` |  | [Defaults to `undefined`] |
| **certificate** | [V1Certificate](V1Certificate.md) |  | |
| **uid** | `string` |  | [Optional] [Defaults to `undefined`] |

### Return type

[**V1PatchResponse**](V1PatchResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | A successful response. |  -  |
| **0** | An unexpected error response. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## certificateServiceUpdate

> V1UpdateResponse certificateServiceUpdate(uid, certificate, name, updateMask)



### Example

```ts
import {
  Configuration,
  CertificateServiceApi,
} from '';
import type { CertificateServiceUpdateRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateServiceApi();

  const body = {
    // string
    uid: uid_example,
    // V1Certificate
    certificate: ...,
    // string (optional)
    name: name_example,
    // string (optional)
    updateMask: updateMask_example,
  } satisfies CertificateServiceUpdateRequest;

  try {
    const data = await api.certificateServiceUpdate(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **uid** | `string` |  | [Defaults to `undefined`] |
| **certificate** | [V1Certificate](V1Certificate.md) |  | |
| **name** | `string` |  | [Optional] [Defaults to `undefined`] |
| **updateMask** | `string` |  | [Optional] [Defaults to `undefined`] |

### Return type

[**V1UpdateResponse**](V1UpdateResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | A successful response. |  -  |
| **0** | An unexpected error response. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## certificateServiceUpdate2

> V1UpdateResponse certificateServiceUpdate2(name, certificate, uid, updateMask)



### Example

```ts
import {
  Configuration,
  CertificateServiceApi,
} from '';
import type { CertificateServiceUpdate2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateServiceApi();

  const body = {
    // string
    name: name_example,
    // V1Certificate
    certificate: ...,
    // string (optional)
    uid: uid_example,
    // string (optional)
    updateMask: updateMask_example,
  } satisfies CertificateServiceUpdate2Request;

  try {
    const data = await api.certificateServiceUpdate2(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **name** | `string` |  | [Defaults to `undefined`] |
| **certificate** | [V1Certificate](V1Certificate.md) |  | |
| **uid** | `string` |  | [Optional] [Defaults to `undefined`] |
| **updateMask** | `string` |  | [Optional] [Defaults to `undefined`] |

### Return type

[**V1UpdateResponse**](V1UpdateResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | A successful response. |  -  |
| **0** | An unexpected error response. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)

