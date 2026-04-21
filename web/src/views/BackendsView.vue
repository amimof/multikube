<script setup lang="ts">
import { onMounted, ref, computed, toRaw } from 'vue'
import { useRouter } from 'vue-router'
import { Plus, Refresh, Delete, Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useBackendStore } from '@/stores/backend'
import { useCaStore } from '@/stores/ca'
import { useCredentialStore } from '@/stores/credential'
import { useResourceTable } from '@/composables/useResourceTable'
import { lbTypeLabels, countHealthyServers, countTotalServers, healthTagType } from '@/utils/backend'
import { formatDate } from '@/utils/format'
import { V1LoadBalancingType } from '@/generated/backend'
import type { V1Backend } from '@/generated/backend'
import LabelEditor from '@/components/LabelEditor.vue'
import ConfirmDelete from '@/components/ConfirmDelete.vue'

const backendStore = useBackendStore()
const caStore = useCaStore()
const credentialStore = useCredentialStore()
const router = useRouter()

const { nameFilter, displayItems } = useResourceTable(computed(() => backendStore.items))

const dialogVisible = ref(false)
const saving = ref(false)
const deleteDialogVisible = ref(false)
const deleteTarget = ref<V1Backend | null>(null)
const selectedRows = ref<V1Backend[]>([])
const bulkDeleteVisible = ref(false)
const bulkDeleting = ref(false)

const form = ref<V1Backend>(createEmptyBackend())

function createEmptyBackend(): V1Backend {
	return {
		version: 'backend/v1',
		meta: { name: '', labels: {} },
		config: {
			enabled: true,
			servers: [],
			caRef: '',
			authRef: '',
			insecureSkipTlsVerify: false,
			cacheTtl: '30s',
			type: V1LoadBalancingType.LoadBalancingTypeRoundRobin,
			impersonationConfig: {
				name: 'default',
				enabled: true,
				usernameClaim: 'sub',
				groupsClaim: 'groups',
				extraClaims: [],
			},
		},
	}
}

const lbTypeOptions = [
	{ label: 'Unspecified', value: V1LoadBalancingType.LoadBalancingTypeUnspecified },
	{ label: 'Round Robin', value: V1LoadBalancingType.LoadBalancingTypeRoundRobin },
	{ label: 'Least Connections', value: V1LoadBalancingType.LoadBalancingTypeLeastConnections },
	{ label: 'Random', value: V1LoadBalancingType.LoadBalancingTypeRandom },
	{ label: 'Weighted Round Robin', value: V1LoadBalancingType.LoadBalancingTypeWeightedRoundRobin },
]

// Form validation
const isFormValid = computed(() => {
	const name = (form.value.meta?.name ?? '').trim()
	const servers = form.value.config?.servers ?? []
	const cacheTtl = (form.value.config?.cacheTtl ?? '').trim()
	return name.length > 0 && servers.length > 0 && cacheTtl.length > 0
})

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

// Extra claims as a newline-separated string for textarea editing
const extraClaimsText = computed({
	get: () => (form.value.config?.impersonationConfig?.extraClaims ?? []).join('\n'),
	set: (val: string) => {
		if (form.value.config?.impersonationConfig) {
			form.value.config.impersonationConfig.extraClaims = val.split('\n').filter((s) => s.trim() !== '')
		}
	},
})

// Sort helpers
function sortByCreated(a: any, b: any): number {
	const ta = new Date(a.meta?.created ?? 0).getTime()
	const tb = new Date(b.meta?.created ?? 0).getTime()
	return ta - tb
}

function sortByReady(a: any, b: any): number {
	const ta = countTotalServers(a.config?.servers ?? [])
	const tb = countTotalServers(b.config?.servers ?? [])
	const ra = ta === 0 ? -1 : countHealthyServers(a.config?.servers ?? [], a.status?.targetStatuses) / ta
	const rb = tb === 0 ? -1 : countHealthyServers(b.config?.servers ?? [], b.status?.targetStatuses) / tb
	return ra - rb
}

function sortByStatus(a: any, b: any): number {
	const sa = a.status?.phase ?? ''
	const sb = b.status?.phase ?? ''
	return sa.localeCompare(sb)
}

function sortByType(a: any, b: any): number {
	const la = lbTypeLabels[a.config?.type ?? ''] ?? a.config?.type ?? ''
	const lb = lbTypeLabels[b.config?.type ?? ''] ?? b.config?.type ?? ''
	return la.localeCompare(lb)
}

