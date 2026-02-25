import { ref } from 'vue'
import { defineStore } from 'pinia'

const DEFAULT_DURATION = 4000

export const useToastStore = defineStore('toast', () => {
  const visible = ref(false)
  const message = ref('')
  const duration = ref(DEFAULT_DURATION)

  function show(msg: string, ms = DEFAULT_DURATION) {
    message.value = msg
    duration.value = ms
    visible.value = false
    // Reset so watch triggers and progress bar restarts when shown again
    queueMicrotask(() => {
      visible.value = true
    })
  }

  function hide() {
    visible.value = false
  }

  return { visible, message, duration, show, hide }
})
