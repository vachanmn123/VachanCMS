import { ref } from 'vue'
import { toast } from 'vue-sonner'

export function useLoading() {
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function withLoading<T>(
    fn: () => Promise<T>,
    options?: {
      errorMessage?: string
      showErrorToast?: boolean
    },
  ): Promise<T | null> {
    const { errorMessage = 'An error occurred', showErrorToast = true } = options ?? {}

    loading.value = true
    error.value = null

    try {
      const result = await fn()
      return result
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : errorMessage
      error.value = message

      if (showErrorToast) {
        toast.error(message)
      }

      return null
    } finally {
      loading.value = false
    }
  }

  function reset() {
    loading.value = false
    error.value = null
  }

  return {
    loading,
    error,
    withLoading,
    reset,
  }
}
