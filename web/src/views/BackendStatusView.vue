<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed, watch, toRaw, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, ArrowRight, Refresh, Document, Plus, Delete } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import type { FormRules } from 'element-plus'
import { useBackendStore } from '@/stores/backend'
import { useCaStore } from '@/stores/ca'
import { useCredentialStore } from '@/stores/credential'
import { stringify as yamlStringify, parse as yamlParse } from 'yaml'
import { Codemirror } from 'vue-codemirror'
import { yaml as yamlLang } from '@codemirror/lang-yaml'
import { oneDark } from '@codemirror/theme-one-dark'
import { EditorState } from '@codemirror/state'
import NetworkTopology from '@/components/NetworkTopology.vue'
import type { NormalizedServer } from '@/components/NetworkTopology.vue'
import LabelEditor from '@/components/LabelEditor.vue'
import MetadataDisplay from '@/components/MetadataDisplay.vue'
import EditYamlModal from '@/components/EditYamlModal.vue'
import {
	lbLabel,
	countHealthyServers,
	countReadyServers,
	countTotalServers,
	healthTagType,
	readinessLabel,
	healthinessLabel,
	booleanStatusTagType,
	targetVisualState,
} from '@/utils/backend'
import { V1LoadBalancingType } from '@/generated/backend'
import type { V1Backend } from '@/generated/backend'
import { formatDate, formatDateFull } from '@/utils/format'
import { normalizeBackendForm } from '@/utils/backendForm'

const route = useRoute()
const router = useRouter()
const backendStore = useBackendStore()
const caStore = useCaStore()
const credentialStore = useCredentialStore()

const backendName = computed(() => route.params.name as string)
const backend = computed(() => backendStore.current)

const saving = ref(false)
const yamlModalVisible = ref(false)

// Editable form, initialized from fetched resource
const form = ref<V1Backend>({})

watch(backend, (val) => {
	if (val) {
		form.value = normalizeBackendForm(structuredClone(toRaw(val)))
	}
}, { immediate: true })

const lbTypeOptions = [
	{ label: 'Unspecified', value: V1LoadBalancingType.LoadBalancingTypeUnspecified },
	{ label: 'Round Robin', value: V1LoadBalancingType.LoadBalancingTypeRoundRobin },
	{ label: 'Least Connections', value: V1LoadBalancingType.LoadBalancingTypeLeastConnections },
	{ label: 'Random', value: V1LoadBalancingType.LoadBalancingTypeRandom },
	{ label: 'Weighted Round Robin', value: V1LoadBalancingType.LoadBalancingTypeWeightedRoundRobin },
]

// Server URL validation
const serverUrlPattern = /^https?:\/\/[a-zA-Z0-9._\-\[\]:]+?(:\d{1,5})?(\/.*)?$/
function isValidServerUrl(url: string): boolean {
	return serverUrlPattern.test(url)
}
function addServer() {
	if (form.value.config) {
		if (!form.value.config.servers) form.value.config.servers = []
		form.value.config.servers.push('')
	}
}
function removeServer(index: number) {
	form.value.config?.servers?.splice(index, 1)
}

// Labels computed for LabelEditor
const formLabels = computed({
	get: () => form.value.meta?.labels ?? {},
	set: (val: Record<string, string>) => {
		if (form.value.meta) {
			form.value.meta.labels = val
		}
	},
})

// Extra claims as a newline-separated string
const extraClaimsText = computed({
	get: () => (form.value.config?.impersonationConfig?.extraClaims ?? []).join('\n'),
	set: (val: string) => {
		if (form.value.config?.impersonationConfig) {
			form.value.config.impersonationConfig.extraClaims = val.split('\n').filter((s) => s.trim() !== '')
		}
	},
})

