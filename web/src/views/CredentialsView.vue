<script setup lang="ts">
import { onMounted, ref, computed, watch, toRaw } from 'vue'
import { useRouter } from 'vue-router'
import { Plus, Refresh, Delete, Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useCredentialStore } from '@/stores/credential'
import { useCertificateStore } from '@/stores/certificate'
import { useResourceTable } from '@/composables/useResourceTable'
import { formatDate } from '@/utils/format'
import type { V1Credential } from '@/generated/credential'
import LabelEditor from '@/components/LabelEditor.vue'
import ConfirmDelete from '@/components/ConfirmDelete.vue'

const credentialStore = useCredentialStore()
const certificateStore = useCertificateStore()
const router = useRouter()

const { nameFilter, displayItems } = useResourceTable(computed(() => credentialStore.items))

const dialogVisible = ref(false)
const saving = ref(false)
const deleteDialogVisible = ref(false)
const deleteTarget = ref<V1Credential | null>(null)
const selectedRows = ref<V1Credential[]>([])
const bulkDeleteVisible = ref(false)
const bulkDeleting = ref(false)

type CredentialMode = '' | 'clientCertificateRef' | 'token' | 'basic'
const credentialMode = ref<CredentialMode>('')

const form = ref<V1Credential>(createEmptyCredential())

function createEmptyCredential(): V1Credential {
	return {
		version: 'credential/v1',
		meta: { name: '', labels: {} },
		config: { enabled: true },
	}
}

const formLabels = computed({
	get: () => form.value.meta?.labels ?? {},
	set: (val: Record<string, string>) => {
		if (form.value.meta) {
			form.value.meta.labels = val
		}
	},
})

// Infer the credential mode from existing config fields
function inferMode(config: V1Credential['config']): CredentialMode {
	if (!config) return ''
	if (config.clientCertificateRef) return 'clientCertificateRef'
	if (config.token) return 'token'
	if (config.basic) return 'basic'
	return ''
}

// When mode changes, reset config auth fields
watch(credentialMode, (newMode, oldMode) => {
	if (newMode === oldMode) return
	const enabled = form.value.config?.enabled
	switch (newMode) {
		case 'clientCertificateRef':
			form.value.config = { enabled, clientCertificateRef: '' }
			break
		case 'token':
			form.value.config = { enabled, token: '' }
			break
		case 'basic':
			form.value.config = { enabled, basic: { username: '', password: '' } }
			break
		default:
			form.value.config = { enabled }
			break
	}
})

// Form validation
const isFormValid = computed(() => {
	const name = (form.value.meta?.name ?? '').trim()
	if (name.length === 0) return false
	if (!credentialMode.value) return false

	const config = form.value.config
	if (!config) return false

	switch (credentialMode.value) {
		case 'clientCertificateRef':
			return (config.clientCertificateRef ?? '').trim().length > 0
		case 'token':
			return (config.token ?? '').trim().length > 0
		case 'basic':
			return (
				(config.basic?.username ?? '').trim().length > 0 &&
				(config.basic?.password ?? '').trim().length > 0
			)
		default:
			return false
	}
})

// Derive credential type label for the table
function credentialTypeLabel(row: V1Credential): string {
	const config = row.config
	if (!config) return '-'
	if (config.clientCertificateRef) return 'Client Certificate'
	if (config.token) return 'Token'
	if (config.basic) return 'Basic Auth'
	return '-'
}

function sortByCreated(a: any, b: any): number {
	const ta = new Date(a.meta?.created ?? 0).getTime()
	const tb = new Date(b.meta?.created ?? 0).getTime()
	return ta - tb
}

function sortByCredentialType(a: any, b: any): number {
	return credentialTypeLabel(a).localeCompare(credentialTypeLabel(b))
}

function sortByHealthy(a: any, b: any): number {
	const rank = (row: any): number => {
		if (row.status?.healthy === true) return 2
		if (row.status?.healthy === false) return 1
		return 0
	}
	return rank(a) - rank(b)
}

// Selection
function handleSelectionChange(rows: V1Credential[]) {
	selectedRows.value = rows
}

function handleRowClick(row: V1Credential, column: any) {
	if (column?.type === 'selection') return
	router.push(`/credentials/${row.meta?.name}`)
}

function confirmBulkDelete() {
	bulkDeleteVisible.value = true
}

