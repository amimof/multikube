<script setup lang="ts">
import { computed } from 'vue'
import ResourceCommonFields from '@/components/resources/forms/ResourceCommonFields.vue'
import ResourceMetadataFields from '@/components/resources/forms/ResourceMetadataFields.vue'
import type { CertificateResource } from '@/types/resources'

const props = defineProps<{
  modelValue: CertificateResource
  isEditing: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: CertificateResource]
}>()

const resource = computed({
  get: () => props.modelValue,
  set: (value: CertificateResource) => emit('update:modelValue', value),
})

function updateConfig(patch: Partial<CertificateResource['config']>) {
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
      <textarea :value="resource.config?.certificateData ?? ''" class="form-control" rows="6" @input="updateConfig({ certificateData: ($event.target as HTMLTextAreaElement).value || undefined })" />
    </div>

    <div class="col-md-6">
      <label class="form-label">Key Path</label>
      <input :value="resource.config?.key ?? ''" class="form-control" @input="updateConfig({ key: ($event.target as HTMLInputElement).value || undefined })" />
    </div>

    <div class="col-12">
      <label class="form-label">Key Data</label>
      <textarea :value="resource.config?.keyData ?? ''" class="form-control" rows="6" @input="updateConfig({ keyData: ($event.target as HTMLTextAreaElement).value || undefined })" />
    </div>

    <ResourceMetadataFields :resource="resource" :is-editing="isEditing" />
  </div>
</template>
