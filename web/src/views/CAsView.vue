<script setup lang="ts">
import { onMounted, ref, computed, toRaw } from 'vue'
import { useRouter } from 'vue-router'
import { Plus, Refresh, Delete, Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useCaStore } from '@/stores/ca'
import { useResourceTable } from '@/composables/useResourceTable'
import { formatDate } from '@/utils/format'
import type { V1CertificateAuthority } from '@/generated/ca'
import LabelEditor from '@/components/LabelEditor.vue'
import ConfirmDelete from '@/components/ConfirmDelete.vue'

const caStore = useCaStore()
const router = useRouter()

const { nameFilter, displayItems } = useResourceTable(computed(() => caStore.items))

const dialogVisible = ref(false)
const saving = ref(false)
const deleteDialogVisible = ref(false)
const deleteTarget = ref<V1CertificateAuthority | null>(null)
const selectedRows = ref<V1CertificateAuthority[]>([])
const bulkDeleteVisible = ref(false)
const bulkDeleting = ref(false)

const form = ref<V1CertificateAuthority>(createEmptyCa())

function createEmptyCa(): V1CertificateAuthority {
	return {
		version: 'certificate_authority/v1',
		meta: { name: '', labels: {} },
		config: {
			enabled: true,
			certificateData: '',
		},
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

// Form validation: name and certificateData required
const isFormValid = computed(() => {
	const name = (form.value.meta?.name ?? '').trim()
	const certData = (form.value.config?.certificateData ?? '').trim()
	return name.length > 0 && certData.length > 0
})

function sortByCreated(a: any, b: any): number {
	const ta = new Date(a.meta?.created ?? 0).getTime()
	const tb = new Date(b.meta?.created ?? 0).getTime()
	return ta - tb
}

function sortByCertificate(a: any, b: any): number {
	const la = a.config?.certificateData ? 'Yes' : ''
	const lb = b.config?.certificateData ? 'Yes' : ''
	return la.localeCompare(lb)
}

// Selection
function handleSelectionChange(rows: V1CertificateAuthority[]) {
	selectedRows.value = rows
}

function handleRowClick(row: V1CertificateAuthority, column: any) {
	if (column?.type === 'selection') return
	router.push(`/cas/${row.meta?.name}`)
}

function confirmBulkDelete() {
	bulkDeleteVisible.value = true
}

async function handleBulkDelete() {
	bulkDeleting.value = true
	try {
		const { succeeded, failed } = await caStore.deleteManyCas(selectedRows.value)
		selectedRows.value = []
		if (failed.length === 0) {
			ElMessage.success(`Deleted ${succeeded} certificate authorit${succeeded === 1 ? 'y' : 'ies'}`)
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
	form.value = createEmptyCa()
	dialogVisible.value = true
}

function confirmDelete(row: V1CertificateAuthority) {
	deleteTarget.value = row
	deleteDialogVisible.value = true
}

async function handleDelete() {
	if (!deleteTarget.value) return
	try {
		await caStore.deleteCa(deleteTarget.value)
		ElMessage.success('Certificate Authority deleted')
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Delete failed')
	}
	deleteTarget.value = null
}

async function handleSave() {
	saving.value = true
	try {
		await caStore.createCa(form.value)
		ElMessage.success('Certificate Authority created')
		dialogVisible.value = false
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Save failed')
	} finally {
		saving.value = false
	}
}

async function handleToggleEnabled(row: V1CertificateAuthority, enabled: boolean) {
	try {
		const updated = structuredClone(toRaw(row))
		if (!updated.config) updated.config = {}
		updated.config.enabled = enabled
		await caStore.updateCa(updated)
		ElMessage.success(`${row.meta?.name} ${enabled ? 'enabled' : 'disabled'}`)
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Update failed')
	}
}

function handleRefresh() {
	caStore.fetchCas().catch(() => { })
}

onMounted(() => {
	caStore.fetchCas().catch(() => { })
})
</script>

<template>
	<div>
		<el-row justify="space-between" align="middle" style="margin-bottom: 16px">
			<el-col :span="12">
				<h2 style="margin: 0">Certificate Authorities</h2>
			</el-col>
			<el-col :span="12" style="text-align: right">
				<el-button :icon="Refresh" @click="handleRefresh">Reload</el-button>
				<el-button type="primary" :icon="Plus" @click="openCreate">New</el-button>
			</el-col>
		</el-row>

		<el-alert v-if="caStore.error" :title="caStore.error" type="error" show-icon style="margin-bottom: 16px" />

		<el-empty v-if="!caStore.loading && caStore.items.length === 0" description="No certificate authorities yet">
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

			<el-table v-loading="caStore.loading" element-loading-text="Loading..." :data="displayItems" style="width: 100%"
				row-key="meta.name" @row-click="handleRowClick" @selection-change="handleSelectionChange"
				:row-class-name="() => 'clickable-row'">
				<el-table-column type="selection" width="48" />
				<el-table-column label="Enabled" width="90">
					<template #default="{ row }">
						<el-switch :model-value="row.config?.enabled ?? true" @update:model-value="handleToggleEnabled(row, $event)"
							@click.stop />
					</template>
				</el-table-column>
				<el-table-column prop="meta.name" label="Name" min-width="200" sortable />
				<el-table-column label="Certificate Data" min-width="200" sortable :sort-method="sortByCertificate">
					<template #default="{ row }">
						<span v-if="row.config?.certificateData" style="font-family: monospace; font-size: 12px">
							{{ row.config.certificateData.length > 40 ? row.config.certificateData.substring(0, 40) + '...' :
								row.config.certificateData }}
						</span>
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
		<el-dialog v-model="dialogVisible" title="Create Certificate Authority" width="700" destroy-on-close>
			<el-form label-width="120px" label-position="right">
				<el-form-item label="Name" required>
					<el-input v-model="form.meta!.name" placeholder="my-ca" />
				</el-form-item>

				<el-form-item label="Labels">
					<LabelEditor v-model="formLabels" />
				</el-form-item>

				<el-divider content-position="left">Config</el-divider>

				<el-form-item label="Enabled">
					<el-switch v-model="form.config!.enabled" />
				</el-form-item>

				<el-form-item label="Data" required>
					<el-input v-model="form.config!.certificateData" type="textarea" :rows="8"
						:input-style="{ fontFamily: 'monospace', fontSize: '13px' }"
						placeholder="Paste PEM certificate data here" />
				</el-form-item>
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
			:message="`Delete ${selectedRows.length} selected certificate authorit${selectedRows.length === 1 ? 'y' : 'ies'}?`"
			@confirm="handleBulkDelete" />
	</div>
</template>
