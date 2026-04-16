<script setup lang="ts">
import { onMounted, onUnmounted, computed, toRaw } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Refresh } from '@element-plus/icons-vue'
import { useBackendStore } from '@/stores/backend'
import { stringify as yamlStringify } from 'yaml'
import { Codemirror } from 'vue-codemirror'
import { yaml as yamlLang } from '@codemirror/lang-yaml'
import { oneDark } from '@codemirror/theme-one-dark'
import { EditorState } from '@codemirror/state'
import NetworkTopology from '@/components/NetworkTopology.vue'
import type { NormalizedServer } from '@/components/NetworkTopology.vue'
import { lbLabel, countHealthyServers, countTotalServers, healthTagType } from '@/utils/backend'
import { formatDate, formatDateFull } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const backendStore = useBackendStore()

const backendName = computed(() => route.params.name as string)

const backend = computed(() => backendStore.current)

// Normalized server list: config.servers joined with status.targetStatuses
const normalizedServers = computed<NormalizedServer[]>(() => {
	const servers = backend.value?.config?.servers ?? []
	const statuses = backend.value?.status?.targetStatuses ?? {}
	return servers.map((url) => {
		const status = statuses[url]
		return {
			url,
			phase: status?.phase ?? 'Unknown',
			reason: status?.reason ?? '',
			lastTransitionTime: status?.lastTransitionTime,
		}
	})
})

const healthyCount = computed(() =>
	countHealthyServers(backend.value?.config?.servers ?? [], backend.value?.status?.targetStatuses),
)
const totalCount = computed(() => countTotalServers(backend.value?.config?.servers ?? []))

const healthTag = computed(() => healthTagType(healthyCount.value, totalCount.value))

const yamlContent = computed(() => {
	if (!backend.value) return ''
	try {
		const raw = structuredClone(toRaw(backend.value))
		return yamlStringify(raw, { lineWidth: 120 })
	} catch {
		return '# Failed to serialize resource'
	}
})

const cmExtensions = [yamlLang(), oneDark, EditorState.readOnly.of(true)]

const labelEntries = computed(() => {
	const labels = backend.value?.meta?.labels
	if (!labels) return []
	return Object.entries(labels)
})

function handleRefresh() {
	backendStore.fetchBackend(backendName.value).catch(() => {})
}

function goBack() {
	router.push('/backends')
}

onMounted(() => {
	backendStore.fetchBackend(backendName.value).catch(() => {})
})

onUnmounted(() => {
	backendStore.clearCurrent()
})
</script>

