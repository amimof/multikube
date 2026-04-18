
# V1CertificateAuthorityStatus


## Properties

Name | Type
------------ | -------------
`subjectCn` | string
`issuer` | string
`notBefore` | Date
`notAfter` | Date
`serialNumber` | string
`sans` | Array&lt;string&gt;
`signatureAlgorithm` | string
`publicKeyAlgorithm` | string
`isCa` | boolean
`ipAddresses` | Array&lt;string&gt;
`uris` | Array&lt;string&gt;

## Example

```typescript
import type { V1CertificateAuthorityStatus } from ''

// TODO: Update the object below with actual values
const example = {
  "subjectCn": null,
  "issuer": null,
  "notBefore": null,
  "notAfter": null,
  "serialNumber": null,
  "sans": null,
  "signatureAlgorithm": null,
  "publicKeyAlgorithm": null,
  "isCa": null,
  "ipAddresses": null,
  "uris": null,
} satisfies V1CertificateAuthorityStatus

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as V1CertificateAuthorityStatus
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


