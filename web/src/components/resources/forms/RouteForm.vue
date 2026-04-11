<script setup lang="ts">
import { computed } from 'vue'
import ResourceCommonFields from '@/components/resources/forms/ResourceCommonFields.vue'
import ResourceMetadataFields from '@/components/resources/forms/ResourceMetadataFields.vue'
import type { RouteResource } from '@/types/resources'

const props = defineProps<{
  modelValue: RouteResource
  isEditing: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: RouteResource]
}>()

const resource = computed({
  get: () => props.modelValue,
  set: (value: RouteResource) => emit('update:modelValue', value),
})

function updateConfig(patch: Partial<RouteResource['config']>) {
  resource.value = {
    ...resource.value,
    config: {
      ...(resource.value.config ?? {}),
      ...patch,
    },
  }
}

const routeMatchMode = computed({
  get: () => {
    const match = resource.value.config?.match ?? {}
    if (match.header) return 'header'
    if (match.jwt) return 'jwt'
    if (match.path) return 'path'
    if (match.pathPrefix) return 'pathPrefix'
    if (match.sni) return 'sni'
    return 'sni'
  },
  set: (value: string) => {
    const match = resource.value.config?.match ?? {}
    updateConfig({
      match: {
        sni: value === 'sni' ? match.sni ?? '' : undefined,
        path: value === 'path' ? match.path ?? '' : undefined,
        pathPrefix: value === 'pathPrefix' ? match.pathPrefix ?? '' : undefined,
        header: value === 'header' ? { name: match.header?.name ?? '', value: match.header?.value ?? '' } : undefined,
        jwt: value === 'jwt' ? { claim: match.jwt?.claim ?? '', value: match.jwt?.value ?? '' } : undefined,
      },
    })
  },
})

const routeMatchJson = computed({
  get: () => JSON.stringify(resource.value.config?.match ?? {}, null, 2),
  set: (value: string) => {
    updateConfig({
      match: value.trim().length === 0 ? {} : JSON.parse(value),
    })
  },
})

function updateMatchField(key: 'sni' | 'path' | 'pathPrefix', value: string) {
  const current = resource.value.config?.match ?? {}
  updateConfig({
    match: {
      ...current,
      [key]: value,
    },
  })
}

function updateHeaderField(key: 'name' | 'value', value: string) {
  const current = resource.value.config?.match ?? {}
  updateConfig({
    match: {
      ...current,
      header: {
        ...(current.header ?? {}),
        [key]: value,
      },
    },
  })
}

function updateJwtField(key: 'claim' | 'value', value: string) {
  const current = resource.value.config?.match ?? {}
  updateConfig({
    match: {
      ...current,
      jwt: {
        ...(current.jwt ?? {}),
        [key]: value,
      },
    },
  })
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
      <label class="form-label">Backend Ref</label>
      <input :value="resource.config?.backendRef ?? ''" class="form-control" @input="updateConfig({ backendRef: ($event.target as HTMLInputElement).value || undefined })" />
    </div>

    <div class="col-md-6">
      <label class="form-label">Match Type</label>
      <select v-model="routeMatchMode" class="form-select">
        <option value="sni">SNI</option>
        <option value="path">Path</option>
        <option value="pathPrefix">Path Prefix</option>
        <option value="header">Header</option>
        <option value="jwt">JWT</option>
      </select>
    </div>

    <div v-if="routeMatchMode === 'sni'" class="col-md-6">
      <label class="form-label">SNI</label>
      <input :value="resource.config?.match?.sni ?? ''" class="form-control" @input="updateMatchField('sni', ($event.target as HTMLInputElement).value)" />
    </div>

    <div v-else-if="routeMatchMode === 'path'" class="col-md-6">
      <label class="form-label">Path</label>
      <input :value="resource.config?.match?.path ?? ''" class="form-control" @input="updateMatchField('path', ($event.target as HTMLInputElement).value)" />
    </div>

    <div v-else-if="routeMatchMode === 'pathPrefix'" class="col-md-6">
      <label class="form-label">Path Prefix</label>
      <input :value="resource.config?.match?.pathPrefix ?? ''" class="form-control" @input="updateMatchField('pathPrefix', ($event.target as HTMLInputElement).value)" />
    </div>

    <template v-else-if="routeMatchMode === 'header'">
      <div class="col-md-6">
        <label class="form-label">Header Name</label>
        <input :value="resource.config?.match?.header?.name ?? ''" class="form-control" @input="updateHeaderField('name', ($event.target as HTMLInputElement).value)" />
      </div>

      <div class="col-md-6">
        <label class="form-label">Header Value</label>
        <input :value="resource.config?.match?.header?.value ?? ''" class="form-control" @input="updateHeaderField('value', ($event.target as HTMLInputElement).value)" />
      </div>
    </template>

    <template v-else-if="routeMatchMode === 'jwt'">
      <div class="col-md-6">
        <label class="form-label">JWT Claim</label>
        <input :value="resource.config?.match?.jwt?.claim ?? ''" class="form-control" @input="updateJwtField('claim', ($event.target as HTMLInputElement).value)" />
      </div>

      <div class="col-md-6">
        <label class="form-label">JWT Value</label>
        <input :value="resource.config?.match?.jwt?.value ?? ''" class="form-control" @input="updateJwtField('value', ($event.target as HTMLInputElement).value)" />
      </div>
    </template>

    <div class="col-12">
      <label class="form-label">Match JSON</label>
      <textarea v-model="routeMatchJson" class="form-control font-monospace" rows="8" />
      <div class="form-text">Use this for advanced match editing.</div>
    </div>

    <ResourceMetadataFields :resource="resource" :is-editing="isEditing" />
  </div>
</template>
