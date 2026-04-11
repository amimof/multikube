<script setup lang="ts">
import ResourcePage from '@/components/resources/ResourcePage.vue'
import { backendSummary } from '@/resources/summaries'
import { useBackendStore } from '@/stores/backend'

const store = useBackendStore()
</script>

<template>
	<ResourcePage
		:summary="backendSummary"
		:items="store.items"
		:loading="store.loading"
		:error="store.error"
		:fetch-items="store.fetchBackends"
		:create-item="(item) => store.createBackend(item as typeof store.items[number])"
		:update-item="(item) => store.updateBackend(item as typeof store.items[number])"
		:delete-item="(item) => store.deleteBackend(item as typeof store.items[number])"
	/>
</template>
