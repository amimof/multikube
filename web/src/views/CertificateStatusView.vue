<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed, watch, toRaw } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Refresh, Document } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useCertificateStore } from '@/stores/certificate'
import { stringify as yamlStringify } from 'yaml'
import { Codemirror } from 'vue-codemirror'
import { yaml as yamlLang } from '@codemirror/lang-yaml'
import { oneDark } from '@codemirror/theme-one-dark'
import { EditorState } from '@codemirror/state'
import { formatDate, formatDateFull } from '@/utils/format'
import type { V1Certificate } from '@/generated/certificate'
import LabelEditor from '@/components/LabelEditor.vue'
import MetadataDisplay from '@/components/MetadataDisplay.vue'
import EditYamlModal from '@/components/EditYamlModal.vue'

const route = useRoute()
const router = useRouter()
const certificateStore = useCertificateStore()

const certName = computed(() => route.params.name as string)
const certificate = computed(() => certificateStore.current)

const saving = ref(false)
const yamlModalVisible = ref(false)

const form = ref<V1Certificate>({})

watch(certificate, (val) => {
	if (val) {
		form.value = structuredClone(toRaw(val))
		if (!form.value.config) form.value.config = {}
	}
}, { immediate: true })

const formLabels = computed({
	get: () => form.value.meta?.labels ?? {},
	set: (val: Record<string, string>) => {
		if (form.value.meta) {
			form.value.meta.labels = val
		}
	},
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
		await certificateStore.updateCertificate(form.value)
		await certificateStore.fetchCertificate(certName.value)
		ElMessage.success('Certificate updated')
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Save failed')
	} finally {
		saving.value = false
	}
}

async function handleYamlSave(parsed: unknown) {
	saving.value = true
	try {
		const resource = parsed as V1Certificate
		if (!resource.meta?.name) {
			resource.meta = { ...resource.meta, name: certName.value }
		}
		await certificateStore.updateCertificate(resource)
		await certificateStore.fetchCertificate(certName.value)
		ElMessage.success('Certificate updated from YAML')
		yamlModalVisible.value = false
	} catch (err) {
		ElMessage.error(err instanceof Error ? err.message : 'Save failed')
	} finally {
		saving.value = false
	}
}

function handleRefresh() {
	certificateStore.fetchCertificate(certName.value).catch(() => { })
}

function goBack() {
	router.push('/certificates')
}

onMounted(() => {
	certificateStore.fetchCertificate(certName.value).catch(() => { })
})

onUnmounted(() => {
	certificateStore.clearCurrent()
})
</script>

<template>
	<div>
		<!-- Header -->
		<el-row justify="space-between" align="middle" style="margin-bottom: 16px">
			<el-col :span="12">
				<div style="display: flex; align-items: center; gap: 12px">
					<el-button :icon="ArrowLeft" @click="goBack" text>Certificates</el-button>
					<h2 style="margin: 0">{{ certName }}</h2>
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

		<el-alert v-if="certificateStore.error" :title="certificateStore.error" type="error" show-icon
			style="margin-bottom: 16px" />

		<!-- Loading -->
		<el-card v-if="certificateStore.loading" shadow="never" style="margin-bottom: 16px">
			<el-skeleton :rows="6" animated />
		</el-card>

		<!-- Not found -->
		<el-empty v-else-if="!certificate" :description="`Certificate '${certName}' not found`">
			<el-button type="primary" @click="goBack">Back to Certificates</el-button>
		</el-empty>

		<!-- Content -->
		<template v-else>
			<el-card shadow="never" style="margin-bottom: 16px">
				<template #header>
					<span style="font-weight: 600">Configuration</span>
				</template>

				<!-- Metadata (read-only) -->
				<el-collapse style="margin-bottom: 20px">
					<el-collapse-item title="Metadata" name="metadata">
						<MetadataDisplay :meta="certificate.meta" />
					</el-collapse-item>
				</el-collapse>

				<el-form label-width="120px" label-position="right">
					<el-form-item label="Name">
						<el-input :model-value="form.meta?.name" disabled />
					</el-form-item>

					<el-form-item label="Labels">
						<LabelEditor v-model="formLabels" />
					</el-form-item>

					<el-divider content-position="left">Certificate</el-divider>

					<el-form-item label="Enabled">
						<el-switch v-model="form.config!.enabled" />
					</el-form-item>

					<el-form-item label="Data">
						<el-input v-model="form.config!.certificateData" type="textarea" :rows="8"
							:input-style="{ fontFamily: 'monospace', fontSize: '13px' }"
							placeholder="Paste PEM certificate data here" />
					</el-form-item>

					<el-divider content-position="left">Private Key</el-divider>

					<el-form-item label="Data">
						<el-input v-model="form.config!.keyData" type="textarea" :rows="8"
							:input-style="{ fontFamily: 'monospace', fontSize: '13px' }"
							placeholder="Paste PEM private key data here" />
					</el-form-item>
				</el-form>
			</el-card>

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
