# PolicyServiceApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**policyServiceCreate**](PolicyServiceApi.md#policyservicecreate) | **POST** /api/v1/policys |  |
| [**policyServiceDelete**](PolicyServiceApi.md#policyservicedelete) | **DELETE** /api/v1/policys/{uid} |  |
| [**policyServiceDelete2**](PolicyServiceApi.md#policyservicedelete2) | **DELETE** /api/v1/policys/{name} |  |
| [**policyServiceGet**](PolicyServiceApi.md#policyserviceget) | **GET** /api/v1/policys/{uid} |  |
| [**policyServiceGet2**](PolicyServiceApi.md#policyserviceget2) | **GET** /api/v1/policys/{name} |  |
| [**policyServiceList**](PolicyServiceApi.md#policyservicelist) | **GET** /api/v1/policys |  |
| [**policyServicePatch**](PolicyServiceApi.md#policyservicepatch) | **PATCH** /api/v1/policys/{uid} |  |
| [**policyServicePatch2**](PolicyServiceApi.md#policyservicepatch2) | **PATCH** /api/v1/policys/{name} |  |
| [**policyServiceUpdate**](PolicyServiceApi.md#policyserviceupdate) | **PUT** /api/v1/policys/{uid} |  |
| [**policyServiceUpdate2**](PolicyServiceApi.md#policyserviceupdate2) | **PUT** /api/v1/policys/{name} |  |



## policyServiceCreate

> V1CreateResponse policyServiceCreate(policy)



### Example

```ts
import {
  Configuration,
  PolicyServiceApi,
} from '';
import type { PolicyServiceCreateRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new PolicyServiceApi();

  const body = {
    // V1Policy
    policy: ...,
  } satisfies PolicyServiceCreateRequest;

  try {
    const data = await api.policyServiceCreate(body);
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
| **policy** | [V1Policy](V1Policy.md) |  | |

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


## policyServiceDelete

> object policyServiceDelete(uid, name, purge)



### Example

```ts
import {
  Configuration,
  PolicyServiceApi,
} from '';
import type { PolicyServiceDeleteRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new PolicyServiceApi();

  const body = {
    // string
    uid: uid_example,
    // string (optional)
    name: name_example,
    // boolean (optional)
    purge: true,
  } satisfies PolicyServiceDeleteRequest;

  try {
    const data = await api.policyServiceDelete(body);
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


## policyServiceDelete2

> object policyServiceDelete2(name, uid, purge)



### Example

```ts
import {
  Configuration,
  PolicyServiceApi,
} from '';
import type { PolicyServiceDelete2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new PolicyServiceApi();

  const body = {
    // string
    name: name_example,
    // string (optional)
    uid: uid_example,
    // boolean (optional)
    purge: true,
  } satisfies PolicyServiceDelete2Request;

  try {
    const data = await api.policyServiceDelete2(body);
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


## policyServiceGet

> V1GetResponse policyServiceGet(uid, name)



### Example

```ts
import {
  Configuration,
  PolicyServiceApi,
} from '';
import type { PolicyServiceGetRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new PolicyServiceApi();

  const body = {
    // string
    uid: uid_example,
    // string (optional)
    name: name_example,
  } satisfies PolicyServiceGetRequest;

  try {
    const data = await api.policyServiceGet(body);
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


## policyServiceGet2

> V1GetResponse policyServiceGet2(name, uid)



### Example

```ts
import {
  Configuration,
  PolicyServiceApi,
} from '';
import type { PolicyServiceGet2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new PolicyServiceApi();

  const body = {
    // string
    name: name_example,
    // string (optional)
    uid: uid_example,
  } satisfies PolicyServiceGet2Request;

  try {
    const data = await api.policyServiceGet2(body);
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


## policyServiceList

> V1ListResponse policyServiceList(limit, selector)



### Example

```ts
import {
  Configuration,
  PolicyServiceApi,
} from '';
import type { PolicyServiceListRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new PolicyServiceApi();

  const body = {
    // number (optional)
    limit: 56,
    // string | This is a request variable of the map type. The query format is \"map_name[key]=value\", e.g. If the map name is Age, the key type is string, and the value type is integer, the query parameter is expressed as Age[\"bob\"]=18 (optional)
    selector: selector_example,
  } satisfies PolicyServiceListRequest;

  try {
    const data = await api.policyServiceList(body);
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


## policyServicePatch

> V1PatchResponse policyServicePatch(uid, policy, name)



### Example

```ts
import {
  Configuration,
  PolicyServiceApi,
} from '';
import type { PolicyServicePatchRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new PolicyServiceApi();

  const body = {
    // string
    uid: uid_example,
    // V1Policy
    policy: ...,
    // string (optional)
    name: name_example,
  } satisfies PolicyServicePatchRequest;

  try {
    const data = await api.policyServicePatch(body);
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
| **policy** | [V1Policy](V1Policy.md) |  | |
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


## policyServicePatch2

> V1PatchResponse policyServicePatch2(name, policy, uid)



### Example

```ts
import {
  Configuration,
  PolicyServiceApi,
} from '';
import type { PolicyServicePatch2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new PolicyServiceApi();

  const body = {
    // string
    name: name_example,
    // V1Policy
    policy: ...,
    // string (optional)
    uid: uid_example,
  } satisfies PolicyServicePatch2Request;

  try {
    const data = await api.policyServicePatch2(body);
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
| **policy** | [V1Policy](V1Policy.md) |  | |
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


## policyServiceUpdate

> V1UpdateResponse policyServiceUpdate(uid, policy, name, updateMask)



### Example

```ts
import {
  Configuration,
  PolicyServiceApi,
} from '';
import type { PolicyServiceUpdateRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new PolicyServiceApi();

  const body = {
    // string
    uid: uid_example,
    // V1Policy
    policy: ...,
    // string (optional)
    name: name_example,
    // string (optional)
    updateMask: updateMask_example,
  } satisfies PolicyServiceUpdateRequest;

  try {
    const data = await api.policyServiceUpdate(body);
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
| **policy** | [V1Policy](V1Policy.md) |  | |
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


## policyServiceUpdate2

> V1UpdateResponse policyServiceUpdate2(name, policy, uid, updateMask)



### Example

```ts
import {
  Configuration,
  PolicyServiceApi,
} from '';
import type { PolicyServiceUpdate2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new PolicyServiceApi();

  const body = {
    // string
    name: name_example,
    // V1Policy
    policy: ...,
    // string (optional)
    uid: uid_example,
    // string (optional)
    updateMask: updateMask_example,
  } satisfies PolicyServiceUpdate2Request;

  try {
    const data = await api.policyServiceUpdate2(body);
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
| **policy** | [V1Policy](V1Policy.md) |  | |
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

