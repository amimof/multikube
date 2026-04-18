<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed, watch, toRaw } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Refresh, Document } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useCredentialStore } from '@/stores/credential'
import { useCertificateStore } from '@/stores/certificate'
import { stringify as yamlStringify } from 'yaml'
import { Codemirror } from 'vue-codemirror'
import { yaml as yamlLang } from '@codemirror/lang-yaml'
import { oneDark } from '@codemirror/theme-one-dark'
import { EditorState } from '@codemirror/state'
import { formatDate, formatDateFull } from '@/utils/format'
import type { V1Credential } from '@/generated/credential'
import LabelEditor from '@/components/LabelEditor.vue'
import MetadataDisplay from '@/components/MetadataDisplay.vue'
import EditYamlModal from '@/components/EditYamlModal.vue'

type CredentialMode = '' | 'clientCertificateRef' | 'token' | 'basic'

const route = useRoute()
const router = useRouter()
const credentialStore = useCredentialStore()
const certificateStore = useCertificateStore()

const credentialName = computed(() => route.params.name as string)
const credential = computed(() => credentialStore.current)

const saving = ref(false)
const yamlModalVisible = ref(false)
const credentialMode = ref<CredentialMode>('')

const form = ref<V1Credential>({})

function inferMode(config: V1Credential['config']): CredentialMode {
	if (!config) return ''
	if (config.clientCertificateRef) return 'clientCertificateRef'
	if (config.token) return 'token'
	if (config.basic) return 'basic'
	return ''
}

// Initialize form from store
watch(credential, (val) => {
	if (val) {
		const raw = structuredClone(toRaw(val))
		if (!raw.config) raw.config = {}
		form.value = raw
		credentialMode.value = inferMode(raw.config)
	}
}, { immediate: true })

// When mode changes, reset config auth fields
watch(credentialMode, (newMode, oldMode) => {
	if (newMode === oldMode) return
	const enabled = form.value.config?.enabled
	switch (newMode) {
		case 'clientCertificateRef':
			form.value.config = { enabled, clientCertificateRef: '' }
			break
		case 'token':
			form.value.config = { enabled, token: '' }
			break
		case 'basic':
			form.value.config = { enabled, basic: { username: '', password: '' } }
			break
		default:
			form.value.config = { enabled }
			break
	}
})

const formLabels = computed({
	get: () => form.value.meta?.labels ?? {},
	set: (val: Record<string, string>) => {
		if (form.value.meta) {
			form.value.meta.labels = val
		}
	},
})

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

const yamlContent = computed(() => {
	if (!form.value || !form.value.version) return ''
	try {
		const raw = structuredClone(toRaw(form.value))
		return yamlStringify(raw, { lineWidth: 120 })
	} catch {
		return '# Failed to serialize resource'
	}
})

const cmExtensions = [yamlLang(), oneDark, EditorState.readOnly.of(true)]

async function handleSave() {
	saving.value = true
	try {
		await credentialStore.updateCredential(form.value)
		await credentialStore.fetchCredential(credentialName.value)
		ElMessage.success('Credential updated')
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Save failed')
	} finally {
		saving.value = false
	}
}

async function handleYamlSave(parsed: unknown) {
	saving.value = true
	try {
		const resource = parsed as V1Credential
		if (!resource.meta?.name) {
			resource.meta = { ...resource.meta, name: credentialName.value }
		}
		await credentialStore.updateCredential(resource)
		await credentialStore.fetchCredential(credentialName.value)
		ElMessage.success('Credential updated from YAML')
		yamlModalVisible.value = false
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Save failed')
	} finally {
		saving.value = false
	}
}

function handleRefresh() {
	credentialStore.fetchCredential(credentialName.value).catch(() => { })
}

function goBack() {
	router.push('/credentials')
}

onMounted(() => {
	credentialStore.fetchCredential(credentialName.value).catch(() => { })
	certificateStore.fetchCertificates().catch(() => { })
})

