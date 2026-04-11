<script setup lang="ts">
import { onMounted, ref } from 'vue'
import ResourceModal from '@/components/resources/ResourceModal.vue'
import type { ResourceItem, ResourceSummary } from '@/types/resources'

type ResourceMutation = (item: ResourceItem) => Promise<unknown>
type ResourceDelete = (item: ResourceItem) => Promise<void>

const props = defineProps<{
  summary: ResourceSummary
  items: ResourceItem[]
  loading: boolean
  error: string | null
  fetchItems: () => Promise<void>
  createItem: ResourceMutation
  updateItem: ResourceMutation
  deleteItem: ResourceDelete
}>()
const saving = ref(false)
const modalError = ref<string | null>(null)
const selectedItem = ref<ResourceItem | null>(null)
const modalOpen = ref(false)

function itemKey(item: ResourceItem) {
  return item.meta?.uid ?? item.meta?.name ?? JSON.stringify(item)
}

function openCreate() {
  selectedItem.value = null
  modalError.value = null
  modalOpen.value = true
}

function openEdit(item: ResourceItem) {
  selectedItem.value = item
  modalError.value = null
  modalOpen.value = true
}

function closeModal() {
  if (saving.value) {
    return
  }

  modalOpen.value = false
  selectedItem.value = null
  modalError.value = null
}

async function saveItem(item: ResourceItem) {
  saving.value = true
  modalError.value = null

  try {
    if (selectedItem.value) {
      await props.updateItem(item)
    } else {
      await props.createItem(item)
    }

    closeModal()
  } catch (err) {
    modalError.value = err instanceof Error ? err.message : 'Failed to save resource'
  } finally {
    saving.value = false
  }
}

async function removeItem(item: ResourceItem) {
  saving.value = true
  modalError.value = null

  try {
    await props.deleteItem(item)
    closeModal()
  } catch (err) {
    modalError.value = err instanceof Error ? err.message : 'Failed to delete resource'
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  props.fetchItems()
})
</script>

<template>
  <section class="d-flex flex-column gap-4">
    <div class="d-flex flex-wrap justify-content-between align-items-center gap-3">
      <div>
        <h1 class="h3 mb-1">{{ summary.label }}</h1>
        <p class="text-body-secondary mb-0">Manage {{ summary.label.toLowerCase() }} from one place.</p>
      </div>

      <button class="btn btn-primary" type="button" @click="openCreate">Create {{ summary.itemLabel }}</button>
    </div>

    <div v-if="error" class="alert alert-danger" role="alert">
      {{ error }}
    </div>

    <div class="card shadow-sm border-0">
      <div class="card-body p-0">
        <div class="table-responsive">
          <table class="table table-hover align-middle mb-0">
            <thead class="table-light">
              <tr>
                <th scope="col">Name</th>
                <th v-for="column in summary.columns" :key="column.key" scope="col">{{ column.label }}</th>
                <th scope="col">UID</th>
                <th scope="col">Updated</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="loading">
                <td :colspan="summary.columns.length + 3" class="text-center py-5 text-body-secondary">Loading...</td>
              </tr>

              <tr v-else-if="items.length === 0">
                <td :colspan="summary.columns.length + 3" class="text-center py-5 text-body-secondary">
                  No {{ summary.label.toLowerCase() }} yet.
                </td>
              </tr>

              <tr
                v-for="item in items"
                :key="itemKey(item)"
                style="cursor: pointer"
                @click="openEdit(item)"
              >
                <td class="fw-semibold">{{ item.meta?.name || item.config?.name || 'Unnamed' }}</td>
                <td v-for="column in summary.columns" :key="column.key">{{ column.value(item) }}</td>
                <td class="font-monospace small">{{ item.meta?.uid || '-' }}</td>
                <td>{{ item.meta?.updated ? new Date(item.meta.updated).toLocaleString() : '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </section>

  <ResourceModal
    :summary="summary"
    :item="selectedItem"
    :open="modalOpen"
    @close="closeModal"
    @save="saveItem"
    @remove="removeItem"
  />

  <div v-if="modalError" class="toast-container position-fixed bottom-0 end-0 p-3">
    <div class="toast show text-bg-danger border-0" role="alert" aria-live="assertive" aria-atomic="true">
      <div class="d-flex">
        <div class="toast-body">{{ modalError }}</div>
        <button type="button" class="btn-close btn-close-white me-2 m-auto" aria-label="Close" @click="modalError = null" />
      </div>
    </div>
  </div>
</template>
