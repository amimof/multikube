
# V1CredentialConfig


## Properties

Name | Type
------------ | -------------
`clientCertificateRef` | string
`token` | string
`basic` | [V1CredentialBasic](V1CredentialBasic.md)

## Example

```typescript
import type { V1CredentialConfig } from ''

// TODO: Update the object below with actual values
const example = {
  "clientCertificateRef": null,
  "token": null,
  "basic": null,
} satisfies V1CredentialConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1CredentialConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