<template>
	<div>
		<!-- Header -->
		<el-row justify="space-between" align="middle" style="margin-bottom: 16px">
			<el-col :span="16">
				<div style="display: flex; align-items: center; gap: 12px">
					<el-button :icon="ArrowLeft" @click="goBack" text>Backends</el-button>
					<h2 style="margin: 0">{{ backendName }}</h2>
          <el-tag :type="healthTag" effect="dark" size="small">
            {{ healthyCount }}/{{ totalCount }} Healthy
          </el-tag>
				</div>
			</el-col>
			<el-col :span="8" style="text-align: right">
				<el-button :icon="Refresh" @click="handleRefresh">Reload</el-button>
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
								{{ backend.meta?.name ?? '-' }}
							</el-descriptions-item>
							<el-descriptions-item label="Load Balancing">
								{{ lbLabel(backend.config?.type as string) }}
							</el-descriptions-item>
							<el-descriptions-item label="Created">
								<el-tooltip :content="formatDateFull(backend.meta?.created)" placement="top">
									<span>{{ formatDate(backend.meta?.created) }}</span>
								</el-tooltip>
							</el-descriptions-item>
							<el-descriptions-item label="Updated">
								<el-tooltip :content="formatDateFull(backend.meta?.updated)" placement="top">
									<span>{{ formatDate(backend.meta?.updated) }}</span>
								</el-tooltip>
							</el-descriptions-item>
							<el-descriptions-item label="UID">
								<span style="font-family: monospace; font-size: 12px">{{ backend.meta?.uid ?? '-' }}</span>
							</el-descriptions-item>
							<el-descriptions-item label="Cache TTL">
								{{ backend.config?.cacheTtl || '-' }}
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

						<!-- Authentication section -->
						<h4 class="section-title">Authentication</h4>
						<el-descriptions :column="2" border size="default">
							<el-descriptions-item label="CA Reference">
								<router-link v-if="backend.config?.caRef" :to="'/cas'" style="text-decoration: none">
									<el-link type="primary">{{ backend.config.caRef }}</el-link>
								</router-link>
								<span v-else style="color: #909399">-</span>
							</el-descriptions-item>
							<el-descriptions-item label="Auth Reference">
								<router-link v-if="backend.config?.authRef" :to="'/credentials'" style="text-decoration: none">
									<el-link type="primary">{{ backend.config.authRef }}</el-link>
								</router-link>
								<span v-else style="color: #909399">-</span>
							</el-descriptions-item>
							<el-descriptions-item label="Skip TLS Verify">
								<el-tag :type="backend.config?.insecureSkipTlsVerify ? 'warning' : 'info'" size="small">
									{{ backend.config?.insecureSkipTlsVerify ? 'Yes' : 'No' }}
								</el-tag>
							</el-descriptions-item>
						</el-descriptions>

						<!-- Impersonation section -->
						<h4 class="section-title">Impersonation</h4>
						<el-descriptions :column="2" border size="default">
							<el-descriptions-item label="Enabled">
								<el-tag :type="backend.config?.impersonationConfig?.enabled ? 'success' : 'info'" size="small">
									{{ backend.config?.impersonationConfig?.enabled ? 'Yes' : 'No' }}
								</el-tag>
							</el-descriptions-item>
							<el-descriptions-item label="Username Claim">
								{{ backend.config?.impersonationConfig?.usernameClaim || '-' }}
							</el-descriptions-item>
							<el-descriptions-item label="Groups Claim">
								{{ backend.config?.impersonationConfig?.groupsClaim || '-' }}
							</el-descriptions-item>
							<el-descriptions-item label="Extra Claims">
								<span v-if="(backend.config?.impersonationConfig?.extraClaims ?? []).length > 0">
									{{ backend.config!.impersonationConfig!.extraClaims!.join(', ') }}
								</span>
								<span v-else style="color: #909399">-</span>
							</el-descriptions-item>
						</el-descriptions>
					</el-card>
				</el-col>

				<!-- Target Health (right) -->
				<el-col :span="10">
					<el-card shadow="never" style="height: 100%">
						<template #header>
							<div style="display: flex; justify-content: space-between; align-items: center">
								<span style="font-weight: 600">Target Health</span>
                <el-tag :type="healthTag" effect="dark" size="small">
                  {{ healthyCount }}/{{ totalCount }}
                </el-tag>
							</div>
						</template>

						<el-empty v-if="normalizedServers.length === 0" description="No servers configured" :image-size="60" />

						<el-table v-else :data="normalizedServers" style="width: 100%" stripe size="small">
							<el-table-column prop="url" label="Target" min-width="180">
								<template #default="{ row }">
									<span style="font-family: monospace; font-size: 12px">{{ row.url }}</span>
								</template>
							</el-table-column>
							<el-table-column prop="phase" label="Phase" width="110">
								<template #default="{ row }">
									<el-tag :type="row.phase === 'Healthy' ? 'success' : row.phase === 'Unhealthy' ? 'danger' : 'warning'"
										effect="dark" size="small">
										{{ row.phase }}
									</el-tag>
								</template>
							</el-table-column>
							<el-table-column label="Error" min-width="140">
								<template #default="{ row }">
									<span v-if="row.reason" style="color: #f56c6c; font-size: 12px">{{ row.reason }}</span>
									<span v-else style="color: #909399">-</span>
								</template>
							</el-table-column>
							<el-table-column label="Last Transition" width="130">
								<template #default="{ row }">
									<el-tooltip v-if="row.lastTransitionTime" :content="formatDateFull(row.lastTransitionTime)"
										placement="top">
										<span style="font-size: 12px">{{ formatDate(row.lastTransitionTime) }}</span>
									</el-tooltip>
									<span v-else style="color: #909399">-</span>
								</template>
							</el-table-column>
						</el-table>
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
</style>
