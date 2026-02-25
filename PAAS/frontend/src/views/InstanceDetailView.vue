<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import type { RedisInstance } from '@/types/instance'
import ConfirmModal from '@/components/ConfirmModal.vue'

const REFRESH_INTERVAL_MS = 10_000

const route = useRoute()
const toast = useToastStore()
const router = useRouter()
const instance = ref<RedisInstance | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)
const showDeleteModal = ref(false)
let refreshTimer: ReturnType<typeof setInterval> | null = null

const showEditModal = ref(false)
const showConfirmEditModal = ref(false)
const editRedisReplicas = ref<number | null>(null)
const editSentinelReplicas = ref<number | null>(null)

const id = computed(() => route.params.id as string)
const namespace = computed(() => (route.query.namespace as string) || 'default')

const deleteModalMessage = computed(
  () =>
    instance.value
      ? `Delete instance "${instance.value.name}"? This cannot be undone.`
      : ''
)

const editConfirmMessage = computed(() => {
  if (!instance.value) return ''

  const parts: string[] = []

  if (
    editRedisReplicas.value !== null &&
    editRedisReplicas.value !== instance.value.redisReplicas
  ) {
    parts.push(
      `Redis replicas: ${instance.value.redisReplicas} → ${editRedisReplicas.value}`
    )
  }

  if (
    editSentinelReplicas.value !== null &&
    editSentinelReplicas.value !== instance.value.sentinelReplicas
  ) {
    parts.push(
      `Sentinel replicas: ${instance.value.sentinelReplicas} → ${editSentinelReplicas.value}`
    )
  }

  if (parts.length === 0) {
    return 'No changes detected.'
  }

  return `Apply the following changes?\n\n${parts.join('\n')}`
})

function openDeleteModal() {
  showDeleteModal.value = true
}

function openEditModal() {
  if (!instance.value) return
  editRedisReplicas.value = instance.value.redisReplicas
  editSentinelReplicas.value = instance.value.sentinelReplicas
  showEditModal.value = true
}

async function performDelete() {
  if (!instance.value) return
  const ns = encodeURIComponent(instance.value.namespace ?? 'default')
  try {
    await api.delete(`/api/instances/${instance.value.id}?namespace=${ns}`)
    toast.show('Instance deleted successfully')
    router.push('/instances')
  } catch (e) {
    alert(e instanceof Error ? e.message : 'Failed to delete instance')
  }
}

function requestEditConfirmation() {
  if (!instance.value) return

  const changedRedis =
    editRedisReplicas.value !== null &&
    editRedisReplicas.value !== instance.value.redisReplicas
  const changedSentinel =
    editSentinelReplicas.value !== null &&
    editSentinelReplicas.value !== instance.value.sentinelReplicas

  if (!changedRedis && !changedSentinel) {
    toast.show('No changes to apply')
    showEditModal.value = false
    return
  }

  showConfirmEditModal.value = true
}

async function performUpdate() {
  if (!instance.value) return

  const payload: Record<string, number> = {}

  if (
    editRedisReplicas.value !== null &&
    editRedisReplicas.value !== instance.value.redisReplicas
  ) {
    payload.redisReplicas = editRedisReplicas.value
  }

  if (
    editSentinelReplicas.value !== null &&
    editSentinelReplicas.value !== instance.value.sentinelReplicas
  ) {
    payload.sentinelReplicas = editSentinelReplicas.value
  }

  if (Object.keys(payload).length === 0) {
    toast.show('No changes to apply')
    showEditModal.value = false
    showConfirmEditModal.value = false
    return
  }

  const ns = encodeURIComponent(instance.value.namespace ?? 'default')

  try {
    const { data } = await api.patch<{ message: string; instance: RedisInstance }>(
      `/api/instances/${instance.value.id}?namespace=${ns}`,
      payload
    )
    instance.value = data.instance
    toast.show('Instance updated successfully')
    showEditModal.value = false
    showConfirmEditModal.value = false
  } catch (e) {
    const message =
      e instanceof Error ? e.message : 'Failed to update instance'
    toast.show(message)
  }
}

