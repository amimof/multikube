<script setup lang="ts">
import ResourcePage from '@/components/resources/ResourcePage.vue'
import { credentialSummary } from '@/resources/summaries'
import { useCredentialStore } from '@/stores/credential'

const store = useCredentialStore()
</script>

<template>
  <ResourcePage
    :summary="credentialSummary"
    :items="store.items"
    :loading="store.loading"
    :error="store.error"
    :fetch-items="store.fetchCredentials"
    :create-item="(item) => store.createCredential(item as typeof store.items[number])"
    :update-item="(item) => store.updateCredential(item as typeof store.items[number])"
    :delete-item="(item) => store.deleteCredential(item as typeof store.items[number])"
  />
</template>
