<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
	Document,
	Guide,
	HomeFilled,
	Key,
	List,
	Lock,
	Monitor,
	Moon,
	Plus,
	Sunny,
	SwitchButton,
	Upload,
	UserFilled,
	ArrowDown,
	DocumentAdd
} from '@element-plus/icons-vue'
import ApplyResourcesModal from '@/components/ApplyResourcesModal.vue'
import ImportKubeconfigModal from '@/components/ImportKubeconfigModal.vue'
import { useTheme } from '@/composables/useTheme'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const showApplyModal = ref(false)
const showImportModal = ref(false)
const { isDark } = useTheme()

function logout() {
	authStore.logout()
	void router.push('/login')
}
</script>

<template>
	<el-container class="app-shell">
		<el-aside width="220px" :style="{ backgroundColor: 'var(--sidebar-bg)' }">
			<div class="brand-block">
				<h2>Multikube</h2>
				<p>Console</p>
			</div>
			<el-menu :default-active="route.path" router :background-color="'var(--sidebar-bg)'" text-color="#ffffffa6"
				active-text-color="#ffffff" style="border-right: none">
				<el-menu-item index="/">
					<el-icon>
						<HomeFilled />
					</el-icon>
					<span>Home</span>
				</el-menu-item>
				<el-menu-item index="/backends">
					<el-icon>
						<Monitor />
					</el-icon>
					<span>Backends</span>
				</el-menu-item>
				<el-menu-item index="/routes">
					<el-icon>
						<Guide />
					</el-icon>
					<span>Routes</span>
				</el-menu-item>
				<el-menu-item index="/cas">
					<el-icon>
						<Key />
					</el-icon>
					<span>Certificate Authorities</span>
				</el-menu-item>
				<el-menu-item index="/credentials">
					<el-icon>
						<Lock />
					</el-icon>
					<span>Credentials</span>
				</el-menu-item>
				<el-menu-item index="/certificates">
					<el-icon>
						<Document />
					</el-icon>
					<span>Certificates</span>
				</el-menu-item>
				<el-menu-item index="/policies">
					<el-icon>
						<List />
					</el-icon>
					<span>Policies</span>
				</el-menu-item>
			</el-menu>
		</el-aside>

		<el-container direction="vertical">
			<el-header class="app-header">
				<div class="header-title-group">
					<div class="status-dot" />
					<span class="header-title">Management Console</span>
				</div>

				<div class="header-actions">
					<el-switch v-model="isDark" :active-action-icon="Moon" :inactive-action-icon="Sunny" />
					<el-dropdown trigger="click">
						<el-button type="success">
							Create<el-icon class="el-icon--right">
								<ArrowDown />
							</el-icon>
						</el-button>
						<template #dropdown>
							<el-dropdown-menu>
								<el-dropdown-item @click="showImportModal = true" :icon="Upload">Import kubeconfig</el-dropdown-item>
								<el-dropdown-item @click="showApplyModal = true" :icon="DocumentAdd">From YAML</el-dropdown-item>
							</el-dropdown-menu>
						</template>
					</el-dropdown>
					<el-dropdown>
						<el-button plain>
							<el-icon>
								<UserFilled />
							</el-icon>
							<span class="user-label">{{ authStore.username ?? 'Account' }}</span>
							<el-icon class="el-icon--right">
								<ArrowDown />
							</el-icon>
						</el-button>
						<template #dropdown>
							<el-dropdown-menu>
								<el-dropdown-item @click="logout" :icon="Lock">Logout</el-dropdown-item>
							</el-dropdown-menu>
						</template>
					</el-dropdown>
				</div>
			</el-header>

			<el-main class="app-main">
				<RouterView />
			</el-main>
		</el-container>

		<ApplyResourcesModal v-model:visible="showApplyModal" />
		<ImportKubeconfigModal v-model:visible="showImportModal" />
	</el-container>
</template>

<style scoped>
.app-shell {
	height: 100vh;
}

.brand-block {
	padding: 22px 20px 18px;
	border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.brand-block h2 {
	margin: 0;
	color: #fff;
	font-size: 18px;
	letter-spacing: 0.08em;
	text-transform: uppercase;
}

.brand-block p {
	margin: 6px 0 0;
	color: rgba(255, 255, 255, 0.56);
	font-size: 12px;
	letter-spacing: 0.16em;
	text-transform: uppercase;
}

:deep(.el-menu-item.is-active) {
	background-color: var(--el-color-primary) !important;
}

.app-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	height: 56px;
	padding: 0 20px;
	background-color: var(--header-bg);
	border-bottom: 1px solid var(--header-border);
	transition: background-color 0.3s, border-color 0.3s;
}

.header-title-group {
	display: flex;
	align-items: center;
	gap: 10px;
}

.status-dot {
	width: 8px;
	height: 8px;
	border-radius: 999px;
	background: linear-gradient(180deg, #37f499 0%, #14b86d 100%);
	box-shadow: 0 0 18px rgba(55, 244, 153, 0.5);
}

.header-title {
	font-size: 13px;
	font-weight: 600;
	letter-spacing: 0.08em;
	text-transform: uppercase;
	color: var(--color-text, #0f172a);
}

.header-actions {
	display: flex;
	align-items: center;
	gap: 8px;
}

.user-label {
	margin: 0 8px;
	max-width: 180px;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.app-main {
	padding: 20px;
	background-color: var(--main-bg);
	transition: background-color 0.3s;
}

@media (max-width: 960px) {
	.app-header {
		height: auto;
		padding: 12px 16px;
		align-items: flex-start;
		flex-direction: column;
		gap: 12px;
	}

	.header-actions {
		width: 100%;
		flex-wrap: wrap;
	}
}
</style>
