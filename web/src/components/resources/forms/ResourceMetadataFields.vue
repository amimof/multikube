<script setup lang="ts">
import { computed } from 'vue'
import type { ResourceItem } from '@/types/resources'

const props = defineProps<{
  resource: ResourceItem
  isEditing: boolean
}>()

function formatMetaDate(value?: string | Date) {
  if (!value) {
    return ''
  }

  const date = value instanceof Date ? value : new Date(value)
  return Number.isNaN(date.getTime()) ? String(value) : date.toLocaleString()
}

const metadataRows = computed(() => {
  const meta = props.resource.meta ?? {}

  return [
    { label: 'UID', value: meta.uid },
    { label: 'Created', value: formatMetaDate(meta.created) },
    { label: 'Updated', value: formatMetaDate(meta.updated) },
    { label: 'Generation', value: meta.generation },
    { label: 'Resource Version', value: meta.resourceVersion },
  ].filter((row) => row.value)
})
</script>

<template>
  <div v-if="isEditing && metadataRows.length > 0" class="col-12">
    <hr />
    <h6>Metadata</h6>
    <div class="row g-3">
      <div v-for="row in metadataRows" :key="row.label" class="col-md-6">
        <label class="form-label">{{ row.label }}</label>
        <input :value="row.value" class="form-control" readonly />
      </div>
    </div>
  </div>
</template>
