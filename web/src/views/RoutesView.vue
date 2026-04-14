<script setup lang="ts">
import { onMounted, ref, computed, watch, toRaw } from 'vue'
import { Plus, Refresh, Delete, Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useRouteStore } from '@/stores/route'
import { useBackendStore } from '@/stores/backend'
import { useResourceTable } from '@/composables/useResourceTable'
import moment from 'moment'
import type { V1Route, V1Match } from '@/generated/route'
import LabelEditor from '@/components/LabelEditor.vue'
import MetadataDisplay from '@/components/MetadataDisplay.vue'
import ConfirmDelete from '@/components/ConfirmDelete.vue'

type RouteMatchMode = '' | 'sni' | 'path' | 'pathPrefix' | 'header' | 'jwt'

const routeStore = useRouteStore()
const backendStore = useBackendStore()

const { nameFilter, displayItems } = useResourceTable(computed(() => routeStore.items))

const dialogVisible = ref(false)
const isEditing = ref(false)
const saving = ref(false)
const deleteDialogVisible = ref(false)
const deleteTarget = ref<V1Route | null>(null)
const selectedRows = ref<V1Route[]>([])
const bulkDeleteVisible = ref(false)
const bulkDeleting = ref(false)

const matchMode = ref<RouteMatchMode>('')
const form = ref<V1Route>(createEmptyRoute())

function createEmptyRoute(): V1Route {
  return {
    version: 'route/v1',
    meta: { name: '', labels: {} },
    config: {
      name: '',
      backendRef: '',
      match: undefined,
    },
  }
}

const matchModeOptions: { label: string; value: RouteMatchMode }[] = [
  { label: 'SNI', value: 'sni' },
  { label: 'Path (exact)', value: 'path' },
  { label: 'Path Prefix', value: 'pathPrefix' },
  { label: 'Header', value: 'header' },
  { label: 'JWT', value: 'jwt' },
]

// When matchMode changes, replace the match object with a clean one for that mode only
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

// Infer match mode from an existing match object
function inferMatchMode(match?: V1Match): RouteMatchMode {
  if (!match) return ''
  if (match.sni) return 'sni'
  if (match.path) return 'path'
  if (match.pathPrefix) return 'pathPrefix'
  if (match.header?.name || match.header?.value) return 'header'
  if (match.jwt?.claim || match.jwt?.value) return 'jwt'
  return ''
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
  const backendRef = (form.value.config?.backendRef ?? '').trim()
  if (name.length === 0 || backendRef.length === 0) return false
  if (!matchMode.value) return false

  const match = form.value.config?.match
  if (!match) return false

  switch (matchMode.value) {
    case 'sni':
      return (match.sni ?? '').trim().length > 0
    case 'path':
      return (match.path ?? '').trim().length > 0
    case 'pathPrefix':
      return (match.pathPrefix ?? '').trim().length > 0
    case 'header':
      return (match.header?.name ?? '').trim().length > 0 && (match.header?.value ?? '').trim().length > 0
    case 'jwt':
      return (match.jwt?.claim ?? '').trim().length > 0 && (match.jwt?.value ?? '').trim().length > 0
    default:
      return false
  }
})

function formatDate(date?: Date): string {
  if (!date) return '-'
  return moment(date).fromNow()
}

function sortByCreated(a: any, b: any): number {
  const ta = new Date(a.meta?.created ?? 0).getTime()
  const tb = new Date(b.meta?.created ?? 0).getTime()
  return ta - tb
}

function sortByStatus(a: any, b: any): number {
  const sa = a.status?.phase ?? ''
  const sb = b.status?.phase ?? ''
  return sa.localeCompare(sb)
}

function sortByBackendRef(a: any, b: any): number {
  const ra = a.config?.backendRef ?? ''
  const rb = b.config?.backendRef ?? ''
  return ra.localeCompare(rb)
}

function sortByMatch(a: any, b: any): number {
  return describeMatch(a.config?.match).localeCompare(describeMatch(b.config?.match))
}

// Selection
function handleSelectionChange(rows: V1Route[]) {
  selectedRows.value = rows
}

