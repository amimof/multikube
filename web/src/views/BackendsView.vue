<script setup lang="ts">
import { onMounted, ref, computed, toRaw } from 'vue'
import { useRouter } from 'vue-router'
import { Plus, Refresh, Delete, Search, View } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useBackendStore } from '@/stores/backend'
import { useCaStore } from '@/stores/ca'
import { useCredentialStore } from '@/stores/credential'
import { useResourceTable } from '@/composables/useResourceTable'
import moment from 'moment'
import { V1LoadBalancingType } from '@/generated/backend'
import type { V1Backend, V1BackendStatus, V1TargetStatus } from '@/generated/backend'
import LabelEditor from '@/components/LabelEditor.vue'
import MetadataDisplay from '@/components/MetadataDisplay.vue'
import ConfirmDelete from '@/components/ConfirmDelete.vue'

const backendStore = useBackendStore()
const caStore = useCaStore()
const credentialStore = useCredentialStore()
const router = useRouter()

const { nameFilter, displayItems } = useResourceTable(computed(() => backendStore.items))

const dialogVisible = ref(false)
const isEditing = ref(false)
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
      servers: [],
      caRef: '',
      authRef: '',
      insecureSkipTlsVerify: false,
      cacheTtl: '',
      type: V1LoadBalancingType.LoadBalancingTypeRoundRobin,
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

const lbTypeLabels: Record<string, string> = {
  LOAD_BALANCING_TYPE_UNSPECIFIED: 'Unspecified',
  LOAD_BALANCING_TYPE_ROUND_ROBIN: 'Round Robin',
  LOAD_BALANCING_TYPE_LEAST_CONNECTIONS: 'Least Connections',
  LOAD_BALANCING_TYPE_RANDOM: 'Random',
  LOAD_BALANCING_TYPE_WEIGHTED_ROUND_ROBIN: 'Weighted Round Robin',
}

// Form validation
const isFormValid = computed(() => {
  const name = (form.value.meta?.name ?? '').trim()
  const servers = form.value.config?.servers ?? []
  return name.length > 0 && servers.length > 0
})

// Servers as a newline-separated string for textarea editing
const serversText = computed({
  get: () => (form.value.config?.servers ?? []).join('\n'),
  set: (val: string) => {
    if (form.value.config) {
      form.value.config.servers = val.split('\n').filter((s) => s.trim() !== '')
    }
  },
})

// Labels computed for LabelEditor
const formLabels = computed({
  get: () => form.value.meta?.labels ?? {},
  set: (val: Record<string, string>) => {
    if (form.value.meta) {
      form.value.meta.labels = val
    }
  },
})

function countHealthy(status?: V1BackendStatus): number {
  let healthy = 0
  if (status?.targetStatuses) {
    for (const key in status.targetStatuses) {
      if (status.targetStatuses[key]?.phase === 'Healthy') {
        healthy += 1
      }
    }
  }
  return healthy
}

function countTotal(status?: V1BackendStatus): number {
  return Object.keys(status?.targetStatuses ?? {}).length
}

function formatDate(date?: Date): string {
  if (!date) return '-'
  return moment(date).fromNow()
}

// Sort helpers
function sortByCreated(a: any, b: any): number {
  const ta = new Date(a.meta?.created ?? 0).getTime()
  const tb = new Date(b.meta?.created ?? 0).getTime()
  return ta - tb
}

function sortByReady(a: any, b: any): number {
  const ra = countTotal(a.status) === 0 ? -1 : countHealthy(a.status) / countTotal(a.status)
  const rb = countTotal(b.status) === 0 ? -1 : countHealthy(b.status) / countTotal(b.status)
  return ra - rb
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
  openEdit(row)
}

function viewStatus(row: V1Backend) {
  const name = row.meta?.name
  if (name) router.push(`/backends/${name}`)
}

function openCreate() {
  form.value = createEmptyBackend()
  isEditing.value = false
  dialogVisible.value = true
}

function openEdit(row: V1Backend) {
  form.value = structuredClone(toRaw(row))
  isEditing.value = true
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
    if (isEditing.value) {
      await backendStore.updateBackend(form.value)
      ElMessage.success('Backend updated')
    } else {
      await backendStore.createBackend(form.value)
      ElMessage.success('Backend created')
    }
    dialogVisible.value = false
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'Save failed')
  } finally {
    saving.value = false
  }
}

function handleRefresh() {
  backendStore.fetchBackends().catch(() => {})
}

onMounted(() => {
  backendStore.fetchBackends().catch(() => {})
  caStore.fetchCas().catch(() => {})
  credentialStore.fetchCredentials().catch(() => {})
})
</script>

