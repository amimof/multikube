
# V1Probe


## Properties

Name | Type
------------ | -------------
`path` | string
`timeoutSeconds` | string
`periodSeconds` | string
`failureThreshold` | string
`successThreshold` | string
`initialDelaySeconds` | string

## Example

```typescript
import type { V1Probe } from ''

// TODO: Update the object below with actual values
const example = {
  "path": null,
  "timeoutSeconds": null,
  "periodSeconds": null,
  "failureThreshold": null,
  "successThreshold": null,
  "initialDelaySeconds": null,
} satisfies V1Probe

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1Probe
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


