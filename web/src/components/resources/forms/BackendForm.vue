<script setup lang="ts">
import { computed } from 'vue'
import ResourceCommonFields from '@/components/resources/forms/ResourceCommonFields.vue'
import ResourceMetadataFields from '@/components/resources/forms/ResourceMetadataFields.vue'
import type { BackendResource } from '@/types/resources'
import type { V1LoadBalancingType } from '@/generated/backend'

const props = defineProps<{
  modelValue: BackendResource
  isEditing: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: BackendResource]
}>()

const resource = computed({
  get: () => props.modelValue,
  set: (value: BackendResource) => emit('update:modelValue', value),
})

function updateConfig(patch: Partial<BackendResource['config']>) {
  resource.value = {
    ...resource.value,
    config: {
      ...(resource.value.config ?? {}),
      ...patch,
    },
  }
}

function updateLoadBalancingType(value: string) {
  updateConfig({
    type: (value || undefined) as V1LoadBalancingType | undefined,
  })
}

const serversText = computed({
  get: () => (resource.value.config?.servers ?? []).join('\n'),
  set: (value: string) => {
    updateConfig({
      servers: value
        .split('\n')
        .map((line) => line.trim())
        .filter(Boolean),
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
      <label class="form-label">Load Balancing</label>
      <select :value="resource.config?.type ?? ''" class="form-select" @change="updateLoadBalancingType(($event.target as HTMLSelectElement).value)">
        <option value="">Select type</option>
        <option value="LOAD_BALANCING_TYPE_ROUND_ROBIN">Round Robin</option>
        <option value="LOAD_BALANCING_TYPE_LEAST_CONNECTIONS">Least Connections</option>
        <option value="LOAD_BALANCING_TYPE_RANDOM">Random</option>
        <option value="LOAD_BALANCING_TYPE_WEIGHTED_ROUND_ROBIN">Weighted Round Robin</option>
      </select>
    </div>

    <div class="col-12">
      <label class="form-label">Servers</label>
      <textarea v-model="serversText" class="form-control" rows="4" placeholder="https://server.example" />
      <div class="form-text">One backend server per line.</div>
    </div>

    <div class="col-md-6">
      <label class="form-label">CA Ref</label>
      <input :value="resource.config?.caRef ?? ''" class="form-control" @input="updateConfig({ caRef: ($event.target as HTMLInputElement).value || undefined })" />
    </div>

    <div class="col-md-6">
      <label class="form-label">Credential Ref</label>
      <input :value="resource.config?.authRef ?? ''" class="form-control" @input="updateConfig({ authRef: ($event.target as HTMLInputElement).value || undefined })" />
    </div>

    <div class="col-md-6">
      <label class="form-label">Cache TTL</label>
      <input :value="resource.config?.cacheTtl ?? ''" class="form-control" placeholder="300s" @input="updateConfig({ cacheTtl: ($event.target as HTMLInputElement).value || undefined })" />
    </div>

    <div class="col-md-6 d-flex align-items-end">
      <div class="form-check mb-2">
        <input :checked="Boolean(resource.config?.insecureSkipTlsVerify)" class="form-check-input" type="checkbox" id="backend-insecure-skip" @change="updateConfig({ insecureSkipTlsVerify: ($event.target as HTMLInputElement).checked })" />
        <label class="form-check-label" for="backend-insecure-skip">Skip TLS verification</label>
      </div>
    </div>

    <ResourceMetadataFields :resource="resource" :is-editing="isEditing" />
  </div>
</template>
