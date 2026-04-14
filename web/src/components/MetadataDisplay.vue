<script setup lang="ts">
import type { V1Meta } from '@/generated/backend'

defineProps<{
  meta?: V1Meta
}>()

function formatDate(date?: Date): string {
  if (!date) return '-'
  return new Date(date).toLocaleString()
}
</script>

<template>
  <el-descriptions :column="2" border size="small" v-if="meta">
    <el-descriptions-item label="UID">{{ meta.uid ?? '-' }}</el-descriptions-item>
    <el-descriptions-item label="Generation">{{ meta.generation ?? '-' }}</el-descriptions-item>
    <el-descriptions-item label="Resource Version">{{ meta.resourceVersion ?? '-' }}</el-descriptions-item>
    <el-descriptions-item label="Created">{{ formatDate(meta.created) }}</el-descriptions-item>
    <el-descriptions-item label="Updated">{{ formatDate(meta.updated) }}</el-descriptions-item>
  </el-descriptions>
</template>
