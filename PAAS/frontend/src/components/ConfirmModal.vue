<script setup lang="ts">
import { watch } from 'vue'

const props = defineProps<{
  show: boolean
  title?: string
  message?: string
  confirmText?: string
  cancelText?: string
  confirmVariant?: 'primary' | 'danger' | 'secondary'
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  confirm: []
  cancel: []
}>()

watch(
  () => props.show,
  (show) => {
    document.body.classList.toggle('modal-open', show)
  },
  { immediate: true }
)

function onConfirm() {
  emit('confirm')
  emit('update:show', false)
}

function onCancel() {
  emit('cancel')
  emit('update:show', false)
}
</script>

<template>
  <Teleport to="body">
    <template v-if="show">
      <div class="modal-backdrop fade show" @click="onCancel" />
      <div
        class="modal fade show"
        style="display: block"
        tabindex="-1"
        role="dialog"
        aria-labelledby="confirmModalLabel"
        aria-modal="true"
      >
        <div class="modal-dialog modal-dialog-centered">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title" id="confirmModalLabel">{{ title ?? 'Confirm' }}</h5>
              <button
                type="button"
                class="btn-close"
                aria-label="Close"
                @click="onCancel"
              />
            </div>
            <div class="modal-body">
              {{ message }}
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="onCancel">
                {{ cancelText ?? 'Cancel' }}
              </button>
              <button
                type="button"
                :class="['btn', `btn-${confirmVariant ?? 'primary'}`]"
                @click="onConfirm"
              >
                {{ confirmText ?? 'Yes' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </template>
  </Teleport>
</template>
