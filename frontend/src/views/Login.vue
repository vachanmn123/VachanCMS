<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Github, Loader2 } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import ThemeToggle from '@/components/ThemeToggle.vue'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)

onMounted(async () => {
  await authStore.checkAuth()
  if (authStore.isAuthenticated) {
    router.push('/repos')
  }
})

function login() {
  loading.value = true
  window.location.href = '/api/auth/login'
}
</script>

<template>
  <div class="flex min-h-screen flex-col items-center justify-center bg-background p-4">
    <!-- Theme toggle in corner -->
    <div class="absolute right-4 top-4">
      <ThemeToggle />
    </div>

    <Card class="w-full max-w-sm">
      <CardHeader class="space-y-1 text-center">
        <div class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-lg bg-primary">
          <span class="text-xl font-bold text-primary-foreground">V</span>
        </div>
        <CardTitle class="text-2xl font-bold">VachanCMS</CardTitle>
        <CardDescription> Sign in to manage your content with GitHub </CardDescription>
      </CardHeader>
      <CardContent>
        <Button @click="login" :disabled="loading" class="w-full" size="lg">
          <Loader2 v-if="loading" class="mr-2 h-4 w-4 animate-spin" />
          <Github v-else class="mr-2 h-4 w-4" />
          Continue with GitHub
        </Button>
      </CardContent>
    </Card>

    <p class="mt-6 text-center text-sm text-muted-foreground">Your content, versioned in GitHub</p>
  </div>
</template>
