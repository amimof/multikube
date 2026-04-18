<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed, watch, toRaw } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Refresh, Document } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useRouteStore } from '@/stores/route'
import { useBackendStore } from '@/stores/backend'
import { stringify as yamlStringify } from 'yaml'
import { Codemirror } from 'vue-codemirror'
import { yaml as yamlLang } from '@codemirror/lang-yaml'
import { oneDark } from '@codemirror/theme-one-dark'
import { EditorState } from '@codemirror/state'
import { formatDate, formatDateFull } from '@/utils/format'
import type { V1Route, V1Match } from '@/generated/route'
import LabelEditor from '@/components/LabelEditor.vue'
import MetadataDisplay from '@/components/MetadataDisplay.vue'
import EditYamlModal from '@/components/EditYamlModal.vue'

type RouteMatchMode = '' | 'sni' | 'path' | 'pathPrefix' | 'header' | 'jwt'

const vueRoute = useRoute()
const router = useRouter()
const routeStore = useRouteStore()
const backendStore = useBackendStore()

const routeName = computed(() => vueRoute.params.name as string)
const routeResource = computed(() => routeStore.current)

const saving = ref(false)
const yamlModalVisible = ref(false)
const matchMode = ref<RouteMatchMode>('')

const form = ref<V1Route>({})

const matchModeOptions: { label: string; value: RouteMatchMode }[] = [
	{ label: 'SNI', value: 'sni' },
	{ label: 'Path (exact)', value: 'path' },
	{ label: 'Path Prefix', value: 'pathPrefix' },
	{ label: 'Header', value: 'header' },
	{ label: 'JWT', value: 'jwt' },
]

function inferMatchMode(match?: V1Match): RouteMatchMode {
	if (!match) return ''
	if (match.sni) return 'sni'
	if (match.path) return 'path'
	if (match.pathPrefix) return 'pathPrefix'
	if (match.header?.name || match.header?.value) return 'header'
	if (match.jwt?.claim || match.jwt?.value) return 'jwt'
	return ''
}

// Initialize form from store
watch(routeResource, (val) => {
	if (val) {
		const raw = structuredClone(toRaw(val))
		if (!raw.config) raw.config = {}
		form.value = raw
		// Infer match mode and restore match data after watcher fires
		const inferred = inferMatchMode(raw.config?.match)
		matchMode.value = inferred
		// Restore actual match data since the matchMode watcher will have cleared it
		if (raw.config?.match) {
			form.value.config!.match = structuredClone(raw.config.match)
		}
	}
}, { immediate: true })

// When matchMode changes, replace match object
watch(matchMode, (mode) => {
	if (!form.value.config) form.value.config = {}
	switch (mode) {
		case 'sni':
			form.value.config.match = { sni: '' }
			break
		case 'path':
			form.value.config.match = { path: '' }
			break
		case 'pathPrefix':
			form.value.config.match = { pathPrefix: '' }
			break
		case 'header':
			form.value.config.match = { header: { name: '', value: '' } }
			break
		case 'jwt':
			form.value.config.match = { jwt: { claim: '', value: '' } }
			break
		default:
			form.value.config.match = undefined
	}
})

const formLabels = computed({
	get: () => form.value.meta?.labels ?? {},
	set: (val: Record<string, string>) => {
		if (form.value.meta) {
			form.value.meta.labels = val
		}
	},
})

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

const yamlContent = computed(() => {
	if (!form.value || !form.value.version) return ''
	try {
		const raw = structuredClone(toRaw(form.value))
		return yamlStringify(raw, { lineWidth: 120 })
	} catch {
		return '# Failed to serialize resource'
	}
})

const cmExtensions = [yamlLang(), oneDark, EditorState.readOnly.of(true)]

async function handleSave() {
	saving.value = true
	try {
		await routeStore.updateRoute(form.value)
		await routeStore.fetchRoute(routeName.value)
		ElMessage.success('Route updated')
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Save failed')
	} finally {
		saving.value = false
	}
}

async function handleYamlSave(parsed: unknown) {
	saving.value = true
	try {
		const resource = parsed as V1Route
		if (!resource.meta?.name) {
			resource.meta = { ...resource.meta, name: routeName.value }
		}
		await routeStore.updateRoute(resource)
		await routeStore.fetchRoute(routeName.value)
		ElMessage.success('Route updated from YAML')
		yamlModalVisible.value = false
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Save failed')
	} finally {
		saving.value = false
	}
}

function handleRefresh() {
	routeStore.fetchRoute(routeName.value).catch(() => { })
}

