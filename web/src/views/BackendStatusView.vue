<script setup lang="ts">
import { onMounted, computed, toRaw } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Refresh, View } from '@element-plus/icons-vue'
import { useBackendStore } from '@/stores/backend'
import { useCaStore } from '@/stores/ca'
import { useCredentialStore } from '@/stores/credential'
import moment from 'moment'
import { stringify as yamlStringify } from 'yaml'
import { Codemirror } from 'vue-codemirror'
import { yaml as yamlLang } from '@codemirror/lang-yaml'
import { oneDark } from '@codemirror/theme-one-dark'
import { EditorState } from '@codemirror/state'
import type { V1Backend } from '@/generated/backend'

const route = useRoute()
const router = useRouter()
const backendStore = useBackendStore()
const caStore = useCaStore()
const credentialStore = useCredentialStore()

const lbTypeLabels: Record<string, string> = {
  LOAD_BALANCING_TYPE_UNSPECIFIED: 'Unspecified',
  LOAD_BALANCING_TYPE_ROUND_ROBIN: 'Round Robin',
  LOAD_BALANCING_TYPE_LEAST_CONNECTIONS: 'Least Connections',
  LOAD_BALANCING_TYPE_RANDOM: 'Random',
  LOAD_BALANCING_TYPE_WEIGHTED_ROUND_ROBIN: 'Weighted Round Robin',
}

const backendName = computed(() => route.params.name as string)

const backend = computed<V1Backend | undefined>(() =>
  backendStore.items.find((b) => b.meta?.name === backendName.value),
)

const targetRows = computed(() => {
  const statuses = backend.value?.status?.targetStatuses
  if (!statuses) return []
  return Object.entries(statuses).map(([target, status]) => ({
    target,
    phase: status.phase ?? '-',
    reason: status.reason ?? '',
    lastTransitionTime: status.lastTransitionTime,
  }))
})

const healthyCount = computed(() => targetRows.value.filter((r) => r.phase === 'Healthy').length)
const totalCount = computed(() => targetRows.value.length)

const healthTagType = computed(() => {
  if (totalCount.value === 0) return 'info'
  if (healthyCount.value === totalCount.value) return 'success'
  if (healthyCount.value === 0) return 'danger'
  return 'warning'
})

// YAML serialization of the full resource
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

function formatDate(date?: Date): string {
  if (!date) return '-'
  return moment(date).fromNow()
}

function formatDateFull(date?: Date): string {
  if (!date) return '-'
  return new Date(date).toLocaleString()
}

function lbLabel(type?: string): string {
  if (!type) return '-'
  return lbTypeLabels[type] ?? type
}

function handleRefresh() {
  backendStore.fetchBackends().catch(() => {})
}

function goBack() {
  router.push('/backends')
}

