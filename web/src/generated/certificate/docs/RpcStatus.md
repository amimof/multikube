
# RpcStatus


## Properties

Name | Type
------------ | -------------
`code` | number
`message` | string
`details` | [Array&lt;ProtobufAny&gt;](ProtobufAny.md)

## Example

```typescript
import type { RpcStatus } from ''

// TODO: Update the object below with actual values
const example = {
  "code": null,
  "message": null,
  "details": null,
} satisfies RpcStatus

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as RpcStatus
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


