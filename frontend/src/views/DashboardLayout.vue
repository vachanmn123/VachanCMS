<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useRepoStore, type ContentType } from '@/stores/repo'
import { useAuthStore } from '@/stores/auth'
import { usePagesStore } from '@/stores/pages'
import axios from 'axios'
import { toast } from 'vue-sonner'
import {
  FileText,
  Plus,
  PanelLeft,
  LogOut,
  Home,
  Trash2,
  ChevronRight,
  Loader2,
  Image as ImageIcon,
  Globe,
  AlertCircle,
} from 'lucide-vue-next'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSkeleton,
  SidebarProvider,
  SidebarRail,
  SidebarTrigger,
} from '@/components/ui/sidebar'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Separator } from '@/components/ui/separator'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import ThemeToggle from '@/components/ThemeToggle.vue'

interface FieldDraft {
  field_name: string
  field_type: string
  is_required: boolean
  options: string[]
}

const route = useRoute()
const router = useRouter()
const repoStore = useRepoStore()
const authStore = useAuthStore()
const pagesStore = usePagesStore()

const contentTypes = ref<ContentType[]>([])
const selectedType = ref('')
const showNewTypeDialog = ref(false)
const loading = ref(true)
const creating = ref(false)

const newType = ref({
  name: '',
  slug: '',
  fields: [] as FieldDraft[],
  items_per_page: 10,
  add_to: 'bottom' as 'top' | 'bottom',
})

const owner = computed(() => String(route.params.owner))
const repo = computed(() => String(route.params.repo))
const ctSlug = computed(() => (route.params.ctSlug ? String(route.params.ctSlug) : ''))
const isMediaRoute = computed(() => route.name === 'media')
const currentTypeName = computed(() => {
  const type = contentTypes.value.find((t) => t.slug === ctSlug.value)
  return type?.name || ''
})
const pagesSettingsUrl = computed(
  () => `https://github.com/${owner.value}/${repo.value}/settings/pages`,
)

onMounted(async () => {
  repoStore.selectRepo(owner.value, repo.value)

  try {
    const response = await axios.get(`/api/${owner.value}/${repo.value}/config`)
    repoStore.setConfig(response.data)
    contentTypes.value = response.data.content_types || []
    // Set selected type from route if present
    if (ctSlug.value) {
      selectedType.value = ctSlug.value
    }
  } catch (error: unknown) {
    const axiosError = error as { response?: { status?: number } }
    if (axiosError.response?.status === 404) {
      toast.error('Repository not initialized', {
        description: 'Please initialize the repository first.',
      })
      router.push('/repos')
    } else {
      toast.error('Failed to load configuration')
    }
  } finally {
    loading.value = false
  }

  // Fetch pages config (non-blocking)
  pagesStore.fetchPagesConfig(owner.value, repo.value)
})

// Watch for route changes to update selected type
watch(
  () => route.params.ctSlug,
  (newSlug) => {
    selectedType.value = newSlug ? String(newSlug) : ''
  },
)

function selectType(slug: string) {
  selectedType.value = slug
  router.push(`/dashboard/${owner.value}/${repo.value}/${slug}`)
}

function addField() {
  newType.value.fields.push({
    field_name: '',
    field_type: 'text',
    is_required: false,
    options: [],
  })
}

function removeField(index: number) {
  newType.value.fields.splice(index, 1)
}

function resetNewTypeForm() {
  newType.value = { name: '', slug: '', fields: [], items_per_page: 10, add_to: 'bottom' }
}

async function createType() {
  const payload = {
    name: newType.value.name,
    slug: newType.value.slug,
    items_per_page: newType.value.items_per_page,
    add_to: newType.value.add_to,
    fields: newType.value.fields.map((f) => {
      let options: string[] = []
      if (f.field_type === 'select') {
        options = f.options
      } else if (f.field_type === 'media' && f.options.includes('multiple')) {
        options = ['multiple']
      }
      return {
        ...f,
        options,
      }
    }),
  }

  creating.value = true
  try {
    await axios.post(`/api/${owner.value}/${repo.value}/content-types`, payload)
    toast.success('Content type created', {
      description: `"${payload.name}" has been created successfully.`,
    })
    showNewTypeDialog.value = false

    // Refresh content types
    const response = await axios.get(`/api/${owner.value}/${repo.value}/config`)
    contentTypes.value = response.data.content_types || []
    resetNewTypeForm()
  } catch {
    toast.error('Failed to create content type')
  } finally {
    creating.value = false
  }
}

function logout() {
  authStore.logout()
  router.push('/')
}

function goToRepoSelect() {
  router.push('/repos')
}

function goToMedia() {
  selectedType.value = ''
  router.push(`/dashboard/${owner.value}/${repo.value}/media`)
}
</script>

