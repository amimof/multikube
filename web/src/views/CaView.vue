<script setup lang="ts">
import ResourcePage from '@/components/resources/ResourcePage.vue'
import { certificateAuthoritySummary } from '@/resources/summaries'
import { useCaStore } from '@/stores/ca'

const store = useCaStore()
</script>

<template>
  <ResourcePage
    :summary="certificateAuthoritySummary"
    :items="store.items"
    :loading="store.loading"
    :error="store.error"
    :fetch-items="store.fetchCas"
    :create-item="(item) => store.createCa(item as typeof store.items[number])"
    :update-item="(item) => store.updateCa(item as typeof store.items[number])"
    :delete-item="(item) => store.deleteCa(item as typeof store.items[number])"
  />
</template>
