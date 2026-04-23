<script setup lang="ts">
import type { V1Meta } from '@/generated/backend'
import { formatDateFull } from '@/utils/format'

defineProps<{
	meta?: V1Meta
}>()
</script>

<template>
	<el-descriptions :column="2" border size="small" v-if="meta">
		<el-descriptions-item label="Generation">{{ meta.generation ?? '-' }}</el-descriptions-item>
		<el-descriptions-item label="Resource Version">{{ meta.resourceVersion ?? '-' }}</el-descriptions-item>
		<el-descriptions-item label="Created">{{ formatDateFull(meta.created) }}</el-descriptions-item>
		<el-descriptions-item label="Updated">{{ formatDateFull(meta.updated) }}</el-descriptions-item>
		<el-descriptions-item min-width="2" label="UID">{{ meta.uid ?? '-' }}</el-descriptions-item>
	</el-descriptions>
</template>
