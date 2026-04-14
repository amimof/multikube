<script setup lang="ts">
import { ref, watch } from 'vue'
import { Plus, Delete } from '@element-plus/icons-vue'

const props = defineProps<{
  modelValue: Record<string, string>
}>()

const emit = defineEmits<{
  'update:modelValue': [value: Record<string, string>]
}>()

interface LabelRow {
  key: string
  value: string
}

const rows = ref<LabelRow[]>([])

function syncFromModel() {
  const labels = props.modelValue ?? {}
  rows.value = Object.entries(labels).map(([key, value]) => ({ key, value }))
}

syncFromModel()

watch(() => props.modelValue, syncFromModel, { deep: true })

function emitUpdate() {
  const labels: Record<string, string> = {}
  for (const row of rows.value) {
    if (row.key.trim()) {
      labels[row.key.trim()] = row.value
    }
  }
  emit('update:modelValue', labels)
}

function addLabel() {
  rows.value.push({ key: '', value: '' })
}

function removeLabel(index: number) {
  rows.value.splice(index, 1)
  emitUpdate()
}
</script>

<template>
  <div>
    <div v-for="(row, index) in rows" :key="index" style="display: flex; gap: 8px; margin-bottom: 8px">
      <el-input
        v-model="row.key"
        placeholder="Key"
        style="flex: 1"
        @change="emitUpdate"
      />
      <el-input
        v-model="row.value"
        placeholder="Value"
        style="flex: 1"
        @change="emitUpdate"
      />
      <el-button :icon="Delete" type="danger" plain @click="removeLabel(index)" />
    </div>
    <el-button :icon="Plus" size="small" @click="addLabel">Add Label</el-button>
  </div>
</template>
