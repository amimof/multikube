
# V1Float64Series

Float64Series represents a single float64 histogram bucket (observation count + sum per interval).

## Properties

Name | Type
------------ | -------------
`start` | Date
`count` | string
`sum` | number

## Example

```typescript
import type { V1Float64Series } from ''

// TODO: Update the object below with actual values
const example = {
  "start": null,
  "count": null,
  "sum": null,
} satisfies V1Float64Series

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1Float64Series
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


