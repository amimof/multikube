<script setup lang="ts">
import { onMounted, onUnmounted, computed, toRaw } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Refresh } from '@element-plus/icons-vue'
import { useCaStore } from '@/stores/ca'
import { stringify as yamlStringify } from 'yaml'
import { Codemirror } from 'vue-codemirror'
import { yaml as yamlLang } from '@codemirror/lang-yaml'
import { oneDark } from '@codemirror/theme-one-dark'
import { EditorState } from '@codemirror/state'
import { formatDate, formatDateFull } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const caStore = useCaStore()

const caName = computed(() => route.params.name as string)
const ca = computed(() => caStore.current)

const yamlContent = computed(() => {
	if (!ca.value) return ''
	try {
		const raw = structuredClone(toRaw(ca.value))
		return yamlStringify(raw, { lineWidth: 120 })
	} catch {
		return '# Failed to serialize resource'
	}
})

const cmExtensions = [yamlLang(), oneDark, EditorState.readOnly.of(true)]

const labelEntries = computed(() => {
	const labels = ca.value?.meta?.labels
	if (!labels) return []
	return Object.entries(labels)
})

const certValue = computed(() => {
	const data = ca.value?.config?.certificateData
	if (!data) return '-'
	if (data.length > 80) return data.substring(0, 80) + '...'
	return data
})

function handleRefresh() {
	caStore.fetchCa(caName.value).catch(() => {})
}

function goBack() {
	router.push('/cas')
}

onMounted(() => {
	caStore.fetchCa(caName.value).catch(() => {})
})

onUnmounted(() => {
	caStore.clearCurrent()
})
</script>

<template>
	<div>
		<!-- Header -->
		<el-row justify="space-between" align="middle" style="margin-bottom: 16px">
			<el-col :span="16">
				<div style="display: flex; align-items: center; gap: 12px">
					<el-button :icon="ArrowLeft" @click="goBack" text>Certificate Authorities</el-button>
					<h2 style="margin: 0">{{ caName }}</h2>
				</div>
			</el-col>
			<el-col :span="8" style="text-align: right">
				<el-button :icon="Refresh" @click="handleRefresh">Reload</el-button>
			</el-col>
		</el-row>

		<el-alert v-if="caStore.error" :title="caStore.error" type="error" show-icon style="margin-bottom: 16px" />

		<!-- Loading -->
		<el-card v-if="caStore.loading" shadow="never" style="margin-bottom: 16px">
			<el-skeleton :rows="6" animated />
		</el-card>

		<!-- Not found -->
		<el-empty v-else-if="!ca" :description="`Certificate Authority '${caName}' not found`">
			<el-button type="primary" @click="goBack">Back to Certificate Authorities</el-button>
		</el-empty>

		<!-- Content -->
		<template v-else>
			<el-card shadow="never" style="margin-bottom: 16px">
				<template #header>
					<span style="font-weight: 600">Configuration</span>
				</template>

				<!-- General section -->
				<h4 class="section-title">General</h4>
				<el-descriptions :column="2" border size="default">
					<el-descriptions-item label="Name">
						{{ ca.meta?.name ?? '-' }}
					</el-descriptions-item>
					<el-descriptions-item label="Created">
						<el-tooltip :content="formatDateFull(ca.meta?.created)" placement="top">
							<span>{{ formatDate(ca.meta?.created) }}</span>
						</el-tooltip>
					</el-descriptions-item>
					<el-descriptions-item label="Updated">
						<el-tooltip :content="formatDateFull(ca.meta?.updated)" placement="top">
							<span>{{ formatDate(ca.meta?.updated) }}</span>
						</el-tooltip>
					</el-descriptions-item>
					<el-descriptions-item label="UID">
						<span style="font-family: monospace; font-size: 12px">{{ ca.meta?.uid ?? '-' }}</span>
					</el-descriptions-item>
					<el-descriptions-item label="Labels" :span="2">
						<template v-if="labelEntries.length > 0">
							<el-tag v-for="[key, value] in labelEntries" :key="key" size="small" style="margin: 0 6px 4px 0">
								{{ key }}={{ value }}
							</el-tag>
						</template>
						<span v-else style="color: #909399">-</span>
					</el-descriptions-item>
				</el-descriptions>

				<!-- Certificate section -->
				<h4 class="section-title">Certificate</h4>
				<el-descriptions :column="1" border size="default">
					<el-descriptions-item label="Certificate Data">
						<span style="font-family: monospace; font-size: 12px">{{ certValue }}</span>
					</el-descriptions-item>
				</el-descriptions>
			</el-card>

			<!-- YAML -->
			<el-card shadow="never" style="margin-bottom: 16px">
				<template #header>
					<span style="font-weight: 600">YAML</span>
				</template>
				<div class="yaml-editor">
					<Codemirror :model-value="yamlContent" :extensions="cmExtensions" :style="{ fontSize: '13px' }" />
				</div>
			</el-card>
		</template>
	</div>
</template>

<style scoped>
.section-title {
	margin: 20px 0 8px 0;
	font-size: 13px;
	font-weight: 600;
	color: #909399;
	text-transform: uppercase;
	letter-spacing: 0.5px;
}

.section-title:first-of-type {
	margin-top: 0;
}

.yaml-editor :deep(.cm-editor) {
	border-radius: 4px;
	max-height: 600px;
	overflow: auto;
}

.yaml-editor :deep(.cm-gutters) {
	border-radius: 4px 0 0 4px;
}
</style>
