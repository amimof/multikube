<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Codemirror } from 'vue-codemirror'
import { yaml as yamlLang } from '@codemirror/lang-yaml'
import { oneDark } from '@codemirror/theme-one-dark'
import { applyResources, SUPPORTED_VERSIONS } from '@/utils/applyResources'
import type { ApplyResult } from '@/utils/applyResources'

const visible = defineModel<boolean>('visible', { required: true })

const DEFAULT_YAML = `# Paste one or more YAML documents separated by ---
# Example:
# version: backend/v1
# meta:
#   name: my-backend
# config:
#   servers:
#     - https://10.0.0.1:6443
`

const yamlText = ref(DEFAULT_YAML)

const applying = ref(false)
const results = ref<ApplyResult[]>([])
const parseError = ref('')

const cmExtensions = [yamlLang(), oneDark]

const hasResults = computed(() => results.value.length > 0)
const allSucceeded = computed(() => hasResults.value && results.value.every((r) => r.action !== 'failed'))
const failedCount = computed(() => results.value.filter((r) => r.action === 'failed').length)
const createdCount = computed(() => results.value.filter((r) => r.action === 'created').length)
const updatedCount = computed(() => results.value.filter((r) => r.action === 'updated').length)

const canApply = computed(() => yamlText.value.trim().length > 0 && !applying.value)

function actionTagType(action: string): 'success' | 'warning' | 'danger' {
  if (action === 'created') return 'success'
  if (action === 'updated') return 'warning'
  return 'danger'
}

function resetState() {
  results.value = []
  parseError.value = ''
}

async function handleApply() {
  resetState()
  applying.value = true

  try {
    const res = await applyResources(yamlText.value)
    results.value = res

    if (res.every((r) => r.action !== 'failed')) {
      // All succeeded — show toast and auto-close
      const parts: string[] = []
      if (createdCount.value > 0) parts.push(`${createdCount.value} created`)
      if (updatedCount.value > 0) parts.push(`${updatedCount.value} updated`)
      ElMessage.success(`Applied ${res.length} resource${res.length > 1 ? 's' : ''}: ${parts.join(', ')}`)
      yamlText.value = DEFAULT_YAML
      visible.value = false
    }
    // If there are failures, keep modal open — results table will show
  } catch (err) {
    // Parse-level error (before any API calls)
    parseError.value = err instanceof Error ? err.message : String(err)
  } finally {
    applying.value = false
  }
}

function handleClose() {
  visible.value = false
}

function handleOpened() {
  // Reset results when modal re-opens, but keep YAML text for convenience
  resetState()
}
</script>

<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="(val: boolean) => (visible = val)"
    title="Create from YAML"
    width="750px"
    :close-on-click-modal="false"
    @opened="handleOpened"
    destroy-on-close
  >
    <!-- Help text -->
    <el-alert
      type="info"
      :closable="false"
      show-icon
      style="margin-bottom: 12px"
    >
      <template #title>
        Paste one or more YAML documents (separated by <code>---</code>).
        Existing resources will be updated automatically.
      </template>
      <div style="margin-top: 4px; font-size: 12px; color: #909399">
        Supported versions:
        <el-tag
          v-for="v in SUPPORTED_VERSIONS"
          :key="v"
          size="small"
          style="margin: 2px 4px 2px 0"
        >
          {{ v }}
        </el-tag>
      </div>
    </el-alert>

    <!-- Parse error -->
    <el-alert
      v-if="parseError"
      type="error"
      :title="parseError"
      show-icon
      :closable="true"
      @close="parseError = ''"
      style="margin-bottom: 12px"
    />

    <!-- YAML Editor -->
    <div class="yaml-editor" v-loading="applying">
      <Codemirror
        v-model="yamlText"
        :extensions="cmExtensions"
        :style="{ fontSize: '13px' }"
        placeholder="Paste YAML here..."
      />
    </div>

    <!-- Results table (shown on partial/all failure) -->
    <div v-if="hasResults && !allSucceeded" style="margin-top: 16px">
      <el-alert
        type="warning"
        show-icon
        :closable="false"
        style="margin-bottom: 8px"
      >
        <template #title>
          {{ failedCount }} of {{ results.length }} resource{{ results.length > 1 ? 's' : '' }} failed to apply
        </template>
      </el-alert>
      <el-table :data="results" size="small" stripe style="width: 100%">
        <el-table-column prop="index" label="#" width="50" />
        <el-table-column prop="version" label="Version" min-width="140" />
        <el-table-column prop="name" label="Name" min-width="140" />
        <el-table-column label="Result" width="100">
          <template #default="{ row }">
            <el-tag :type="actionTagType(row.action)" size="small" effect="dark">
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
    </div>

    <!-- Footer -->
    <template #footer>
      <el-button @click="handleClose">Cancel</el-button>
      <el-button
        type="primary"
        :loading="applying"
        :disabled="!canApply"
        @click="handleApply"
      >
        {{ applying ? 'Applying...' : 'Apply' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.yaml-editor :deep(.cm-editor) {
  border-radius: 4px;
  min-height: 250px;
  max-height: 400px;
  overflow: auto;
}

.yaml-editor :deep(.cm-gutters) {
  border-radius: 4px 0 0 4px;
}

code {
  background-color: #f0f0f0;
  padding: 1px 4px;
  border-radius: 3px;
  font-size: 12px;
}
</style>
