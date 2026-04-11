<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'

const navigationItems = [
  { label: 'Backends', to: '/backends' },
  { label: "CA's", to: '/cas' },
  { label: 'Certificates', to: '/certificates' },
  { label: 'Credentials', to: '/credentials' },
  { label: 'Policies', to: '/policies' },
  { label: 'Routes', to: '/routes' },
]
</script>

<template>
  <div class="container-fluid px-0">
    <nav class="navbar navbar-expand-lg bg-body-tertiary border-bottom shadow-sm">
      <div class="container">
        <RouterLink class="navbar-brand fw-semibold" to="/">Multikube Admin</RouterLink>
        <button
          class="navbar-toggler"
          type="button"
          data-bs-toggle="collapse"
          data-bs-target="#navbarText"
          aria-controls="navbarText"
          aria-expanded="false"
          aria-label="Toggle navigation"
        >
          <span class="navbar-toggler-icon"></span>
        </button>
        <div class="collapse navbar-collapse" id="navbarText">
          <ul class="navbar-nav me-auto mb-2 mb-lg-0 gap-lg-1">
            <li class="nav-item">
              <RouterLink class="nav-link" to="/">Home</RouterLink>
            </li>
            <li v-for="item in navigationItems" :key="item.to" class="nav-item">
              <RouterLink class="nav-link" :to="item.to">{{ item.label }}</RouterLink>
            </li>
          </ul>
          <span class="navbar-text">
            Resource console
          </span>
        </div>
      </div>
    </nav>
  </div>

  <div class="container">
    <main class="main my-4">
      <RouterView />
    </main>

    <footer class="text-center py-3 mt-auto text-body-secondary small">
      <span>&copy; 2026 Multikube</span>
    </footer>
  </div>
</template>

<style scoped></style>
