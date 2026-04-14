<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Upload, UploadFilled } from '@element-plus/icons-vue'
import type { UploadFile } from 'element-plus'
import {
  parseKubeconfig,
  getContextNames,
  getDefaultContext,
  defaultImportResourceNames,
  buildImportPlan,
  executeImportPlan,
} from '@/utils/importKubeconfig'
import type { ImportPlan, ImportResult, ImportResourceNames } from '@/utils/importKubeconfig'

const visible = defineModel<boolean>('visible', { required: true })

// --- State ---
type Step = 'upload' | 'configure' | 'results'
const step = ref<Step>('upload')

// Upload state
const rawText = ref('')
const parseError = ref('')

// Config state
const contextNames = ref<string[]>([])
const selectedContext = ref('')
const names = ref<ImportResourceNames>({
  backend: '',
  credential: '',
  certificate: '',
  certificateAuthority: '',
})
const force = ref(false)
const planError = ref('')
const plan = ref<ImportPlan | null>(null)

// Import state
const importing = ref(false)
const results = ref<ImportResult[]>([])

// --- Kubeconfig reference (kept after parse) ---
let parsedKubeconfig: ReturnType<typeof parseKubeconfig> | null = null

// --- Computed ---
const allSucceeded = computed(() => results.value.length > 0 && results.value.every((r) => r.action !== 'failed'))
const failedCount = computed(() => results.value.filter((r) => r.action === 'failed').length)

const canImport = computed(() => {
  return step.value === 'configure' && selectedContext.value && plan.value && !planError.value && !importing.value
})

// --- Watchers ---
// When context changes, regenerate default names and rebuild plan
watch(selectedContext, (ctx) => {
  if (!ctx) {
    plan.value = null
    return
  }
  const defaults = defaultImportResourceNames(ctx)
  names.value = { ...defaults }
  rebuildPlan()
})

// When names change, rebuild plan
watch(names, () => {
  if (selectedContext.value) {
    rebuildPlan()
  }
}, { deep: true })

function rebuildPlan() {
  planError.value = ''
  plan.value = null
  if (!parsedKubeconfig || !selectedContext.value) return

  try {
    plan.value = buildImportPlan(parsedKubeconfig, selectedContext.value, names.value)
  } catch (err) {
    planError.value = err instanceof Error ? err.message : String(err)
  }
}

// --- Handlers ---

function handleFileChange(uploadFile: UploadFile) {
  parseError.value = ''
  const file = uploadFile.raw
  if (!file) return

  file.text().then((text) => {
    processKubeconfigText(text)
  }).catch(() => {
    parseError.value = 'Failed to read file'
  })
}

function handleDrop(event: DragEvent) {
  event.preventDefault()
  const file = event.dataTransfer?.files?.[0]
  if (!file) return

  parseError.value = ''
  file.text().then((text) => {
    processKubeconfigText(text)
  }).catch(() => {
    parseError.value = 'Failed to read file'
  })
}

function processKubeconfigText(text: string) {
  parseError.value = ''
  rawText.value = text

  try {
    parsedKubeconfig = parseKubeconfig(text)
  } catch (err) {
    parseError.value = err instanceof Error ? err.message : 'Failed to parse kubeconfig YAML'
    return
  }

  const ctxNames = getContextNames(parsedKubeconfig)
  if (ctxNames.length === 0) {
    parseError.value = 'No contexts found in kubeconfig'
    return
  }

  contextNames.value = ctxNames
  const defaultCtx = getDefaultContext(parsedKubeconfig)
  selectedContext.value = defaultCtx && ctxNames.includes(defaultCtx) ? defaultCtx : ctxNames[0]!

  step.value = 'configure'
}

async function handleImport() {
  if (!plan.value) return
  importing.value = true
  results.value = []

  try {
    const res = await executeImportPlan(plan.value, force.value)
    results.value = res
    step.value = 'results'

    if (res.every((r) => r.action !== 'failed')) {
      const created = res.filter((r) => r.action === 'created').length
      const updated = res.filter((r) => r.action === 'updated').length
      const parts: string[] = []
      if (created > 0) parts.push(`${created} created`)
      if (updated > 0) parts.push(`${updated} updated`)
      ElMessage.success(`Import complete: ${parts.join(', ')}`)
      visible.value = false
    }
  } catch (err) {
    planError.value = err instanceof Error ? err.message : String(err)
  } finally {
    importing.value = false
  }
}

function handleBack() {
  step.value = 'upload'
  resetUploadState()
}

function handleClose() {
  visible.value = false
}

function handleOpened() {
  step.value = 'upload'
  resetUploadState()
  results.value = []
  planError.value = ''
  importing.value = false
}

function resetUploadState() {
  rawText.value = ''
  parseError.value = ''
  parsedKubeconfig = null
  contextNames.value = []
  selectedContext.value = ''
  names.value = { backend: '', credential: '', certificate: '', certificateAuthority: '' }
  force.value = false
  plan.value = null
}

function resultTagType(action: string): 'success' | 'warning' | 'danger' | 'info' {
  if (action === 'created') return 'success'
  if (action === 'updated') return 'warning'
  if (action === 'skipped') return 'info'
  return 'danger'
}

const authTypeLabel: Record<string, string> = {
  none: 'None',
  token: 'Bearer Token',
  basic: 'Basic Auth',
  'client-certificate': 'Client Certificate (mTLS)',
}
</script>

