<script setup lang="ts">
import { computed } from 'vue'
import ResourceCommonFields from '@/components/resources/forms/ResourceCommonFields.vue'
import ResourceMetadataFields from '@/components/resources/forms/ResourceMetadataFields.vue'
import type { CredentialResource } from '@/types/resources'

const props = defineProps<{
  modelValue: CredentialResource
  isEditing: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: CredentialResource]
}>()

const resource = computed({
  get: () => props.modelValue,
  set: (value: CredentialResource) => emit('update:modelValue', value),
})

function updateConfig(patch: Partial<CredentialResource['config']>) {
  resource.value = {
    ...resource.value,
    config: {
      ...(resource.value.config ?? {}),
      ...patch,
    },
  }
}

const credentialMode = computed({
  get: () => {
    const config = resource.value.config ?? {}

    if (config.basic) return 'basic'
    if (config.token) return 'token'
    if (config.clientCertificateRef) return 'clientCertificateRef'
    return 'token'
  },
  set: (value: string) => {
    const current = resource.value.config ?? {}
    updateConfig({
      ...current,
      clientCertificateRef: value === 'clientCertificateRef' ? current.clientCertificateRef ?? '' : undefined,
      token: value === 'token' ? current.token ?? '' : undefined,
      basic:
        value === 'basic'
          ? {
              username: current.basic?.username ?? '',
              password: current.basic?.password ?? '',
            }
          : undefined,
    })
  },
})
</script>

<template>
  <div class="row g-3">
    <ResourceCommonFields v-model="resource" />

    <div class="col-md-6">
      <label class="form-label">Config Name</label>
      <input :value="resource.config?.name ?? ''" class="form-control" @input="updateConfig({ name: ($event.target as HTMLInputElement).value })" />
    </div>

    <div class="col-md-6">
      <label class="form-label">Credential Type</label>
      <select v-model="credentialMode" class="form-select">
        <option value="token">Token</option>
        <option value="basic">Basic</option>
        <option value="clientCertificateRef">Client Certificate Ref</option>
      </select>
    </div>

    <div v-if="credentialMode === 'token'" class="col-12">
      <label class="form-label">Token</label>
      <textarea :value="resource.config?.token ?? ''" class="form-control" rows="4" @input="updateConfig({ token: ($event.target as HTMLTextAreaElement).value || undefined })" />
    </div>

    <template v-else-if="credentialMode === 'basic'">
      <div class="col-md-6">
        <label class="form-label">Username</label>
        <input :value="resource.config?.basic?.username ?? ''" class="form-control" @input="updateConfig({ basic: { ...(resource.config?.basic ?? {}), username: ($event.target as HTMLInputElement).value } })" />
      </div>

      <div class="col-md-6">
        <label class="form-label">Password</label>
        <input :value="resource.config?.basic?.password ?? ''" class="form-control" @input="updateConfig({ basic: { ...(resource.config?.basic ?? {}), password: ($event.target as HTMLInputElement).value } })" />
      </div>
    </template>

    <div v-else class="col-12">
      <label class="form-label">Client Certificate Ref</label>
      <input :value="resource.config?.clientCertificateRef ?? ''" class="form-control" @input="updateConfig({ clientCertificateRef: ($event.target as HTMLInputElement).value || undefined })" />
    </div>

    <ResourceMetadataFields :resource="resource" :is-editing="isEditing" />
  </div>
</template>
