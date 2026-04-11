<script setup lang="ts">
import ResourcePage from '@/components/resources/ResourcePage.vue'
import { certificateSummary } from '@/resources/summaries'
import { useCertificateStore } from '@/stores/certificate'

const store = useCertificateStore()
</script>

<template>
  <ResourcePage
    :summary="certificateSummary"
    :items="store.items"
    :loading="store.loading"
    :error="store.error"
    :fetch-items="store.fetchCertificates"
    :create-item="(item) => store.createCertificate(item as typeof store.items[number])"
    :update-item="(item) => store.updateCertificate(item as typeof store.items[number])"
    :delete-item="(item) => store.deleteCertificate(item as typeof store.items[number])"
  />
</template>
