<script setup lang="ts">
import { ref, watch } from 'vue'
import { Codemirror } from 'vue-codemirror'
import { yaml as yamlLang } from '@codemirror/lang-yaml'
import { oneDark } from '@codemirror/theme-one-dark'
import { parse as yamlParse } from 'yaml'

const visible = defineModel<boolean>('visible', { required: true })

const props = defineProps<{
	yamlContent: string
}>()

const emit = defineEmits<{
	save: [parsed: unknown]
}>()

const localYaml = ref('')
const parseError = ref('')
const cmExtensions = [yamlLang(), oneDark]

watch(
	() => props.yamlContent,
	(val) => {
		localYaml.value = val
		parseError.value = ''
	},
	{ immediate: true },
)

function handleSave() {
	parseError.value = ''
	try {
		const parsed = yamlParse(localYaml.value)
		if (!parsed || typeof parsed !== 'object') {
			parseError.value = 'YAML must represent an object'
			return
		}
		emit('save', parsed)
	} catch (err) {
		parseError.value = err instanceof Error ? err.message : 'Invalid YAML'
	}
}

function handleClose() {
	visible.value = false
}

function handleOpened() {
	localYaml.value = props.yamlContent
	parseError.value = ''
}
</script>

<template>
	<el-dialog :model-value="visible" @update:model-value="(val: boolean) => (visible = val)" title="Edit YAML"
		width="750px" :close-on-click-modal="false" @opened="handleOpened" destroy-on-close>
		<el-alert v-if="parseError" type="error" :title="parseError" show-icon :closable="true"
			@close="parseError = ''" style="margin-bottom: 12px" />

		<div class="yaml-editor">
			<Codemirror v-model="localYaml" :extensions="cmExtensions" :style="{ fontSize: '13px' }"
				placeholder="Edit YAML..." />
		</div>

		<template #footer>
			<el-button @click="handleClose">Cancel</el-button>
			<el-button type="primary" @click="handleSave">Save</el-button>
		</template>
	</el-dialog>
</template>

<style scoped>
.yaml-editor :deep(.cm-editor) {
	border-radius: 4px;
	min-height: 300px;
	max-height: 500px;
	overflow: auto;
}

.yaml-editor :deep(.cm-gutters) {
	border-radius: 4px 0 0 4px;
}
</style>