<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="(val: boolean) => (visible = val)"
    title="Import Kubeconfig"
    width="650px"
    :close-on-click-modal="false"
    @opened="handleOpened"
    destroy-on-close
  >
    <!-- Step 1: Upload -->
    <template v-if="step === 'upload'">
      <el-alert
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 16px"
      >
        <template #title>
          Upload a kubeconfig file to import a Kubernetes context into Multikube.
        </template>
        <div style="margin-top: 4px; font-size: 12px; color: #909399">
          Only kubeconfigs with inline data are supported. For file-based references,
          use <code>multikubectl import</code>.
        </div>
      </el-alert>

      <el-alert
        v-if="parseError"
        type="error"
        :title="parseError"
        show-icon
        :closable="true"
        @close="parseError = ''"
        style="margin-bottom: 16px"
      />

      <div
        class="upload-drop-zone"
        @drop="handleDrop"
        @dragover.prevent
        @dragenter.prevent
      >
        <el-upload
          drag
          :auto-upload="false"
          :show-file-list="false"
          :on-change="handleFileChange"
          accept=".yaml,.yml,.conf,.config,*"
        >
          <el-icon class="el-icon--upload" :size="48" style="color: #909399">
            <UploadFilled />
          </el-icon>
          <div class="el-upload__text">
            Drop kubeconfig file here or <em>click to browse</em>
          </div>
          <template #tip>
            <div class="el-upload__tip">
              Typically <code>~/.kube/config</code>
            </div>
          </template>
        </el-upload>
      </div>
    </template>

    <!-- Step 2: Configure -->
    <template v-if="step === 'configure'">
      <el-alert
        v-if="planError"
        type="error"
        :title="planError"
        show-icon
        :closable="false"
        style="margin-bottom: 16px"
      />

      <el-form label-position="top" size="default">
        <!-- Context selector -->
        <el-form-item label="Context">
          <el-select
            v-model="selectedContext"
            placeholder="Select context"
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="ctx in contextNames"
              :key="ctx"
              :label="ctx"
              :value="ctx"
            />
          </el-select>
        </el-form-item>

        <!-- Auth type (read-only info) -->
        <el-form-item v-if="plan" label="Detected Auth">
          <el-tag type="info" effect="plain">{{ authTypeLabel[plan.authType] ?? plan.authType }}</el-tag>
        </el-form-item>

        <!-- Resource name overrides -->
        <el-divider content-position="left">Resource Names</el-divider>
        <el-form-item label="Backend">
          <el-input v-model="names.backend" placeholder="Backend name" clearable />
        </el-form-item>
        <el-form-item label="Certificate Authority" v-if="plan?.resources.some((r) => r.kind === 'Certificate Authority')">
          <el-input v-model="names.certificateAuthority" placeholder="CA name" clearable />
        </el-form-item>
        <el-form-item label="Certificate" v-if="plan?.resources.some((r) => r.kind === 'Certificate')">
          <el-input v-model="names.certificate" placeholder="Certificate name" clearable />
        </el-form-item>
        <el-form-item label="Credential" v-if="plan?.resources.some((r) => r.kind === 'Credential')">
          <el-input v-model="names.credential" placeholder="Credential name" clearable />
        </el-form-item>

        <!-- Force toggle -->
        <el-divider content-position="left">Options</el-divider>
        <el-form-item>
          <el-checkbox v-model="force">
            Force update existing resources
          </el-checkbox>
        </el-form-item>

        <!-- Preview -->
        <el-divider content-position="left">Preview</el-divider>
        <el-table v-if="plan" :data="plan.resources" size="small" stripe style="width: 100%">
          <el-table-column prop="kind" label="Kind" width="180" />
          <el-table-column prop="name" label="Name" min-width="200" />
        </el-table>
      </el-form>
    </template>

    <!-- Step 3: Results (shown on failure only — success auto-closes) -->
    <template v-if="step === 'results'">
      <el-alert
        v-if="!allSucceeded"
        type="warning"
        show-icon
        :closable="false"
        style="margin-bottom: 12px"
      >
        <template #title>
          {{ failedCount }} of {{ results.length }} resource{{ results.length > 1 ? 's' : '' }} failed to import
        </template>
      </el-alert>

      <el-table :data="results" size="small" stripe style="width: 100%">
        <el-table-column prop="kind" label="Kind" width="180" />
        <el-table-column prop="name" label="Name" min-width="160" />
        <el-table-column label="Result" width="100">
          <template #default="{ row }">
            <el-tag :type="resultTagType(row.action)" size="small" effect="dark">
              {{ row.action }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Error" min-width="200">
          <template #default="{ row }">
            <span v-if="row.error" style="color: #f56c6c; font-size: 12px">{{ row.error }}</span>
            <span v-else style="color: #909399">-</span>
          </template>
        </el-table-column>
      </el-table>
    </template>

    <!-- Footer -->
    <template #footer>
      <div style="display: flex; justify-content: space-between">
        <div>
          <el-button v-if="step === 'configure'" @click="handleBack" text>
            Back
          </el-button>
        </div>
        <div>
          <el-button @click="handleClose">Cancel</el-button>
          <el-button
            v-if="step === 'configure'"
            type="primary"
            :icon="Upload"
            :loading="importing"
            :disabled="!canImport"
            @click="handleImport"
          >
            {{ importing ? 'Importing...' : 'Import' }}
          </el-button>
          <el-button
            v-if="step === 'results' && !allSucceeded"
            type="primary"
            @click="step = 'configure'"
          >
            Back to Configure
          </el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<style scoped>
.upload-drop-zone {
  width: 100%;
}

.upload-drop-zone :deep(.el-upload) {
  width: 100%;
}

.upload-drop-zone :deep(.el-upload-dragger) {
  width: 100%;
}

code {
  background-color: #f0f0f0;
  padding: 1px 4px;
  border-radius: 3px;
  font-size: 12px;
}
</style>
