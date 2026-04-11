
# V1ResourceSelector


## Properties

Name | Type
------------ | -------------
`apiGroup` | string
`resource` | string
`subResource` | string
`namespaces` | Array&lt;string&gt;
`names` | Array&lt;string&gt;
`labelSelector` | [V1LabelSelector](V1LabelSelector.md)

## Example

```typescript
import type { V1ResourceSelector } from ''

// TODO: Update the object below with actual values
const example = {
  "apiGroup": null,
  "resource": null,
  "subResource": null,
  "namespaces": null,
  "names": null,
  "labelSelector": null,
} satisfies V1ResourceSelector

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1ResourceSelector
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