<template>
  <SidebarProvider>
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" as-child>
              <a href="#" @click.prevent="goToRepoSelect">
                <div
                  class="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground"
                >
                  <span class="text-sm font-bold">V</span>
                </div>
                <div class="grid flex-1 text-left text-sm leading-tight">
                  <span class="truncate font-semibold">VachanCMS</span>
                  <span class="truncate text-xs text-muted-foreground">{{ owner }}/{{ repo }}</span>
                </div>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem class="group-data-[collapsible=icon]:hidden">
            <a
              :href="pagesSettingsUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="flex w-full"
            >
              <Badge
                v-if="!pagesStore.loading"
                :variant="pagesStore.isInitialized ? 'default' : 'destructive'"
                class="w-full justify-center gap-1.5 py-1"
              >
                <Globe v-if="pagesStore.isInitialized" class="h-3 w-3" />
                <AlertCircle v-else class="h-3 w-3" />
                {{ pagesStore.isInitialized ? 'Pages Active' : 'Pages Off' }}
              </Badge>
              <Badge v-else variant="secondary" class="w-full justify-center py-1">
                <Loader2 class="h-3 w-3 animate-spin" />
              </Badge>
            </a>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Content Types</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <!-- Loading State -->
              <template v-if="loading">
                <SidebarMenuSkeleton v-for="i in 4" :key="i" />
              </template>

              <!-- Empty State -->
              <template v-else-if="contentTypes.length === 0">
                <div class="px-2 py-4 text-center">
                  <FileText class="mx-auto h-8 w-8 text-muted-foreground" />
                  <p class="mt-2 text-xs text-muted-foreground">No content types yet</p>
                </div>
              </template>

              <!-- Content Types List -->
              <template v-else>
                <SidebarMenuItem v-for="type in contentTypes" :key="type.id">
                  <SidebarMenuButton
                    :is-active="selectedType === type.slug"
                    @click="selectType(type.slug)"
                  >
                    <FileText class="h-4 w-4" />
                    <span>{{ type.name }}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </template>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        <SidebarGroup>
          <SidebarGroupLabel>Assets</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton :is-active="isMediaRoute" @click="goToMedia">
                  <ImageIcon class="h-4 w-4" />
                  <span>Media Library</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton @click="showNewTypeDialog = true">
              <Plus class="h-4 w-4" />
              <span>New Content Type</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>

      <SidebarRail />
    </Sidebar>

    <SidebarInset>
      <!-- Header -->
      <header
        class="flex h-14 shrink-0 items-center gap-2 border-b bg-background/95 px-4 backdrop-blur supports-backdrop-filter:bg-background/60"
      >
        <SidebarTrigger class="-ml-1">
          <PanelLeft class="h-4 w-4" />
        </SidebarTrigger>

        <Separator orientation="vertical" class="mr-2 h-4" />

        <Breadcrumb>
          <BreadcrumbList>
            <BreadcrumbItem class="hidden md:block">
              <BreadcrumbLink href="#" @click.prevent="goToRepoSelect">
                <Home class="h-4 w-4" />
              </BreadcrumbLink>
            </BreadcrumbItem>
            <BreadcrumbSeparator class="hidden md:block">
              <ChevronRight class="h-4 w-4" />
            </BreadcrumbSeparator>
            <BreadcrumbItem class="hidden md:block">
              <BreadcrumbLink href="#">{{ owner }}/{{ repo }}</BreadcrumbLink>
            </BreadcrumbItem>
            <template v-if="currentTypeName">
              <BreadcrumbSeparator class="hidden md:block">
                <ChevronRight class="h-4 w-4" />
              </BreadcrumbSeparator>
              <BreadcrumbItem>
                <BreadcrumbPage>{{ currentTypeName }}</BreadcrumbPage>
              </BreadcrumbItem>
            </template>
            <template v-else-if="isMediaRoute">
              <BreadcrumbSeparator class="hidden md:block">
                <ChevronRight class="h-4 w-4" />
              </BreadcrumbSeparator>
              <BreadcrumbItem>
                <BreadcrumbPage>Media Library</BreadcrumbPage>
              </BreadcrumbItem>
            </template>
          </BreadcrumbList>
        </Breadcrumb>

        <div class="ml-auto flex items-center gap-2">
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
              <DropdownMenuItem @click="goToRepoSelect">
                <Home class="mr-2 h-4 w-4" />
                Switch Repository
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem @click="logout" class="text-destructive">
                <LogOut class="mr-2 h-4 w-4" />
                Log out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </header>

      <!-- Main Content -->
      <main class="flex-1 overflow-auto p-4 md:p-6">
        <router-view />
      </main>
    </SidebarInset>

    <!-- New Type Dialog -->
    <Dialog v-model:open="showNewTypeDialog" @update:open="(open) => !open && resetNewTypeForm()">
      <DialogContent class="max-h-[90vh] overflow-y-auto sm:max-w-137.5">
        <DialogHeader>
          <DialogTitle>Create New Content Type</DialogTitle>
          <DialogDescription>
            Define the structure for your new content type. Add fields to specify what data it will
            hold.
          </DialogDescription>
        </DialogHeader>

        <form @submit.prevent="createType" class="space-y-6">
          <div class="grid gap-4 sm:grid-cols-2">
            <div class="space-y-2">
              <Label for="typeName">Name</Label>
              <Input id="typeName" v-model="newType.name" placeholder="e.g., Blog Post" required />
            </div>
            <div class="space-y-2">
              <Label for="typeSlug">Slug</Label>
              <Input id="typeSlug" v-model="newType.slug" placeholder="e.g., blog-post" required />
            </div>
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <div class="space-y-2">
              <Label for="itemsPerPage">Items Per Page</Label>
              <Input
                id="itemsPerPage"
                type="number"
                v-model.number="newType.items_per_page"
                min="1"
                max="100"
                placeholder="10"
              />
              <p class="text-xs text-muted-foreground">Number of items per page (1-100)</p>
            </div>
            <div class="space-y-2">
              <Label for="addTo">Add New Items To</Label>
              <Select v-model="newType.add_to">
                <SelectTrigger id="addTo">
                  <SelectValue placeholder="Select position" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="bottom">Bottom (oldest first)</SelectItem>
                  <SelectItem value="top">Top (newest first)</SelectItem>
                </SelectContent>
              </Select>
              <p class="text-xs text-muted-foreground">Where new entries appear in the list</p>
            </div>
          </div>

          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <Label>Fields</Label>
              <Button type="button" variant="outline" size="sm" @click="addField">
                <Plus class="mr-1 h-3 w-3" />
                Add Field
              </Button>
            </div>

            <div v-if="newType.fields.length === 0" class="rounded-lg border border-dashed p-6">
              <div class="text-center">
                <FileText class="mx-auto h-8 w-8 text-muted-foreground" />
                <p class="mt-2 text-sm text-muted-foreground">No fields added yet</p>
                <Button type="button" variant="outline" size="sm" class="mt-3" @click="addField">
                  <Plus class="mr-1 h-3 w-3" />
                  Add your first field
                </Button>
              </div>
            </div>

            <div v-else class="space-y-3">
              <Card v-for="(field, index) in newType.fields" :key="index">
                <CardContent class="p-4">
                  <div class="grid gap-3">
                    <div class="grid gap-3 sm:grid-cols-2">
                      <div class="space-y-2">
                        <Label :for="`field-name-${index}`">Field Name</Label>
                        <Input
                          :id="`field-name-${index}`"
                          v-model="field.field_name"
                          placeholder="e.g., title"
                          required
                        />
                      </div>
                      <div class="space-y-2">
                        <Label :for="`field-type-${index}`">Field Type</Label>
                        <Select v-model="field.field_type">
                          <SelectTrigger :id="`field-type-${index}`">
                            <SelectValue placeholder="Select type" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="text">Text</SelectItem>
                            <SelectItem value="textarea">Textarea</SelectItem>
                            <SelectItem value="number">Number</SelectItem>
                            <SelectItem value="boolean">Boolean</SelectItem>
                            <SelectItem value="select">Select</SelectItem>
                            <SelectItem value="media">Media</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                    </div>

                    <div v-if="field.field_type === 'select'" class="space-y-2">
                      <Label :for="`field-options-${index}`">Options (comma separated)</Label>
                      <Input
                        :id="`field-options-${index}`"
                        :model-value="field.options.join(', ')"
                        @update:model-value="
                          (val: string | number) =>
                            (field.options = String(val)
                              .split(',')
                              .map((s) => s.trim())
                              .filter((s) => s !== ''))
                        "
                        placeholder="e.g., draft, published, archived"
                      />
                    </div>

                    <div v-if="field.field_type === 'media'" class="space-y-2">
                      <div class="flex items-center space-x-2">
                        <Checkbox
                          :id="`field-multiple-${index}`"
                          :checked="field.options.includes('multiple')"
                          @update:model-value="
                            (val) => {
                              if (val === true) {
                                field.options = ['multiple']
                              } else {
                                field.options = []
                              }
                            }
                          "
                        />
                        <Label :for="`field-multiple-${index}`" class="text-sm font-normal">
                          Allow multiple media selection
                        </Label>
                      </div>
                    </div>

                    <div class="flex items-center justify-between">
                      <div class="flex items-center space-x-2">
                        <Checkbox
                          :id="`field-required-${index}`"
                          :checked="field.is_required"
                          @update:model-value="
                            (v) => {
                              if (v === true) {
                                field.is_required = true
                              } else {
                                field.is_required = false
                              }
                            }
                          "
                        />
                        <Label :for="`field-required-${index}`" class="text-sm font-normal">
                          Required field
                        </Label>
                      </div>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        class="text-destructive hover:bg-destructive/10 hover:text-destructive"
                        @click="removeField(index)"
                      >
                        <Trash2 class="mr-1 h-3 w-3" />
                        Remove
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              @click="showNewTypeDialog = false"
              :disabled="creating"
            >
              Cancel
            </Button>
            <Button type="submit" :disabled="creating || !newType.name || !newType.slug">
              <Loader2 v-if="creating" class="mr-2 h-4 w-4 animate-spin" />
              Create Content Type
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </SidebarProvider>
</template>
