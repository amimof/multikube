<script setup lang="ts">
import { onMounted, onUnmounted, computed, toRaw } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Refresh } from '@element-plus/icons-vue'
import { useRouteStore } from '@/stores/route'
import { stringify as yamlStringify } from 'yaml'
import { Codemirror } from 'vue-codemirror'
import { yaml as yamlLang } from '@codemirror/lang-yaml'
import { oneDark } from '@codemirror/theme-one-dark'
import { EditorState } from '@codemirror/state'
import { formatDate, formatDateFull } from '@/utils/format'
import type { V1Match } from '@/generated/route'

const vueRoute = useRoute()
const router = useRouter()
const routeStore = useRouteStore()

const routeName = computed(() => vueRoute.params.name as string)

const routeResource = computed(() => routeStore.current)

const statusPhase = computed(() => routeResource.value?.status?.phase ?? 'Unknown')

const statusTagType = computed(() => {
	switch (statusPhase.value) {
		case 'Active':
			return 'success'
		case 'Inactive':
			return 'info'
		default:
			return 'warning'
	}
})

const matchType = computed(() => {
	const match = routeResource.value?.config?.match
	if (!match) return '-'
	if (match.sni) return 'SNI'
	if (match.path) return 'Path (exact)'
	if (match.pathPrefix) return 'Path Prefix'
	if (match.header?.name) return 'Header'
	if (match.jwt?.claim) return 'JWT'
	return '-'
})

const matchValue = computed(() => {
	const match = routeResource.value?.config?.match
	if (!match) return '-'
	if (match.sni) return match.sni
	if (match.path) return match.path
	if (match.pathPrefix) return match.pathPrefix
	if (match.header?.name) return `${match.header.name}=${match.header.value}`
	if (match.jwt?.claim) return `${match.jwt.claim}=${match.jwt.value}`
	return '-'
})

const yamlContent = computed(() => {
	if (!routeResource.value) return ''
	try {
		const raw = structuredClone(toRaw(routeResource.value))
		return yamlStringify(raw, { lineWidth: 120 })
	} catch {
		return '# Failed to serialize resource'
	}
})

const cmExtensions = [yamlLang(), oneDark, EditorState.readOnly.of(true)]

const labelEntries = computed(() => {
	const labels = routeResource.value?.meta?.labels
	if (!labels) return []
	return Object.entries(labels)
})

function handleRefresh() {
	routeStore.fetchRoute(routeName.value).catch(() => {})
}

function goBack() {
	router.push('/routes')
}

onMounted(() => {
	routeStore.fetchRoute(routeName.value).catch(() => {})
})

onUnmounted(() => {
	routeStore.clearCurrent()
})
</script>

<template>
	<div>
		<!-- Header -->
		<el-row justify="space-between" align="middle" style="margin-bottom: 16px">
			<el-col :span="16">
				<div style="display: flex; align-items: center; gap: 12px">
					<el-button :icon="ArrowLeft" @click="goBack" text>Routes</el-button>
					<h2 style="margin: 0">{{ routeName }}</h2>
					<el-tag :type="statusTagType" effect="dark" size="small">
						{{ statusPhase }}
					</el-tag>
				</div>
			</el-col>
			<el-col :span="8" style="text-align: right">
				<el-button :icon="Refresh" @click="handleRefresh">Reload</el-button>
			</el-col>
		</el-row>

		<el-alert v-if="routeStore.error" :title="routeStore.error" type="error" show-icon
			style="margin-bottom: 16px" />

		<!-- Loading -->
		<el-card v-if="routeStore.loading" shadow="never" style="margin-bottom: 16px">
			<el-skeleton :rows="6" animated />
		</el-card>

		<!-- Not found -->
		<el-empty v-else-if="!routeResource" :description="`Route '${routeName}' not found`">
			<el-button type="primary" @click="goBack">Back to Routes</el-button>
		</el-empty>

		<!-- Content -->
		<template v-else>
			<!-- Two-column layout: Configuration + Status -->
			<el-row :gutter="16" style="margin-bottom: 16px">
				<!-- Configuration (left) -->
				<el-col :span="14">
					<el-card shadow="never" style="height: 100%">
						<template #header>
							<span style="font-weight: 600">Configuration</span>
						</template>

						<!-- General section -->
						<h4 class="section-title">General</h4>
						<el-descriptions :column="2" border size="default">
							<el-descriptions-item label="Name">
								{{ routeResource.meta?.name ?? '-' }}
							</el-descriptions-item>
							<el-descriptions-item label="Created">
								<el-tooltip :content="formatDateFull(routeResource.meta?.created)" placement="top">
									<span>{{ formatDate(routeResource.meta?.created) }}</span>
								</el-tooltip>
							</el-descriptions-item>
							<el-descriptions-item label="Updated">
								<el-tooltip :content="formatDateFull(routeResource.meta?.updated)" placement="top">
									<span>{{ formatDate(routeResource.meta?.updated) }}</span>
								</el-tooltip>
							</el-descriptions-item>
							<el-descriptions-item label="UID">
								<span style="font-family: monospace; font-size: 12px">{{ routeResource.meta?.uid ?? '-' }}</span>
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

						<!-- Routing section -->
						<h4 class="section-title">Routing</h4>
						<el-descriptions :column="2" border size="default">
							<el-descriptions-item label="Backend Ref">
								<router-link v-if="routeResource.config?.backendRef" :to="`/backends/${routeResource.config.backendRef}`" style="text-decoration: none">
									<el-link type="primary">{{ routeResource.config.backendRef }}</el-link>
								</router-link>
								<span v-else style="color: #909399">-</span>
							</el-descriptions-item>
							<el-descriptions-item label="Match Type">
								{{ matchType }}
							</el-descriptions-item>
							<el-descriptions-item label="Match Value" :span="2">
								<span style="font-family: monospace; font-size: 12px">{{ matchValue }}</span>
							</el-descriptions-item>
						</el-descriptions>
					</el-card>
				</el-col>

				<!-- Status (right) -->
				<el-col :span="10">
					<el-card shadow="never" style="height: 100%">
						<template #header>
							<div style="display: flex; justify-content: space-between; align-items: center">
								<span style="font-weight: 600">Status</span>
								<el-tag :type="statusTagType" effect="dark" size="small">
									{{ statusPhase }}
								</el-tag>
							</div>
						</template>

						<el-descriptions :column="1" border size="default">
							<el-descriptions-item label="Phase">
								<el-tag :type="statusTagType" effect="dark" size="small">
									{{ statusPhase }}
								</el-tag>
							</el-descriptions-item>
							<el-descriptions-item label="Reason">
								<span v-if="routeResource.status?.reason" style="color: #f56c6c; font-size: 12px">
									{{ routeResource.status.reason }}
								</span>
								<span v-else style="color: #909399">-</span>
							</el-descriptions-item>
							<el-descriptions-item label="Last Transition">
								<el-tooltip v-if="routeResource.status?.lastTransitionTime"
									:content="formatDateFull(routeResource.status.lastTransitionTime)" placement="top">
									<span>{{ formatDate(routeResource.status.lastTransitionTime) }}</span>
								</el-tooltip>
								<span v-else style="color: #909399">-</span>
							</el-descriptions-item>
						</el-descriptions>
					</el-card>
				</el-col>
			</el-row>

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