onMounted(() => {
  if (backendStore.items.length === 0) {
    backendStore.fetchBackends().catch(() => {})
  }
  if (caStore.items.length === 0) {
    caStore.fetchCas().catch(() => {})
  }
  if (credentialStore.items.length === 0) {
    credentialStore.fetchCredentials().catch(() => {})
  }
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
        </div>
      </el-col>
      <el-col :span="8" style="text-align: right">
        <el-button :icon="Refresh" @click="handleRefresh">Reload</el-button>
      </el-col>
    </el-row>

    <el-alert
      v-if="backendStore.error"
      :title="backendStore.error"
      type="error"
      show-icon
      style="margin-bottom: 16px"
    />

    <!-- Loading -->
    <el-card v-if="backendStore.loading" shadow="never" style="margin-bottom: 16px">
      <el-skeleton :rows="6" animated />
    </el-card>

    <!-- Not found -->
    <el-empty
      v-else-if="!backend"
      :description="`Backend '${backendName}' not found`"
    >
      <el-button type="primary" @click="goBack">Back to Backends</el-button>
    </el-empty>

    <!-- Content -->
    <template v-else>
      <!-- Overview -->
      <el-card shadow="never" style="margin-bottom: 16px">
        <template #header>
          <span style="font-weight: 600">Overview</span>
        </template>
        <el-descriptions :column="2" border size="default">
          <el-descriptions-item label="Name">{{ backend.meta?.name ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="Load Balancing">{{ lbLabel(backend.config?.type as string) }}</el-descriptions-item>
          <el-descriptions-item label="Servers">{{ (backend.config?.servers ?? []).length }}</el-descriptions-item>
          <el-descriptions-item label="Cache TTL">{{ backend.config?.cacheTtl || '-' }}</el-descriptions-item>
          <el-descriptions-item label="Skip TLS Verify">
            <el-tag :type="backend.config?.insecureSkipTlsVerify ? 'warning' : 'info'" size="small">
              {{ backend.config?.insecureSkipTlsVerify ? 'Yes' : 'No' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="Health">
            <el-tag :type="healthTagType" effect="dark" size="small">
              {{ healthyCount }}/{{ totalCount }} Healthy
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="Created">
            <el-tooltip :content="formatDateFull(backend.meta?.created)" placement="top">
              <span>{{ formatDate(backend.meta?.created) }}</span>
            </el-tooltip>
          </el-descriptions-item>
          <el-descriptions-item label="UID">
            <span style="font-family: monospace; font-size: 12px">{{ backend.meta?.uid ?? '-' }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="Resource Version">{{ backend.meta?.resourceVersion ?? '-' }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- Servers -->
      <el-card shadow="never" style="margin-bottom: 16px">
        <template #header>
          <span style="font-weight: 600">Servers</span>
        </template>
        <div v-if="(backend.config?.servers ?? []).length === 0" style="color: #909399">
          No servers configured
        </div>
        <el-tag
          v-for="server in backend.config?.servers ?? []"
          :key="server"
          style="margin: 0 8px 8px 0; font-family: monospace"
          size="default"
        >
          {{ server }}
        </el-tag>
      </el-card>

      <!-- References -->
      <el-card shadow="never" style="margin-bottom: 16px">
        <template #header>
          <span style="font-weight: 600">References</span>
        </template>
        <el-descriptions :column="1" border size="default">
          <el-descriptions-item label="CA Reference">
            <router-link
              v-if="backend.config?.caRef"
              :to="'/cas'"
              style="text-decoration: none"
            >
              <el-link type="primary" :icon="View">{{ backend.config.caRef }}</el-link>
            </router-link>
            <span v-else style="color: #909399">-</span>
          </el-descriptions-item>
          <el-descriptions-item label="Auth Reference (Credential)">
            <router-link
              v-if="backend.config?.authRef"
              :to="'/credentials'"
              style="text-decoration: none"
            >
              <el-link type="primary" :icon="View">{{ backend.config.authRef }}</el-link>
            </router-link>
            <span v-else style="color: #909399">-</span>
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- Target Health -->
      <el-card shadow="never" style="margin-bottom: 16px">
        <template #header>
          <div style="display: flex; justify-content: space-between; align-items: center">
            <span style="font-weight: 600">Target Health</span>
            <el-tag :type="healthTagType" effect="dark" size="small">
              {{ healthyCount }}/{{ totalCount }}
            </el-tag>
          </div>
        </template>

        <el-empty
          v-if="targetRows.length === 0"
          description="No target status data available"
          :image-size="60"
        />

        <el-table
          v-else
          :data="targetRows"
          style="width: 100%"
          stripe
        >
          <el-table-column prop="target" label="Target" min-width="250">
            <template #default="{ row }">
              <span style="font-family: monospace">{{ row.target }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="phase" label="Phase" width="140">
            <template #default="{ row }">
              <el-tag
                :type="row.phase === 'Healthy' ? 'success' : row.phase === 'Unhealthy' ? 'danger' : 'warning'"
                effect="dark"
                size="small"
              >
                {{ row.phase }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="Error" min-width="200">
            <template #default="{ row }">
              <span v-if="row.reason" style="color: #f56c6c">{{ row.reason }}</span>
              <span v-else style="color: #909399">-</span>
            </template>
          </el-table-column>
          <el-table-column label="Last Transition" width="200">
            <template #default="{ row }">
              <el-tooltip v-if="row.lastTransitionTime" :content="formatDateFull(row.lastTransitionTime)" placement="top">
                <span>{{ formatDate(row.lastTransitionTime) }}</span>
              </el-tooltip>
              <span v-else style="color: #909399">-</span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- YAML -->
      <el-card shadow="never" style="margin-bottom: 16px">
        <template #header>
          <span style="font-weight: 600">YAML</span>
        </template>
        <div class="yaml-editor">
          <Codemirror
            :model-value="yamlContent"
            :extensions="cmExtensions"
            :style="{ fontSize: '13px' }"
          />
        </div>
      </el-card>
    </template>
  </div>
</template>

<style scoped>
.yaml-editor :deep(.cm-editor) {
  border-radius: 4px;
  max-height: 600px;
  overflow: auto;
}

.yaml-editor :deep(.cm-gutters) {
  border-radius: 4px 0 0 4px;
}
</style>