function goBack() {
	router.push('/routes')
}

onMounted(() => {
	routeStore.fetchRoute(routeName.value).catch(() => { })
	backendStore.fetchBackends().catch(() => { })
})

onUnmounted(() => {
	routeStore.clearCurrent()
})
</script>

<template>
	<div>
		<!-- Header -->
		<el-row justify="space-between" align="middle" style="margin-bottom: 16px">
			<el-col :span="12">
				<div style="display: flex; align-items: center; gap: 12px">
					<el-button :icon="ArrowLeft" @click="goBack" text>Routes</el-button>
					<h2 style="margin: 0">{{ routeName }}</h2>
					<el-tag :type="statusTagType" effect="dark" size="small">
						{{ statusPhase }}
					</el-tag>
				</div>
			</el-col>
			<el-col :span="12" style="text-align: right">
				<el-button :icon="Refresh" @click="handleRefresh">Reload</el-button>
				<el-button :icon="Document" @click="yamlModalVisible = true">Edit YAML</el-button>
				<el-button type="primary" :loading="saving" @click="handleSave">
					{{ saving ? 'Saving...' : 'Save' }}
				</el-button>
			</el-col>
		</el-row>

		<el-alert v-if="routeStore.error" :title="routeStore.error" type="error" show-icon style="margin-bottom: 16px" />

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
				<!-- Configuration (left) - editable form -->
				<el-col :span="14">
					<el-card shadow="never" style="height: 100%">
						<template #header>
							<span style="font-weight: 600">Configuration</span>
						</template>

						<!-- Metadata (read-only) -->
						<el-collapse style="margin-bottom: 20px">
							<el-collapse-item title="Metadata" name="metadata">
								<MetadataDisplay :meta="routeResource.meta" />
							</el-collapse-item>
						</el-collapse>

						<el-form label-width="140px" label-position="right">
							<el-form-item label="Name">
								<el-input :model-value="form.meta?.name" disabled />
							</el-form-item>

							<el-form-item label="Labels">
								<LabelEditor v-model="formLabels" />
							</el-form-item>

							<el-divider content-position="left">Config</el-divider>

							<el-form-item label="Enabled">
								<el-switch v-model="form.config!.enabled" />
							</el-form-item>

							<el-form-item label="Backend Ref">
								<el-select v-model="form.config!.backendRef" placeholder="Select Backend" style="width: 100%" clearable
									filterable :loading="backendStore.loading">
									<el-option v-for="item in backendStore.items" :key="item.meta?.name" :label="item.meta?.name"
										:value="item.meta?.name" />
								</el-select>
							</el-form-item>

							<el-divider content-position="left">Match</el-divider>

							<el-form-item label="Match Type">
								<el-select v-model="matchMode" placeholder="Select match type" style="width: 100%">
									<el-option v-for="opt in matchModeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
								</el-select>
							</el-form-item>

							<!-- SNI -->
							<el-form-item v-if="matchMode === 'sni'" label="SNI">
								<el-input v-model="form.config!.match!.sni" placeholder="example.com" />
							</el-form-item>

							<!-- Path (exact) -->
							<el-form-item v-if="matchMode === 'path'" label="Path">
								<el-input v-model="form.config!.match!.path" placeholder="/exact/path" />
							</el-form-item>

							<!-- Path Prefix -->
							<el-form-item v-if="matchMode === 'pathPrefix'" label="Path Prefix">
								<el-input v-model="form.config!.match!.pathPrefix" placeholder="/api/" />
							</el-form-item>

							<!-- Header -->
							<template v-if="matchMode === 'header'">
								<el-form-item label="Header Name">
									<el-input v-model="form.config!.match!.header!.name" placeholder="X-Custom-Header" />
								</el-form-item>
								<el-form-item label="Header Value">
									<el-input v-model="form.config!.match!.header!.value" placeholder="expected-value" />
								</el-form-item>
							</template>

							<!-- JWT -->
							<template v-if="matchMode === 'jwt'">
								<el-form-item label="JWT Claim">
									<el-input v-model="form.config!.match!.jwt!.claim" placeholder="sub" />
								</el-form-item>
								<el-form-item label="JWT Value">
									<el-input v-model="form.config!.match!.jwt!.value" placeholder="expected-value" />
								</el-form-item>
							</template>
						</el-form>
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

		</template>

		<!-- Edit YAML Modal -->
		<EditYamlModal v-model:visible="yamlModalVisible" :yaml-content="yamlContent" @save="handleYamlSave" />
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
