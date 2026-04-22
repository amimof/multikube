
# V1ProbeConfig


## Properties

Name | Type
------------ | -------------
`healthiness` | [V1Probe](V1Probe.md)
`readiness` | [V1Probe](V1Probe.md)

## Example

```typescript
import type { V1ProbeConfig } from ''

// TODO: Update the object below with actual values
const example = {
  "healthiness": null,
  "readiness": null,
} satisfies V1ProbeConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1ProbeConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


