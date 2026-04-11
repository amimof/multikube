
# V1TargetStatus


## Properties

Name | Type
------------ | -------------
`phase` | string
`reason` | string
`lastTransitionTime` | Date

## Example

```typescript
import type { V1TargetStatus } from ''

// TODO: Update the object below with actual values
const example = {
  "phase": null,
  "reason": null,
  "lastTransitionTime": null,
} satisfies V1TargetStatus

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1TargetStatus
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


