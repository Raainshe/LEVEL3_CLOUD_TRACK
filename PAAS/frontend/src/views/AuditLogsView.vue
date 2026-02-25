<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { getAuditLogs } from '@/api/audit'
import type { AuditLog } from '@/types/audit'
import { AUDIT_ACTION_TYPES } from '@/types/audit'

const PAGE_SIZE = 50

const logs = ref<AuditLog[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const page = ref(1)
const total = ref(0)
const filterType = ref<string>('')

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

async function fetchLogs() {
  loading.value = true
  error.value = null
  try {
    const data = await getAuditLogs({
      page: page.value,
      type: filterType.value || undefined,
    })
    logs.value = data.audit_logs ?? []
    total.value = data.total
    totalPages.value = Math.max(1, Math.ceil(data.total / PAGE_SIZE))
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load audit logs'
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

watch(filterType, onFilterChange)

onMounted(fetchLogs)
</script>

<template>
  <div>
    <h1 class="mb-4">Audit Logs</h1>

    <div class="d-flex flex-wrap align-items-center gap-3 mb-3">
      <label for="audit-type-filter" class="form-label mb-0">Filter by type</label>
      <select
        id="audit-type-filter"
        v-model="filterType"
        class="form-select form-select-sm"
        style="width: auto"
      >
        <option value="">All</option>
        <option
          v-for="t in AUDIT_ACTION_TYPES"
          :key="t"
          :value="t"
        >
          {{ t }}
        </option>
      </select>
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
      No audit logs found.
      <template v-if="filterType"> Try changing the filter.</template>
    </div>

    <div v-else>
      <div class="table-responsive">
        <table class="table table-striped table-hover">
          <thead>
            <tr>
              <th>Timestamp</th>
              <th>User</th>
              <th>Type</th>
              <th>Name</th>
              <th>Namespace</th>
              <th>Details</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="log in logs" :key="log.id">
              <td>{{ formatDateTime(log.timestamp) }}</td>
              <td>{{ log.user_email }}</td>
              <td>
                <span v-if="log.action.action === 'login'" class="badge bg-primary">{{ log.action.action }}</span>
                <span v-if="log.action.action === 'register'" class="badge bg-success">{{ log.action.action }}</span>
                <span v-if="log.action.action === 'create'" class="badge bg-info">{{ log.action.action }}</span>
                <span v-if="log.action.action === 'update'" class="badge bg-warning">{{ log.action.action }}</span>
                <span v-if="log.action.action === 'delete'" class="badge bg-danger">{{ log.action.action }}</span>
              </td>
              <td>{{ log.action.name || '—' }}</td>
              <td>{{ log.action.namespace || '—' }}</td>
              <td class="text-break" style="max-width: 200px">
                {{ log.action.details || '—' }}
              </td>
              <td>
                <span v-if="log.admin_info" class="badge bg-info">Admin</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <nav aria-label="Audit logs pagination" class="d-flex justify-content-between align-items-center flex-wrap gap-2 mt-3">
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
