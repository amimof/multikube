<script setup lang="ts">
import ResourcePage from '@/components/resources/ResourcePage.vue'
import { routeSummary } from '@/resources/summaries'
import { useRouteStore } from '@/stores/route'

const store = useRouteStore()
</script>

<template>
  <ResourcePage
    :summary="routeSummary"
    :items="store.items"
    :loading="store.loading"
    :error="store.error"
    :fetch-items="store.fetchRoutes"
    :create-item="(item) => store.createRoute(item as typeof store.items[number])"
    :update-item="(item) => store.updateRoute(item as typeof store.items[number])"
    :delete-item="(item) => store.deleteRoute(item as typeof store.items[number])"
  />
</template>
