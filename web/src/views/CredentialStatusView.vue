<script setup lang="ts">
import { onMounted, onUnmounted, computed, toRaw } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Refresh } from '@element-plus/icons-vue'
import { useCredentialStore } from '@/stores/credential'
import { stringify as yamlStringify } from 'yaml'
import { Codemirror } from 'vue-codemirror'
import { yaml as yamlLang } from '@codemirror/lang-yaml'
import { oneDark } from '@codemirror/theme-one-dark'
import { EditorState } from '@codemirror/state'
import { formatDate, formatDateFull } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const credentialStore = useCredentialStore()

const credentialName = computed(() => route.params.name as string)
const credential = computed(() => credentialStore.current)

const healthyTag = computed(() => {
	if (credential.value?.status?.healthy === true) return 'success'
	if (credential.value?.status?.healthy === false) return 'danger'
	return 'info'
})

const healthyLabel = computed(() => {
	if (credential.value?.status?.healthy === true) return 'Healthy'
	if (credential.value?.status?.healthy === false) return 'Unhealthy'
	return 'Unknown'
})

const credentialType = computed(() => {
	const config = credential.value?.config
	if (!config) return '-'
	if (config.clientCertificateRef) return 'Client Certificate'
	if (config.token) return 'Token'
	if (config.basic) return 'Basic Auth'
	return '-'
})

const yamlContent = computed(() => {
	if (!credential.value) return ''
	try {
		const raw = structuredClone(toRaw(credential.value))
		return yamlStringify(raw, { lineWidth: 120 })
	} catch {
		return '# Failed to serialize resource'
	}
})

const cmExtensions = [yamlLang(), oneDark, EditorState.readOnly.of(true)]

const labelEntries = computed(() => {
	const labels = credential.value?.meta?.labels
	if (!labels) return []
	return Object.entries(labels)
})

function handleRefresh() {
	credentialStore.fetchCredential(credentialName.value).catch(() => {})
}

function goBack() {
	router.push('/credentials')
}

onMounted(() => {
	credentialStore.fetchCredential(credentialName.value).catch(() => {})
})

onUnmounted(() => {
	credentialStore.clearCurrent()
})
</script>