function handleRowClick(row: V1Route, column: any) {
  if (column?.type === 'selection') return
  openEdit(row)
}

function confirmBulkDelete() {
  bulkDeleteVisible.value = true
}

async function handleBulkDelete() {
  bulkDeleting.value = true
  try {
    const { succeeded, failed } = await routeStore.deleteManyRoutes(selectedRows.value)
    selectedRows.value = []
    if (failed.length === 0) {
      ElMessage.success(`Deleted ${succeeded} route${succeeded === 1 ? '' : 's'}`)
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
  form.value = createEmptyRoute()
  matchMode.value = ''
  isEditing.value = false
  dialogVisible.value = true
}

function openEdit(row: V1Route) {
  const raw = structuredClone(toRaw(row))
  if (!raw.config) raw.config = {}
  // Infer match mode BEFORE setting form to avoid watch overwriting match data
  const inferred = inferMatchMode(raw.config.match)
  form.value = raw
  // Set matchMode without triggering the watcher to overwrite existing data.
  // We temporarily remove the watch effect by setting the value after form is assigned.
  // Since the watch triggers on matchMode change, and we need the existing match data
  // preserved, we set matchMode only if it differs from current (which it always will
  // on a fresh open). To avoid the watcher clearing the data, we set form first then mode.
  // Actually the watcher WILL fire and overwrite. So we need a guard.
  matchMode.value = inferred
  // Restore the actual match data from the resource after watcher ran
  if (raw.config.match) {
    form.value.config!.match = structuredClone(raw.config.match)
  }
  isEditing.value = true
  dialogVisible.value = true
}

function confirmDelete(row: V1Route) {
  deleteTarget.value = row
  deleteDialogVisible.value = true
}

async function handleDelete() {
  if (!deleteTarget.value) return
  try {
    await routeStore.deleteRoute(deleteTarget.value)
    ElMessage.success('Route deleted')
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'Delete failed')
  }
  deleteTarget.value = null
}

async function handleSave() {
  saving.value = true
  try {
    if (isEditing.value) {
      await routeStore.updateRoute(form.value)
      ElMessage.success('Route updated')
    } else {
      await routeStore.createRoute(form.value)
      ElMessage.success('Route created')
    }
    dialogVisible.value = false
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'Save failed')
  } finally {
    saving.value = false
  }
}

function handleRefresh() {
  routeStore.fetchRoutes().catch(() => {})
}

// Describe the active match for the table column
function describeMatch(match?: V1Match): string {
  if (!match) return '-'
  if (match.sni) return `SNI: ${match.sni}`
  if (match.path) return `Path: ${match.path}`
  if (match.pathPrefix) return `Prefix: ${match.pathPrefix}`
  if (match.header?.name) return `Header: ${match.header.name}=${match.header.value}`
  if (match.jwt?.claim) return `JWT: ${match.jwt.claim}=${match.jwt.value}`
  return '-'
}

onMounted(() => {
  routeStore.fetchRoutes().catch(() => {})
  backendStore.fetchBackends().catch(() => {})
})
</script>

