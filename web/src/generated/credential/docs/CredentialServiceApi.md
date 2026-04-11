# CredentialServiceApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**credentialServiceCreate**](CredentialServiceApi.md#credentialservicecreate) | **POST** /api/v1/credentials |  |
| [**credentialServiceDelete**](CredentialServiceApi.md#credentialservicedelete) | **DELETE** /api/v1/credentials/{uid} |  |
| [**credentialServiceDelete2**](CredentialServiceApi.md#credentialservicedelete2) | **DELETE** /api/v1/credentials/{name} |  |
| [**credentialServiceGet**](CredentialServiceApi.md#credentialserviceget) | **GET** /api/v1/credentials/{uid} |  |
| [**credentialServiceGet2**](CredentialServiceApi.md#credentialserviceget2) | **GET** /api/v1/credentials/{name} |  |
| [**credentialServiceList**](CredentialServiceApi.md#credentialservicelist) | **GET** /api/v1/credentials |  |
| [**credentialServicePatch**](CredentialServiceApi.md#credentialservicepatch) | **PATCH** /api/v1/credentials/{uid} |  |
| [**credentialServicePatch2**](CredentialServiceApi.md#credentialservicepatch2) | **PATCH** /api/v1/credentials/{name} |  |
| [**credentialServiceUpdate**](CredentialServiceApi.md#credentialserviceupdate) | **PUT** /api/v1/credentials/{uid} |  |
| [**credentialServiceUpdate2**](CredentialServiceApi.md#credentialserviceupdate2) | **PUT** /api/v1/credentials/{name} |  |



## credentialServiceCreate

> V1CreateResponse credentialServiceCreate(credential)



### Example

```ts
import {
  Configuration,
  CredentialServiceApi,
} from '';
import type { CredentialServiceCreateRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CredentialServiceApi();

  const body = {
    // V1Credential
    credential: ...,
  } satisfies CredentialServiceCreateRequest;

  try {
    const data = await api.credentialServiceCreate(body);
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
| **credential** | [V1Credential](V1Credential.md) |  | |

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


## credentialServiceDelete

> object credentialServiceDelete(uid, name, purge)



### Example

```ts
import {
  Configuration,
  CredentialServiceApi,
} from '';
import type { CredentialServiceDeleteRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CredentialServiceApi();

  const body = {
    // string
    uid: uid_example,
    // string (optional)
    name: name_example,
    // boolean (optional)
    purge: true,
  } satisfies CredentialServiceDeleteRequest;

  try {
    const data = await api.credentialServiceDelete(body);
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


## credentialServiceDelete2

> object credentialServiceDelete2(name, uid, purge)



### Example

```ts
import {
  Configuration,
  CredentialServiceApi,
} from '';
import type { CredentialServiceDelete2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CredentialServiceApi();

  const body = {
    // string
    name: name_example,
    // string (optional)
    uid: uid_example,
    // boolean (optional)
    purge: true,
  } satisfies CredentialServiceDelete2Request;

  try {
    const data = await api.credentialServiceDelete2(body);
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


## credentialServiceGet

> V1GetResponse credentialServiceGet(uid, name)



### Example

```ts
import {
  Configuration,
  CredentialServiceApi,
} from '';
import type { CredentialServiceGetRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CredentialServiceApi();

  const body = {
    // string
    uid: uid_example,
    // string (optional)
    name: name_example,
  } satisfies CredentialServiceGetRequest;

  try {
    const data = await api.credentialServiceGet(body);
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


## credentialServiceGet2

> V1GetResponse credentialServiceGet2(name, uid)



### Example

```ts
import {
  Configuration,
  CredentialServiceApi,
} from '';
import type { CredentialServiceGet2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CredentialServiceApi();

  const body = {
    // string
    name: name_example,
    // string (optional)
    uid: uid_example,
  } satisfies CredentialServiceGet2Request;

  try {
    const data = await api.credentialServiceGet2(body);
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


## credentialServiceList

> V1ListResponse credentialServiceList(limit, selector)



### Example

```ts
import {
  Configuration,
  CredentialServiceApi,
} from '';
import type { CredentialServiceListRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CredentialServiceApi();

  const body = {
    // number (optional)
    limit: 56,
    // string | This is a request variable of the map type. The query format is \"map_name[key]=value\", e.g. If the map name is Age, the key type is string, and the value type is integer, the query parameter is expressed as Age[\"bob\"]=18 (optional)
    selector: selector_example,
  } satisfies CredentialServiceListRequest;

  try {
    const data = await api.credentialServiceList(body);
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


## credentialServicePatch

> V1PatchResponse credentialServicePatch(uid, credential, name)



### Example

```ts
import {
  Configuration,
  CredentialServiceApi,
} from '';
import type { CredentialServicePatchRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CredentialServiceApi();

  const body = {
    // string
    uid: uid_example,
    // V1Credential
    credential: ...,
    // string (optional)
    name: name_example,
  } satisfies CredentialServicePatchRequest;

  try {
    const data = await api.credentialServicePatch(body);
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
| **credential** | [V1Credential](V1Credential.md) |  | |
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


## credentialServicePatch2

> V1PatchResponse credentialServicePatch2(name, credential, uid)



### Example

```ts
import {
  Configuration,
  CredentialServiceApi,
} from '';
import type { CredentialServicePatch2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CredentialServiceApi();

  const body = {
    // string
    name: name_example,
    // V1Credential
    credential: ...,
    // string (optional)
    uid: uid_example,
  } satisfies CredentialServicePatch2Request;

  try {
    const data = await api.credentialServicePatch2(body);
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
| **credential** | [V1Credential](V1Credential.md) |  | |
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


## credentialServiceUpdate

> V1UpdateResponse credentialServiceUpdate(uid, credential, name, updateMask)



### Example

```ts
import {
  Configuration,
  CredentialServiceApi,
} from '';
import type { CredentialServiceUpdateRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CredentialServiceApi();

  const body = {
    // string
    uid: uid_example,
    // V1Credential
    credential: ...,
    // string (optional)
    name: name_example,
    // string (optional)
    updateMask: updateMask_example,
  } satisfies CredentialServiceUpdateRequest;

  try {
    const data = await api.credentialServiceUpdate(body);
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
| **credential** | [V1Credential](V1Credential.md) |  | |
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


## credentialServiceUpdate2

> V1UpdateResponse credentialServiceUpdate2(name, credential, uid, updateMask)



### Example

```ts
import {
  Configuration,
  CredentialServiceApi,
} from '';
import type { CredentialServiceUpdate2Request } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new CredentialServiceApi();

  const body = {
    // string
    name: name_example,
    // V1Credential
    credential: ...,
    // string (optional)
    uid: uid_example,
    // string (optional)
    updateMask: updateMask_example,
  } satisfies CredentialServiceUpdate2Request;

  try {
    const data = await api.credentialServiceUpdate2(body);
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
| **credential** | [V1Credential](V1Credential.md) |  | |
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