function sortByServers(a: any, b: any): number {
	const sa = (a.config?.servers ?? []).join(', ')
	const sb = (b.config?.servers ?? []).join(', ')
	return sa.localeCompare(sb)
}

// Selection
function handleSelectionChange(rows: V1Backend[]) {
	selectedRows.value = rows
}

function handleRowClick(row: V1Backend, column: any) {
	if (column?.type === 'selection') return
	// openEdit(row)
	viewStatus(row)
}

function viewStatus(row: V1Backend) {
	const name = row.meta?.name
	if (name) router.push(`/backends/${name}`)
}

function openCreate() {
	form.value = createEmptyBackend()
	dialogVisible.value = true
}

function confirmDelete(row: V1Backend) {
	deleteTarget.value = row
	deleteDialogVisible.value = true
}

async function handleDelete() {
	if (!deleteTarget.value) return
	try {
		await backendStore.deleteBackend(deleteTarget.value)
		ElMessage.success('Backend deleted')
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Delete failed')
	}
	deleteTarget.value = null
}

function confirmBulkDelete() {
	bulkDeleteVisible.value = true
}

async function handleBulkDelete() {
	bulkDeleting.value = true
	try {
		const { succeeded, failed } = await backendStore.deleteManyBackends(selectedRows.value)
		selectedRows.value = []
		if (failed.length === 0) {
			ElMessage.success(`Deleted ${succeeded} backend${succeeded === 1 ? '' : 's'}`)
		} else if (succeeded > 0) {
			ElMessage.warning(`Deleted ${succeeded}, failed ${failed.length}: ${failed.map((f) => f.name).join(', ')}`)
		} else {
			ElMessage.error(`All ${failed.length} deletes failed`)
		}
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Bulk delete failed')
	} finally {
		bulkDeleting.value = false
	}
}

async function handleSave() {
	saving.value = true
	try {
		await backendStore.createBackend(form.value)
		ElMessage.success('Backend created')
		dialogVisible.value = false
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Save failed')
	} finally {
		saving.value = false
	}
}

async function handleToggleEnabled(row: V1Backend, enabled: boolean) {
	try {
		const updated = structuredClone(toRaw(row))
		if (!updated.config) updated.config = {}
		updated.config.enabled = enabled
		await backendStore.updateBackend(updated)
		ElMessage.success(`${row.meta?.name} ${enabled ? 'enabled' : 'disabled'}`)
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Update failed')
	}
}

function handleRefresh() {
	backendStore.fetchBackends().catch(() => { })
}

onMounted(() => {
	backendStore.fetchBackends().catch(() => { })
	caStore.fetchCas().catch(() => { })
	credentialStore.fetchCredentials().catch(() => { })
})
</script>

