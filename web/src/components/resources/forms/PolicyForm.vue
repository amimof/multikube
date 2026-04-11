<script setup lang="ts">
import { computed } from 'vue'
import ResourceCommonFields from '@/components/resources/forms/ResourceCommonFields.vue'
import ResourceMetadataFields from '@/components/resources/forms/ResourceMetadataFields.vue'
import type { PolicyResource } from '@/types/resources'

const props = defineProps<{
  modelValue: PolicyResource
  isEditing: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: PolicyResource]
}>()

const resource = computed({
  get: () => props.modelValue,
  set: (value: PolicyResource) => emit('update:modelValue', value),
})

function updateConfig(patch: Partial<PolicyResource['config']>) {
  resource.value = {
    ...resource.value,
    config: {
      ...(resource.value.config ?? {}),
      ...patch,
    },
  }
}

const rulesText = computed({
  get: () => JSON.stringify(resource.value.config?.rules ?? [], null, 2),
  set: (value: string) => {
    updateConfig({
      rules: value.trim().length === 0 ? [] : JSON.parse(value),
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

    <div class="col-12">
      <label class="form-label">Rules JSON</label>
      <textarea v-model="rulesText" class="form-control font-monospace" rows="12" />
      <div class="form-text">Provide a JSON array of policy rules.</div>
    </div>

    <ResourceMetadataFields :resource="resource" :is-editing="isEditing" />
  </div>
</template>
