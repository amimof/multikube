
# V1UserConfig


## Properties

Name | Type
------------ | -------------
`id` | string
`created` | Date
`updated` | Date
`labels` | { [key: string]: string; }
`email` | string
`password` | string
`tokenHash` | string
`roles` | Array&lt;string&gt;
`enabled` | boolean

## Example

```typescript
import type { V1UserConfig } from ''

// TODO: Update the object below with actual values
const example = {
  "id": null,
  "created": null,
  "updated": null,
  "labels": null,
  "email": null,
  "password": null,
  "tokenHash": null,
  "roles": null,
  "enabled": null,
} satisfies V1UserConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1UserConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


