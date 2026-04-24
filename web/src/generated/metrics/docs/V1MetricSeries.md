
# V1MetricSeries


## Properties

Name | Type
------------ | -------------
`metric` | string
`kind` | string
`labels` | [Array&lt;Metricsv1Label&gt;](Metricsv1Label.md)
`buckets` | [Array&lt;V1MetricBucket&gt;](V1MetricBucket.md)

## Example

```typescript
import type { V1MetricSeries } from ''

// TODO: Update the object below with actual values
const example = {
  "metric": null,
  "kind": null,
  "labels": null,
  "buckets": null,
} satisfies V1MetricSeries

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1MetricSeries
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


