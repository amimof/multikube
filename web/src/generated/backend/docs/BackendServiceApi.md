# BackendServiceApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**backendServiceCreate**](BackendServiceApi.md#backendservicecreate) | **POST** /api/v1/backends |  |
| [**backendServiceDelete**](BackendServiceApi.md#backendservicedelete) | **DELETE** /api/v1/backends/{uid} |  |
| [**backendServiceDelete2**](BackendServiceApi.md#backendservicedelete2) | **DELETE** /api/v1/backends/{name} |  |
| [**backendServiceGet**](BackendServiceApi.md#backendserviceget) | **GET** /api/v1/backends/{uid} |  |
| [**backendServiceGet2**](BackendServiceApi.md#backendserviceget2) | **GET** /api/v1/backends/{name} |  |
| [**backendServiceList**](BackendServiceApi.md#backendservicelist) | **GET** /api/v1/backends |  |
| [**backendServicePatch**](BackendServiceApi.md#backendservicepatch) | **PATCH** /api/v1/backends/{uid} |  |
| [**backendServicePatch2**](BackendServiceApi.md#backendservicepatch2) | **PATCH** /api/v1/backends/{name} |  |
| [**backendServiceUpdate**](BackendServiceApi.md#backendserviceupdate) | **PUT** /api/v1/backends/{uid} |  |
| [**backendServiceUpdate2**](BackendServiceApi.md#backendserviceupdate2) | **PUT** /api/v1/backends/{name} |  |
| [**backendServiceUpdateStatus**](BackendServiceApi.md#backendserviceupdatestatus) | **PUT** /api/v1/backends/{uid}/status |  |
| [**backendServiceUpdateStatus2**](BackendServiceApi.md#backendserviceupdatestatus2) | **PUT** /api/v1/backends/{name}/status |  |



## backendServiceCreate

> V1CreateResponse backendServiceCreate(backend)



### Example

```ts
import {
  Configuration,
  BackendServiceApi,
} from '';
import type { BackendServiceCreateRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new BackendServiceApi();

  const body = {
    // V1Backend
    backend: ...,
  } satisfies BackendServiceCreateRequest;

  try {
    const data = await api.backendServiceCreate(body);
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
| **backend** | [V1Backend](V1Backend.md) |  | |

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


## backendServiceDelete

> object backendServiceDelete(uid, name, purge)



### Example

```ts
import {
  Configuration,
  BackendServiceApi,
} from '';
import type { BackendServiceDeleteRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new BackendServiceApi();

  const body = {
    // string
    uid: uid_example,
    // string (optional)
    name: name_example,
    // boolean (optional)
    purge: true,
  } satisfies BackendServiceDeleteRequest;

  try {
    const data = await api.backendServiceDelete(body);
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


## backendServiceDelete2

> object backendServiceDelete2(name, uid, purge)



### Example

```ts
import {
  Configuration,
  BackendServiceApi,
} from '';
import type { BackendServiceDelete2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new BackendServiceApi();

  const body = {
    // string
    name: name_example,
    // string (optional)
    uid: uid_example,
    // boolean (optional)
    purge: true,
  } satisfies BackendServiceDelete2Request;

  try {
    const data = await api.backendServiceDelete2(body);
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


## backendServiceGet

> V1GetResponse backendServiceGet(uid, name)



### Example

```ts
import {
  Configuration,
  BackendServiceApi,
} from '';
import type { BackendServiceGetRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new BackendServiceApi();

  const body = {
    // string
    uid: uid_example,
    // string (optional)
    name: name_example,
  } satisfies BackendServiceGetRequest;

  try {
    const data = await api.backendServiceGet(body);
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


## backendServiceGet2

> V1GetResponse backendServiceGet2(name, uid)



### Example

```ts
import {
  Configuration,
  BackendServiceApi,
} from '';
import type { BackendServiceGet2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new BackendServiceApi();

  const body = {
    // string
    name: name_example,
    // string (optional)
    uid: uid_example,
  } satisfies BackendServiceGet2Request;

  try {
    const data = await api.backendServiceGet2(body);
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


## backendServiceList

> V1ListResponse backendServiceList(limit, selector)



### Example

```ts
import {
  Configuration,
  BackendServiceApi,
} from '';
import type { BackendServiceListRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new BackendServiceApi();

  const body = {
    // number (optional)
    limit: 56,
    // string | This is a request variable of the map type. The query format is \"map_name[key]=value\", e.g. If the map name is Age, the key type is string, and the value type is integer, the query parameter is expressed as Age[\"bob\"]=18 (optional)
    selector: selector_example,
  } satisfies BackendServiceListRequest;

  try {
    const data = await api.backendServiceList(body);
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


## backendServicePatch

> V1PatchResponse backendServicePatch(uid, backend, name)



### Example

```ts
import {
  Configuration,
  BackendServiceApi,
} from '';
import type { BackendServicePatchRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new BackendServiceApi();

  const body = {
    // string
    uid: uid_example,
    // V1Backend
    backend: ...,
    // string (optional)
    name: name_example,
  } satisfies BackendServicePatchRequest;

  try {
    const data = await api.backendServicePatch(body);
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
| **backend** | [V1Backend](V1Backend.md) |  | |
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


## backendServicePatch2

> V1PatchResponse backendServicePatch2(name, backend, uid)



### Example

```ts
import {
  Configuration,
  BackendServiceApi,
} from '';
import type { BackendServicePatch2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new BackendServiceApi();

  const body = {
    // string
    name: name_example,
    // V1Backend
    backend: ...,
    // string (optional)
    uid: uid_example,
  } satisfies BackendServicePatch2Request;

  try {
    const data = await api.backendServicePatch2(body);
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
| **backend** | [V1Backend](V1Backend.md) |  | |
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


## backendServiceUpdate

> V1UpdateResponse backendServiceUpdate(uid, backend, name, updateMask)



### Example

```ts
import {
  Configuration,
  BackendServiceApi,
} from '';
import type { BackendServiceUpdateRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new BackendServiceApi();

  const body = {
    // string
    uid: uid_example,
    // V1Backend
    backend: ...,
    // string (optional)
    name: name_example,
    // string (optional)
    updateMask: updateMask_example,
  } satisfies BackendServiceUpdateRequest;

  try {
    const data = await api.backendServiceUpdate(body);
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
| **backend** | [V1Backend](V1Backend.md) |  | |
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


## backendServiceUpdate2

> V1UpdateResponse backendServiceUpdate2(name, backend, uid, updateMask)



### Example

```ts
import {
  Configuration,
  BackendServiceApi,
} from '';
import type { BackendServiceUpdate2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new BackendServiceApi();

  const body = {
    // string
    name: name_example,
    // V1Backend
    backend: ...,
    // string (optional)
    uid: uid_example,
    // string (optional)
    updateMask: updateMask_example,
  } satisfies BackendServiceUpdate2Request;

  try {
    const data = await api.backendServiceUpdate2(body);
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
| **backend** | [V1Backend](V1Backend.md) |  | |
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


## backendServiceUpdateStatus

> object backendServiceUpdateStatus(uid, status, name, updateMask)



### Example

```ts
import {
  Configuration,
  BackendServiceApi,
} from '';
import type { BackendServiceUpdateStatusRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new BackendServiceApi();

  const body = {
    // string
    uid: uid_example,
    // V1BackendStatus
    status: ...,
    // string (optional)
    name: name_example,
    // string (optional)
    updateMask: updateMask_example,
  } satisfies BackendServiceUpdateStatusRequest;

  try {
    const data = await api.backendServiceUpdateStatus(body);
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
| **status** | [V1BackendStatus](V1BackendStatus.md) |  | |
| **name** | `string` |  | [Optional] [Defaults to `undefined`] |
| **updateMask** | `string` |  | [Optional] [Defaults to `undefined`] |

### Return type

**object**

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


## backendServiceUpdateStatus2

> object backendServiceUpdateStatus2(name, status, uid, updateMask)



### Example

```ts
import {
  Configuration,
  BackendServiceApi,
} from '';
import type { BackendServiceUpdateStatus2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new BackendServiceApi();

  const body = {
    // string
    name: name_example,
    // V1BackendStatus
    status: ...,
    // string (optional)
    uid: uid_example,
    // string (optional)
    updateMask: updateMask_example,
  } satisfies BackendServiceUpdateStatus2Request;

  try {
    const data = await api.backendServiceUpdateStatus2(body);
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
| **status** | [V1BackendStatus](V1BackendStatus.md) |  | |
| **uid** | `string` |  | [Optional] [Defaults to `undefined`] |
| **updateMask** | `string` |  | [Optional] [Defaults to `undefined`] |

### Return type

**object**

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

