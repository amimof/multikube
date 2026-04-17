<script setup lang="ts">
import { onMounted, ref, computed, toRaw } from 'vue'
import { useRouter } from 'vue-router'
import { Plus, Refresh, Delete, Search, EditPen } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useCertificateStore } from '@/stores/certificate'
import { useResourceTable } from '@/composables/useResourceTable'
import { formatDate } from '@/utils/format'
import type { V1Certificate } from '@/generated/certificate'
import LabelEditor from '@/components/LabelEditor.vue'
import MetadataDisplay from '@/components/MetadataDisplay.vue'
import ConfirmDelete from '@/components/ConfirmDelete.vue'

const certificateStore = useCertificateStore()
const router = useRouter()

const { nameFilter, displayItems } = useResourceTable(computed(() => certificateStore.items))

const dialogVisible = ref(false)
const isEditing = ref(false)
const saving = ref(false)
const deleteDialogVisible = ref(false)
const deleteTarget = ref<V1Certificate | null>(null)
const selectedRows = ref<V1Certificate[]>([])
const bulkDeleteVisible = ref(false)
const bulkDeleting = ref(false)

const form = ref<V1Certificate>(createEmptyCertificate())

function createEmptyCertificate(): V1Certificate {
	return {
		version: 'certificate/v1',
		meta: { name: '', labels: {} },
		config: {},
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

// Form validation
const isFormValid = computed(() => {
	const name = (form.value.meta?.name ?? '').trim()
	if (name.length === 0) return false

	const config = form.value.config
	if (!config) return false

	if (!(config.certificateData ?? '').trim()) return false
	if (!(config.keyData ?? '').trim()) return false

	return true
})

// Table helpers
function sortByCreated(a: any, b: any): number {
	const ta = new Date(a.meta?.created ?? 0).getTime()
	const tb = new Date(b.meta?.created ?? 0).getTime()
	return ta - tb
}

function sortByCertificate(a: any, b: any): number {
	const va = (a.config?.certificateData ?? '').substring(0, 40)
	const vb = (b.config?.certificateData ?? '').substring(0, 40)
	return va.localeCompare(vb)
}

// Selection
function handleSelectionChange(rows: V1Certificate[]) {
	selectedRows.value = rows
}

function handleRowClick(row: V1Certificate, column: any) {
	if (column?.type === 'selection') return
	router.push(`/certificates/${row.meta?.name}`)
}

function confirmBulkDelete() {
	bulkDeleteVisible.value = true
}

async function handleBulkDelete() {
	bulkDeleting.value = true
	try {
		const { succeeded, failed } = await certificateStore.deleteManyCertificates(selectedRows.value)
		selectedRows.value = []
		if (failed.length === 0) {
			ElMessage.success(`Deleted ${succeeded} certificate${succeeded === 1 ? '' : 's'}`)
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
	form.value = createEmptyCertificate()
	isEditing.value = false
	dialogVisible.value = true
}

function openEdit(row: V1Certificate) {
	form.value = structuredClone(toRaw(row))
	if (!form.value.config) form.value.config = {}
	isEditing.value = true
	dialogVisible.value = true
}

function confirmDelete(row: V1Certificate) {
	deleteTarget.value = row
	deleteDialogVisible.value = true
}

async function handleDelete() {
	if (!deleteTarget.value) return
	try {
		await certificateStore.deleteCertificate(deleteTarget.value)
		ElMessage.success('Certificate deleted')
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Delete failed')
	}
	deleteTarget.value = null
}

async function handleSave() {
	saving.value = true
	try {
		if (isEditing.value) {
			await certificateStore.updateCertificate(form.value)
			ElMessage.success('Certificate updated')
		} else {
			await certificateStore.createCertificate(form.value)
			ElMessage.success('Certificate created')
		}
		dialogVisible.value = false
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Save failed')
	} finally {
		saving.value = false
	}
}

function handleRefresh() {
	certificateStore.fetchCertificates().catch(() => { })
}

onMounted(() => {
	certificateStore.fetchCertificates().catch(() => { })
})
</script>

<template>
	<div>
		<el-row justify="space-between" align="middle" style="margin-bottom: 16px">
			<el-col :span="12">
				<h2 style="margin: 0">Certificates</h2>
			</el-col>
			<el-col :span="12" style="text-align: right">
				<el-button :icon="Refresh" @click="handleRefresh">Reload</el-button>
				<el-button type="primary" :icon="Plus" @click="openCreate">New</el-button>
			</el-col>
		</el-row>

		<el-alert v-if="certificateStore.error" :title="certificateStore.error" type="error" show-icon
			style="margin-bottom: 16px" />

		<el-empty v-if="!certificateStore.loading && certificateStore.items.length === 0" description="No certificates yet">
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

			<el-table v-loading="certificateStore.loading" element-loading-text="Loading..." :data="displayItems"
				style="width: 100%" row-key="meta.name" @row-click="handleRowClick" @selection-change="handleSelectionChange"
				:row-class-name="() => 'clickable-row'">
				<el-table-column type="selection" width="48" />
				<el-table-column prop="meta.name" label="Name" min-width="180" sortable />
				<el-table-column label="Certificate" min-width="200" sortable :sort-method="sortByCertificate">
					<template #default="{ row }">
						<span v-if="row.config?.certificateData" style="font-family: monospace; font-size: 12px">
							{{ row.config.certificateData.substring(0, 40) }}...
						</span>
						<span v-else>-</span>
					</template>
				</el-table-column>
				<el-table-column label="Key" min-width="120">
					<template #default="{ row }">
						<span v-if="row.config?.keyData">********</span>
						<span v-else>-</span>
					</template>
				</el-table-column>
				<el-table-column label="Created" width="180" sortable :sort-method="sortByCreated">
					<template #default="{ row }">
						{{ formatDate(row.meta?.created) }}
					</template>
				</el-table-column>
				<el-table-column label="Actions" width="120" fixed="right">
					<template #default="{ row }">
						<el-button :icon="EditPen" type="primary" size="small" plain @click.stop="openEdit(row)" />
						<el-button :icon="Delete" type="danger" size="small" plain @click.stop="confirmDelete(row)" />
					</template>
				</el-table-column>
			</el-table>
		</template>

		<!-- Create / Edit Dialog -->
		<el-dialog v-model="dialogVisible" :title="isEditing ? 'Edit Certificate' : 'Create Certificate'" width="700"
			destroy-on-close>
			<el-form label-width="120px" label-position="right">
				<el-collapse v-if="isEditing" style="margin-bottom: 20px">
					<el-collapse-item title="Metadata" name="metadata">
						<MetadataDisplay :meta="form.meta" />
					</el-collapse-item>
				</el-collapse>

				<el-form-item label="Name" required>
					<el-input v-model="form.meta!.name" :disabled="isEditing" placeholder="my-certificate" />
				</el-form-item>

				<el-form-item label="Labels">
					<LabelEditor v-model="formLabels" />
				</el-form-item>

				<el-divider content-position="left">Certificate</el-divider>

				<el-form-item label="Data" required>
					<el-input v-model="form.config!.certificateData" type="textarea" :rows="8"
						:input-style="{ fontFamily: 'monospace', fontSize: '13px' }"
						placeholder="Paste PEM certificate data here" />
				</el-form-item>

				<el-divider content-position="left">Private Key</el-divider>

				<el-form-item label="Data" required>
					<el-input v-model="form.config!.keyData" type="textarea" :rows="8"
						:input-style="{ fontFamily: 'monospace', fontSize: '13px' }"
						placeholder="Paste PEM private key data here" />
				</el-form-item>
			</el-form>

			<template #footer>
				<el-button @click="dialogVisible = false">Cancel</el-button>
				<el-button type="primary" :loading="saving" :disabled="!isFormValid" @click="handleSave">
					{{ saving ? 'Saving...' : isEditing ? 'Update' : 'Create' }}
				</el-button>
			</template>
		</el-dialog>

		<!-- Delete confirmation -->
		<ConfirmDelete v-model:visible="deleteDialogVisible" :item-name="deleteTarget?.meta?.name ?? ''"
			@confirm="handleDelete" />

		<!-- Bulk delete confirmation -->
		<ConfirmDelete v-model:visible="bulkDeleteVisible"
			:message="`Delete ${selectedRows.length} selected certificate${selectedRows.length === 1 ? '' : 's'}?`"
			@confirm="handleBulkDelete" />
	</div>
</template>
