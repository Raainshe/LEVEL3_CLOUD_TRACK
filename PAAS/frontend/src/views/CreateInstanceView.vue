<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import { useAuth } from '@/stores/auth'
import type { CreateInstanceRequest } from '@/types/instance'

const router = useRouter()
const toast = useToastStore()
const { user } = useAuth()
const loading = ref(false)
const error = ref<string | null>(null)

const isAdmin = computed(() => user.value?.isAdmin ?? false)

const form = ref<CreateInstanceRequest>({
  name: '',
  namespace: 'default',
  redisReplicas: 3,
  sentinelReplicas: 3,
})

async function onSubmit() {
  loading.value = true
  error.value = null
  try {
    const payload: CreateInstanceRequest = {
      redisReplicas: form.value.redisReplicas,
      sentinelReplicas: form.value.sentinelReplicas,
    }
    if (form.value.name?.trim()) payload.name = form.value.name.trim()
    if (isAdmin.value && form.value.namespace?.trim()) payload.namespace = form.value.namespace.trim()

    await api.post('/api/instances', payload)
    toast.show('Instance created successfully')
    router.push('/instances')
  } catch (e: unknown) {
    const err = e as { response?: { data?: { error?: string; details?: string } } }
    const msg = err.response?.data?.error ?? (e instanceof Error ? e.message : 'Failed to create instance')
    const details = err.response?.data?.details
    error.value = details ? `${msg}: ${details}` : msg
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="mb-4">Create Instance</h1>

    <form @submit.prevent="onSubmit" class="card shadow-sm">
      <div class="card-body">
        <div v-if="error" class="alert alert-danger" role="alert">
          {{ error }}
        </div>

        <div class="mb-3">
          <label for="name" class="form-label">Name (optional)</label>
          <input
            id="name"
            v-model="form.name"
            type="text"
            class="form-control"
            placeholder="Leave empty to auto-generate"
          />
        </div>

        <div v-if="isAdmin" class="mb-3">
          <label for="namespace" class="form-label">Namespace (optional)</label>
          <input
            id="namespace"
            v-model="form.namespace"
            type="text"
            class="form-control"
            placeholder="default"
          />
          <div class="form-text">As an admin you can create instances in any namespace.</div>
        </div>
        <div v-else class="mb-3">
          <p class="text-muted small mb-0">
            Instances will be created in your namespace (tied to your account).
          </p>
        </div>

        <div class="mb-3">
          <label for="redisReplicas" class="form-label">Redis replicas</label>
          <input
            id="redisReplicas"
            v-model.number="form.redisReplicas"
            type="number"
            min="1"
            max="10"
            class="form-control"
            required
          />
          <div class="form-text">Number of Redis replicas (1–10). Defaults to 3 if empty.</div>
        </div>

        <div class="mb-4">
          <label for="sentinelReplicas" class="form-label">Sentinel replicas</label>
          <input
            id="sentinelReplicas"
            v-model.number="form.sentinelReplicas"
            type="number"
            min="1"
            max="10"
            class="form-control"
            required
          />
          <div class="form-text">Number of Sentinel replicas (1–10). Defaults to 3 if empty.</div>
        </div>

        <div class="d-flex gap-2">
          <button type="submit" class="btn btn-primary" :disabled="loading">
            {{ loading ? 'Creating...' : 'Create Instance' }}
          </button>
          <router-link to="/instances" class="btn btn-outline-secondary">Cancel</router-link>
        </div>
      </div>
    </form>
  </div>
</template>
