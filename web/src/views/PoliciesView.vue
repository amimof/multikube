<script setup lang="ts">
import { onMounted, ref, computed, toRaw } from 'vue'
import { Plus, Refresh, Delete, Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { usePolicyStore } from '@/stores/policy'
import { useResourceTable } from '@/composables/useResourceTable'
import moment from 'moment'
import type { V1Policy } from '@/generated/policy'
import type { Policyv1Rule } from '@/generated/policy'
import type { V1SubjectSelector } from '@/generated/policy'
import type { V1ClusterSelector } from '@/generated/policy'
import type { V1ResourceSelector } from '@/generated/policy'
import type { V1Condition } from '@/generated/policy'
import type { V1Claim } from '@/generated/policy'
import { V1Effect, V1Action, V1ConditionType } from '@/generated/policy'
import LabelEditor from '@/components/LabelEditor.vue'
import MetadataDisplay from '@/components/MetadataDisplay.vue'
import ConfirmDelete from '@/components/ConfirmDelete.vue'

const policyStore = usePolicyStore()

const { nameFilter, displayItems } = useResourceTable(computed(() => policyStore.items))

const dialogVisible = ref(false)
const isEditing = ref(false)
const saving = ref(false)
const deleteDialogVisible = ref(false)
const deleteTarget = ref<V1Policy | null>(null)
const selectedRows = ref<V1Policy[]>([])
const bulkDeleteVisible = ref(false)
const bulkDeleting = ref(false)

const form = ref<V1Policy>(createEmptyPolicy())

function createEmptyPolicy(): V1Policy {
  return {
    version: 'policy/v1',
    meta: { name: '', labels: {} },
    config: { rules: [] },
  }
}

function createEmptyRule(): Policyv1Rule {
  return {
    effect: V1Effect.EffectUnspecified,
    subjects: [],
    clusters: [],
    resources: [],
    actions: [],
    conditions: [],
  }
}

function createEmptySubject(): V1SubjectSelector {
  return { users: [], groups: [], serviceAccounts: [], claims: [] }
}

function createEmptyCluster(): V1ClusterSelector {
  return { names: [], labels: {} }
}

function createEmptyResource(): V1ResourceSelector {
  return {
    apiGroup: '',
    resource: '',
    subResource: '',
    namespaces: [],
    names: [],
    labelSelector: { matchLabels: {} },
  }
}

function createEmptyClaim(): V1Claim {
  return { name: '', value: '' }
}

function createEmptyCondition(): V1Condition {
  return { type: V1ConditionType.ConditionTypeUnspecified }
}

// Labels computed
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

  const rules = form.value.config?.rules
  if (!rules || rules.length === 0) return false

  for (const rule of rules) {
    if (!rule.effect || rule.effect === V1Effect.EffectUnspecified) return false
    if (!rule.actions || rule.actions.length === 0) return false
  }

  return true
})

// Enum display helpers
const effectOptions = [
  { label: 'Allow', value: V1Effect.EffectAllow },
  { label: 'Deny', value: V1Effect.EffectDeny },
]

const actionOptions = Object.entries(V1Action)
  .filter(([, v]) => v !== V1Action.ActionUnspecified)
  .map(([k, v]) => ({
    label: k.replace('Action', ''),
    value: v,
  }))

const conditionTypeOptions = Object.entries(V1ConditionType)
  .filter(([, v]) => v !== V1ConditionType.ConditionTypeUnspecified)
  .map(([k, v]) => ({
    label: k.replace('ConditionType', ''),
    value: v,
  }))

// Textarea helpers: convert between string[] and newline-separated text
function arrToText(arr?: string[]): string {
  return (arr ?? []).join('\n')
}

function textToArr(text: string): string[] {
  return text
    .split('\n')
    .map((s) => s.trim())
    .filter((s) => s.length > 0)
}

// Table helpers
function formatDate(date?: Date): string {
  if (!date) return '-'
  return moment(date).fromNow()
}

function sortByCreated(a: any, b: any): number {
  const ta = new Date(a.meta?.created ?? 0).getTime()
  const tb = new Date(b.meta?.created ?? 0).getTime()
  return ta - tb
}

function sortByRules(a: any, b: any): number {
  return (a.config?.rules?.length ?? 0) - (b.config?.rules?.length ?? 0)
}

// Selection
function handleSelectionChange(rows: V1Policy[]) {
  selectedRows.value = rows
}

function handleRowClick(row: V1Policy, column: any) {
  if (column?.type === 'selection') return
  openEdit(row)
}

function confirmBulkDelete() {
  bulkDeleteVisible.value = true
}

