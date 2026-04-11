import { BackendServiceApi, Configuration as BackendConfiguration } from '@/generated/backend'
import { CertificateAuthorityServiceApi, Configuration as CaConfiguration } from '@/generated/ca'
import { CertificateServiceApi, Configuration as CertificateConfiguration } from '@/generated/certificate'
import { CredentialServiceApi, Configuration as CredentialConfiguration } from '@/generated/credential'
import { PolicyServiceApi, Configuration as PolicyConfiguration } from '@/generated/policy'
import { RouteServiceApi, Configuration as RouteConfiguration } from '@/generated/route'

const clientOptions = {
  basePath: '',
  credentials: 'include' as const,
}

export const api = {
  backendService: new BackendServiceApi(new BackendConfiguration(clientOptions)),
  certificateAuthorityService: new CertificateAuthorityServiceApi(new CaConfiguration(clientOptions)),
  certificateService: new CertificateServiceApi(new CertificateConfiguration(clientOptions)),
  credentialService: new CredentialServiceApi(new CredentialConfiguration(clientOptions)),
  policyService: new PolicyServiceApi(new PolicyConfiguration(clientOptions)),
  routeService: new RouteServiceApi(new RouteConfiguration(clientOptions)),
}
