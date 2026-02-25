<script setup lang="ts">
import { RouterView } from 'vue-router'
import AppHeader from './components/AppHeader.vue'
import AppFooter from './components/AppFooter.vue'
import AppToast from './components/AppToast.vue'
</script>

<template>
  <div class="app-layout d-flex flex-column min-vh-100">
    <AppToast />
    <AppHeader />
    <main class="flex-grow-1 py-4">
      <div class="content-container">
        <RouterView />
      </div>
    </main>
    <AppFooter />
  </div>
</template>
