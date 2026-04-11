# RouteServiceApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**routeServiceCreate**](RouteServiceApi.md#routeservicecreate) | **POST** /api/v1/routes |  |
| [**routeServiceDelete**](RouteServiceApi.md#routeservicedelete) | **DELETE** /api/v1/routes/{uid} |  |
| [**routeServiceDelete2**](RouteServiceApi.md#routeservicedelete2) | **DELETE** /api/v1/routes/{name} |  |
| [**routeServiceGet**](RouteServiceApi.md#routeserviceget) | **GET** /api/v1/routes/{uid} |  |
| [**routeServiceGet2**](RouteServiceApi.md#routeserviceget2) | **GET** /api/v1/routes/{name} |  |
| [**routeServiceList**](RouteServiceApi.md#routeservicelist) | **GET** /api/v1/routes |  |
| [**routeServicePatch**](RouteServiceApi.md#routeservicepatch) | **PATCH** /api/v1/routes/{uid} |  |
| [**routeServicePatch2**](RouteServiceApi.md#routeservicepatch2) | **PATCH** /api/v1/routes/{name} |  |
| [**routeServiceUpdate**](RouteServiceApi.md#routeserviceupdate) | **PUT** /api/v1/routes/{uid} |  |
| [**routeServiceUpdate2**](RouteServiceApi.md#routeserviceupdate2) | **PUT** /api/v1/routes/{name} |  |
| [**routeServiceUpdateStatus**](RouteServiceApi.md#routeserviceupdatestatus) | **PUT** /api/v1/routes/{uid}/status |  |
| [**routeServiceUpdateStatus2**](RouteServiceApi.md#routeserviceupdatestatus2) | **PUT** /api/v1/routes/{name}/status |  |



## routeServiceCreate

> V1CreateResponse routeServiceCreate(route)



### Example

```ts
import {
  Configuration,
  RouteServiceApi,
} from '';
import type { RouteServiceCreateRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new RouteServiceApi();

  const body = {
    // V1Route
    route: ...,
  } satisfies RouteServiceCreateRequest;

  try {
    const data = await api.routeServiceCreate(body);
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
| **route** | [V1Route](V1Route.md) |  | |

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


## routeServiceDelete

> object routeServiceDelete(uid, name, purge)



### Example

```ts
import {
  Configuration,
  RouteServiceApi,
} from '';
import type { RouteServiceDeleteRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new RouteServiceApi();

  const body = {
    // string
    uid: uid_example,
    // string (optional)
    name: name_example,
    // boolean (optional)
    purge: true,
  } satisfies RouteServiceDeleteRequest;

  try {
    const data = await api.routeServiceDelete(body);
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


## routeServiceDelete2

> object routeServiceDelete2(name, uid, purge)



### Example

```ts
import {
  Configuration,
  RouteServiceApi,
} from '';
import type { RouteServiceDelete2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new RouteServiceApi();

  const body = {
    // string
    name: name_example,
    // string (optional)
    uid: uid_example,
    // boolean (optional)
    purge: true,
  } satisfies RouteServiceDelete2Request;

  try {
    const data = await api.routeServiceDelete2(body);
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


## routeServiceGet

> V1GetResponse routeServiceGet(uid, name)



### Example

```ts
import {
  Configuration,
  RouteServiceApi,
} from '';
import type { RouteServiceGetRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new RouteServiceApi();

  const body = {
    // string
    uid: uid_example,
    // string (optional)
    name: name_example,
  } satisfies RouteServiceGetRequest;

  try {
    const data = await api.routeServiceGet(body);
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


## routeServiceGet2

> V1GetResponse routeServiceGet2(name, uid)



### Example

```ts
import {
  Configuration,
  RouteServiceApi,
} from '';
import type { RouteServiceGet2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new RouteServiceApi();

  const body = {
    // string
    name: name_example,
    // string (optional)
    uid: uid_example,
  } satisfies RouteServiceGet2Request;

  try {
    const data = await api.routeServiceGet2(body);
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


## routeServiceList

> V1ListResponse routeServiceList(limit, selector)



### Example

```ts
import {
  Configuration,
  RouteServiceApi,
} from '';
import type { RouteServiceListRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new RouteServiceApi();

  const body = {
    // number (optional)
    limit: 56,
    // string | This is a request variable of the map type. The query format is \"map_name[key]=value\", e.g. If the map name is Age, the key type is string, and the value type is integer, the query parameter is expressed as Age[\"bob\"]=18 (optional)
    selector: selector_example,
  } satisfies RouteServiceListRequest;

  try {
    const data = await api.routeServiceList(body);
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


## routeServicePatch

> V1PatchResponse routeServicePatch(uid, route, name)



### Example

```ts
import {
  Configuration,
  RouteServiceApi,
} from '';
import type { RouteServicePatchRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new RouteServiceApi();

  const body = {
    // string
    uid: uid_example,
    // V1Route
    route: ...,
    // string (optional)
    name: name_example,
  } satisfies RouteServicePatchRequest;

  try {
    const data = await api.routeServicePatch(body);
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
| **route** | [V1Route](V1Route.md) |  | |
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


## routeServicePatch2

> V1PatchResponse routeServicePatch2(name, route, uid)



### Example

```ts
import {
  Configuration,
  RouteServiceApi,
} from '';
import type { RouteServicePatch2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new RouteServiceApi();

  const body = {
    // string
    name: name_example,
    // V1Route
    route: ...,
    // string (optional)
    uid: uid_example,
  } satisfies RouteServicePatch2Request;

  try {
    const data = await api.routeServicePatch2(body);
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
| **route** | [V1Route](V1Route.md) |  | |
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


## routeServiceUpdate

> V1UpdateResponse routeServiceUpdate(uid, route, name, updateMask)



### Example

```ts
import {
  Configuration,
  RouteServiceApi,
} from '';
import type { RouteServiceUpdateRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new RouteServiceApi();

  const body = {
    // string
    uid: uid_example,
    // V1Route
    route: ...,
    // string (optional)
    name: name_example,
    // string (optional)
    updateMask: updateMask_example,
  } satisfies RouteServiceUpdateRequest;

  try {
    const data = await api.routeServiceUpdate(body);
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
| **route** | [V1Route](V1Route.md) |  | |
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


## routeServiceUpdate2

> V1UpdateResponse routeServiceUpdate2(name, route, uid, updateMask)



### Example

```ts
import {
  Configuration,
  RouteServiceApi,
} from '';
import type { RouteServiceUpdate2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new RouteServiceApi();

  const body = {
    // string
    name: name_example,
    // V1Route
    route: ...,
    // string (optional)
    uid: uid_example,
    // string (optional)
    updateMask: updateMask_example,
  } satisfies RouteServiceUpdate2Request;

  try {
    const data = await api.routeServiceUpdate2(body);
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
| **route** | [V1Route](V1Route.md) |  | |
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


## routeServiceUpdateStatus

> object routeServiceUpdateStatus(uid, name, updateMask, statusPhase, statusReason, statusLastTransitionTime)



### Example

```ts
import {
  Configuration,
  RouteServiceApi,
} from '';
import type { RouteServiceUpdateStatusRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new RouteServiceApi();

  const body = {
    // string
    uid: uid_example,
    // string (optional)
    name: name_example,
    // string (optional)
    updateMask: updateMask_example,
    // string (optional)
    statusPhase: statusPhase_example,
    // string (optional)
    statusReason: statusReason_example,
    // Date (optional)
    statusLastTransitionTime: 2013-10-20T19:20:30+01:00,
  } satisfies RouteServiceUpdateStatusRequest;

  try {
    const data = await api.routeServiceUpdateStatus(body);
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
| **updateMask** | `string` |  | [Optional] [Defaults to `undefined`] |
| **statusPhase** | `string` |  | [Optional] [Defaults to `undefined`] |
| **statusReason** | `string` |  | [Optional] [Defaults to `undefined`] |
| **statusLastTransitionTime** | `Date` |  | [Optional] [Defaults to `undefined`] |

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


## routeServiceUpdateStatus2

> object routeServiceUpdateStatus2(name, uid, updateMask, statusPhase, statusReason, statusLastTransitionTime)



### Example

```ts
import {
  Configuration,
  RouteServiceApi,
} from '';
import type { RouteServiceUpdateStatus2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new RouteServiceApi();

  const body = {
    // string
    name: name_example,
    // string (optional)
    uid: uid_example,
    // string (optional)
    updateMask: updateMask_example,
    // string (optional)
    statusPhase: statusPhase_example,
    // string (optional)
    statusReason: statusReason_example,
    // Date (optional)
    statusLastTransitionTime: 2013-10-20T19:20:30+01:00,
  } satisfies RouteServiceUpdateStatus2Request;

  try {
    const data = await api.routeServiceUpdateStatus2(body);
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
| **updateMask** | `string` |  | [Optional] [Defaults to `undefined`] |
| **statusPhase** | `string` |  | [Optional] [Defaults to `undefined`] |
| **statusReason** | `string` |  | [Optional] [Defaults to `undefined`] |
| **statusLastTransitionTime** | `Date` |  | [Optional] [Defaults to `undefined`] |

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

