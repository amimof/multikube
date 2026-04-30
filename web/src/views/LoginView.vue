<script setup lang="ts">
import { computed, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Moon, Sunny, User, Lock } from '@element-plus/icons-vue'
import { useTheme } from '@/composables/useTheme'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { isDark } = useTheme()

const form = reactive({
	username: '',
	password: '',
})

const redirectTarget = computed(() => {
	const redirect = route.query.redirect
	return typeof redirect === 'string' && redirect.length > 0 ? redirect : '/'
})

const canSubmit = computed(() => form.username.trim().length > 0 && form.password.length > 0)

async function submit() {
	if (!canSubmit.value || authStore.loginLoading) {
		return
	}

	try {
		await authStore.login({
			username: form.username.trim(),
			password: form.password,
		})

		await router.push(redirectTarget.value)
	} catch {
		// The store exposes the backend error text for the form.
	}
}
</script>

<template>
	<div class="login-page">
		<div class="login-shell">
			<div class="login-form-shell">
				<div class="brand-eyebrow">
					<span class="brand-eyebrow-dot" />
					Multikube Control Plane
				</div>

				<div class="login-form-header">
					<div>
						<div class="console-kicker">Console Entry</div>
						<h2>Login</h2>
					</div>

					<el-switch v-model="isDark" :active-action-icon="Moon" :inactive-action-icon="Sunny" />
				</div>

				<el-alert v-if="authStore.error" :title="authStore.error" type="error" :closable="false" show-icon
					class="login-alert" />

				<el-form label-position="top" @submit.prevent="submit">
					<el-form-item label="Username">
						<el-input v-model="form.username" :prefix-icon="User" size="large" />
					</el-form-item>

					<el-form-item label="Password">
						<el-input v-model="form.password" :prefix-icon="Lock" type="password" show-password size="large"
							@keyup.enter="submit" />
					</el-form-item>

					<el-button class="login-submit" type="primary" size="large" :loading="authStore.loginLoading"
						:disabled="!canSubmit" @click="submit">
						Login
					</el-button>
				</el-form>
			</div>
		</div>
	</div>
</template>

<style scoped>
.login-page {
	min-height: 100vh;
	min-height: 100dvh;
	box-sizing: border-box;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 32px 24px;
	color: var(--color-text, #0f172a);
	background:
		radial-gradient(circle at top left, var(--login-bg-accent-1), transparent 32%),
		radial-gradient(circle at bottom right, var(--login-bg-accent-2), transparent 28%),
		var(--login-bg-base);
}

.login-shell {
	width: 100%;
	display: flex;
	justify-content: center;
}

.brand-eyebrow {
	display: inline-flex;
	align-items: center;
	gap: 10px;
	width: fit-content;
	margin-bottom: 18px;
	padding: 8px 12px;
	border-radius: 999px;
	background: var(--login-chip-bg);
	color: var(--login-chip-text);
	font-size: 12px;
	font-weight: 600;
	letter-spacing: 0.1em;
	text-transform: uppercase;
}

.brand-eyebrow-dot {
	width: 8px;
	height: 8px;
	border-radius: 999px;
	background: #37f499;
	box-shadow: 0 0 18px rgba(55, 244, 153, 0.5);
}

.login-form-shell {
	width: 100%;
	max-width: 420px;
	padding: 32px;
	border-radius: 24px;
	border: 1px solid var(--login-card-border);
	background: var(--login-card-bg);
	box-shadow: var(--login-card-shadow);
	backdrop-filter: blur(20px);
}

.login-form-header {
	display: flex;
	align-items: flex-start;
	justify-content: space-between;
	gap: 16px;
	margin-bottom: 24px;
}

.console-kicker {
	color: #905ce1;
	font-size: 12px;
	font-weight: 700;
	letter-spacing: 0.16em;
	text-transform: uppercase;
}

.login-form-header h2 {
	margin: 8px 0 8px;
	color: var(--login-heading);
	font-size: 30px;
}

.login-form-header p {
	margin: 0;
	color: var(--login-muted);
}

.login-alert {
	margin-bottom: 20px;
}

.login-submit {
	width: 100%;
	margin-top: 8px;
}

.login-footer {
	display: flex;
	flex-direction: column;
	gap: 6px;
	margin-top: 20px;
	color: var(--color-text-muted);
	font-size: 12px;
}

.login-footer code {
	color: var(--login-code);
}

@media (max-width: 1100px) {
	.login-page {
		padding: 24px;
	}
}

@media (max-width: 720px) {
	.login-form-shell {
		padding: 24px;
	}

	.login-form-header {
		align-items: stretch;
		flex-direction: column;
	}
}
</style>
