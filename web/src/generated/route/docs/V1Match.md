
# V1Match


## Properties

Name | Type
------------ | -------------
`sni` | string
`header` | [V1HeaderMatch](V1HeaderMatch.md)
`path` | string
`pathPrefix` | string
`jwt` | [V1JWTMatch](V1JWTMatch.md)

## Example

```typescript
import type { V1Match } from ''

// TODO: Update the object below with actual values
const example = {
  "sni": null,
  "header": null,
  "path": null,
  "pathPrefix": null,
  "jwt": null,
} satisfies V1Match

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1Match
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


