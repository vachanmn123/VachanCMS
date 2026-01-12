import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import axios from 'axios'

interface User {
  login: string
  name: string
  avatar_url: string
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const isAuthenticated = computed(() => !!user.value)

  async function checkAuth() {
    try {
      const response = await axios.get('/api/me')
      user.value = response.data
    } catch {
      user.value = null
    }
  }

  function logout() {
    user.value = null
    // Clear cookie by redirect or something
  }

  return { user, isAuthenticated, checkAuth, logout }
})
