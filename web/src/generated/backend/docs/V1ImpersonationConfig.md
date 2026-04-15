
# V1ImpersonationConfig


## Properties

Name | Type
------------ | -------------
`name` | string
`enabled` | boolean
`usernameClaim` | string
`groupsClaim` | string
`extraClaims` | Array&lt;string&gt;

## Example

```typescript
import type { V1ImpersonationConfig } from ''

// TODO: Update the object below with actual values
const example = {
  "name": null,
  "enabled": null,
  "usernameClaim": null,
  "groupsClaim": null,
  "extraClaims": null,
} satisfies V1ImpersonationConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1ImpersonationConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


