
# V1GetResponse


## Properties

Name | Type
------------ | -------------
`requestsTotal` | [V1CounterMetric](V1CounterMetric.md)
`requestDuration` | [V1HistogramMetric](V1HistogramMetric.md)
`activeRequests` | [V1GaugeMetric](V1GaugeMetric.md)
`requestSizeBytes` | [V1Int64HistogramMetric](V1Int64HistogramMetric.md)
`responseSizeBytes` | [V1Int64HistogramMetric](V1Int64HistogramMetric.md)
`backendRequestsTotal` | [V1CounterMetric](V1CounterMetric.md)
`backendRequestDuration` | [V1HistogramMetric](V1HistogramMetric.md)
`backendActiveRequests` | [V1GaugeMetric](V1GaugeMetric.md)
`authRequestsTotal` | [V1CounterMetric](V1CounterMetric.md)
`policyEvaluationsTotal` | [V1CounterMetric](V1CounterMetric.md)
`routeMatchesTotal` | [V1CounterMetric](V1CounterMetric.md)
`routeNoMatchTotal` | [V1CounterMetric](V1CounterMetric.md)

## Example

```typescript
import type { V1GetResponse } from ''

// TODO: Update the object below with actual values
const example = {
  "requestsTotal": null,
  "requestDuration": null,
  "activeRequests": null,
  "requestSizeBytes": null,
  "responseSizeBytes": null,
  "backendRequestsTotal": null,
  "backendRequestDuration": null,
  "backendActiveRequests": null,
  "authRequestsTotal": null,
  "policyEvaluationsTotal": null,
  "routeMatchesTotal": null,
  "routeNoMatchTotal": null,
} satisfies V1GetResponse

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1GetResponse
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