<template>
  <div>
    <el-row justify="space-between" align="middle" style="margin-bottom: 16px">
      <el-col :span="12">
        <h2 style="margin: 0">Backends</h2>
      </el-col>
      <el-col :span="12" style="text-align: right">
        <el-button :icon="Refresh" @click="handleRefresh">Reload</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">Create</el-button>
      </el-col>
    </el-row>

    <el-alert v-if="backendStore.error" :title="backendStore.error" type="error" show-icon style="margin-bottom: 16px" />

    <el-empty v-if="!backendStore.loading && backendStore.items.length === 0" description="No backends yet">
      <el-button type="primary" :icon="Plus" @click="openCreate">Create</el-button>
    </el-empty>

    <template v-else>
      <el-row :gutter="12" align="middle" style="margin-bottom: 12px">
        <el-col :span="12">
          <el-input
            v-model="nameFilter"
            placeholder="Filter by name..."
            clearable
            :prefix-icon="Search"
          />
        </el-col>
        <el-col :span="12" v-if="selectedRows.length > 0">
          <el-button type="danger" :icon="Delete" @click="confirmBulkDelete">
            Delete ({{ selectedRows.length }})
          </el-button>
        </el-col>
      </el-row>

      <el-table
        v-loading="backendStore.loading"
        element-loading-text="Loading..."
        :data="displayItems"
        style="width: 100%"
        row-key="meta.name"
        @row-click="handleRowClick"
        @selection-change="handleSelectionChange"
        :row-class-name="() => 'clickable-row'"
      >
      <el-table-column type="selection" width="48" />
      <el-table-column prop="meta.name" label="Name" min-width="150" sortable />
      <el-table-column label="Ready" width="100" sortable :sort-method="sortByReady">
        <template #default="{ row }">
          <el-tag
            :type="countHealthy(row.status) < countTotal(row.status) ? 'warning' : 'success'"
            effect="dark"
            size="small"
          >
            {{ countHealthy(row.status) }}/{{ countTotal(row.status) }}
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
      <el-table-column label="Actions" width="120" fixed="right">
        <template #default="{ row }">
          <el-button
            :icon="View"
            type="primary"
            size="small"
            plain
            @click.stop="viewStatus(row)"
          />
          <el-button
            :icon="Delete"
            type="danger"
            size="small"
            plain
            @click.stop="confirmDelete(row)"
          />
        </template>
      </el-table-column>
    </el-table>
    </template>

    <!-- Create / Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? 'Edit Backend' : 'Create Backend'"
      width="600"
      destroy-on-close
    >
      <el-form label-width="160px" label-position="right">
        <!-- Metadata section (read-only when editing) -->
        <el-collapse v-if="isEditing" style="margin-bottom: 20px">
          <el-collapse-item title="Metadata" name="metadata">
            <MetadataDisplay :meta="form.meta" />
          </el-collapse-item>
        </el-collapse>

        <el-form-item label="Name" required>
          <el-input v-model="form.meta!.name" :disabled="isEditing" placeholder="my-backend" />
        </el-form-item>

        <el-form-item label="Labels">
          <LabelEditor v-model="formLabels" />
        </el-form-item>

        <el-divider content-position="left">Config</el-divider>

        <el-form-item label="Servers">
          <el-input
            v-model="serversText"
            type="textarea"
            :rows="3"
            placeholder="One server per line (e.g. https://10.0.0.1:6443)"
          />
        </el-form-item>

        <el-form-item label="Load Balancing Type">
          <el-select v-model="form.config!.type" style="width: 100%">
            <el-option
              v-for="opt in lbTypeOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="CA Ref">
          <el-select
            v-model="form.config!.caRef"
            placeholder="Select Certificate Authority"
            style="width: 100%"
            clearable
            filterable
            :loading="caStore.loading"
          >
            <el-option
              v-for="item in caStore.items"
              :key="item.meta?.name"
              :label="item.meta?.name"
              :value="item.meta?.name"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="Auth Ref">
          <el-select
            v-model="form.config!.authRef"
            placeholder="Select Credential"
            style="width: 100%"
            clearable
            filterable
            :loading="credentialStore.loading"
          >
            <el-option
              v-for="item in credentialStore.items"
              :key="item.meta?.name"
              :label="item.meta?.name"
              :value="item.meta?.name"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="Cache TTL">
          <el-input v-model="form.config!.cacheTtl" placeholder="e.g. 30s, 5m" />
        </el-form-item>

        <el-form-item label="Skip TLS Verify">
          <el-switch v-model="form.config!.insecureSkipTlsVerify" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="saving" :disabled="!isFormValid" @click="handleSave">
          {{ saving ? 'Saving...' : isEditing ? 'Update' : 'Create' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- Single delete confirmation -->
    <ConfirmDelete
      v-model:visible="deleteDialogVisible"
      :item-name="deleteTarget?.meta?.name ?? ''"
      @confirm="handleDelete"
    />

    <!-- Bulk delete confirmation -->
    <ConfirmDelete
      v-model:visible="bulkDeleteVisible"
      :message="`Delete ${selectedRows.length} selected backend${selectedRows.length === 1 ? '' : 's'}?`"
      @confirm="handleBulkDelete"
    />
  </div>
</template>
