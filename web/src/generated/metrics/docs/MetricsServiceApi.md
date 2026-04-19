# MetricsServiceApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**metricsServiceGet**](MetricsServiceApi.md#metricsserviceget) | **GET** /api/v1/metrics |  |



## metricsServiceGet

> V1GetResponse metricsServiceGet()



### Example

```ts
import {
  Configuration,
  MetricsServiceApi,
} from '';
import type { MetricsServiceGetRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new MetricsServiceApi();

  try {
    const data = await api.metricsServiceGet();
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters

This endpoint does not need any parameter.

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

