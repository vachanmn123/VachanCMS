<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Toaster } from '@/components/ui/sonner'

const router = useRouter()
const authStore = useAuthStore()

onMounted(async () => {
  await authStore.checkAuth()
  if (!authStore.isAuthenticated) {
    router.push('/')
  }
})
</script>

<template>
  <router-view />
  <Toaster position="top-right" :expand="true" :rich-colors="true" />
</template>
