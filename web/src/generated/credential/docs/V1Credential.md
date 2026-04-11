
# V1Credential


## Properties

Name | Type
------------ | -------------
`version` | string
`meta` | [V1Meta](V1Meta.md)
`config` | [V1CredentialConfig](V1CredentialConfig.md)
`status` | [V1CredentialStatus](V1CredentialStatus.md)

## Example

```typescript
import type { V1Credential } from ''

// TODO: Update the object below with actual values
const example = {
  "version": null,
  "meta": null,
  "config": null,
  "status": null,
} satisfies V1Credential

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1Credential
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


