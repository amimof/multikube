
# V1Meta


## Properties

Name | Type
------------ | -------------
`name` | string
`labels` | { [key: string]: string; }
`created` | Date
`updated` | Date
`generation` | string
`resourceVersion` | string
`uid` | string

## Example

```typescript
import type { V1Meta } from ''

// TODO: Update the object below with actual values
const example = {
  "name": null,
  "labels": null,
  "created": null,
  "updated": null,
  "generation": null,
  "resourceVersion": null,
  "uid": null,
} satisfies V1Meta

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1Meta
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