// Normalized server list for topology display (from store, not form)
const normalizedServers = computed<NormalizedServer[]>(() => {
	const servers = backend.value?.config?.servers ?? []
	const statuses = backend.value?.status?.targetStatuses ?? {}
	return servers.map((url) => {
		const status = statuses[url]
		return {
			url,
			readiness: readinessLabel(status?.readiness?.isReady),
			healthiness: healthinessLabel(status?.healthiness?.isHealthy),
			readinessReason: status?.readiness?.reason ?? '',
			healthinessReason: status?.healthiness?.reason ?? '',
			readinessLastTransitionTime: status?.readiness?.lastTransitionTime,
			healthinessLastTransitionTime: status?.healthiness?.lastTransitionTime,
			visualState: targetVisualState(status),
		}
	})
})

const readyCount = computed(() =>
	countReadyServers(backend.value?.config?.servers ?? [], backend.value?.status?.targetStatuses),
)
const healthyCount = computed(() =>
	countHealthyServers(backend.value?.config?.servers ?? [], backend.value?.status?.targetStatuses),
)
const totalCount = computed(() => countTotalServers(backend.value?.config?.servers ?? []))
const readinessTag = computed(() => healthTagType(readyCount.value, totalCount.value))
const healthTag = computed(() => healthTagType(healthyCount.value, totalCount.value))

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

const statusPhase = computed(() => backend.value?.status?.phase ?? 'Unknown')

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
async function handleSave() {
	saving.value = true
	try {
		await backendStore.updateBackend(form.value)
		await backendStore.fetchBackend(backendName.value)
		ElMessage.success('Backend updated')
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Save failed')
	} finally {
		saving.value = false
	}
}

async function handleYamlSave(parsed: unknown) {
	saving.value = true
	try {
		const resource = parsed as V1Backend
		if (!resource.meta?.name) {
			resource.meta = { ...resource.meta, name: backendName.value }
		}
		await backendStore.updateBackend(resource)
		await backendStore.fetchBackend(backendName.value)
		ElMessage.success('Backend updated from YAML')
		yamlModalVisible.value = false
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Save failed')
	} finally {
		saving.value = false
	}
}

function handleRefresh() {
	backendStore.fetchBackend(backendName.value).catch(() => { })
}

function goBack() {
	router.push('/backends')
}

function visitRef(type: string, name: string) {
	if (name) router.push(`/${type}/${name}`)
}


onMounted(() => {
	backendStore.fetchBackend(backendName.value).catch(() => { })
	caStore.fetchCas().catch(() => { })
	credentialStore.fetchCredentials().catch(() => { })
})

onUnmounted(() => {
	backendStore.clearCurrent()
})
</script>