<template>
	<div>
		<!-- Header -->
		<el-row justify="space-between" align="middle" style="margin-bottom: 16px">
			<el-col :span="16">
				<div style="display: flex; align-items: center; gap: 12px">
					<el-button :icon="ArrowLeft" @click="goBack" text>Credentials</el-button>
					<h2 style="margin: 0">{{ credentialName }}</h2>
					<el-tag :type="healthyTag" effect="dark" size="small">
						{{ healthyLabel }}
					</el-tag>
				</div>
			</el-col>
			<el-col :span="8" style="text-align: right">
				<el-button :icon="Refresh" @click="handleRefresh">Reload</el-button>
			</el-col>
		</el-row>

		<el-alert v-if="credentialStore.error" :title="credentialStore.error" type="error" show-icon
			style="margin-bottom: 16px" />

		<!-- Loading -->
		<el-card v-if="credentialStore.loading" shadow="never" style="margin-bottom: 16px">
			<el-skeleton :rows="6" animated />
		</el-card>

		<!-- Not found -->
		<el-empty v-else-if="!credential" :description="`Credential '${credentialName}' not found`">
			<el-button type="primary" @click="goBack">Back to Credentials</el-button>
		</el-empty>

		<!-- Content -->
		<template v-else>
			<el-row :gutter="16" style="margin-bottom: 16px">
				<!-- Configuration (left) -->
				<el-col :span="14">
					<el-card shadow="never" style="height: 100%">
						<template #header>
							<span style="font-weight: 600">Configuration</span>
						</template>

						<!-- General section -->
						<h4 class="section-title">General</h4>
						<el-descriptions :column="2" border size="default">
							<el-descriptions-item label="Name">
								{{ credential.meta?.name ?? '-' }}
							</el-descriptions-item>
							<el-descriptions-item label="Created">
								<el-tooltip :content="formatDateFull(credential.meta?.created)" placement="top">
									<span>{{ formatDate(credential.meta?.created) }}</span>
								</el-tooltip>
							</el-descriptions-item>
							<el-descriptions-item label="Updated">
								<el-tooltip :content="formatDateFull(credential.meta?.updated)" placement="top">
									<span>{{ formatDate(credential.meta?.updated) }}</span>
								</el-tooltip>
							</el-descriptions-item>
							<el-descriptions-item label="UID">
								<span style="font-family: monospace; font-size: 12px">{{ credential.meta?.uid ?? '-' }}</span>
							</el-descriptions-item>
							<el-descriptions-item label="Labels" :span="2">
								<template v-if="labelEntries.length > 0">
									<el-tag v-for="[key, value] in labelEntries" :key="key" size="small" style="margin: 0 6px 4px 0">
										{{ key }}={{ value }}
									</el-tag>
								</template>
								<span v-else style="color: #909399">-</span>
							</el-descriptions-item>
						</el-descriptions>

						<!-- Authentication section -->
						<h4 class="section-title">Authentication</h4>
						<el-descriptions :column="2" border size="default">
							<el-descriptions-item label="Type">
								<el-tag size="small">{{ credentialType }}</el-tag>
							</el-descriptions-item>

							<!-- Client Certificate Ref -->
							<el-descriptions-item v-if="credential.config?.clientCertificateRef" label="Certificate Ref">
								<router-link :to="`/certificates/${credential.config.clientCertificateRef}`" style="text-decoration: none">
									<el-link type="primary">{{ credential.config.clientCertificateRef }}</el-link>
								</router-link>
							</el-descriptions-item>

							<!-- Token -->
							<el-descriptions-item v-if="credential.config?.token" label="Token">
								<span style="font-family: monospace; font-size: 12px">********</span>
							</el-descriptions-item>

							<!-- Basic Auth -->
							<template v-if="credential.config?.basic">
								<el-descriptions-item label="Username">
									{{ credential.config.basic.username ?? '-' }}
								</el-descriptions-item>
								<el-descriptions-item label="Password">
									<span style="font-family: monospace; font-size: 12px">********</span>
								</el-descriptions-item>
							</template>
						</el-descriptions>
					</el-card>
				</el-col>

				<!-- Status (right) -->
				<el-col :span="10">
					<el-card shadow="never" style="height: 100%">
						<template #header>
							<div style="display: flex; justify-content: space-between; align-items: center">
								<span style="font-weight: 600">Status</span>
								<el-tag :type="healthyTag" effect="dark" size="small">
									{{ healthyLabel }}
								</el-tag>
							</div>
						</template>

						<el-descriptions :column="1" border size="default">
							<el-descriptions-item label="Healthy">
								<el-tag v-if="credential.status?.healthy === true" type="success" effect="dark" size="small">Yes</el-tag>
								<el-tag v-else-if="credential.status?.healthy === false" type="danger" effect="dark" size="small">No</el-tag>
								<span v-else style="color: #909399">-</span>
							</el-descriptions-item>
						</el-descriptions>
					</el-card>
				</el-col>
			</el-row>

			<!-- YAML -->
			<el-card shadow="never" style="margin-bottom: 16px">
				<template #header>
					<span style="font-weight: 600">YAML</span>
				</template>
				<div class="yaml-editor">
					<Codemirror :model-value="yamlContent" :extensions="cmExtensions" :style="{ fontSize: '13px' }" />
				</div>
			</el-card>
		</template>
	</div>
</template>

<style scoped>
.section-title {
	margin: 20px 0 8px 0;
	font-size: 13px;
	font-weight: 600;
	color: #909399;
	text-transform: uppercase;
	letter-spacing: 0.5px;
}

.section-title:first-of-type {
	margin-top: 0;
}

.yaml-editor :deep(.cm-editor) {
	border-radius: 4px;
	max-height: 600px;
	overflow: auto;
}

.yaml-editor :deep(.cm-gutters) {
	border-radius: 4px 0 0 4px;
}
</style>
