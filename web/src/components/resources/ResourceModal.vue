<script setup lang="ts">
import { computed, ref, toRaw, watch } from 'vue'
import BackendForm from '@/components/resources/forms/BackendForm.vue'
import CaForm from '@/components/resources/forms/CaForm.vue'
import CertificateForm from '@/components/resources/forms/CertificateForm.vue'
import CredentialForm from '@/components/resources/forms/CredentialForm.vue'
import PolicyForm from '@/components/resources/forms/PolicyForm.vue'
import RouteForm from '@/components/resources/forms/RouteForm.vue'
import type { ResourceItem, ResourceSummary } from '@/types/resources'

const props = defineProps<{
  summary: ResourceSummary
  item: ResourceItem | null
  open: boolean
}>()

const emit = defineEmits<{
  close: []
  save: [value: ResourceItem]
  remove: [value: ResourceItem]
}>()

const draft = ref<ResourceItem>(props.summary.createEmpty())
const error = ref<string | null>(null)

function cloneResource(value: ResourceItem) {
  return structuredClone(toRaw(value)) as ResourceItem
}

watch(
  () => [props.item, props.open, props.summary],
  () => {
    draft.value = cloneResource(props.item ?? props.summary.createEmpty())
    error.value = null
  },
  { immediate: true },
)

const isEditing = computed(() => Boolean(props.item?.meta?.uid || props.item?.meta?.name))

const formComponent = computed(() => {
  switch (props.summary.name) {
    case 'backends':
      return BackendForm
    case 'cas':
      return CaForm
    case 'certificates':
      return CertificateForm
    case 'credentials':
      return CredentialForm
    case 'policies':
      return PolicyForm
    case 'routes':
      return RouteForm
    default:
      return BackendForm
  }
})

function save() {
  try {
    error.value = null
    emit('save', cloneResource(draft.value))
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to prepare resource'
  }
}

function remove() {
  if (!isEditing.value) {
    return
  }

  const confirmed = window.confirm(`Delete ${props.summary.itemLabel} \"${draft.value.meta?.name ?? 'unnamed'}\"?`)

  if (confirmed) {
    emit('remove', cloneResource(draft.value))
  }
}
</script>

<template>
  <div v-if="open" class="modal fade show d-block" tabindex="-1" role="dialog" aria-modal="true">
    <div class="modal-dialog modal-xl modal-dialog-scrollable">
      <div class="modal-content">
        <div class="modal-header">
          <div>
            <h5 class="modal-title mb-0">
              {{ isEditing ? `Edit ${summary.itemLabel}` : `Create ${summary.itemLabel}` }}
            </h5>
            <small class="text-body-secondary">{{ draft.meta?.name ?? 'New resource' }}</small>
          </div>
          <button type="button" class="btn-close" aria-label="Close" @click="emit('close')" />
        </div>

        <div class="modal-body">
          <div v-if="error" class="alert alert-danger" role="alert">
            {{ error }}
          </div>

          <component :is="formComponent" v-model="draft" :is-editing="isEditing" />
        </div>

        <div class="modal-footer justify-content-between">
          <button v-if="isEditing" type="button" class="btn btn-outline-danger" @click="remove">Delete</button>
          <div class="d-flex gap-2 ms-auto">
            <button type="button" class="btn btn-outline-secondary" @click="emit('close')">Cancel</button>
            <button type="button" class="btn btn-primary" @click="save">
              {{ isEditing ? 'Save Changes' : 'Create' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
  <div v-if="open" class="modal-backdrop fade show" @click="emit('close')" />
</template>
