
# V1User


## Properties

Name | Type
------------ | -------------
`version` | string
`meta` | [V1Meta](V1Meta.md)
`config` | [V1UserConfig](V1UserConfig.md)
`status` | [V1UserStatus](V1UserStatus.md)

## Example

```typescript
import type { V1User } from ''

// TODO: Update the object below with actual values
const example = {
  "version": null,
  "meta": null,
  "config": null,
  "status": null,
} satisfies V1User

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1User
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


