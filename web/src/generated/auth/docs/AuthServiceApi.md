# AuthServiceApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**authServiceLogin**](AuthServiceApi.md#authservicelogin) | **POST** /api/v1/auth/login |  |
| [**authServiceLogout**](AuthServiceApi.md#authservicelogout) | **POST** /api/v1/auth/logout |  |
| [**authServiceRefresh**](AuthServiceApi.md#authservicerefresh) | **POST** /api/v1/auth/refresh |  |



## authServiceLogin

> V1LoginResponse authServiceLogin(body)



### Example

```ts
import {
  Configuration,
  AuthServiceApi,
} from '';
import type { AuthServiceLoginRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new AuthServiceApi();

  const body = {
    // V1LoginRequest
    body: ...,
  } satisfies AuthServiceLoginRequest;

  try {
    const data = await api.authServiceLogin(body);
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
| **body** | [V1LoginRequest](V1LoginRequest.md) |  | |

### Return type

[**V1LoginResponse**](V1LoginResponse.md)

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


## authServiceLogout

> object authServiceLogout(accessToken)



### Example

```ts
import {
  Configuration,
  AuthServiceApi,
} from '';
import type { AuthServiceLogoutRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new AuthServiceApi();

  const body = {
    // string (optional)
    accessToken: accessToken_example,
  } satisfies AuthServiceLogoutRequest;

  try {
    const data = await api.authServiceLogout(body);
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
| **accessToken** | `string` |  | [Optional] [Defaults to `undefined`] |

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


## authServiceRefresh

> V1RefreshResponse authServiceRefresh(body)



### Example

```ts
import {
  Configuration,
  AuthServiceApi,
} from '';
import type { AuthServiceRefreshRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new AuthServiceApi();

  const body = {
    // V1RefreshRequest
    body: ...,
  } satisfies AuthServiceRefreshRequest;

  try {
    const data = await api.authServiceRefresh(body);
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
| **body** | [V1RefreshRequest](V1RefreshRequest.md) |  | |

### Return type

[**V1RefreshResponse**](V1RefreshResponse.md)

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

