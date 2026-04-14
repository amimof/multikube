<script setup lang="ts">
import { onMounted, ref, computed, watch, toRaw } from 'vue'
import { Plus, Refresh, Delete } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useCredentialStore } from '@/stores/credential'
import { useCertificateStore } from '@/stores/certificate'
import type { V1Credential } from '@/generated/credential'
import LabelEditor from '@/components/LabelEditor.vue'
import MetadataDisplay from '@/components/MetadataDisplay.vue'
import ConfirmDelete from '@/components/ConfirmDelete.vue'

const credentialStore = useCredentialStore()
const certificateStore = useCertificateStore()

const dialogVisible = ref(false)
const isEditing = ref(false)
const saving = ref(false)
const deleteDialogVisible = ref(false)
const deleteTarget = ref<V1Credential | null>(null)

type CredentialMode = '' | 'clientCertificateRef' | 'token' | 'basic'
const credentialMode = ref<CredentialMode>('')

const form = ref<V1Credential>(createEmptyCredential())

function createEmptyCredential(): V1Credential {
  return {
    version: 'credential/v1',
    meta: { name: '', labels: {} },
    config: { name: '' },
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

// When mode changes, reset config auth fields while preserving config.name
watch(credentialMode, (newMode, oldMode) => {
  if (newMode === oldMode) return
  const configName = form.value.config?.name ?? ''
  switch (newMode) {
    case 'clientCertificateRef':
      form.value.config = { name: configName, clientCertificateRef: '' }
      break
    case 'token':
      form.value.config = { name: configName, token: '' }
      break
    case 'basic':
      form.value.config = { name: configName, basic: { username: '', password: '' } }
      break
    default:
      form.value.config = { name: configName }
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

function formatDate(date?: Date): string {
  if (!date) return '-'
  return new Date(date).toLocaleString()
}

function openCreate() {
  form.value = createEmptyCredential()
  credentialMode.value = ''
  isEditing.value = false
  dialogVisible.value = true
}

function openEdit(row: V1Credential) {
  form.value = structuredClone(toRaw(row))
  if (!form.value.config) form.value.config = {}
  credentialMode.value = inferMode(form.value.config)
  isEditing.value = true
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
    if (isEditing.value) {
      await credentialStore.updateCredential(form.value)
      ElMessage.success('Credential updated')
    } else {
      await credentialStore.createCredential(form.value)
      ElMessage.success('Credential created')
    }
    dialogVisible.value = false
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'Save failed')
  } finally {
    saving.value = false
  }
}

function handleRefresh() {
  credentialStore.fetchCredentials().catch(() => {})
}

onMounted(() => {
  credentialStore.fetchCredentials().catch(() => {})
  certificateStore.fetchCertificates().catch(() => {})
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
        <el-button type="primary" :icon="Plus" @click="openCreate">Create</el-button>
      </el-col>
    </el-row>

    <el-alert v-if="credentialStore.error" :title="credentialStore.error" type="error" show-icon style="margin-bottom: 16px" />

    <el-empty v-if="!credentialStore.loading && credentialStore.items.length === 0" description="No credentials yet">
      <el-button type="primary" :icon="Plus" @click="openCreate">Create</el-button>
    </el-empty>

    <el-table
      v-else
      v-loading="credentialStore.loading"
      :data="credentialStore.items"
      style="width: 100%"
      @row-click="openEdit"
      :row-class-name="() => 'clickable-row'"
    >
      <el-table-column prop="meta.name" label="Name" min-width="200" />
      <el-table-column label="Config Name" min-width="160">
        <template #default="{ row }">
          {{ row.config?.name || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="Type" min-width="150">
        <template #default="{ row }">
          <el-tag size="small">{{ credentialTypeLabel(row) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="Healthy" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.status?.healthy === true" type="success" size="small">Yes</el-tag>
          <el-tag v-else-if="row.status?.healthy === false" type="danger" size="small">No</el-tag>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="Created" width="180">
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

    <!-- Create / Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? 'Edit Credential' : 'Create Credential'"
      width="600"
      destroy-on-close
    >
      <el-form label-width="180px" label-position="right">
        <el-collapse v-if="isEditing" style="margin-bottom: 20px">
          <el-collapse-item title="Metadata" name="metadata">
            <MetadataDisplay :meta="form.meta" />
          </el-collapse-item>
        </el-collapse>

        <el-form-item label="Name" required>
          <el-input v-model="form.meta!.name" :disabled="isEditing" placeholder="my-credential" />
        </el-form-item>

        <el-form-item label="Labels">
          <LabelEditor v-model="formLabels" />
        </el-form-item>

        <el-divider content-position="left">Config</el-divider>

        <el-form-item label="Config Name">
          <el-input v-model="form.config!.name" placeholder="Config name" />
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
          <el-select
            v-model="form.config!.clientCertificateRef"
            placeholder="Select a certificate"
            style="width: 100%"
            filterable
            clearable
            :loading="certificateStore.loading"
          >
            <el-option
              v-for="cert in certificateStore.items"
              :key="cert.meta?.name"
              :label="cert.meta?.name"
              :value="cert.meta?.name ?? ''"
            />
          </el-select>
        </el-form-item>

        <!-- Token mode -->
        <el-form-item v-if="credentialMode === 'token'" label="Token" required>
          <el-input
            v-model="form.config!.token"
            type="textarea"
            :rows="4"
            placeholder="Bearer token"
          />
        </el-form-item>

        <!-- Basic Auth mode -->
        <template v-if="credentialMode === 'basic'">
          <el-form-item label="Username" required>
            <el-input v-model="form.config!.basic!.username" placeholder="Username" />
          </el-form-item>
          <el-form-item label="Password" required>
            <el-input
              v-model="form.config!.basic!.password"
              type="password"
              placeholder="Password"
              show-password
            />
          </el-form-item>
        </template>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="saving" :disabled="!isFormValid" @click="handleSave">
          {{ isEditing ? 'Update' : 'Create' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- Delete confirmation -->
    <ConfirmDelete
      v-model:visible="deleteDialogVisible"
      :item-name="deleteTarget?.meta?.name ?? ''"
      @confirm="handleDelete"
    />
  </div>
</template>
