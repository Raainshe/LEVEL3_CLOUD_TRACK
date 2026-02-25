<script setup lang="ts">
import { watch, ref } from 'vue'
import { useToastStore } from '@/stores/toast'

const store = useToastStore()
const progress = ref(100)

watch(
  () => store.visible,
  (visible) => {
    if (!visible) return
    progress.value = 100
    const duration = store.duration
    const start = Date.now()

    const tick = () => {
      const elapsed = Date.now() - start
      progress.value = Math.max(0, 100 - (elapsed / duration) * 100)

      if (progress.value > 0) {
        requestAnimationFrame(tick)
      } else {
        store.hide()
      }
    }

    requestAnimationFrame(tick)
  }
)
</script>

<template>
  <Teleport to="body">
    <Transition name="toast">
      <div
        v-if="store.visible"
        class="app-toast position-fixed top-0 end-0 m-3 p-3 bg-primary text-white rounded shadow-lg"
        role="alert"
      >
        <div class="app-toast__message">{{ store.message }}</div>
        <div class="app-toast__progress progress mt-2" style="height: 4px">
          <div
            class="progress-bar bg-light"
            role="progressbar"
            :style="{ width: `${progress}%` }"
            aria-valuenow="100"
            aria-valuemin="0"
            aria-valuemax="100"
          />
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.app-toast {
  min-width: 280px;
  max-width: 400px;
  top: 70px;
  right: 1rem;
}

.app-toast__message {
  font-size: 0.95rem;
}

.toast-enter-active,
.toast-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(1rem);
}
</style>
