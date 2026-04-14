
# V1RouteConfig


## Properties

Name | Type
------------ | -------------
`match` | [V1Match](V1Match.md)
`backendRef` | string

## Example

```typescript
import type { V1RouteConfig } from ''

// TODO: Update the object below with actual values
const example = {
  "match": null,
  "backendRef": null,
} satisfies V1RouteConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1RouteConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


