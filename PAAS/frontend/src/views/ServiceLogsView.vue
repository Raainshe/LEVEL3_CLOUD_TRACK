<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { getServiceLogs } from '@/api/serviceLog'
import type { ServiceLog } from '@/types/serviceLog'

const PAGE_SIZE = 50

const logs = ref<ServiceLog[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const page = ref(1)
const total = ref(0)
const filterInstance = ref('')
const filterNamespace = ref('')

const totalPages = ref(0)

function formatDateTime(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function statusLabel(log: ServiceLog) {
  if (log.from_status && log.to_status) {
    return `${log.from_status} → ${log.to_status}`
  }
  return log.to_status || '—'
}

async function fetchLogs() {
  loading.value = true
  error.value = null
  try {
    const data = await getServiceLogs({
      page: page.value,
      instance: filterInstance.value.trim() || undefined,
      namespace: filterNamespace.value.trim() || undefined,
    })
    logs.value = data.service_logs ?? []
    total.value = data.total
    totalPages.value = Math.max(1, Math.ceil(data.total / PAGE_SIZE))
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load service logs'
    logs.value = []
    total.value = 0
    totalPages.value = 1
  } finally {
    loading.value = false
  }
}

function onFilterChange() {
  page.value = 1
  fetchLogs()
}

function goToPage(p: number) {
  if (p < 1 || p > totalPages.value) return
  page.value = p
  fetchLogs()
}

watch([filterInstance, filterNamespace], onFilterChange)

onMounted(fetchLogs)
</script>

<template>
  <div>
    <h1 class="mb-4">Service Logs</h1>

    <div class="d-flex flex-wrap align-items-center gap-3 mb-3">
      <label for="service-instance-filter" class="form-label mb-0">Instance</label>
      <input
        id="service-instance-filter"
        v-model="filterInstance"
        type="text"
        class="form-control form-control-sm"
        placeholder="Filter by instance name"
        style="width: 180px"
      />
      <label for="service-namespace-filter" class="form-label mb-0">Namespace</label>
      <input
        id="service-namespace-filter"
        v-model="filterNamespace"
        type="text"
        class="form-control form-control-sm"
        placeholder="Filter by namespace"
        style="width: 180px"
      />
    </div>

    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
    </div>

    <div v-else-if="error" class="alert alert-danger" role="alert">
      {{ error }}
    </div>

    <div v-else-if="logs.length === 0" class="alert alert-info" role="alert">
      No service logs found.
      <template v-if="filterInstance || filterNamespace"> Try changing the filters.</template>
    </div>

    <div v-else>
      <div class="table-responsive">
        <table class="table table-striped table-hover">
          <thead>
            <tr>
              <th>Timestamp</th>
              <th>Instance</th>
              <th>Namespace</th>
              <th>Type</th>
              <th>Status</th>
              <th>Message</th>
              <th>Details</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="log in logs" :key="log.id">
              <td>{{ formatDateTime(log.timestamp) }}</td>
              <td>{{ log.instance_name }}</td>
              <td>{{ log.namespace }}</td>
              <td>
  
                <span v-if="log.event_type === 'status_change'" class="badge bg-primary">{{ log.event_type }}</span>
                <span v-if="log.event_type === 'failure'" class="badge bg-danger">{{ log.event_type }}</span>
              </td>
              <td>{{ statusLabel(log) }}</td>
              <td class="text-break" style="max-width: 220px">
                {{ log.message || '—' }}
              </td>
              <td class="text-break" style="max-width: 200px">
                {{ log.details || '—' }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <nav aria-label="Service logs pagination" class="d-flex justify-content-between align-items-center flex-wrap gap-2 mt-3">
        <div class="text-muted small">
          Page {{ page }} of {{ totalPages }} ({{ total }} total)
        </div>
        <ul class="pagination pagination-sm mb-0">
          <li class="page-item" :class="{ disabled: page <= 1 }">
            <button
              type="button"
              class="page-link"
              :disabled="page <= 1"
              @click="goToPage(page - 1)"
            >
              Previous
            </button>
          </li>
          <li class="page-item" :class="{ disabled: page >= totalPages }">
            <button
              type="button"
              class="page-link"
              :disabled="page >= totalPages"
              @click="goToPage(page + 1)"
            >
              Next
            </button>
          </li>
        </ul>
      </nav>
    </div>
  </div>
</template>