async function fetchInstance(silent = false) {
  if (!id.value) return
  if (!silent) {
    loading.value = true
    error.value = null
  }
  try {
    const ns = encodeURIComponent(namespace.value)
    const { data } = await api.get<{ message: string; instance: RedisInstance }>(
      `/api/instances/${id.value}?namespace=${ns}`
    )
    instance.value = data.instance
  } catch (e) {
    if (!silent) {
      error.value = e instanceof Error ? e.message : 'Failed to load instance'
      instance.value = null
    }
  } finally {
    if (!silent) loading.value = false
  }
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function statusClass(status: string) {
  const s = (status ?? '').toLowerCase()
  if (s === 'running' || s === 'ready') return 'success'
  if (s === 'provisioning' || s === 'pending') return 'warning'
  return 'secondary'
}

const copied = ref<string | null>(null)

async function copyToClipboard(text: string, key: string) {
  try {
    await navigator.clipboard.writeText(text)
    copied.value = key
    setTimeout(() => { copied.value = null }, 1500)
  } catch {
    copied.value = null
  }
}

watch([id, namespace], () => {
  fetchInstance()
})

onMounted(() => {
  fetchInstance()
  refreshTimer = setInterval(() => fetchInstance(true), REFRESH_INTERVAL_MS)
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
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h1 class="mb-0">Instance Details</h1>
      <div class="d-flex gap-2">
        <router-link to="/instances" class="btn btn-outline-secondary">Back to Instances</router-link>
        <button
          v-if="instance"
          type="button"
          class="btn btn-outline-primary icon-button-wrapper"
          aria-label="Edit instance configuration"
          @click="openEditModal"
        >
          <FontAwesomeIcon :icon="['fas', 'pen']" />
          <span class="icon-tooltip">Edit instance configuration</span>
        </button>
        <button
          v-if="instance"
          class="btn btn-outline-danger"
          @click="openDeleteModal"
        >
          Delete
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
    </div>

    <div v-else-if="error" class="alert alert-danger" role="alert">
      {{ error }}
    </div>

    <div v-else-if="instance" class="card shadow-sm">
      <div class="card-header d-flex justify-content-between align-items-center bg-white py-3">
        <span class="fs-5 fw-bold redis-accent">{{ instance.name }}</span>
        <span class="badge" :class="`bg-${statusClass(instance.status)}`">{{ instance.status }}</span>
      </div>
      <div class="card-body">
        <div class="row">
          <div class="col-md-6">
            <h6 class="text-muted text-uppercase small mb-3">Configuration</h6>
            <dl class="row mb-0">
              <dt class="col-sm-4">ID</dt>
              <dd class="col-sm-8">{{ instance.id }}</dd>

              <dt class="col-sm-4">Name</dt>
              <dd class="col-sm-8">{{ instance.name }}</dd>

              <dt class="col-sm-4">Namespace</dt>
              <dd class="col-sm-8">{{ instance.namespace }}</dd>

              <dt class="col-sm-4">Redis replicas</dt>
              <dd class="col-sm-8">{{ instance.redisReplicas }}</dd>

              <dt class="col-sm-4">Sentinel replicas</dt>
              <dd class="col-sm-8">{{ instance.sentinelReplicas }}</dd>
            </dl>
          </div>
          <div class="col-md-6">
            <h6 class="text-muted text-uppercase small mb-3">Timestamps</h6>
            <dl class="row mb-0">
              <dt class="col-sm-4">Created</dt>
              <dd class="col-sm-8">{{ formatDate(instance.createdAt) }}</dd>

              <dt class="col-sm-4">Updated</dt>
              <dd class="col-sm-8">{{ formatDate(instance.updatedAt) }}</dd>
            </dl>
          </div>
        </div>

        <div v-if="instance.externalHost || instance.redisCli" class="mt-4 pt-4 border-top">
          <h6 class="text-muted text-uppercase small mb-3">Connection</h6>
          <dl class="row mb-0">
            <dt v-if="instance.externalHost" class="col-sm-3">Host</dt>
            <dd v-if="instance.externalHost" class="col-sm-9 d-flex align-items-center gap-2">
              <code class="bg-light px-2 py-1 rounded flex-grow-1">{{ instance.externalHost }}</code>
              <button
                type="button"
                class="btn btn-sm btn-link p-1"
                :class="copied === 'host' ? 'text-success' : 'text-secondary'"
                title="Copy host"
                @click="copyToClipboard(instance.externalHost!, 'host')"
              >
                <FontAwesomeIcon :icon="copied === 'host' ? ['fas', 'check'] : ['fas', 'copy']" />
              </button>
            </dd>

            <dt v-if="instance.externalPort" class="col-sm-3">Port</dt>
            <dd v-if="instance.externalPort" class="col-sm-9 d-flex align-items-center gap-2">
              <code class="bg-light px-2 py-1 rounded flex-grow-1">{{ instance.externalPort }}</code>
              <button
                type="button"
                class="btn btn-sm btn-link p-1"
                :class="copied === 'port' ? 'text-success' : 'text-secondary'"
                title="Copy port"
                @click="copyToClipboard(String(instance.externalPort), 'port')"
              >
                <FontAwesomeIcon :icon="copied === 'port' ? ['fas', 'check'] : ['fas', 'copy']" />
              </button>
            </dd>

            <dt v-if="instance.redisCli" class="col-sm-3">CLI command</dt>
            <dd v-if="instance.redisCli" class="col-sm-9 d-flex align-items-center gap-2">
              <code class="bg-light px-2 py-1 rounded flex-grow-1">{{ instance.redisCli }}</code>
              <button
                type="button"
                class="btn btn-sm btn-link p-1"
                :class="copied === 'cli' ? 'text-success' : 'text-secondary'"
                title="Copy CLI command"
                @click="copyToClipboard(instance.redisCli!, 'cli')"
              >
                <FontAwesomeIcon :icon="copied === 'cli' ? ['fas', 'check'] : ['fas', 'copy']" />
              </button>
            </dd>
          </dl>
        </div>
      </div>
    </div>

    <!-- Edit configuration modal -->
    <Teleport to="body">
      <template v-if="showEditModal">
        <div class="modal-backdrop fade show" @click="showEditModal = false" />
        <div
          class="modal fade show"
          style="display: block"
          tabindex="-1"
          role="dialog"
          aria-modal="true"
        >
          <div class="modal-dialog modal-dialog-centered">
            <div class="modal-content">
              <div class="modal-header">
                <h5 class="modal-title">Edit Instance</h5>
                <button
                  type="button"
                  class="btn-close"
                  aria-label="Close"
                  @click="showEditModal = false"
                />
              </div>
              <div class="modal-body">
                <div class="mb-3">
                  <label for="edit-redis-replicas" class="form-label">Redis replicas</label>
                  <input
                    id="edit-redis-replicas"
                    v-model.number="editRedisReplicas"
                    type="number"
                    min="1"
                    class="form-control"
                  />
                </div>
                <div class="mb-3">
                  <label for="edit-sentinel-replicas" class="form-label">Sentinel replicas</label>
                  <input
                    id="edit-sentinel-replicas"
                    v-model.number="editSentinelReplicas"
                    type="number"
                    min="1"
                    class="form-control"
                  />
                </div>
              </div>
              <div class="modal-footer">
                <button
                  type="button"
                  class="btn btn-secondary"
                  @click="showEditModal = false"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  class="btn btn-primary"
                  @click="requestEditConfirmation"
                >
                  Review changes
                </button>
              </div>
            </div>
          </div>
        </div>
      </template>
    </Teleport>

    <ConfirmModal
      v-model:show="showDeleteModal"
      title="Delete Instance"
      :message="deleteModalMessage"
      confirm-text="Yes"
      cancel-text="Cancel"
      confirm-variant="danger"
      @confirm="performDelete"
    />
    <ConfirmModal
      v-model:show="showConfirmEditModal"
      title="Apply Changes"
      :message="editConfirmMessage"
      confirm-text="Apply"
      cancel-text="Cancel"
      confirm-variant="primary"
      @confirm="performUpdate"
    />
  </div>
</template>

<style scoped>
.icon-button-wrapper {
  position: relative;
}

.icon-tooltip {
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%) translateY(-4px);
  padding: 4px 8px;
  background: #212529;
  color: #fff;
  font-size: 0.75rem;
  white-space: nowrap;
  border-radius: 4px;
  opacity: 0;
  visibility: hidden;
  transition: opacity 0.15s ease, visibility 0.15s ease;
  z-index: 1000;
}

.icon-tooltip::after {
  content: '';
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  border: 5px solid transparent;
  border-top-color: #212529;
}

.icon-button-wrapper:hover .icon-tooltip {
  opacity: 1;
  visibility: visible;
}
</style>
