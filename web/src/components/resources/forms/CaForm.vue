<script setup lang="ts">
import { computed } from 'vue'
import ResourceCommonFields from '@/components/resources/forms/ResourceCommonFields.vue'
import ResourceMetadataFields from '@/components/resources/forms/ResourceMetadataFields.vue'
import type { CertificateAuthorityResource } from '@/types/resources'

const props = defineProps<{
  modelValue: CertificateAuthorityResource
  isEditing: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: CertificateAuthorityResource]
}>()

const resource = computed({
  get: () => props.modelValue,
  set: (value: CertificateAuthorityResource) => emit('update:modelValue', value),
})

function updateConfig(patch: Partial<CertificateAuthorityResource['config']>) {
  resource.value = {
    ...resource.value,
    config: {
      ...(resource.value.config ?? {}),
      ...patch,
    },
  }
}
</script>

<template>
  <div class="row g-3">
    <ResourceCommonFields v-model="resource" />

    <div class="col-md-6">
      <label class="form-label">Config Name</label>
      <input :value="resource.config?.name ?? ''" class="form-control" @input="updateConfig({ name: ($event.target as HTMLInputElement).value })" />
    </div>

    <div class="col-md-6">
      <label class="form-label">Certificate Path</label>
      <input :value="resource.config?.certificate ?? ''" class="form-control" @input="updateConfig({ certificate: ($event.target as HTMLInputElement).value || undefined })" />
    </div>

    <div class="col-12">
      <label class="form-label">Certificate Data</label>
      <textarea :value="resource.config?.certificateData ?? ''" class="form-control" rows="8" @input="updateConfig({ certificateData: ($event.target as HTMLTextAreaElement).value || undefined })" />
    </div>

    <ResourceMetadataFields :resource="resource" :is-editing="isEditing" />
  </div>
</template>
