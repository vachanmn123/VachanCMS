<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { toast } from 'vue-sonner'
import { Search, GitBranch, Loader2, FolderGit2, LogOut } from 'lucide-vue-next'
import { useAuthStore } from '@/stores/auth'
import { useLoading } from '@/composables/useLoading'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import ThemeToggle from '@/components/ThemeToggle.vue'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

interface Repo {
  full_name: string
  description: string
}

const router = useRouter()
const authStore = useAuthStore()
const { loading, withLoading } = useLoading()

const repos = ref<Repo[]>([])
const search = ref('')
const selectedRepo = ref<Repo | null>(null)
const showInitDialog = ref(false)
const showSiteNameDialog = ref(false)
const siteName = ref('')
const initLoading = ref(false)

const filteredRepos = computed(() =>
  repos.value.filter((repo) => repo.full_name.toLowerCase().includes(search.value.toLowerCase())),
)

onMounted(async () => {
  await withLoading(
    async () => {
      const response = await axios.get('/api/repos')
      repos.value = response.data
    },
    { errorMessage: 'Failed to fetch repositories' },
  )
})

async function selectRepo(repo: Repo) {
  const [owner, repoName] = repo.full_name.split('/')
  try {
    await axios.get(`/api/${owner}/${repoName}/config`)
    router.push(`/dashboard/${owner}/${repoName}`)
  } catch (error: unknown) {
    const axiosError = error as { response?: { status?: number } }
    if (axiosError.response?.status === 404) {
      selectedRepo.value = repo
      showInitDialog.value = true
    } else {
      toast.error('Failed to check repository status')
    }
  }
}

function confirmInit() {
  showInitDialog.value = false
  showSiteNameDialog.value = true
}

async function initializeRepo() {
  if (!selectedRepo.value || !siteName.value.trim()) return

  const [owner, repoName] = selectedRepo.value.full_name.split('/')
  initLoading.value = true

  try {
    await axios.post(`/api/${owner}/${repoName}/init`, {
      site_name: siteName.value.trim(),
    })
    toast.success('Repository initialized successfully!')
    showSiteNameDialog.value = false
    router.push(`/dashboard/${owner}/${repoName}`)
  } catch {
    toast.error('Failed to initialize repository')
  } finally {
    initLoading.value = false
  }
}

function logout() {
  authStore.logout()
  router.push('/')
}
</script>

<template>
  <div class="min-h-screen bg-background">
    <!-- Header -->
    <header
      class="sticky top-0 z-50 border-b bg-background/95 backdrop-blur supports-backdrop-filter:bg-background/60"
    >
      <div class="container flex h-14 items-center justify-between px-4">
        <div class="flex items-center gap-2">
          <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
            <span class="text-sm font-bold text-primary-foreground">V</span>
          </div>
          <span class="font-semibold">VachanCMS</span>
        </div>

        <div class="flex items-center gap-2">
          <ThemeToggle />
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <Button variant="ghost" class="relative h-8 w-8 rounded-full">
                <Avatar class="h-8 w-8">
                  <AvatarImage
                    :src="authStore.user?.avatar_url ?? ''"
                    :alt="authStore.user?.login ?? ''"
                  />
                  <AvatarFallback>{{
                    authStore.user?.login?.charAt(0)?.toUpperCase()
                  }}</AvatarFallback>
                </Avatar>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent class="w-56" align="end">
              <DropdownMenuLabel class="font-normal">
                <div class="flex flex-col space-y-1">
                  <p class="text-sm font-medium">
                    {{ authStore.user?.name || authStore.user?.login }}
                  </p>
                  <p class="text-xs text-muted-foreground">@{{ authStore.user?.login }}</p>
                </div>
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem @click="logout" class="text-destructive">
                <LogOut class="mr-2 h-4 w-4" />
                Log out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="container mx-auto px-4 py-8">
      <div class="mb-8">
        <h1 class="text-3xl font-bold tracking-tight">Select Repository</h1>
        <p class="mt-2 text-muted-foreground">Choose a repository to manage its content</p>
      </div>

      <!-- Search -->
      <div class="relative mb-6">
        <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input v-model="search" type="text" placeholder="Search repositories..." class="pl-10" />
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Skeleton v-for="i in 6" :key="i" class="h-32" />
      </div>

      <!-- Empty State -->
      <div v-else-if="repos.length === 0" class="flex flex-col items-center justify-center py-16">
        <FolderGit2 class="h-12 w-12 text-muted-foreground" />
        <h3 class="mt-4 text-lg font-semibold">No repositories found</h3>
        <p class="mt-2 text-center text-sm text-muted-foreground">
          We couldn't find any repositories in your GitHub account.
        </p>
      </div>

      <!-- No Search Results -->
      <div
        v-else-if="filteredRepos.length === 0"
        class="flex flex-col items-center justify-center py-16"
      >
        <Search class="h-12 w-12 text-muted-foreground" />
        <h3 class="mt-4 text-lg font-semibold">No matching repositories</h3>
        <p class="mt-2 text-center text-sm text-muted-foreground">Try adjusting your search term</p>
      </div>

      <!-- Repository Grid -->
      <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Card
          v-for="repo in filteredRepos"
          :key="repo.full_name"
          class="cursor-pointer transition-colors hover:bg-accent"
          @click="selectRepo(repo)"
        >
          <CardHeader>
            <CardTitle class="flex items-center gap-2 text-base">
              <GitBranch class="h-4 w-4 text-muted-foreground" />
              {{ repo.full_name }}
            </CardTitle>
            <CardDescription class="line-clamp-2">
              {{ repo.description || 'No description' }}
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    </main>

    <!-- Init Confirmation Dialog -->
    <AlertDialog v-model:open="showInitDialog">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Initialize Repository?</AlertDialogTitle>
          <AlertDialogDescription>
            Repository <strong>{{ selectedRepo?.full_name }}</strong> is not initialized for
            VachanCMS. This will create configuration files and directories in the repository.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction @click="confirmInit">Continue</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

    <!-- Site Name Dialog -->
    <Dialog v-model:open="showSiteNameDialog">
      <DialogContent class="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Site Configuration</DialogTitle>
          <DialogDescription>
            Enter a name for your site. This will be used as the display name in the CMS.
          </DialogDescription>
        </DialogHeader>
        <form @submit.prevent="initializeRepo">
          <div class="grid gap-4 py-4">
            <div class="space-y-2">
              <Label for="siteName">Site Name</Label>
              <Input id="siteName" v-model="siteName" placeholder="My Awesome Site" required />
            </div>
          </div>
          <DialogFooter>
            <Button type="submit" :disabled="initLoading || !siteName.trim()">
              <Loader2 v-if="initLoading" class="mr-2 h-4 w-4 animate-spin" />
              Initialize
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </div>
</template>
