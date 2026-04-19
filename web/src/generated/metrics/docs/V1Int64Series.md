
# V1Int64Series

Int64Series represents a single counter bucket (sum per interval).

## Properties

Name | Type
------------ | -------------
`start` | Date
`value` | string

## Example

```typescript
import type { V1Int64Series } from ''

// TODO: Update the object below with actual values
const example = {
  "start": null,
  "value": null,
} satisfies V1Int64Series

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1Int64Series
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


