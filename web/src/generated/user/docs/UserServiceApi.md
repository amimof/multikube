# UserServiceApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**userServiceCreate**](UserServiceApi.md#userservicecreate) | **POST** /api/v1/users |  |
| [**userServiceDelete**](UserServiceApi.md#userservicedelete) | **DELETE** /api/v1/users/{uid} |  |
| [**userServiceDelete2**](UserServiceApi.md#userservicedelete2) | **DELETE** /api/v1/users/{name} |  |
| [**userServiceGet**](UserServiceApi.md#userserviceget) | **GET** /api/v1/users/{uid} |  |
| [**userServiceGet2**](UserServiceApi.md#userserviceget2) | **GET** /api/v1/users/{name} |  |
| [**userServiceList**](UserServiceApi.md#userservicelist) | **GET** /api/v1/users |  |
| [**userServicePatch**](UserServiceApi.md#userservicepatch) | **PATCH** /api/v1/users/{uid} |  |
| [**userServicePatch2**](UserServiceApi.md#userservicepatch2) | **PATCH** /api/v1/users/{name} |  |
| [**userServiceUpdate**](UserServiceApi.md#userserviceupdate) | **PUT** /api/v1/users/{uid} |  |
| [**userServiceUpdate2**](UserServiceApi.md#userserviceupdate2) | **PUT** /api/v1/users/{name} |  |



## userServiceCreate

> V1CreateResponse userServiceCreate(user)



### Example

```ts
import {
  Configuration,
  UserServiceApi,
} from '';
import type { UserServiceCreateRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new UserServiceApi();

  const body = {
    // V1User
    user: ...,
  } satisfies UserServiceCreateRequest;

  try {
    const data = await api.userServiceCreate(body);
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
| **user** | [V1User](V1User.md) |  | |

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


## userServiceDelete

> object userServiceDelete(uid, name, purge)



### Example

```ts
import {
  Configuration,
  UserServiceApi,
} from '';
import type { UserServiceDeleteRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new UserServiceApi();

  const body = {
    // string
    uid: uid_example,
    // string (optional)
    name: name_example,
    // boolean (optional)
    purge: true,
  } satisfies UserServiceDeleteRequest;

  try {
    const data = await api.userServiceDelete(body);
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


## userServiceDelete2

> object userServiceDelete2(name, uid, purge)



### Example

```ts
import {
  Configuration,
  UserServiceApi,
} from '';
import type { UserServiceDelete2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new UserServiceApi();

  const body = {
    // string
    name: name_example,
    // string (optional)
    uid: uid_example,
    // boolean (optional)
    purge: true,
  } satisfies UserServiceDelete2Request;

  try {
    const data = await api.userServiceDelete2(body);
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


## userServiceGet

> V1GetResponse userServiceGet(uid, name)



### Example

```ts
import {
  Configuration,
  UserServiceApi,
} from '';
import type { UserServiceGetRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new UserServiceApi();

  const body = {
    // string
    uid: uid_example,
    // string (optional)
    name: name_example,
  } satisfies UserServiceGetRequest;

  try {
    const data = await api.userServiceGet(body);
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


## userServiceGet2

> V1GetResponse userServiceGet2(name, uid)



### Example

```ts
import {
  Configuration,
  UserServiceApi,
} from '';
import type { UserServiceGet2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new UserServiceApi();

  const body = {
    // string
    name: name_example,
    // string (optional)
    uid: uid_example,
  } satisfies UserServiceGet2Request;

  try {
    const data = await api.userServiceGet2(body);
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


## userServiceList

> V1ListResponse userServiceList(limit, selector)



### Example

```ts
import {
  Configuration,
  UserServiceApi,
} from '';
import type { UserServiceListRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new UserServiceApi();

  const body = {
    // number (optional)
    limit: 56,
    // string | This is a request variable of the map type. The query format is \"map_name[key]=value\", e.g. If the map name is Age, the key type is string, and the value type is integer, the query parameter is expressed as Age[\"bob\"]=18 (optional)
    selector: selector_example,
  } satisfies UserServiceListRequest;

  try {
    const data = await api.userServiceList(body);
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


## userServicePatch

> V1PatchResponse userServicePatch(uid, user, name)



### Example

```ts
import {
  Configuration,
  UserServiceApi,
} from '';
import type { UserServicePatchRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new UserServiceApi();

  const body = {
    // string
    uid: uid_example,
    // V1User
    user: ...,
    // string (optional)
    name: name_example,
  } satisfies UserServicePatchRequest;

  try {
    const data = await api.userServicePatch(body);
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
| **user** | [V1User](V1User.md) |  | |
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


## userServicePatch2

> V1PatchResponse userServicePatch2(name, user, uid)



### Example

```ts
import {
  Configuration,
  UserServiceApi,
} from '';
import type { UserServicePatch2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new UserServiceApi();

  const body = {
    // string
    name: name_example,
    // V1User
    user: ...,
    // string (optional)
    uid: uid_example,
  } satisfies UserServicePatch2Request;

  try {
    const data = await api.userServicePatch2(body);
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
| **user** | [V1User](V1User.md) |  | |
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


## userServiceUpdate

> V1UpdateResponse userServiceUpdate(uid, user, name, updateMask)



### Example

```ts
import {
  Configuration,
  UserServiceApi,
} from '';
import type { UserServiceUpdateRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new UserServiceApi();

  const body = {
    // string
    uid: uid_example,
    // V1User
    user: ...,
    // string (optional)
    name: name_example,
    // string (optional)
    updateMask: updateMask_example,
  } satisfies UserServiceUpdateRequest;

  try {
    const data = await api.userServiceUpdate(body);
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
| **user** | [V1User](V1User.md) |  | |
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


## userServiceUpdate2

> V1UpdateResponse userServiceUpdate2(name, user, uid, updateMask)



### Example

```ts
import {
  Configuration,
  UserServiceApi,
} from '';
import type { UserServiceUpdate2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new UserServiceApi();

  const body = {
    // string
    name: name_example,
    // V1User
    user: ...,
    // string (optional)
    uid: uid_example,
    // string (optional)
    updateMask: updateMask_example,
  } satisfies UserServiceUpdate2Request;

  try {
    const data = await api.userServiceUpdate2(body);
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
| **user** | [V1User](V1User.md) |  | |
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