onUnmounted(() => {
	credentialStore.clearCurrent()
})
</script>

<template>
	<div>
		<!-- Header -->
		<el-row justify="space-between" align="middle" style="margin-bottom: 16px">
			<el-col :span="12">
				<div style="display: flex; align-items: center; gap: 12px">
					<el-button :icon="ArrowLeft" @click="goBack" text>Credentials</el-button>
					<h2 style="margin: 0">{{ credentialName }}</h2>
					<el-tag :type="healthyTag" effect="dark" size="small">
						{{ healthyLabel }}
					</el-tag>
				</div>
			</el-col>
			<el-col :span="12" style="text-align: right">
				<el-button :icon="Refresh" @click="handleRefresh">Reload</el-button>
				<el-button :icon="Document" @click="yamlModalVisible = true">Edit YAML</el-button>
				<el-button type="primary" :loading="saving" @click="handleSave">
					{{ saving ? 'Saving...' : 'Save' }}
				</el-button>
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
				<!-- Configuration (left) - editable form -->
				<el-col :span="14">
					<el-card shadow="never" style="height: 100%">
						<template #header>
							<span style="font-weight: 600">Configuration</span>
						</template>

						<!-- Metadata (read-only) -->
						<el-collapse style="margin-bottom: 20px">
							<el-collapse-item title="Metadata" name="metadata">
								<MetadataDisplay :meta="credential.meta" />
							</el-collapse-item>
						</el-collapse>

						<el-form label-width="180px" label-position="right">
							<el-form-item label="Name">
								<el-input :model-value="form.meta?.name" disabled />
							</el-form-item>

							<el-form-item label="Labels">
								<LabelEditor v-model="formLabels" />
							</el-form-item>

							<el-divider content-position="left">Config</el-divider>

							<el-form-item label="Enabled">
								<el-switch v-model="form.config!.enabled" />
							</el-form-item>

							<el-form-item label="Credential Type">
								<el-select v-model="credentialMode" placeholder="Select credential type" style="width: 100%">
									<el-option label="Client Certificate" value="clientCertificateRef" />
									<el-option label="Token" value="token" />
									<el-option label="Basic Auth" value="basic" />
								</el-select>
							</el-form-item>

							<!-- Client Certificate Ref mode -->
							<el-form-item v-if="credentialMode === 'clientCertificateRef'" label="Client Certificate">
								<el-select v-model="form.config!.clientCertificateRef" placeholder="Select a certificate"
									style="width: 100%" filterable clearable :loading="certificateStore.loading">
									<el-option v-for="cert in certificateStore.items" :key="cert.meta?.name" :label="cert.meta?.name"
										:value="cert.meta?.name ?? ''" />
								</el-select>
							</el-form-item>

							<!-- Token mode -->
							<el-form-item v-if="credentialMode === 'token'" label="Token">
								<el-input v-model="form.config!.token" type="textarea" :rows="4" placeholder="Bearer token" />
							</el-form-item>

							<!-- Basic Auth mode -->
							<template v-if="credentialMode === 'basic'">
								<el-form-item label="Username">
									<el-input v-model="form.config!.basic!.username" placeholder="Username" />
								</el-form-item>
								<el-form-item label="Password">
									<el-input v-model="form.config!.basic!.password" type="password" placeholder="Password"
										show-password />
								</el-form-item>
							</template>
						</el-form>
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
								<el-tag v-if="credential.status?.healthy === true" type="success" effect="dark"
									size="small">Yes</el-tag>
								<el-tag v-else-if="credential.status?.healthy === false" type="danger" effect="dark"
									size="small">No</el-tag>
								<span v-else style="color: #909399">-</span>
							</el-descriptions-item>
						</el-descriptions>
					</el-card>
				</el-col>
			</el-row>

		</template>

		<!-- Edit YAML Modal -->
		<EditYamlModal v-model:visible="yamlModalVisible" :yaml-content="yamlContent" @save="handleYamlSave" />
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
