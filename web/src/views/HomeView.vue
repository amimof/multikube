<script setup lang="ts">
import { onMounted } from 'vue'
import { Monitor, Guide, Key } from '@element-plus/icons-vue'
import { useBackendStore } from '@/stores/backend'
import { useRouteStore } from '@/stores/route'
import { useCaStore } from '@/stores/ca'

const backendStore = useBackendStore()
const routeStore = useRouteStore()
const caStore = useCaStore()

onMounted(() => {
  backendStore.fetchBackends().catch(() => {})
  routeStore.fetchRoutes().catch(() => {})
  caStore.fetchCas().catch(() => {})
})
</script>

<template>
  <div>
    <h1 style="margin-bottom: 24px">Dashboard</h1>
    <el-row :gutter="20">
      <el-col :span="8">
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
      <el-col :span="8">
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
      <el-col :span="8">
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
    </el-row>
  </div>
</template>
