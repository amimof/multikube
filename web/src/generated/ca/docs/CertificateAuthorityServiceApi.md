# CertificateAuthorityServiceApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**certificateAuthorityServiceCreate**](CertificateAuthorityServiceApi.md#certificateauthorityservicecreate) | **POST** /api/v1/certificate_authoritys |  |
| [**certificateAuthorityServiceDelete**](CertificateAuthorityServiceApi.md#certificateauthorityservicedelete) | **DELETE** /api/v1/certificate_authoritys/{uid} |  |
| [**certificateAuthorityServiceDelete2**](CertificateAuthorityServiceApi.md#certificateauthorityservicedelete2) | **DELETE** /api/v1/certificate_authoritys/{name} |  |
| [**certificateAuthorityServiceGet**](CertificateAuthorityServiceApi.md#certificateauthorityserviceget) | **GET** /api/v1/certificate_authoritys/{uid} |  |
| [**certificateAuthorityServiceGet2**](CertificateAuthorityServiceApi.md#certificateauthorityserviceget2) | **GET** /api/v1/certificate_authoritys/{name} |  |
| [**certificateAuthorityServiceList**](CertificateAuthorityServiceApi.md#certificateauthorityservicelist) | **GET** /api/v1/certificate_authoritys |  |
| [**certificateAuthorityServicePatch**](CertificateAuthorityServiceApi.md#certificateauthorityservicepatch) | **PATCH** /api/v1/certificate_authoritys/{uid} |  |
| [**certificateAuthorityServicePatch2**](CertificateAuthorityServiceApi.md#certificateauthorityservicepatch2) | **PATCH** /api/v1/certificate_authoritys/{name} |  |
| [**certificateAuthorityServiceUpdate**](CertificateAuthorityServiceApi.md#certificateauthorityserviceupdate) | **PUT** /api/v1/certificate_authoritys/{uid} |  |
| [**certificateAuthorityServiceUpdate2**](CertificateAuthorityServiceApi.md#certificateauthorityserviceupdate2) | **PUT** /api/v1/certificate_authoritys/{name} |  |



## certificateAuthorityServiceCreate

> V1CreateResponse certificateAuthorityServiceCreate(certificateAuthority)



### Example

```ts
import {
  Configuration,
  CertificateAuthorityServiceApi,
} from '';
import type { CertificateAuthorityServiceCreateRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateAuthorityServiceApi();

  const body = {
    // V1CertificateAuthority
    certificateAuthority: ...,
  } satisfies CertificateAuthorityServiceCreateRequest;

  try {
    const data = await api.certificateAuthorityServiceCreate(body);
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
| **certificateAuthority** | [V1CertificateAuthority](V1CertificateAuthority.md) |  | |

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


## certificateAuthorityServiceDelete

> object certificateAuthorityServiceDelete(uid, name, purge)



### Example

```ts
import {
  Configuration,
  CertificateAuthorityServiceApi,
} from '';
import type { CertificateAuthorityServiceDeleteRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateAuthorityServiceApi();

  const body = {
    // string
    uid: uid_example,
    // string (optional)
    name: name_example,
    // boolean (optional)
    purge: true,
  } satisfies CertificateAuthorityServiceDeleteRequest;

  try {
    const data = await api.certificateAuthorityServiceDelete(body);
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


## certificateAuthorityServiceDelete2

> object certificateAuthorityServiceDelete2(name, uid, purge)



### Example

```ts
import {
  Configuration,
  CertificateAuthorityServiceApi,
} from '';
import type { CertificateAuthorityServiceDelete2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateAuthorityServiceApi();

  const body = {
    // string
    name: name_example,
    // string (optional)
    uid: uid_example,
    // boolean (optional)
    purge: true,
  } satisfies CertificateAuthorityServiceDelete2Request;

  try {
    const data = await api.certificateAuthorityServiceDelete2(body);
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


## certificateAuthorityServiceGet

> V1GetResponse certificateAuthorityServiceGet(uid, name)



### Example

```ts
import {
  Configuration,
  CertificateAuthorityServiceApi,
} from '';
import type { CertificateAuthorityServiceGetRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateAuthorityServiceApi();

  const body = {
    // string
    uid: uid_example,
    // string (optional)
    name: name_example,
  } satisfies CertificateAuthorityServiceGetRequest;

  try {
    const data = await api.certificateAuthorityServiceGet(body);
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


## certificateAuthorityServiceGet2

> V1GetResponse certificateAuthorityServiceGet2(name, uid)



### Example

```ts
import {
  Configuration,
  CertificateAuthorityServiceApi,
} from '';
import type { CertificateAuthorityServiceGet2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateAuthorityServiceApi();

  const body = {
    // string
    name: name_example,
    // string (optional)
    uid: uid_example,
  } satisfies CertificateAuthorityServiceGet2Request;

  try {
    const data = await api.certificateAuthorityServiceGet2(body);
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


## certificateAuthorityServiceList

> V1ListResponse certificateAuthorityServiceList(limit, selector)



### Example

```ts
import {
  Configuration,
  CertificateAuthorityServiceApi,
} from '';
import type { CertificateAuthorityServiceListRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateAuthorityServiceApi();

  const body = {
    // number (optional)
    limit: 56,
    // string | This is a request variable of the map type. The query format is \"map_name[key]=value\", e.g. If the map name is Age, the key type is string, and the value type is integer, the query parameter is expressed as Age[\"bob\"]=18 (optional)
    selector: selector_example,
  } satisfies CertificateAuthorityServiceListRequest;

  try {
    const data = await api.certificateAuthorityServiceList(body);
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


## certificateAuthorityServicePatch

> V1PatchResponse certificateAuthorityServicePatch(uid, certificateAuthority, name)



### Example

```ts
import {
  Configuration,
  CertificateAuthorityServiceApi,
} from '';
import type { CertificateAuthorityServicePatchRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateAuthorityServiceApi();

  const body = {
    // string
    uid: uid_example,
    // V1CertificateAuthority
    certificateAuthority: ...,
    // string (optional)
    name: name_example,
  } satisfies CertificateAuthorityServicePatchRequest;

  try {
    const data = await api.certificateAuthorityServicePatch(body);
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
| **certificateAuthority** | [V1CertificateAuthority](V1CertificateAuthority.md) |  | |
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


## certificateAuthorityServicePatch2

> V1PatchResponse certificateAuthorityServicePatch2(name, certificateAuthority, uid)



### Example

```ts
import {
  Configuration,
  CertificateAuthorityServiceApi,
} from '';
import type { CertificateAuthorityServicePatch2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateAuthorityServiceApi();

  const body = {
    // string
    name: name_example,
    // V1CertificateAuthority
    certificateAuthority: ...,
    // string (optional)
    uid: uid_example,
  } satisfies CertificateAuthorityServicePatch2Request;

  try {
    const data = await api.certificateAuthorityServicePatch2(body);
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
| **certificateAuthority** | [V1CertificateAuthority](V1CertificateAuthority.md) |  | |
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


## certificateAuthorityServiceUpdate

> V1UpdateResponse certificateAuthorityServiceUpdate(uid, certificateAuthority, name, updateMask)



### Example

```ts
import {
  Configuration,
  CertificateAuthorityServiceApi,
} from '';
import type { CertificateAuthorityServiceUpdateRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateAuthorityServiceApi();

  const body = {
    // string
    uid: uid_example,
    // V1CertificateAuthority
    certificateAuthority: ...,
    // string (optional)
    name: name_example,
    // string (optional)
    updateMask: updateMask_example,
  } satisfies CertificateAuthorityServiceUpdateRequest;

  try {
    const data = await api.certificateAuthorityServiceUpdate(body);
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
| **certificateAuthority** | [V1CertificateAuthority](V1CertificateAuthority.md) |  | |
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


## certificateAuthorityServiceUpdate2

> V1UpdateResponse certificateAuthorityServiceUpdate2(name, certificateAuthority, uid, updateMask)



### Example

```ts
import {
  Configuration,
  CertificateAuthorityServiceApi,
} from '';
import type { CertificateAuthorityServiceUpdate2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CertificateAuthorityServiceApi();

  const body = {
    // string
    name: name_example,
    // V1CertificateAuthority
    certificateAuthority: ...,
    // string (optional)
    uid: uid_example,
    // string (optional)
    updateMask: updateMask_example,
  } satisfies CertificateAuthorityServiceUpdate2Request;

  try {
    const data = await api.certificateAuthorityServiceUpdate2(body);
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
| **certificateAuthority** | [V1CertificateAuthority](V1CertificateAuthority.md) |  | |
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

