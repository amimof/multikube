
<script setup lang="ts">
import { ref } from 'vue';
import { useAuthStore } from '@/stores';
import { useRouter } from 'vue-router';

const username = ref('');
const password = ref('');

const auth = useAuthStore();
const router = useRouter();

function handleLogin() {
  auth.login(username.value, password.value).then(() => {
    router.push('/');
  }).catch((err) => {
    console.log(err);
  });
}

</script>

<template>
  <div class="row">
    <div class="col-md-4 offset-md-4">
      <div class="card">
        <div class="card-body">
          <div class="mb-3 row">
            <label for="staticEmail" class="col-sm-4 col-form-label">Email</label>
            <div class="col-sm-8">
              <input type="text" v-model="username" class="form-control" id="staticEmail">
            </div>
          </div>
          <div class="mb-3 row">
            <label for="inputPassword" class="col-sm-4 col-form-label">Password</label>
            <div class="col-sm-8">
              <input type="password" v-model="password" class="form-control" id="inputPassword">
            </div>
          </div>
          <div class="row">
            <div class="col-sm-12 d-grid gap-2">
              <button type="button" class="btn btn-primary" @click="handleLogin()">Login</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
