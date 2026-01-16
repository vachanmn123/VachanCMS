import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import axios from 'axios'

interface PagesConfigResponse {
  initialized: boolean
  baseUrl?: string
}

export const usePagesStore = defineStore('pages', () => {
  const pagesConfig = ref<PagesConfigResponse | null>(null)
  const loading = ref(false)
  const error = ref(false)
  const cachedRepoKey = ref<string | null>(null)

  const isInitialized = computed(() => pagesConfig.value?.initialized ?? false)
  const baseUrl = computed(() => pagesConfig.value?.baseUrl ?? '')

  async function fetchPagesConfig(owner: string, repo: string) {
    const repoKey = `${owner}/${repo}`

    // Return cached data if we already fetched for this repo
    if (cachedRepoKey.value === repoKey && pagesConfig.value !== null) {
      return pagesConfig.value
    }

    loading.value = true
    error.value = false

    try {
      const response = await axios.get<PagesConfigResponse>(`/api/${owner}/${repo}/pages`)
      pagesConfig.value = response.data
      cachedRepoKey.value = repoKey
      return response.data
    } catch (err) {
      console.error('Failed to fetch pages config', err)
      error.value = true
      pagesConfig.value = { initialized: false }
      cachedRepoKey.value = repoKey
      return pagesConfig.value
    } finally {
      loading.value = false
    }
  }

  function clearCache() {
    pagesConfig.value = null
    cachedRepoKey.value = null
    error.value = false
  }

  return {
    pagesConfig,
    loading,
    error,
    isInitialized,
    baseUrl,
    fetchPagesConfig,
    clearCache,
  }
})