<template>
	<div>
		<!-- Header -->
		<el-row justify="space-between" align="middle" style="margin-bottom: 16px">
			<el-col :span="12">
				<div style="display: flex; align-items: center; gap: 12px">
					<el-button :icon="ArrowLeft" @click="goBack" text>Backends</el-button>
					<h2 style="margin: 0">{{ backendName }}</h2>
					<el-tag :type="readinessTag" effect="dark" size="small">
						{{ readyCount }}/{{ totalCount }} Ready
					</el-tag>
					<el-tag :type="healthTag" effect="dark" size="small">
						{{ healthyCount }}/{{ totalCount }} Healthy
					</el-tag>
					<el-tag :type="statusPhase === 'READY' ? 'success' : statusPhase === 'Inactive' ? 'info' : 'warning'"
						effect="dark" size="small">
						{{ statusPhase }}
					</el-tag>
				</div>
			</el-col>
			<el-col :span="12" style="text-align: right">
				<el-button plain :icon="Refresh" @click="handleRefresh">Reload</el-button>
				<el-button plain :icon="Document" @click="yamlModalVisible = true">Edit YAML</el-button>
				<el-button type="primary" :loading="saving" @click="handleSave">
					{{ saving ? 'Saving...' : 'Save' }}
				</el-button>
			</el-col>
		</el-row>

		<el-alert v-if="backendStore.error" :title="backendStore.error" type="error" show-icon
			style="margin-bottom: 16px" />

		<!-- Loading -->
		<el-card v-if="backendStore.loading" shadow="never" style="margin-bottom: 16px">
			<el-skeleton :rows="6" animated />
		</el-card>

		<!-- Not found -->
		<el-empty v-else-if="!backend" :description="`Backend '${backendName}' not found`">
			<el-button type="primary" @click="goBack">Back to Backends</el-button>
		</el-empty>

		<!-- Content -->
		<template v-else>
			<!-- Network Topology Map -->
			<el-card shadow="never" style="margin-bottom: 16px; background: #141414">
				<div class="topology-container">
					<NetworkTopology :backendName="backend.meta?.name ?? '-'" :lbType="lbLabel(backend.config?.type as string)"
						:servers="normalizedServers" />
				</div>
			</el-card>

			<!-- Two-column layout: Configuration + Target Health -->
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
								<MetadataDisplay :meta="backend.meta" />
							</el-collapse-item>
						</el-collapse>

						<el-form label-width="160px" label-position="right">
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

							<el-form-item label="Servers">
								<div style="width: 100%">
									<div v-for="(server, idx) in form.config!.servers" :key="idx"
										style="display: flex; gap: 8px; margin-bottom: 8px; align-items: start;">
										<div style="flex: 1">
											<el-input v-model="form.config!.servers![idx]" placeholder="https://10.0.0.1:6443" />
											<div v-if="server && !isValidServerUrl(server)"
												style="color: var(--el-color-danger); font-size: 12px; margin-top: 2px;">
												Invalid URL format (e.g. https://host:port/path)
											</div>
										</div>
										<el-button type="danger" :icon="Delete" plain @click="removeServer(idx)" />
									</div>
									<el-button type="primary" size="small" plain :icon="Plus" @click="addServer()">
										Add Server
									</el-button>
								</div>
							</el-form-item>

							<el-form-item label="Load Balancing Type">
								<el-select v-model="form.config!.type" style="width: 100%">
									<el-option v-for="opt in lbTypeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
								</el-select>
							</el-form-item>

							<el-form-item label="CA Ref">
								<div class="input-group">
									<el-select v-model="form.config!.caRef" placeholder="Select Certificate Authority" style="width: 100%"
										clearable filterable :loading="caStore.loading">
										<el-option v-for="item in caStore.items" :key="item.meta?.name" :label="item.meta?.name"
											:value="item.meta?.name" />
									</el-select>
									<el-button link @click="visitRef('cas', form.config.caRef)" href="#" type="primary"
										v-if="form.config?.caRef">
										Visit
										<el-icon class="el-icon--right">
											<ArrowRight />
										</el-icon>
									</el-button>


								</div>
							</el-form-item>

							<el-form-item label="Auth Ref">
								<div class="input-group">
									<el-select v-model="form.config!.authRef" placeholder="Select Credential" style="width: 100%"
										clearable filterable :loading="credentialStore.loading">
										<el-option v-for="item in credentialStore.items" :key="item.meta?.name" :label="item.meta?.name"
											:value="item.meta?.name" />
									</el-select>
									<el-button link @click="visitRef('credentials', form.config?.authRef)" href="#" type="primary"
										v-if="form.config?.authRef">
										Visit
										<el-icon class="el-icon--right">
											<ArrowRight />
										</el-icon>
									</el-button>

								</div>
							</el-form-item>

							<el-form-item label="Cache TTL">
								<el-input v-model="form.config!.cacheTtl" placeholder="e.g. 30s, 5m" />
							</el-form-item>

							<el-form-item label="Skip TLS Verify">
								<el-switch v-model="form.config!.insecureSkipTlsVerify" />
							</el-form-item>

							<!-- Advanced section -->
							<el-collapse style="margin-top: 12px">
								<el-collapse-item title="Advanced" name="advanced">
									<el-form-item label="Enable Impersonation" style="margin-top: 12px">
										<el-switch v-model="form.config!.impersonationConfig!.enabled" />
									</el-form-item>

									<el-form-item label="Username Claim">
										<el-input v-model="form.config!.impersonationConfig!.usernameClaim" placeholder="sub"
											:disabled="!form.config!.impersonationConfig!.enabled" />
									</el-form-item>

									<el-form-item label="Groups Claim">
										<el-input v-model="form.config!.impersonationConfig!.groupsClaim" placeholder="groups"
											:disabled="!form.config!.impersonationConfig!.enabled" />
									</el-form-item>

							<el-form-item label="Extra Claims">
								<el-input v-model="extraClaimsText" type="textarea" :rows="3" placeholder="One claim per line"
									:disabled="!form.config!.impersonationConfig!.enabled" />
							</el-form-item>

							<el-divider content-position="left">Health Probe</el-divider>

							<el-form-item label="Path">
								<el-input v-model="form.config!.probes!.healthiness!.path" placeholder="/healthz" />
							</el-form-item>

							<el-form-item label="Timeout Seconds">
								<el-input v-model="form.config!.probes!.healthiness!.timeoutSeconds" placeholder="1" />
							</el-form-item>

							<el-form-item label="Period Seconds">
								<el-input v-model="form.config!.probes!.healthiness!.periodSeconds" placeholder="5" />
							</el-form-item>

							<el-form-item label="Failure Threshold">
								<el-input v-model="form.config!.probes!.healthiness!.failureThreshold" placeholder="3" />
							</el-form-item>

							<el-form-item label="Success Threshold">
								<el-input v-model="form.config!.probes!.healthiness!.successThreshold" placeholder="3" />
							</el-form-item>

							<el-form-item label="Initial Delay Seconds">
								<el-input v-model="form.config!.probes!.healthiness!.initialDelaySeconds" placeholder="1" />
							</el-form-item>

							<el-divider content-position="left">Readiness Probe</el-divider>

							<el-form-item label="Path">
								<el-input v-model="form.config!.probes!.readiness!.path" placeholder="/readyz" />
							</el-form-item>

							<el-form-item label="Timeout Seconds">
								<el-input v-model="form.config!.probes!.readiness!.timeoutSeconds" placeholder="1" />
							</el-form-item>

							<el-form-item label="Period Seconds">
								<el-input v-model="form.config!.probes!.readiness!.periodSeconds" placeholder="5" />
							</el-form-item>

							<el-form-item label="Failure Threshold">
								<el-input v-model="form.config!.probes!.readiness!.failureThreshold" placeholder="3" />
							</el-form-item>

							<el-form-item label="Success Threshold">
								<el-input v-model="form.config!.probes!.readiness!.successThreshold" placeholder="3" />
							</el-form-item>

							<el-form-item label="Initial Delay Seconds">
								<el-input v-model="form.config!.probes!.readiness!.initialDelaySeconds" placeholder="1" />
							</el-form-item>
						</el-collapse-item>
					</el-collapse>
						</el-form>
					</el-card>
				</el-col>

				<!-- Target Health (right) -->
				<el-col :span="10">

					<el-card shadow="never">
						<template #header>
							<div style="display: flex; justify-content: space-between; align-items: center">
								<span style="font-weight: 600">Status</span>
								<el-tag :type="statusPhase === 'READY' ? 'success' : statusPhase === 'Inactive' ? 'info' : 'warning'"
									effect="dark" size="small">
									{{ statusPhase }}
								</el-tag>
							</div>
						</template>

						<el-descriptions :column="1" border size="default">
							<el-descriptions-item label="Phase">
								<el-tag :type="statusPhase === 'READY' ? 'success' : statusPhase === 'Inactive' ? 'info' : 'warning'"
									effect="dark" size="small">
									{{ statusPhase }}
								</el-tag>
							</el-descriptions-item>
							<el-descriptions-item label="Reason">
								<span v-if="backend.status?.reason" style="color: #f56c6c; font-size: 12px">
									{{ backend.status.reason }}
								</span>
								<span v-else style="color: #909399">-</span>
							</el-descriptions-item>
							<el-descriptions-item label="Last Transition">
								<el-tooltip v-if="backend.status?.lastTransitionTime"
									:content="formatDateFull(backend.status.lastTransitionTime)" placement="top">
									<span>{{ formatDate(backend.status.lastTransitionTime) }}</span>
								</el-tooltip>
								<span v-else style="color: #909399">-</span>
							</el-descriptions-item>
						</el-descriptions>
					</el-card>

					<el-divider content-position="left"></el-divider>

					<el-card shadow="never">
						<template #header>
							<div style="display: flex; justify-content: space-between; align-items: center">
								<span style="font-weight: 600">Target Status</span>
								<div style="display: flex; gap: 8px">
									<el-tag :type="readinessTag" effect="dark" size="small">
										{{ readyCount }}/{{ totalCount }} Ready
									</el-tag>
									<el-tag :type="healthTag" effect="dark" size="small">
										{{ healthyCount }}/{{ totalCount }} Healthy
									</el-tag>
								</div>
							</div>
						</template>

						<el-empty v-if="normalizedServers.length === 0" description="No servers configured" :image-size="60" />

						<el-table v-else :data="normalizedServers" style="width: 100%" stripe size="small">
							<el-table-column prop="url" label="Target" min-width="180">
								<template #default="{ row }">
									<span style="font-family: monospace; font-size: 12px">{{ row.url }}</span>
								</template>
							</el-table-column>
						<el-table-column label="Readiness" width="120">
							<template #default="{ row }">
								<el-tag :type="booleanStatusTagType(row.readiness === 'Ready' ? true : row.readiness === 'Not Ready' ? false : undefined)"
									effect="dark" size="small">
									{{ row.readiness }}
								</el-tag>
							</template>
						</el-table-column>
						<el-table-column label="Health" width="120">
							<template #default="{ row }">
								<el-tag
									:type="booleanStatusTagType(row.healthiness === 'Healthy' ? true : row.healthiness === 'Unhealthy' ? false : undefined)"
									effect="dark" size="small">
									{{ row.healthiness }}
								</el-tag>
							</template>
						</el-table-column>
						<el-table-column label="Readiness Reason" min-width="140">
							<template #default="{ row }">
								<span v-if="row.readinessReason" style="color: #f56c6c; font-size: 12px">{{ row.readinessReason }}</span>
								<span v-else style="color: #909399">-</span>
							</template>
						</el-table-column>
						<el-table-column label="Health Reason" min-width="140">
							<template #default="{ row }">
								<span v-if="row.healthinessReason" style="color: #f56c6c; font-size: 12px">{{ row.healthinessReason }}</span>
								<span v-else style="color: #909399">-</span>
							</template>
						</el-table-column>
						<el-table-column label="Ready Since" width="130">
							<template #default="{ row }">
								<el-tooltip v-if="row.readinessLastTransitionTime"
									:content="formatDateFull(row.readinessLastTransitionTime)" placement="top">
									<span style="font-size: 12px">{{ formatDate(row.readinessLastTransitionTime) }}</span>
								</el-tooltip>
								<span v-else style="color: #909399">-</span>
							</template>
						</el-table-column>
						<el-table-column label="Healthy Since" width="130">
							<template #default="{ row }">
								<el-tooltip v-if="row.healthinessLastTransitionTime"
									:content="formatDateFull(row.healthinessLastTransitionTime)" placement="top">
									<span style="font-size: 12px">{{ formatDate(row.healthinessLastTransitionTime) }}</span>
								</el-tooltip>
								<span v-else style="color: #909399">-</span>
							</template>
						</el-table-column>
					</el-table>
				</el-card>
				</el-col>
			</el-row>
		</template>

		<!-- Edit YAML Modal -->
		<EditYamlModal v-model:visible="yamlModalVisible" :yaml-content="yamlContent" @save="handleYamlSave" />
	</div>
</template>

<style scoped>
.topology-container {
	max-height: 450px;
	overflow-y: auto;
	overflow-x: auto;
	text-align: center;
}

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

.input-group {
	width: 100%;
	display: flex;
	align-items: center;
	gap: 1em;
}
</style>
