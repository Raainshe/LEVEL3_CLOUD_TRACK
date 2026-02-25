<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import type { RedisInstance, ListInstancesResponse } from '@/types/instance'
import ConfirmModal from '@/components/ConfirmModal.vue'

const REFRESH_INTERVAL_MS = 10_000

const router = useRouter()
const toast = useToastStore()
const instances = ref<RedisInstance[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const showDeleteModal = ref(false)
const instanceToDelete = ref<string | null>(null)
const instanceToDeleteNamespace = ref<string>('default')
let refreshTimer: ReturnType<typeof setInterval> | null = null

const deleteModalMessage = computed(
  () =>
    `Delete instance "${instanceToDelete.value ?? ''}"? This cannot be undone.`
)

function openDeleteModal(inst: RedisInstance) {
  instanceToDelete.value = inst.id
  instanceToDeleteNamespace.value = inst.namespace ?? 'default'
  showDeleteModal.value = true
}

async function performDelete() {
  const id = instanceToDelete.value
  if (!id) return
  const ns = instanceToDeleteNamespace.value || 'default'
  try {
    await api.delete(`/api/instances/${id}?namespace=${encodeURIComponent(ns)}`)
    toast.show('Instance deleted successfully')
    await fetchInstances()
  } catch (e) {
    alert(e instanceof Error ? e.message : 'Failed to delete instance')
  } finally {
    instanceToDelete.value = null
  }
}

async function fetchInstances(silent = false) {
  if (!silent) {
    loading.value = true
    error.value = null
  }
  try {
    const { data } = await api.get<ListInstancesResponse>('/api/instances')
    instances.value = data.instances ?? []
  } catch (e) {
    if (!silent) {
      error.value = e instanceof Error ? e.message : 'Failed to load instances'
      instances.value = []
    }
  } finally {
    if (!silent) loading.value = false
  }
}

function viewInstance(inst: RedisInstance) {
  const ns = inst.namespace ?? 'default'
  router.push(`/instances/${inst.id}?namespace=${encodeURIComponent(ns)}`)
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

function statusClass(status: string) {
  const s = (status ?? '').toLowerCase()
  if (s === 'running' || s === 'ready') return 'success'
  if (s === 'provisioning' || s === 'pending') return 'warning'
  if (s === 'error' || s === 'failed') return 'danger'
  return 'secondary'
}

function statusIcon(status: string) {
  const s = (status ?? '').toLowerCase()
  if (s === 'running' || s === 'ready') return ['fas', 'circle-check']
  if (s === 'provisioning' || s === 'pending') return ['fas', 'spinner']
  if (s === 'error' || s === 'failed') return ['fas', 'circle-xmark']
  return ['fas', 'circle-question']
}

function statusTooltip(status: string) {
  const s = (status ?? '').toLowerCase()
  if (s === 'running' || s === 'ready')
    return 'Instance is running and ready to accept connections'
  if (s === 'provisioning' || s === 'pending')
    return 'Instance is being created or is pending deployment'
  if (s === 'error' || s === 'failed')
    return 'Instance has encountered an error or failed to start'
  return 'Unknown status – the instance state could not be determined'
}

onMounted(() => {
  fetchInstances()
  refreshTimer = setInterval(() => fetchInstances(true), REFRESH_INTERVAL_MS)
})

onUnmounted(() => {
  if (refreshTimer != null) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
})
</script>

<template>
  <div>
    <h1 class="mb-4">Redis Instances</h1>

    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
    </div>

    <div v-else-if="error" class="alert alert-danger" role="alert">
      {{ error }}
    </div>

    <div v-else-if="instances.length === 0" class="alert alert-info" role="alert">
      No Redis instances found. Create one to get started. <router-link to="/instances/new" class="btn btn-primary btn-sm">Create Instance</router-link>
    </div>

    <div v-else class="row row-cols-1 row-cols-md-2 row-cols-lg-3 g-4">
      <div v-for="inst in instances" :key="inst.id" class="col">
        <div class="card h-100 shadow-sm">
          <div class="card-header d-flex justify-content-between align-items-center bg-white">
            <span class="fw-bold redis-accent">{{ inst.name }}</span>
            <span
              class="status-indicator d-flex align-items-center gap-2"
              :class="`text-${statusClass(inst.status)}`"
            >
              <span class="status-icon-wrapper">
                <FontAwesomeIcon
                  :icon="statusIcon(inst.status)"
                  size="xl"
                  :spin="['provisioning', 'pending'].includes((inst.status ?? '').toLowerCase())"
                />
                <span class="status-tooltip">{{ statusTooltip(inst.status) }}</span>
              </span>
            </span>
          </div>
          <div class="card-body">
            <ul class="list-unstyled mb-0 small">
              <li><strong>Namespace:</strong> {{ inst.namespace }}</li>
              <li><strong>Redis replicas:</strong> {{ inst.redisReplicas }}</li>
              <li><strong>Sentinel replicas:</strong> {{ inst.sentinelReplicas }}</li>
              <li><strong>Created:</strong> {{ formatDate(inst.createdAt) }}</li>
            </ul>
          </div>
          <div class="card-footer bg-white border-top-0">
            <div class="d-flex gap-2">
              <button class="btn btn-primary btn-sm" @click="viewInstance(inst)">View</button>
              <button class="btn btn-outline-secondary btn-sm" @click="openDeleteModal(inst)">
                Delete
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <ConfirmModal
      v-model:show="showDeleteModal"
      title="Delete Instance"
      :message="deleteModalMessage"
      confirm-text="Yes"
      cancel-text="Cancel"
      confirm-variant="danger"
      @confirm="performDelete"
    />
  </div>
</template>

<style scoped>
.status-icon-wrapper {
  position: relative;
  display: inline-flex;
  cursor: help;
}

.status-tooltip {
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%) translateY(-4px);
  padding: 6px 10px;
  background: #212529;
  color: #fff;
  font-size: 0.75rem;
  font-weight: normal;
  white-space: normal;
  max-width: 220px;
  border-radius: 4px;
  opacity: 0;
  visibility: hidden;
  transition: opacity 0.2s, visibility 0.2s;
  pointer-events: none;
  z-index: 1000;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.status-tooltip::after {
  content: '';
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  border: 5px solid transparent;
  border-top-color: #212529;
}

.status-icon-wrapper:hover .status-tooltip {
  opacity: 1;
  visibility: visible;
}
</style>
