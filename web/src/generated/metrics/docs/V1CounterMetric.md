
# V1CounterMetric


## Properties

Name | Type
------------ | -------------
`total` | string
`buckets` | [Array&lt;V1Int64Series&gt;](V1Int64Series.md)

## Example

```typescript
import type { V1CounterMetric } from ''

// TODO: Update the object below with actual values
const example = {
  "total": null,
  "buckets": null,
} satisfies V1CounterMetric

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1CounterMetric
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


