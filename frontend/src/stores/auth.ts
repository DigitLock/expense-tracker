import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api'
import { config } from '@/config'
import type { User, LoginRequest } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const isAuthenticated = computed(() => !!token.value)
  const userName = computed(() => user.value?.name || '')
  const userEmail = computed(() => user.value?.email || '')
  const familyId = computed(() => user.value?.family_id || '')

  // Actions
  async function login(credentials: LoginRequest): Promise<boolean> {
    loading.value = true
    error.value = null

    try {
      const response = await authApi.login(credentials)
      
      if (response.success) {
        token.value = response.data.token
        user.value = response.data.user

        // Persist to localStorage
        localStorage.setItem(config.auth.tokenKey, response.data.token)
        localStorage.setItem(config.auth.userKey, JSON.stringify(response.data.user))

        return true
      }
      
      return false
    } catch (err: any) {
      error.value = err.response?.data?.error?.message || 'Login failed. Please try again.'
      return false
    } finally {
      loading.value = false
    }
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem(config.auth.tokenKey)
    localStorage.removeItem(config.auth.userKey)
  }

  function initialize() {
    // Restore auth state from localStorage
    const storedToken = localStorage.getItem(config.auth.tokenKey)
    const storedUser = localStorage.getItem(config.auth.userKey)

    if (storedToken && storedUser) {
      token.value = storedToken
      try {
        user.value = JSON.parse(storedUser)
      } catch {
        // Invalid stored user, clear everything
        logout()
      }
    }
  }

  function clearError() {
    error.value = null
  }

  return {
    // State
    user,
    token,
    loading,
    error,
    // Getters
    isAuthenticated,
    userName,
    userEmail,
    familyId,
    // Actions
    login,
    logout,
    initialize,
    clearError,
  }
})
