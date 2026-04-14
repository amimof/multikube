<script setup lang="ts">
import { onMounted } from 'vue'
import { Monitor, Guide, Key, Lock, Document, List } from '@element-plus/icons-vue'
import { useBackendStore } from '@/stores/backend'
import { useRouteStore } from '@/stores/route'
import { useCaStore } from '@/stores/ca'
import { useCredentialStore } from '@/stores/credential'
import { useCertificateStore } from '@/stores/certificate'
import { usePolicyStore } from '@/stores/policy'

const backendStore = useBackendStore()
const routeStore = useRouteStore()
const caStore = useCaStore()
const credentialStore = useCredentialStore()
const certificateStore = useCertificateStore()
const policyStore = usePolicyStore()

onMounted(() => {
  backendStore.fetchBackends().catch(() => {})
  routeStore.fetchRoutes().catch(() => {})
  caStore.fetchCas().catch(() => {})
  credentialStore.fetchCredentials().catch(() => {})
  certificateStore.fetchCertificates().catch(() => {})
  policyStore.fetchPolicies().catch(() => {})
})
</script>

<template>
  <div>
    <h1 style="margin-bottom: 24px">Dashboard</h1>
    <el-row :gutter="20">
      <el-col :xs="24" :sm="12" :md="8" style="margin-bottom: 20px">
        <router-link to="/backends" style="text-decoration: none">
          <el-card shadow="hover">
            <template #header>
              <div style="display: flex; align-items: center; gap: 8px">
                <el-icon :size="20"><Monitor /></el-icon>
                <span>Backends</span>
              </div>
            </template>
            <div style="text-align: center">
              <el-statistic :value="backendStore.items.length" />
            </div>
          </el-card>
        </router-link>
      </el-col>
      <el-col :xs="24" :sm="12" :md="8" style="margin-bottom: 20px">
        <router-link to="/routes" style="text-decoration: none">
          <el-card shadow="hover">
            <template #header>
              <div style="display: flex; align-items: center; gap: 8px">
                <el-icon :size="20"><Guide /></el-icon>
                <span>Routes</span>
              </div>
            </template>
            <div style="text-align: center">
              <el-statistic :value="routeStore.items.length" />
            </div>
          </el-card>
        </router-link>
      </el-col>
      <el-col :xs="24" :sm="12" :md="8" style="margin-bottom: 20px">
        <router-link to="/cas" style="text-decoration: none">
          <el-card shadow="hover">
            <template #header>
              <div style="display: flex; align-items: center; gap: 8px">
                <el-icon :size="20"><Key /></el-icon>
                <span>Certificate Authorities</span>
              </div>
            </template>
            <div style="text-align: center">
              <el-statistic :value="caStore.items.length" />
            </div>
          </el-card>
        </router-link>
      </el-col>
      <el-col :xs="24" :sm="12" :md="8" style="margin-bottom: 20px">
        <router-link to="/credentials" style="text-decoration: none">
          <el-card shadow="hover">
            <template #header>
              <div style="display: flex; align-items: center; gap: 8px">
                <el-icon :size="20"><Lock /></el-icon>
                <span>Credentials</span>
              </div>
            </template>
            <div style="text-align: center">
              <el-statistic :value="credentialStore.items.length" />
            </div>
          </el-card>
        </router-link>
      </el-col>
      <el-col :xs="24" :sm="12" :md="8" style="margin-bottom: 20px">
        <router-link to="/certificates" style="text-decoration: none">
          <el-card shadow="hover">
            <template #header>
              <div style="display: flex; align-items: center; gap: 8px">
                <el-icon :size="20"><Document /></el-icon>
                <span>Certificates</span>
              </div>
            </template>
            <div style="text-align: center">
              <el-statistic :value="certificateStore.items.length" />
            </div>
          </el-card>
        </router-link>
      </el-col>
      <el-col :xs="24" :sm="12" :md="8" style="margin-bottom: 20px">
        <router-link to="/policies" style="text-decoration: none">
          <el-card shadow="hover">
            <template #header>
              <div style="display: flex; align-items: center; gap: 8px">
                <el-icon :size="20"><List /></el-icon>
                <span>Policies</span>
              </div>
            </template>
            <div style="text-align: center">
              <el-statistic :value="policyStore.items.length" />
            </div>
          </el-card>
        </router-link>
      </el-col>
    </el-row>
  </div>
</template>