async function handleBulkDelete() {
  bulkDeleting.value = true
  try {
    const { succeeded, failed } = await policyStore.deleteManyPolicies(selectedRows.value)
    selectedRows.value = []
    if (failed.length === 0) {
      ElMessage.success(`Deleted ${succeeded} polic${succeeded === 1 ? 'y' : 'ies'}`)
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

// Rule management
function addRule() {
  if (!form.value.config) form.value.config = { rules: [] }
  if (!form.value.config.rules) form.value.config.rules = []
  form.value.config.rules.push(createEmptyRule())
}

function removeRule(index: number) {
  form.value.config?.rules?.splice(index, 1)
}

// Subject management
function addSubject(rule: Policyv1Rule) {
  if (!rule.subjects) rule.subjects = []
  rule.subjects.push(createEmptySubject())
}

function removeSubject(rule: Policyv1Rule, index: number) {
  rule.subjects?.splice(index, 1)
}

// Claim management
function addClaim(subject: V1SubjectSelector) {
  if (!subject.claims) subject.claims = []
  subject.claims.push(createEmptyClaim())
}

function removeClaim(subject: V1SubjectSelector, index: number) {
  subject.claims?.splice(index, 1)
}

// Cluster management
function addCluster(rule: Policyv1Rule) {
  if (!rule.clusters) rule.clusters = []
  rule.clusters.push(createEmptyCluster())
}

function removeCluster(rule: Policyv1Rule, index: number) {
  rule.clusters?.splice(index, 1)
}

// Resource management
function addResource(rule: Policyv1Rule) {
  if (!rule.resources) rule.resources = []
  rule.resources.push(createEmptyResource())
}

function removeResource(rule: Policyv1Rule, index: number) {
  rule.resources?.splice(index, 1)
}

// Condition management
function addCondition(rule: Policyv1Rule) {
  if (!rule.conditions) rule.conditions = []
  rule.conditions.push(createEmptyCondition())
}

function removeCondition(rule: Policyv1Rule, index: number) {
  rule.conditions?.splice(index, 1)
}

// Dialog actions
function openCreate() {
  form.value = createEmptyPolicy()
  isEditing.value = false
  dialogVisible.value = true
}

function openEdit(row: V1Policy) {
  form.value = structuredClone(toRaw(row))
  if (!form.value.config) form.value.config = { rules: [] }
  if (!form.value.config.rules) form.value.config.rules = []
  // Ensure all rule sub-arrays exist
  for (const rule of form.value.config.rules) {
    if (!rule.subjects) rule.subjects = []
    if (!rule.clusters) rule.clusters = []
    if (!rule.resources) rule.resources = []
    if (!rule.actions) rule.actions = []
    if (!rule.conditions) rule.conditions = []
    // Ensure sub-arrays inside subjects
    for (const subject of rule.subjects) {
      if (!subject.users) subject.users = []
      if (!subject.groups) subject.groups = []
      if (!subject.serviceAccounts) subject.serviceAccounts = []
      if (!subject.claims) subject.claims = []
    }
    // Ensure sub-arrays inside clusters
    for (const cluster of rule.clusters) {
      if (!cluster.names) cluster.names = []
      if (!cluster.labels) cluster.labels = {}
    }
    // Ensure sub-arrays inside resources
    for (const resource of rule.resources) {
      if (!resource.namespaces) resource.namespaces = []
      if (!resource.names) resource.names = []
      if (!resource.labelSelector) resource.labelSelector = { matchLabels: {} }
      if (!resource.labelSelector.matchLabels) resource.labelSelector.matchLabels = {}
    }
  }
  isEditing.value = true
  dialogVisible.value = true
}

function confirmDelete(row: V1Policy) {
  deleteTarget.value = row
  deleteDialogVisible.value = true
}

async function handleDelete() {
  if (!deleteTarget.value) return
  try {
    await policyStore.deletePolicy(deleteTarget.value)
    ElMessage.success('Policy deleted')
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'Delete failed')
  }
  deleteTarget.value = null
}

async function handleSave() {
  saving.value = true
  try {
    if (isEditing.value) {
      await policyStore.updatePolicy(form.value)
      ElMessage.success('Policy updated')
    } else {
      await policyStore.createPolicy(form.value)
      ElMessage.success('Policy created')
    }
    dialogVisible.value = false
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'Save failed')
  } finally {
    saving.value = false
  }
}

function handleRefresh() {
  policyStore.fetchPolicies().catch(() => {})
}

onMounted(() => {
  policyStore.fetchPolicies().catch(() => {})
})
</script>

<template>
  <div>
    <el-row justify="space-between" align="middle" style="margin-bottom: 16px">
      <el-col :span="12">
        <h2 style="margin: 0">Policies</h2>
      </el-col>
      <el-col :span="12" style="text-align: right">
        <el-button :icon="Refresh" @click="handleRefresh">Reload</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">Create</el-button>
      </el-col>
    </el-row>

    <el-alert v-if="policyStore.error" :title="policyStore.error" type="error" show-icon style="margin-bottom: 16px" />

    <el-empty v-if="!policyStore.loading && policyStore.items.length === 0" description="No policies yet">
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
        v-loading="policyStore.loading"
        element-loading-text="Loading..."
        :data="displayItems"
        style="width: 100%"
        row-key="meta.name"
        @row-click="handleRowClick"
        @selection-change="handleSelectionChange"
        :row-class-name="() => 'clickable-row'"
      >
      <el-table-column type="selection" width="48" />
      <el-table-column prop="meta.name" label="Name" min-width="180" sortable />
      <el-table-column label="Rules" min-width="80" sortable :sort-method="sortByRules">
        <template #default="{ row }">
          <el-tag size="small">{{ row.config?.rules?.length ?? 0 }}</el-tag>
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
      :title="isEditing ? 'Edit Policy' : 'Create Policy'"
      width="900"
      destroy-on-close
    >
      <el-form label-width="160px" label-position="right">
        <el-collapse v-if="isEditing" style="margin-bottom: 20px">
          <el-collapse-item title="Metadata" name="metadata">
            <MetadataDisplay :meta="form.meta" />
          </el-collapse-item>
        </el-collapse>

        <el-form-item label="Name" required>
          <el-input v-model="form.meta!.name" :disabled="isEditing" placeholder="my-policy" />
        </el-form-item>

        <el-form-item label="Labels">
          <LabelEditor v-model="formLabels" />
        </el-form-item>

        <el-divider content-position="left">Config</el-divider>

        <!-- Rules -->
        <el-divider content-position="left">Rules</el-divider>

        <div v-if="!form.config?.rules?.length" style="margin-bottom: 16px">
          <el-empty description="No rules defined" :image-size="60">
            <el-button type="primary" size="small" :icon="Plus" @click="addRule">Add Rule</el-button>
          </el-empty>
        </div>

        <div v-for="(rule, rIdx) in form.config?.rules" :key="rIdx" style="margin-bottom: 24px">
          <el-card shadow="never">
            <template #header>
              <div style="display: flex; justify-content: space-between; align-items: center">
                <span style="font-weight: 600">Rule {{ rIdx + 1 }}</span>
                <el-button type="danger" size="small" plain :icon="Delete" @click="removeRule(rIdx)">Remove</el-button>
              </div>
            </template>

            <!-- Effect -->
            <el-form-item label="Effect" required>
              <el-select v-model="rule.effect" placeholder="Select effect" style="width: 200px">
                <el-option
                  v-for="opt in effectOptions"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </el-select>
            </el-form-item>

            <!-- Actions -->
            <el-form-item label="Actions" required>
              <el-select
                v-model="rule.actions"
                multiple
                filterable
                placeholder="Select actions"
                style="width: 100%"
              >
                <el-option
                  v-for="opt in actionOptions"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </el-select>
            </el-form-item>

            <!-- Subjects -->
            <el-divider content-position="left" style="margin: 12px 0">Subjects</el-divider>
            <div v-for="(subject, sIdx) in rule.subjects" :key="sIdx" style="margin-bottom: 12px">
              <el-card shadow="never" style="background-color: #fafafa">
                <template #header>
                  <div style="display: flex; justify-content: space-between; align-items: center">
                    <span>Subject {{ sIdx + 1 }}</span>
                    <el-button size="small" type="danger" plain @click="removeSubject(rule, sIdx)">Remove</el-button>
                  </div>
                </template>
                <el-form-item label="Users">
                  <el-input
                    type="textarea"
                    :rows="2"
                    :model-value="arrToText(subject.users)"
                    @update:model-value="subject.users = textToArr($event)"
                    placeholder="One user per line"
                  />
                </el-form-item>
                <el-form-item label="Groups">
                  <el-input
                    type="textarea"
                    :rows="2"
                    :model-value="arrToText(subject.groups)"
                    @update:model-value="subject.groups = textToArr($event)"
                    placeholder="One group per line"
                  />
                </el-form-item>
                <el-form-item label="Service Accounts">
                  <el-input
                    type="textarea"
                    :rows="2"
                    :model-value="arrToText(subject.serviceAccounts)"
                    @update:model-value="subject.serviceAccounts = textToArr($event)"
                    placeholder="One service account per line"
                  />
                </el-form-item>
                <!-- Claims -->
                <div style="margin-bottom: 8px; font-weight: 500; font-size: 13px">Claims</div>
                <div v-for="(claim, cIdx) in subject.claims" :key="cIdx" style="display: flex; gap: 8px; margin-bottom: 8px">
                  <el-input v-model="claim.name" placeholder="Claim name" style="flex: 1" />
                  <el-input v-model="claim.value" placeholder="Claim value" style="flex: 1" />
                  <el-button type="danger" plain size="small" :icon="Delete" @click="removeClaim(subject, cIdx)" />
                </div>
                <el-button size="small" :icon="Plus" @click="addClaim(subject)">Add Claim</el-button>
              </el-card>
            </div>
            <el-button size="small" :icon="Plus" @click="addSubject(rule)">Add Subject</el-button>

            <!-- Clusters -->
            <el-divider content-position="left" style="margin: 12px 0">Clusters</el-divider>
            <div v-for="(cluster, cIdx) in rule.clusters" :key="cIdx" style="margin-bottom: 12px">
              <el-card shadow="never" style="background-color: #fafafa">
                <template #header>
                  <div style="display: flex; justify-content: space-between; align-items: center">
                    <span>Cluster {{ cIdx + 1 }}</span>
                    <el-button size="small" type="danger" plain @click="removeCluster(rule, cIdx)">Remove</el-button>
                  </div>
                </template>
                <el-form-item label="Names">
                  <el-input
                    type="textarea"
                    :rows="2"
                    :model-value="arrToText(cluster.names)"
                    @update:model-value="cluster.names = textToArr($event)"
                    placeholder="One cluster name per line"
                  />
                </el-form-item>
                <el-form-item label="Labels">
                  <LabelEditor v-model="cluster.labels!" />
                </el-form-item>
              </el-card>
            </div>
            <el-button size="small" :icon="Plus" @click="addCluster(rule)">Add Cluster</el-button>

            <!-- Resources -->
            <el-divider content-position="left" style="margin: 12px 0">Resources</el-divider>
            <div v-for="(resource, resIdx) in rule.resources" :key="resIdx" style="margin-bottom: 12px">
              <el-card shadow="never" style="background-color: #fafafa">
                <template #header>
                  <div style="display: flex; justify-content: space-between; align-items: center">
                    <span>Resource {{ resIdx + 1 }}</span>
                    <el-button size="small" type="danger" plain @click="removeResource(rule, resIdx)">Remove</el-button>
                  </div>
                </template>
                <el-form-item label="API Group">
                  <el-input v-model="resource.apiGroup" placeholder="e.g. apps" />
                </el-form-item>
                <el-form-item label="Resource">
                  <el-input v-model="resource.resource" placeholder="e.g. deployments" />
                </el-form-item>
                <el-form-item label="Sub Resource">
                  <el-input v-model="resource.subResource" placeholder="e.g. status" />
                </el-form-item>
                <el-form-item label="Namespaces">
                  <el-input
                    type="textarea"
                    :rows="2"
                    :model-value="arrToText(resource.namespaces)"
                    @update:model-value="resource.namespaces = textToArr($event)"
                    placeholder="One namespace per line"
                  />
                </el-form-item>
                <el-form-item label="Names">
                  <el-input
                    type="textarea"
                    :rows="2"
                    :model-value="arrToText(resource.names)"
                    @update:model-value="resource.names = textToArr($event)"
                    placeholder="One resource name per line"
                  />
                </el-form-item>
                <el-form-item label="Match Labels">
                  <LabelEditor v-model="resource.labelSelector!.matchLabels!" />
                </el-form-item>
              </el-card>
            </div>
            <el-button size="small" :icon="Plus" @click="addResource(rule)">Add Resource</el-button>

            <!-- Conditions -->
            <el-divider content-position="left" style="margin: 12px 0">Conditions</el-divider>
            <div v-for="(condition, condIdx) in rule.conditions" :key="condIdx" style="display: flex; gap: 8px; margin-bottom: 8px">
              <el-select v-model="condition.type" placeholder="Condition type" style="flex: 1">
                <el-option
                  v-for="opt in conditionTypeOptions"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </el-select>
              <el-button type="danger" plain size="small" :icon="Delete" @click="removeCondition(rule, condIdx)" />
            </div>
            <el-button size="small" :icon="Plus" @click="addCondition(rule)">Add Condition</el-button>
          </el-card>
        </div>

        <el-button v-if="form.config?.rules?.length" :icon="Plus" @click="addRule" style="margin-bottom: 16px">
          Add Rule
        </el-button>
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
      :message="`Delete ${selectedRows.length} selected polic${selectedRows.length === 1 ? 'y' : 'ies'}?`"
      @confirm="handleBulkDelete"
    />
  </div>
</template>
