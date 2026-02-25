import { computed, ref } from 'vue'

const token = ref<string | null>(localStorage.getItem('jwt_token'))

const USER_KEY = 'paas_user'

function loadUser(): { email: string; isAdmin: boolean } | null {
  try {
    const raw = localStorage.getItem(USER_KEY)
    if (!raw) return null
    const data = JSON.parse(raw) as { email?: string; is_admin?: boolean }
    if (!data?.email) return null
    return { email: data.email, isAdmin: !!data.is_admin }
  } catch {
    return null
  }
}

const user = ref<{ email: string; isAdmin: boolean } | null>(loadUser())

export function useAuth() {
  const isAuthenticated = computed(() => !!token.value)

  function setToken(value: string, userPayload?: { email: string; is_admin: boolean }) {
    token.value = value
    localStorage.setItem('jwt_token', value)
    if (userPayload) {
      const u = { email: userPayload.email, isAdmin: !!userPayload.is_admin }
      user.value = u
      localStorage.setItem(USER_KEY, JSON.stringify({ email: u.email, is_admin: userPayload.is_admin }))
    }
  }

  function clearToken() {
    token.value = null
    user.value = null
    localStorage.removeItem('jwt_token')
    localStorage.removeItem(USER_KEY)
  }

  return {
    token,
    user,
    isAuthenticated,
    setToken,
    clearToken,
  }
}

