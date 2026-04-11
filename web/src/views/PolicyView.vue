<script setup lang="ts">
import ResourcePage from '@/components/resources/ResourcePage.vue'
import { policySummary } from '@/resources/summaries'
import { usePolicyStore } from '@/stores/policy'

const store = usePolicyStore()
</script>

<template>
  <ResourcePage
    :summary="policySummary"
    :items="store.items"
    :loading="store.loading"
    :error="store.error"
    :fetch-items="store.fetchPolicies"
    :create-item="(item) => store.createPolicy(item as typeof store.items[number])"
    :update-item="(item) => store.updatePolicy(item as typeof store.items[number])"
    :delete-item="(item) => store.deletePolicy(item as typeof store.items[number])"
  />
</template>
