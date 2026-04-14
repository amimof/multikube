<script setup lang="ts">
import { useRoute } from 'vue-router'
import {
  HomeFilled,
  Monitor,
  Guide,
  Key,
  Lock,
  Document,
  List,
} from '@element-plus/icons-vue'

const route = useRoute()
</script>

<template>
  <el-container style="height: 100vh">
    <el-aside width="220px" style="background-color: #001529">
      <div style="padding: 20px; text-align: center">
        <h2 style="color: #fff; margin: 0; font-size: 18px">Multikube</h2>
      </div>
      <el-menu
        :default-active="route.path"
        router
        background-color="#001529"
        text-color="#ffffffa6"
        active-text-color="#ffffff"
        style="border-right: none"
      >
        <el-menu-item index="/">
          <el-icon><HomeFilled /></el-icon>
          <span>Home</span>
        </el-menu-item>
        <el-menu-item index="/backends">
          <el-icon><Monitor /></el-icon>
          <span>Backends</span>
        </el-menu-item>
        <el-menu-item index="/routes">
          <el-icon><Guide /></el-icon>
          <span>Routes</span>
        </el-menu-item>
        <el-menu-item index="/cas">
          <el-icon><Key /></el-icon>
          <span>Certificate Authorities</span>
        </el-menu-item>
        <el-menu-item index="/credentials">
          <el-icon><Lock /></el-icon>
          <span>Credentials</span>
        </el-menu-item>
        <el-menu-item index="/certificates">
          <el-icon><Document /></el-icon>
          <span>Certificates</span>
        </el-menu-item>
        <el-menu-item index="/policies">
          <el-icon><List /></el-icon>
          <span>Policies</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-main style="background-color: #f5f7fa; padding: 20px">
        <RouterView />
      </el-main>
    </el-container>
  </el-container>
</template>

<style>
body {
  margin: 0;
  padding: 0;
}

.el-menu-item.is-active {
  background-color: #1890ff !important;
}

.clickable-row {
  cursor: pointer;
}
</style>
