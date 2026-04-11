<script setup lang="ts">
const resources = [
  {
    title: 'Backends',
    description: 'Configure upstream targets, TLS behavior, and load balancing strategy.',
    to: '/backends',
  },
  {
    title: "CA's",
    description: 'Manage certificate authority material used to validate backend connections.',
    to: '/cas',
  },
  {
    title: 'Certificates',
    description: 'Store certificate and key pairs for downstream or upstream authentication.',
    to: '/certificates',
  },
  {
    title: 'Credentials',
    description: 'Manage token, basic auth, and client certificate based credentials.',
    to: '/credentials',
  },
  {
    title: 'Policies',
    description: 'Edit access rules and conditions applied to resource operations.',
    to: '/policies',
  },
  {
    title: 'Routes',
    description: 'Define request matching rules and map them to backend references.',
    to: '/routes',
  },
]
</script>

<template>
  <section class="d-flex flex-column gap-4">
    <div class="p-4 p-lg-5 bg-body-tertiary rounded-4 border shadow-sm">
      <p class="text-uppercase small text-body-secondary mb-2">Admin Console</p>
      <h1 class="display-6 fw-semibold mb-3">Manage Multikube resources from one frontend.</h1>
      <p class="lead mb-0 text-body-secondary">
        Each section provides list, create, update, and delete flows against the `/api/v1` API.
      </p>
    </div>

    <div class="row g-4">
      <div v-for="resource in resources" :key="resource.to" class="col-md-6 col-xl-4">
        <RouterLink :to="resource.to" class="card h-100 border-0 shadow-sm text-decoration-none text-body">
          <div class="card-body d-flex flex-column gap-3">
            <div>
              <h2 class="h5 mb-2">{{ resource.title }}</h2>
              <p class="mb-0 text-body-secondary">{{ resource.description }}</p>
            </div>

            <div class="mt-auto fw-semibold">Open section</div>
          </div>
        </RouterLink>
      </div>
    </div>
  </section>
</template>
