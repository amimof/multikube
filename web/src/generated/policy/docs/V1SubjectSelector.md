
# V1SubjectSelector


## Properties

Name | Type
------------ | -------------
`users` | Array&lt;string&gt;
`groups` | Array&lt;string&gt;
`serviceAccounts` | Array&lt;string&gt;
`claims` | [Array&lt;V1Claim&gt;](V1Claim.md)

## Example

```typescript
import type { V1SubjectSelector } from ''

// TODO: Update the object below with actual values
const example = {
  "users": null,
  "groups": null,
  "serviceAccounts": null,
  "claims": null,
} satisfies V1SubjectSelector

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1SubjectSelector
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