<template>
	<div>
		<el-row justify="space-between" align="middle" style="margin-bottom: 16px">
			<el-col :span="12">
				<h2 style="margin: 0">Backends</h2>
			</el-col>
			<el-col :span="12" style="text-align: right">
				<el-button :icon="Refresh" plain @click="handleRefresh">Reload</el-button>
				<el-button :icon="Plus" plain @click="openCreate">New</el-button>
			</el-col>
		</el-row>

		<el-alert v-if="backendStore.error" :title="backendStore.error" type="error" show-icon
			style="margin-bottom: 16px" />

		<el-empty v-if="!backendStore.loading && backendStore.items.length === 0" description="No backends yet">
			<el-button type="primary" :icon="Plus" @click="openCreate">Create</el-button>
		</el-empty>

		<template v-else>
			<el-row :gutter="12" align="middle" style="margin-bottom: 12px">
				<el-col :span="12">
					<el-input v-model="nameFilter" placeholder="Filter by name..." clearable :prefix-icon="Search" />
				</el-col>
				<el-col :span="12" v-if="selectedRows.length > 0">
					<el-button type="danger" :icon="Delete" @click="confirmBulkDelete">
						Delete ({{ selectedRows.length }})
					</el-button>
				</el-col>
			</el-row>

			<el-table v-loading="backendStore.loading" element-loading-text="Loading..." :data="displayItems"
				style="width: 100%" row-key="meta.name" @row-click="handleRowClick" @selection-change="handleSelectionChange"
				:row-class-name="() => 'clickable-row'">
				<el-table-column type="selection" width="48" />
				<el-table-column label="Enabled" width="90">
					<template #default="{ row }">
						<el-switch :model-value="row.config?.enabled ?? true" @update:model-value="handleToggleEnabled(row, $event)"
							@click.stop />
					</template>
				</el-table-column>
				<el-table-column prop="meta.name" label="Name" min-width="150" sortable />
				<el-table-column label="Status" width="100" sortable :sort-method="sortByStatus">
					<template #default="{ row }">
						<el-tag v-if="row.status?.phase"
							:type="row.status.phase === 'READY' ? 'success' : row.status.phase === 'Inactive' ? 'info' : 'warning'"
							effect="dark" size="small">
							{{ row.status.phase }}
						</el-tag>
						<span v-else>-</span>
					</template>
				</el-table-column>

				<el-table-column label="Ready" width="100" sortable :sort-method="sortByReady">
					<template #default="{ row }">
						<el-tag
							:type="healthTagType(countHealthyServers(row.config?.servers ?? [], row.status?.targetStatuses), countTotalServers(row.config?.servers ?? []))"
							effect="dark" size="small">
							{{ countHealthyServers(row.config?.servers ?? [], row.status?.targetStatuses) }}/{{
								countTotalServers(row.config?.servers ?? []) }}
						</el-tag>
					</template>
				</el-table-column>
				<el-table-column label="Type" width="180" sortable :sort-method="sortByType">
					<template #default="{ row }">
						{{ lbTypeLabels[row.config?.type ?? ''] ?? row.config?.type ?? '-' }}
					</template>
				</el-table-column>
				<el-table-column label="Servers" min-width="200" sortable :sort-method="sortByServers">
					<template #default="{ row }">
						{{ (row.config?.servers ?? []).join(', ') || '-' }}
					</template>
				</el-table-column>
				<el-table-column label="Created" width="180" sortable :sort-method="sortByCreated">
					<template #default="{ row }">
						{{ formatDate(row.meta?.created) }}
					</template>
				</el-table-column>
				<el-table-column label="Actions" width="80" fixed="right">
					<template #default="{ row }">
						<el-button :icon="Delete" type="danger" size="small" plain @click.stop="confirmDelete(row)" />
					</template>
				</el-table-column>
			</el-table>
		</template>

		<!-- Create Dialog -->
		<el-dialog v-model="dialogVisible" title="Create Backend" width="600" destroy-on-close>
			<el-form label-width="160px" label-position="right">
				<el-form-item label="Name" required>
					<el-input v-model="form.meta!.name" placeholder="my-backend" />
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
						<el-button type="primary" size="small" :icon="Plus" @click="addServer()">
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
					<el-select v-model="form.config!.caRef" placeholder="Select Certificate Authority" style="width: 100%"
						clearable filterable :loading="caStore.loading">
						<el-option v-for="item in caStore.items" :key="item.meta?.name" :label="item.meta?.name"
							:value="item.meta?.name" />
					</el-select>
				</el-form-item>

				<el-form-item label="Auth Ref">
					<el-select v-model="form.config!.authRef" placeholder="Select Credential" style="width: 100%" clearable
						filterable :loading="credentialStore.loading">
						<el-option v-for="item in credentialStore.items" :key="item.meta?.name" :label="item.meta?.name"
							:value="item.meta?.name" />
					</el-select>
				</el-form-item>

				<el-form-item label="Cache TTL" required>
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
					</el-collapse-item>
				</el-collapse>
			</el-form>

			<template #footer>
				<el-button @click="dialogVisible = false">Cancel</el-button>
				<el-button type="primary" :loading="saving" :disabled="!isFormValid" @click="handleSave">
					{{ saving ? 'Saving...' : 'Create' }}
				</el-button>
			</template>
		</el-dialog>

		<!-- Single delete confirmation -->
		<ConfirmDelete v-model:visible="deleteDialogVisible" :item-name="deleteTarget?.meta?.name ?? ''"
			@confirm="handleDelete" />

		<!-- Bulk delete confirmation -->
		<ConfirmDelete v-model:visible="bulkDeleteVisible"
			:message="`Delete ${selectedRows.length} selected backend${selectedRows.length === 1 ? '' : 's'}?`"
			@confirm="handleBulkDelete" />
	</div>
</template>
