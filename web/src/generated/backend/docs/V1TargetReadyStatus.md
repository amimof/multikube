
# V1TargetReadyStatus


## Properties

Name | Type
------------ | -------------
`isReady` | boolean
`reason` | string
`lastTransitionTime` | Date

## Example

```typescript
import type { V1TargetReadyStatus } from ''

// TODO: Update the object below with actual values
const example = {
  "isReady": null,
  "reason": null,
  "lastTransitionTime": null,
} satisfies V1TargetReadyStatus

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1TargetReadyStatus
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


