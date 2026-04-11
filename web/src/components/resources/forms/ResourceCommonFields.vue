<script setup lang="ts">
import { computed } from 'vue'
import type { ResourceItem, ResourceMeta } from '@/types/resources'

const props = defineProps<{
  modelValue: ResourceItem
}>()

const emit = defineEmits<{
  'update:modelValue': [value: ResourceItem]
}>()

const resource = computed({
  get: () => props.modelValue,
  set: (value: ResourceItem) => emit('update:modelValue', value),
})

function updateResource(patch: Partial<ResourceItem>) {
  resource.value = {
    ...(resource.value as object),
    ...(patch as object),
  } as ResourceItem
}

function updateMeta(patch: Partial<ResourceMeta>) {
  updateResource({
    meta: {
      ...(resource.value.meta ?? {}),
      ...patch,
    },
  })
}

function parseLabelText(value: string) {
  return value
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
    .reduce<Record<string, string>>((labels, line) => {
      const index = line.indexOf('=')

      if (index === -1) {
        return labels
      }

      const key = line.slice(0, index).trim()
      const labelValue = line.slice(index + 1).trim()

      if (key) {
        labels[key] = labelValue
      }

      return labels
    }, {})
}

const labelsText = computed({
  get: () =>
    Object.entries(resource.value.meta?.labels ?? {})
      .map(([key, value]) => `${key}=${value}`)
      .join('\n'),
  set: (value: string) => updateMeta({ labels: parseLabelText(value) }),
})
</script>

<template>
  <div class="col-md-6">
    <label class="form-label">Version</label>
    <input :value="resource.version ?? ''" class="form-control" placeholder="group/version" readonly disabled />
  </div>

  <div class="col-md-6">
    <label class="form-label">Display Name</label>
    <input
      :value="resource.meta?.name ?? ''"
      class="form-control"
      placeholder="Resource name"
      @input="updateMeta({ name: ($event.target as HTMLInputElement).value })"
    />
  </div>

  <div class="col-12">
    <label class="form-label">Labels</label>
    <textarea v-model="labelsText" class="form-control" rows="3" placeholder="key=value" />
    <div class="form-text">One label per line in `key=value` format.</div>
  </div>
</template>
