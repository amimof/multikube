
# V1Int64HistogramSeries

Int64HistogramSeries represents a single int64 histogram bucket (observation count + sum per interval).

## Properties

Name | Type
------------ | -------------
`start` | Date
`count` | string
`sum` | string

## Example

```typescript
import type { V1Int64HistogramSeries } from ''

// TODO: Update the object below with actual values
const example = {
  "start": null,
  "count": null,
  "sum": null,
} satisfies V1Int64HistogramSeries

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1Int64HistogramSeries
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


