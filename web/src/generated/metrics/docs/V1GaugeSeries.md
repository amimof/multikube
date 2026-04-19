
# V1GaugeSeries

GaugeSeries represents a single gauge bucket (max value per interval).

## Properties

Name | Type
------------ | -------------
`start` | Date
`max` | string

## Example

```typescript
import type { V1GaugeSeries } from ''

// TODO: Update the object below with actual values
const example = {
  "start": null,
  "max": null,
} satisfies V1GaugeSeries

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1GaugeSeries
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