async function handleBulkDelete() {
	bulkDeleting.value = true
	try {
		const { succeeded, failed } = await credentialStore.deleteManyCredentials(selectedRows.value)
		selectedRows.value = []
		if (failed.length === 0) {
			ElMessage.success(`Deleted ${succeeded} credential${succeeded === 1 ? '' : 's'}`)
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

function openCreate() {
	form.value = createEmptyCredential()
	credentialMode.value = ''
	dialogVisible.value = true
}

function confirmDelete(row: V1Credential) {
	deleteTarget.value = row
	deleteDialogVisible.value = true
}

async function handleDelete() {
	if (!deleteTarget.value) return
	try {
		await credentialStore.deleteCredential(deleteTarget.value)
		ElMessage.success('Credential deleted')
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Delete failed')
	}
	deleteTarget.value = null
}

async function handleSave() {
	saving.value = true
	try {
		await credentialStore.createCredential(form.value)
		ElMessage.success('Credential created')
		dialogVisible.value = false
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Save failed')
	} finally {
		saving.value = false
	}
}

async function handleToggleEnabled(row: V1Credential, enabled: boolean) {
	try {
		const updated = structuredClone(toRaw(row))
		if (!updated.config) updated.config = {}
		updated.config.enabled = enabled
		await credentialStore.updateCredential(updated)
		ElMessage.success(`${row.meta?.name} ${enabled ? 'enabled' : 'disabled'}`)
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Update failed')
	}
}

function handleRefresh() {
	credentialStore.fetchCredentials().catch(() => { })
}

onMounted(() => {
	credentialStore.fetchCredentials().catch(() => { })
	certificateStore.fetchCertificates().catch(() => { })
})
</script>

<template>
	<div>
		<el-row justify="space-between" align="middle" style="margin-bottom: 16px">
			<el-col :span="12">
				<h2 style="margin: 0">Credentials</h2>
			</el-col>
			<el-col :span="12" style="text-align: right">
				<el-button :icon="Refresh" @click="handleRefresh">Reload</el-button>
				<el-button type="primary" :icon="Plus" @click="openCreate">New</el-button>
			</el-col>
		</el-row>

		<el-alert v-if="credentialStore.error" :title="credentialStore.error" type="error" show-icon
			style="margin-bottom: 16px" />

		<el-empty v-if="!credentialStore.loading && credentialStore.items.length === 0" description="No credentials yet">
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

			<el-table v-loading="credentialStore.loading" element-loading-text="Loading..." :data="displayItems"
				style="width: 100%" row-key="meta.name" @row-click="handleRowClick" @selection-change="handleSelectionChange"
				:row-class-name="() => 'clickable-row'">
				<el-table-column type="selection" width="48" />
				<el-table-column label="Enabled" width="90">
					<template #default="{ row }">
						<el-switch :model-value="row.config?.enabled ?? true" @update:model-value="handleToggleEnabled(row, $event)"
							@click.stop />
					</template>
				</el-table-column>
				<el-table-column prop="meta.name" label="Name" min-width="200" sortable />
				<el-table-column label="Type" min-width="150" sortable :sort-method="sortByCredentialType">
					<template #default="{ row }">
						<el-tag size="small">{{ credentialTypeLabel(row) }}</el-tag>
					</template>
				</el-table-column>
				<el-table-column label="Healthy" width="100" sortable :sort-method="sortByHealthy">
					<template #default="{ row }">
						<el-tag v-if="row.status?.healthy === true" type="success" size="small">Yes</el-tag>
						<el-tag v-else-if="row.status?.healthy === false" type="danger" size="small">No</el-tag>
						<span v-else>-</span>
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
		<el-dialog v-model="dialogVisible" title="Create Credential" width="600" destroy-on-close>
			<el-form label-width="180px" label-position="right">
				<el-form-item label="Name" required>
					<el-input v-model="form.meta!.name" placeholder="my-credential" />
				</el-form-item>

				<el-form-item label="Labels">
					<LabelEditor v-model="formLabels" />
				</el-form-item>

				<el-divider content-position="left">Config</el-divider>

				<el-form-item label="Enabled">
					<el-switch v-model="form.config!.enabled" />
				</el-form-item>

				<el-form-item label="Credential Type" required>
					<el-select v-model="credentialMode" placeholder="Select credential type" style="width: 100%">
						<el-option label="Client Certificate" value="clientCertificateRef" />
						<el-option label="Token" value="token" />
						<el-option label="Basic Auth" value="basic" />
					</el-select>
				</el-form-item>

				<!-- Client Certificate Ref mode -->
				<el-form-item v-if="credentialMode === 'clientCertificateRef'" label="Client Certificate" required>
					<el-select v-model="form.config!.clientCertificateRef" placeholder="Select a certificate" style="width: 100%"
						filterable clearable :loading="certificateStore.loading">
						<el-option v-for="cert in certificateStore.items" :key="cert.meta?.name" :label="cert.meta?.name"
							:value="cert.meta?.name ?? ''" />
					</el-select>
				</el-form-item>

				<!-- Token mode -->
				<el-form-item v-if="credentialMode === 'token'" label="Token" required>
					<el-input v-model="form.config!.token" type="textarea" :rows="4" placeholder="Bearer token" />
				</el-form-item>

				<!-- Basic Auth mode -->
				<template v-if="credentialMode === 'basic'">
					<el-form-item label="Username" required>
						<el-input v-model="form.config!.basic!.username" placeholder="Username" />
					</el-form-item>
					<el-form-item label="Password" required>
						<el-input v-model="form.config!.basic!.password" type="password" placeholder="Password" show-password />
					</el-form-item>
				</template>
			</el-form>

			<template #footer>
				<el-button @click="dialogVisible = false">Cancel</el-button>
				<el-button type="primary" :loading="saving" :disabled="!isFormValid" @click="handleSave">
					{{ saving ? 'Saving...' : 'Create' }}
				</el-button>
			</template>
		</el-dialog>

		<!-- Delete confirmation -->
		<ConfirmDelete v-model:visible="deleteDialogVisible" :item-name="deleteTarget?.meta?.name ?? ''"
			@confirm="handleDelete" />

		<!-- Bulk delete confirmation -->
		<ConfirmDelete v-model:visible="bulkDeleteVisible"
			:message="`Delete ${selectedRows.length} selected credential${selectedRows.length === 1 ? '' : 's'}?`"
			@confirm="handleBulkDelete" />
	</div>
</template>