<template>
  <div>
    <el-row justify="space-between" align="middle" style="margin-bottom: 16px">
      <el-col :span="12">
        <h2 style="margin: 0">Routes</h2>
      </el-col>
      <el-col :span="12" style="text-align: right">
        <el-button :icon="Refresh" @click="handleRefresh">Reload</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">Create</el-button>
      </el-col>
    </el-row>

    <el-alert v-if="routeStore.error" :title="routeStore.error" type="error" show-icon style="margin-bottom: 16px" />

    <el-empty v-if="!routeStore.loading && routeStore.items.length === 0" description="No routes yet">
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
        v-loading="routeStore.loading"
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
      <el-table-column label="Status" width="100" sortable :sort-method="sortByStatus">
        <template #default="{ row }">
          <el-tag
            v-if="row.status?.phase"
            :type="row.status.phase === 'Active' ? 'success' : row.status.phase === 'Inactive' ? 'info' : 'warning'"
            effect="dark"
            size="small"
          >
            {{ row.status.phase }}
          </el-tag>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="Backend Ref" min-width="150" sortable :sort-method="sortByBackendRef">
        <template #default="{ row }">
          {{ row.config?.backendRef || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="Match" min-width="200" sortable :sort-method="sortByMatch">
        <template #default="{ row }">
          {{ describeMatch(row.config?.match) }}
        </template>
      </el-table-column>
      <el-table-column label="Created" width="180" sortable :sort-method="sortByCreated">
        <template #default="{ row }">
          {{ formatDate(row.meta?.created) }}
        </template>
      </el-table-column>
      <el-table-column label="Actions" width="80" fixed="right">
        <template #default="{ row }">
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
      :title="isEditing ? 'Edit Route' : 'Create Route'"
      width="600"
      destroy-on-close
    >
      <el-form label-width="140px" label-position="right">
        <el-collapse v-if="isEditing" style="margin-bottom: 20px">
          <el-collapse-item title="Metadata" name="metadata">
            <MetadataDisplay :meta="form.meta" />
          </el-collapse-item>
        </el-collapse>

        <el-form-item label="Name" required>
          <el-input v-model="form.meta!.name" :disabled="isEditing" placeholder="my-route" />
        </el-form-item>

        <el-form-item label="Labels">
          <LabelEditor v-model="formLabels" />
        </el-form-item>

        <el-divider content-position="left">Config</el-divider>

        <el-form-item label="Config Name">
          <el-input v-model="form.config!.name" placeholder="Config name" />
        </el-form-item>

        <el-form-item label="Backend Ref" required>
          <el-select
            v-model="form.config!.backendRef"
            placeholder="Select Backend"
            style="width: 100%"
            clearable
            filterable
            :loading="backendStore.loading"
          >
            <el-option
              v-for="item in backendStore.items"
              :key="item.meta?.name"
              :label="item.meta?.name"
              :value="item.meta?.name"
            />
          </el-select>
        </el-form-item>

        <el-divider content-position="left">Match</el-divider>

        <el-form-item label="Match Type" required>
          <el-select v-model="matchMode" placeholder="Select match type" style="width: 100%">
            <el-option
              v-for="opt in matchModeOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>

        <!-- SNI -->
        <el-form-item v-if="matchMode === 'sni'" label="SNI" required>
          <el-input v-model="form.config!.match!.sni" placeholder="example.com" />
        </el-form-item>

        <!-- Path (exact) -->
        <el-form-item v-if="matchMode === 'path'" label="Path" required>
          <el-input v-model="form.config!.match!.path" placeholder="/exact/path" />
        </el-form-item>

        <!-- Path Prefix -->
        <el-form-item v-if="matchMode === 'pathPrefix'" label="Path Prefix" required>
          <el-input v-model="form.config!.match!.pathPrefix" placeholder="/api/" />
        </el-form-item>

        <!-- Header -->
        <template v-if="matchMode === 'header'">
          <el-form-item label="Header Name" required>
            <el-input v-model="form.config!.match!.header!.name" placeholder="X-Custom-Header" />
          </el-form-item>
          <el-form-item label="Header Value" required>
            <el-input v-model="form.config!.match!.header!.value" placeholder="expected-value" />
          </el-form-item>
        </template>

        <!-- JWT -->
        <template v-if="matchMode === 'jwt'">
          <el-form-item label="JWT Claim" required>
            <el-input v-model="form.config!.match!.jwt!.claim" placeholder="sub" />
          </el-form-item>
          <el-form-item label="JWT Value" required>
            <el-input v-model="form.config!.match!.jwt!.value" placeholder="expected-value" />
          </el-form-item>
        </template>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="saving" :disabled="!isFormValid" @click="handleSave">
          {{ saving ? 'Saving...' : isEditing ? 'Update' : 'Create' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- Delete confirmation -->
    <ConfirmDelete
      v-model:visible="deleteDialogVisible"
      :item-name="deleteTarget?.meta?.name ?? ''"
      @confirm="handleDelete"
    />

    <!-- Bulk delete confirmation -->
    <ConfirmDelete
      v-model:visible="bulkDeleteVisible"
      :message="`Delete ${selectedRows.length} selected route${selectedRows.length === 1 ? '' : 's'}?`"
      @confirm="handleBulkDelete"
    />
  </div>
</template>
