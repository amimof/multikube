
# V1BackendConfig


## Properties

Name | Type
------------ | -------------
`servers` | Array&lt;string&gt;
`caRef` | string
`authRef` | string
`insecureSkipTlsVerify` | boolean
`cacheTtl` | string
`type` | [V1LoadBalancingType](V1LoadBalancingType.md)

## Example

```typescript
import type { V1BackendConfig } from ''

// TODO: Update the object below with actual values
const example = {
  "servers": null,
  "caRef": null,
  "authRef": null,
  "insecureSkipTlsVerify": null,
  "cacheTtl": null,
  "type": null,
} satisfies V1BackendConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1BackendConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


