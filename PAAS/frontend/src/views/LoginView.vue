<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import { useAuth } from '@/stores/auth'

const router = useRouter()
const { setToken } = useAuth()

const email = ref(import.meta.env.VITE_ADMIN_EMAIL ?? '')
const password = ref('')
const isLoading = ref(false)
const error = ref<string | null>(null)

// Register modal state
const showRegisterModal = ref(false)
const registerEmail = ref('')
const registerPassword = ref('')
const registerLoading = ref(false)
const registerError = ref<string | null>(null)
const registerSuccess = ref(false)

function openRegisterModal() {
  showRegisterModal.value = true
  registerEmail.value = ''
  registerPassword.value = ''
  registerError.value = null
  registerSuccess.value = false
}

function closeRegisterModal() {
  showRegisterModal.value = false
}

async function onRegisterSubmit() {
  if (!registerEmail.value || !registerPassword.value) return

  registerLoading.value = true
  registerError.value = null

  try {
    await api.post('/auth/register', {
      email: registerEmail.value,
      password: registerPassword.value,
    })
    registerSuccess.value = true
    // Close modal after a short delay so user sees success message
    setTimeout(() => {
      closeRegisterModal()
    }, 1500)
  } catch (err: unknown) {
    const msg =
      err && typeof err === 'object' && 'response' in err
        ? (err as { response?: { data?: { message?: string; error?: string } } }).response?.data
            ?.message ||
          (err as { response?: { data?: { error?: string } } }).response?.data?.error ||
          'Registration failed.'
        : 'Registration failed.'
    registerError.value = msg
  } finally {
    registerLoading.value = false
  }
}

async function onSubmit() {
  if (!email.value || !password.value) return

  isLoading.value = true
  error.value = null

  try {
    const response = await api.post('/auth/login', {
      email: email.value,
      password: password.value,
    })

    const token = response.data?.token as string | undefined
    if (!token) {
      throw new Error('No token returned from server')
    }

    const userPayload = response.data?.user as { email?: string; is_admin?: boolean } | undefined
    setToken(token, userPayload ? { email: userPayload.email ?? '', is_admin: !!userPayload.is_admin } : undefined)
    router.push('/instances')
  } catch (err) {
    console.error(err)
    error.value = 'Login failed. Please check your credentials.'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="login-container d-flex justify-content-center align-items-center py-5">
    <div class="card shadow-sm" style="width: 100%; max-width: 400px">
      <div class="card-body p-4">
        <h2 class="card-title text-center mb-4">Login</h2>
        <form @submit.prevent="onSubmit">
          <div class="mb-3">
            <label for="email" class="form-label">Email</label>
            <input
              id="email"
              v-model="email"
              type="email"
              class="form-control"
              autocomplete="username"
              required
            />
          </div>
          <div class="mb-4">
            <label for="password" class="form-label">Password</label>
            <input
              id="password"
              v-model="password"
              type="password"
              class="form-control"
              autocomplete="current-password"
              required
            />
          </div>
          <div v-if="error" class="alert alert-danger mb-3">
            {{ error }}
          </div>
          <button type="submit" class="btn btn-primary w-100" :disabled="isLoading">
            <span v-if="isLoading">Logging in...</span>
            <span v-else>Login</span>
          </button>
          <p class="text-center text-muted mt-3 mb-0">
            Don't have an account?
            <button
              type="button"
              class="btn btn-link p-0 ms-1 align-baseline"
              @click="openRegisterModal"
            >
              Create one
            </button>
          </p>
        </form>
      </div>
    </div>
  </div>

  <!-- Register modal -->
  <Teleport to="body">
    <div
      v-if="showRegisterModal"
      class="modal d-block"
      tabindex="-1"
      role="dialog"
      style="background: rgba(0,0,0,0.5)"
      @click.self="closeRegisterModal"
    >
      <div class="modal-dialog modal-dialog-centered" role="document">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Create account</h5>
            <button
              type="button"
              class="btn-close"
              aria-label="Close"
              @click="closeRegisterModal"
            />
          </div>
          <div class="modal-body">
            <div v-if="registerSuccess" class="alert alert-success mb-0">
              Account created. You can now log in.
            </div>
            <form v-else @submit.prevent="onRegisterSubmit">
              <div class="mb-3">
                <label for="register-email" class="form-label">Email</label>
                <input
                  id="register-email"
                  v-model="registerEmail"
                  type="email"
                  class="form-control"
                  autocomplete="email"
                  required
                />
              </div>
              <div class="mb-3">
                <label for="register-password" class="form-label">Password</label>
                <input
                  id="register-password"
                  v-model="registerPassword"
                  type="password"
                  class="form-control"
                  autocomplete="new-password"
                  required
                />
              </div>
              <div v-if="registerError" class="alert alert-danger mb-3">
                {{ registerError }}
              </div>
              <button
                type="submit"
                class="btn btn-primary w-100"
                :disabled="registerLoading"
              >
                <span v-if="registerLoading">Creating account...</span>
                <span v-else>Create account</span>
              </button>
            </form>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>
