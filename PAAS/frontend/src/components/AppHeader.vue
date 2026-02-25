<script setup lang="ts">
import { RouterLink, useRouter } from 'vue-router'
import { useAuth } from '@/stores/auth'

const router = useRouter()
const { isAuthenticated, clearToken } = useAuth()

function onLogout() {
  clearToken()
  router.push({ name: 'login' })
}
</script>

<template>
  <nav class="navbar navbar-expand-lg navbar-dark bg-primary">
    <div class="container-fluid">
      <RouterLink to="/instances" class="navbar-brand d-flex align-items-center">
        <img src="/logo.png" alt="RedisPaaS on STACKIT" class="me-2" height="36" />
        <span class="d-none d-sm-inline">PAAS LEVEL3</span>
      </RouterLink>
      <button
        class="navbar-toggler"
        type="button"
        data-bs-toggle="collapse"
        data-bs-target="#navbarNav"
        aria-controls="navbarNav"
        aria-expanded="false"
        aria-label="Toggle navigation"
      >
        <span class="navbar-toggler-icon"></span>
      </button>
      <div class="collapse navbar-collapse" id="navbarNav">
        <ul class="navbar-nav me-auto">
          <li class="nav-item">
            <RouterLink to="/instances" class="nav-link" exact-active-class="active">Instances</RouterLink>
          </li>
          <li class="nav-item">
            <RouterLink to="/instances/new" class="nav-link" exact-active-class="active">Create Instance</RouterLink>
          </li>
          <li class="nav-item">
            <RouterLink to="/audit-logs" class="nav-link" exact-active-class="active">Audit Logs</RouterLink>
          </li>
          <li class="nav-item">
            <RouterLink to="/service-logs" class="nav-link" exact-active-class="active">Service Logs</RouterLink>
          </li>
        </ul>
        <div class="d-flex">
          <button
            v-if="isAuthenticated"
            type="button"
            class="btn btn-outline-light"
            @click="onLogout"
          >
            Logout
          </button>
          <RouterLink
            v-else
            to="/"
            class="btn btn-light"
            active-class="active"
          >
            Login
          </RouterLink>
        </div>
      </div>
    </div>
  </nav>
</template>
