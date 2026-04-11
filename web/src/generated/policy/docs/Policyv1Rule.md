
# Policyv1Rule


## Properties

Name | Type
------------ | -------------
`effect` | [V1Effect](V1Effect.md)
`subjects` | [Array&lt;V1SubjectSelector&gt;](V1SubjectSelector.md)
`clusters` | [Array&lt;V1ClusterSelector&gt;](V1ClusterSelector.md)
`resources` | [Array&lt;V1ResourceSelector&gt;](V1ResourceSelector.md)
`actions` | [Array&lt;V1Action&gt;](V1Action.md)
`conditions` | [Array&lt;V1Condition&gt;](V1Condition.md)

## Example

```typescript
import type { Policyv1Rule } from ''

// TODO: Update the object below with actual values
const example = {
  "effect": null,
  "subjects": null,
  "clusters": null,
  "resources": null,
  "actions": null,
  "conditions": null,
} satisfies Policyv1Rule

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as Policyv1Rule
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


