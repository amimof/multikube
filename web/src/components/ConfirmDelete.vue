<script setup lang="ts">
defineProps<{
  visible: boolean
  itemName?: string
  message?: string
}>()

const emit = defineEmits<{
  confirm: []
  'update:visible': [value: boolean]
}>()

function handleClose() {
  emit('update:visible', false)
}

function handleConfirm() {
  emit('confirm')
  emit('update:visible', false)
}
</script>

<template>
  <el-dialog
    :model-value="visible"
    title="Confirm Delete"
    width="400"
    @update:model-value="$emit('update:visible', $event)"
  >
    <p v-if="message">{{ message }}</p>
    <p v-else>Are you sure you want to delete <strong>{{ itemName }}</strong>?</p>
    <p>This action cannot be undone.</p>
    <template #footer>
      <el-button @click="handleClose">Cancel</el-button>
      <el-button type="danger" @click="handleConfirm">Delete</el-button>
    </template>
  </el-dialog>
</template>
