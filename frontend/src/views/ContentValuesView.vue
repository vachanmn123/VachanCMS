<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useRepoStore, type ContentType, type ContentTypeField } from '@/stores/repo'
import { usePagesStore } from '@/stores/pages'
import axios from 'axios'
import { toast } from 'vue-sonner'
import {
  Plus,
  Pencil,
  FileText,
  Loader2,
  Copy,
  Check,
  Trash2,
  MoveVertical,
  ChevronLeft,
  ChevronRight,
} from 'lucide-vue-next'

import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { fieldComponents } from '@/utils/fieldComponents'
import ApiInfoPanel from '@/components/ApiInfoPanel.vue'

interface ContentValue {
  id: string
  slug?: string
  values: Record<string, unknown>
}

interface PaginatedResponse {
  page: number
  items: ContentValue[]
  total_pages: number
  total_items: number
}

const route = useRoute()
const repoStore = useRepoStore()
const pagesStore = usePagesStore()

const selectedTypeName = ref('')
const selectedTypeFields = ref<ContentTypeField[]>([])
const values = ref<ContentValue[]>([])
const showAddValueDialog = ref(false)
const editingValue = ref<ContentValue | null>(null)
const newValue = ref<Record<string, unknown>>({})
const newSlug = ref('')
const loading = ref(true)
const saving = ref(false)
const dataFetched = ref(false)
const copiedUrlIndex = ref<number | null>(null)
const slugError = ref('')

// Pagination state
const currentPage = ref(1)
const totalPages = ref(1)
const totalItems = ref(0)
const itemsPerPage = ref(10)

// Delete state
const showDeleteDialog = ref(false)
const deletingItem = ref<ContentValue | null>(null)
const deleting = ref(false)

// Move/Reorder state
const showMoveDialog = ref(false)
const movingItem = ref<ContentValue | null>(null)
const movePosition = ref(1)
const moving = ref(false)
const moveError = ref('')

const dialogTitle = computed(() => (editingValue.value ? 'Edit Entry' : 'Add New Entry'))
const dialogDescription = computed(() =>
  editingValue.value
    ? 'Update the values for this content entry.'
    : `Create a new entry for "${selectedTypeName.value}".`,
)

const ctSlug = computed(() => String(route.params.ctSlug || ''))
const owner = computed(() => String(route.params.owner || ''))
const repo = computed(() => String(route.params.repo || ''))
const pagesSettingsUrl = computed(
  () => `https://github.com/${owner.value}/${repo.value}/settings/pages`,
)

// Calculate the current item's position in the overall order
function getItemPosition(index: number): number {
  return (currentPage.value - 1) * itemsPerPage.value + index + 1
}

// Validate slug format: lowercase alphanumeric with hyphens
function validateSlugFormat(slug: string): boolean {
  if (slug === '') return true // empty is valid (optional)
  return /^[a-z0-9]+(-[a-z0-9]+)*$/.test(slug)
}

function handleSlugInput(value: string | number) {
  const stringValue = String(value)
  newSlug.value = stringValue
  if (stringValue && !validateSlugFormat(stringValue)) {
    slugError.value = 'Slug must be lowercase alphanumeric with hyphens (e.g., my-blog-post)'
  } else {
    slugError.value = ''
  }
}

function getEntryUrl(item: ContentValue): string {
  if (!pagesStore.baseUrl) return ''
  const base = pagesStore.baseUrl.endsWith('/') ? pagesStore.baseUrl : `${pagesStore.baseUrl}/`
  // Prefer slug over ID for URL
  const identifier = item.slug || item.id
  return `${base}data/${ctSlug.value}/${identifier}.json`
}

function getEntryIdentifier(item: ContentValue): string {
  // Prefer slug over ID for display
  return item.slug || item.id
}

async function copyEntryUrl(item: ContentValue, index: number) {
  const url = getEntryUrl(item)
  if (!url) return

  try {
    await navigator.clipboard.writeText(url)
    copiedUrlIndex.value = index
    setTimeout(() => {
      copiedUrlIndex.value = null
    }, 2000)
  } catch {
    console.error('Failed to copy to clipboard')
  }
}

// Main data loading function
async function loadData() {
  if (!repoStore.config || !route.params.ctSlug) {
    return
  }

  loading.value = true
  await fetchTypeInfo()
  await fetchValues(currentPage.value)
  loading.value = false
  dataFetched.value = true
}

// Watch for route changes to refetch data
watch(
  () => route.params.ctSlug,
  async (newCtSlug, oldSlug) => {
    if (newCtSlug && newCtSlug !== oldSlug) {
      dataFetched.value = false
      currentPage.value = 1
      await loadData()
    }
  },
)

// Watch for config to be loaded (handles direct page load/refresh)
watch(
  () => repoStore.config,
  async (newConfig) => {
    if (newConfig && !dataFetched.value) {
      await loadData()
    }
  },
)

onMounted(async () => {
  await loadData()
})

async function fetchTypeInfo() {
  const { ctSlug } = route.params
  const type = repoStore.config?.content_types.find((t: ContentType) => t.slug === String(ctSlug))
  if (type) {
    selectedTypeName.value = type.name
    selectedTypeFields.value = type.fields
    itemsPerPage.value = type.items_per_page || 10
  }
}

async function fetchValues(page = 1) {
  const { owner, repo, ctSlug } = route.params
  try {
    const response = await axios.get<PaginatedResponse>(
      `/api/${String(owner)}/${String(repo)}/${String(ctSlug)}?page=${page}`,
    )
    values.value = response.data.items || []
    totalPages.value = response.data.total_pages || 1
    totalItems.value = response.data.total_items || 0
    currentPage.value = page
  } catch (error: unknown) {
    console.error('Failed to fetch values', error)
    toast.error('Failed to load content entries')
    values.value = []
  }
}

async function goToPage(page: number) {
  if (page < 1 || page > totalPages.value) return
  loading.value = true
  await fetchValues(page)
  loading.value = false
}

function openAddDialog() {
  editingValue.value = null
  newValue.value = {}
  newSlug.value = ''
  slugError.value = ''
  // Initialize default values for fields
  selectedTypeFields.value.forEach((field) => {
    if (field.field_type === 'boolean') {
      newValue.value[field.field_name] = false
    } else if (field.field_type === 'number') {
      newValue.value[field.field_name] = 0
    } else if (field.field_type === 'media') {
      // Initialize as empty array for multiple, empty string for single
      const isMultiple = field.options?.includes('multiple')
      newValue.value[field.field_name] = isMultiple ? [] : ''
    } else {
      newValue.value[field.field_name] = ''
    }
  })
  showAddValueDialog.value = true
}

function editValue(item: ContentValue) {
  editingValue.value = item
  newValue.value = { ...item.values }
  newSlug.value = item.slug || ''
  slugError.value = ''
  showAddValueDialog.value = true
}

function closeDialog() {
  showAddValueDialog.value = false
  editingValue.value = null
  newValue.value = {}
  newSlug.value = ''
  slugError.value = ''
}

async function saveValue() {
  // Validate slug before saving
  if (newSlug.value && !validateSlugFormat(newSlug.value)) {
    slugError.value = 'Slug must be lowercase alphanumeric with hyphens (e.g., my-blog-post)'
    return
  }

  const { owner, repo, ctSlug } = route.params
  saving.value = true

  try {
    const payload: { values: Record<string, unknown>; slug?: string } = {
      values: newValue.value,
    }
    if (newSlug.value) {
      payload.slug = newSlug.value
    }

    if (editingValue.value) {
      await axios.put(
        `/api/${String(owner)}/${String(repo)}/${String(ctSlug)}/${editingValue.value.id}`,
        payload,
      )
      toast.success('Entry updated', {
        description: 'The content entry has been updated successfully.',
      })
    } else {
      await axios.post(`/api/${String(owner)}/${String(repo)}/${String(ctSlug)}`, payload)
      toast.success('Entry created', {
        description: 'A new content entry has been created.',
      })
    }
    closeDialog()
    await fetchValues(currentPage.value)
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { error?: string } } }
    const errorMessage =
      axiosError.response?.data?.error || 'Please try again or check your connection.'
    toast.error('Failed to save entry', {
      description: errorMessage,
    })
  } finally {
    saving.value = false
  }
}

// Delete functionality
function openDeleteDialog(item: ContentValue) {
  deletingItem.value = item
  showDeleteDialog.value = true
}

function closeDeleteDialog() {
  showDeleteDialog.value = false
  deletingItem.value = null
}

async function confirmDelete() {
  if (!deletingItem.value) return

  const { owner, repo, ctSlug } = route.params
  deleting.value = true

  try {
    await axios.delete(
      `/api/${String(owner)}/${String(repo)}/${String(ctSlug)}/${deletingItem.value.id}`,
    )
    toast.success('Entry deleted', {
      description: 'The content entry has been deleted successfully.',
    })
    closeDeleteDialog()
    // If we deleted the last item on the current page and it's not page 1, go to previous page
    if (values.value.length === 1 && currentPage.value > 1) {
      await fetchValues(currentPage.value - 1)
    } else {
      await fetchValues(currentPage.value)
    }
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { error?: string } } }
    const errorMessage =
      axiosError.response?.data?.error || 'Please try again or check your connection.'
    toast.error('Failed to delete entry', {
      description: errorMessage,
    })
  } finally {
    deleting.value = false
  }
}

// Move/Reorder functionality
function openMoveDialog(item: ContentValue, index: number) {
  movingItem.value = item
  movePosition.value = getItemPosition(index)
  moveError.value = ''
  showMoveDialog.value = true
}

function closeMoveDialog() {
  showMoveDialog.value = false
  movingItem.value = null
  movePosition.value = 1
  moveError.value = ''
}

function handleMovePositionInput(value: string | number) {
  const numValue = typeof value === 'string' ? parseInt(value, 10) : value
  movePosition.value = numValue

  if (isNaN(numValue) || numValue < 1 || numValue > totalItems.value) {
    moveError.value = `Position must be between 1 and ${totalItems.value}`
  } else {
    moveError.value = ''
  }
}

async function confirmMove() {
  if (!movingItem.value) return

  if (movePosition.value < 1 || movePosition.value > totalItems.value) {
    moveError.value = `Position must be between 1 and ${totalItems.value}`
    return
  }

  const { owner, repo, ctSlug } = route.params
  moving.value = true

  try {
    await axios.put(
      `/api/${String(owner)}/${String(repo)}/${String(ctSlug)}/${movingItem.value.id}/reorder`,
      { position: movePosition.value },
    )
    toast.success('Entry moved', {
      description: `The entry has been moved to position ${movePosition.value}.`,
    })
    closeMoveDialog()

    // Navigate to the page where the item is now located
    const newPage = Math.ceil(movePosition.value / itemsPerPage.value)
    await fetchValues(newPage)
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { error?: string } } }
    const errorMessage =
      axiosError.response?.data?.error || 'Please try again or check your connection.'
    toast.error('Failed to move entry', {
      description: errorMessage,
    })
  } finally {
    moving.value = false
  }
}

function formatCellValue(value: unknown): string {
  if (value === null || value === undefined) return '-'
  if (typeof value === 'boolean') return value ? 'Yes' : 'No'
  if (Array.isArray(value)) {
    if (value.length === 0) return '-'
    return `${value.length} item${value.length === 1 ? '' : 's'}`
  }
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}
</script>

<template>
  <div class="flex gap-5 h-full">
    <!-- Main Content -->
    <div class="flex-1 space-y-6 overflow-auto">
      <!-- Header -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">{{ selectedTypeName }}</h1>
          <p class="text-sm text-muted-foreground">
            Manage content entries for this type
            <Badge v-if="!loading" variant="secondary" class="ml-2">
              {{ totalItems }} {{ totalItems === 1 ? 'entry' : 'entries' }}
            </Badge>
          </p>
        </div>
        <Button @click="openAddDialog" :disabled="loading">
          <Plus class="mr-2 h-4 w-4" />
          Add Entry
        </Button>
      </div>

      <!-- Loading State -->
      <Card v-if="loading">
        <CardHeader>
          <Skeleton class="h-6 w-48" />
          <Skeleton class="h-4 w-32" />
        </CardHeader>
        <CardContent>
          <div class="space-y-3">
            <Skeleton class="h-10 w-full" />
            <Skeleton class="h-10 w-full" />
            <Skeleton class="h-10 w-full" />
          </div>
        </CardContent>
      </Card>

      <!-- Empty State -->
      <Card v-else-if="values.length === 0 && currentPage === 1">
        <CardContent class="flex flex-col items-center justify-center py-16">
          <div class="rounded-full bg-muted p-4">
            <FileText class="h-8 w-8 text-muted-foreground" />
          </div>
          <CardTitle class="mt-4 text-lg">No entries yet</CardTitle>
          <CardDescription class="mt-2 max-w-sm text-center">
            Start adding content entries for "{{ selectedTypeName }}". Each entry will be stored in
            your GitHub repository.
          </CardDescription>
          <Button class="mt-6" @click="openAddDialog">
            <Plus class="mr-2 h-4 w-4" />
            Add your first entry
          </Button>
        </CardContent>
      </Card>

      <!-- Data Table -->
      <Card v-else>
        <ScrollArea class="w-full">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead class="w-12 text-center">#</TableHead>
                <TableHead class="whitespace-nowrap">Slug</TableHead>
                <TableHead
                  v-for="field in selectedTypeFields"
                  :key="field.field_name"
                  class="whitespace-nowrap"
                >
                  {{ field.field_name }}
                  <Badge v-if="field.is_required" variant="outline" class="ml-1 text-xs">
                    req
                  </Badge>
                </TableHead>
                <TableHead class="w-40">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="(item, index) in values" :key="item.id">
                <TableCell class="text-center text-muted-foreground text-xs">
                  {{ getItemPosition(index) }}
                </TableCell>
                <TableCell class="max-w-32 truncate font-mono text-xs">
                  {{ item.slug || '-' }}
                </TableCell>
                <TableCell
                  v-for="field in selectedTypeFields"
                  :key="field.field_name"
                  class="max-w-50 truncate"
                >
                  {{ formatCellValue(item.values[field.field_name]) }}
                </TableCell>
                <TableCell>
                  <div class="flex items-center gap-1">
                    <Button @click="editValue(item)" variant="ghost" size="sm" title="Edit">
                      <Pencil class="h-3 w-3" />
                    </Button>
                    <Button
                      @click="openMoveDialog(item, index)"
                      variant="ghost"
                      size="sm"
                      title="Move to position"
                    >
                      <MoveVertical class="h-3 w-3" />
                    </Button>
                    <Button
                      v-if="pagesStore.isInitialized"
                      @click="copyEntryUrl(item, index)"
                      variant="ghost"
                      size="sm"
                      :title="
                        copiedUrlIndex === index
                          ? 'Copied!'
                          : `Copy URL (${getEntryIdentifier(item)})`
                      "
                    >
                      <Check v-if="copiedUrlIndex === index" class="h-3 w-3 text-green-500" />
                      <Copy v-else class="h-3 w-3" />
                    </Button>
                    <Button
                      @click="openDeleteDialog(item)"
                      variant="ghost"
                      size="sm"
                      class="text-destructive hover:text-destructive hover:bg-destructive/10"
                      title="Delete"
                    >
                      <Trash2 class="h-3 w-3" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
          <ScrollBar orientation="horizontal" />
        </ScrollArea>

        <!-- Pagination Controls -->
        <div v-if="totalPages > 1" class="flex items-center justify-between border-t px-4 py-3">
          <div class="text-sm text-muted-foreground">
            Page {{ currentPage }} of {{ totalPages }}
          </div>
          <div class="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              :disabled="currentPage <= 1 || loading"
              @click="goToPage(currentPage - 1)"
            >
              <ChevronLeft class="h-4 w-4" />
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              :disabled="currentPage >= totalPages || loading"
              @click="goToPage(currentPage + 1)"
            >
              Next
              <ChevronRight class="h-4 w-4" />
            </Button>
          </div>
        </div>
      </Card>

      <!-- Add/Edit Value Dialog -->
      <Dialog v-model:open="showAddValueDialog" @update:open="(open) => !open && closeDialog()">
        <DialogContent class="max-h-[90vh] overflow-y-auto sm:max-w-125">
          <DialogHeader>
            <DialogTitle>{{ dialogTitle }}</DialogTitle>
            <DialogDescription>{{ dialogDescription }}</DialogDescription>
          </DialogHeader>

          <form @submit.prevent="saveValue" class="space-y-4">
            <div class="grid gap-4 py-2">
              <!-- Slug Field -->
              <div class="space-y-2">
                <Label for="slug">
                  Slug
                  <span class="text-muted-foreground text-xs ml-1">(optional)</span>
                </Label>
                <Input
                  id="slug"
                  :model-value="newSlug"
                  @update:model-value="handleSlugInput"
                  placeholder="my-blog-post"
                  :class="{ 'border-destructive': slugError }"
                />
                <p v-if="slugError" class="text-xs text-destructive">{{ slugError }}</p>
                <p v-else class="text-xs text-muted-foreground">
                  A URL-friendly identifier. Use lowercase letters, numbers, and hyphens.
                </p>
              </div>

              <!-- Dynamic Fields -->
              <component
                v-for="field in selectedTypeFields"
                :key="field.field_name"
                :is="fieldComponents[field.field_type]"
                :label="field.field_name"
                :required="field.is_required"
                :model-value="newValue[field.field_name]"
                :options="field.options"
                @update:model-value="newValue[field.field_name] = $event"
              />
            </div>

            <DialogFooter class="gap-2 sm:gap-0">
              <Button type="button" variant="outline" @click="closeDialog" :disabled="saving">
                Cancel
              </Button>
              <Button type="submit" :disabled="saving || !!slugError">
                <Loader2 v-if="saving" class="mr-2 h-4 w-4 animate-spin" />
                {{ editingValue ? 'Update' : 'Create' }}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <!-- Delete Confirmation Dialog -->
      <Dialog v-model:open="showDeleteDialog">
        <DialogContent class="sm:max-w-100">
          <DialogHeader>
            <DialogTitle>Delete Entry</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this entry? This action cannot be undone.
              <span v-if="deletingItem?.slug" class="block mt-2 font-mono text-sm">
                Slug: {{ deletingItem.slug }}
              </span>
            </DialogDescription>
          </DialogHeader>
          <DialogFooter class="gap-2 sm:gap-0">
            <Button variant="outline" :disabled="deleting" @click="closeDeleteDialog">
              Cancel
            </Button>
            <Button :disabled="deleting" variant="destructive" @click="confirmDelete">
              <Loader2 v-if="deleting" class="mr-2 h-4 w-4 animate-spin" />
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <!-- Move/Reorder Dialog -->
      <Dialog v-model:open="showMoveDialog" @update:open="(open) => !open && closeMoveDialog()">
        <DialogContent class="sm:max-w-100">
          <DialogHeader>
            <DialogTitle>Move Entry</DialogTitle>
            <DialogDescription>
              Enter the new position for this entry (1 to {{ totalItems }}).
              <span v-if="movingItem?.slug" class="block mt-1 font-mono text-sm">
                Moving: {{ movingItem.slug }}
              </span>
            </DialogDescription>
          </DialogHeader>

          <form @submit.prevent="confirmMove" class="space-y-4">
            <div class="space-y-2">
              <Label for="movePosition">New Position</Label>
              <Input
                id="movePosition"
                type="number"
                :model-value="movePosition"
                @update:model-value="handleMovePositionInput"
                :min="1"
                :max="totalItems"
                :class="{ 'border-destructive': moveError }"
              />
              <p v-if="moveError" class="text-xs text-destructive">{{ moveError }}</p>
              <p v-else class="text-xs text-muted-foreground">
                Position 1 is the first item, {{ totalItems }} is the last.
              </p>
            </div>

            <DialogFooter class="gap-2 sm:gap-0">
              <Button type="button" variant="outline" @click="closeMoveDialog" :disabled="moving">
                Cancel
              </Button>
              <Button type="submit" :disabled="moving || !!moveError">
                <Loader2 v-if="moving" class="mr-2 h-4 w-4 animate-spin" />
                Move
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>

    <!-- API Info Panel -->
    <ApiInfoPanel
      :ct-slug="ctSlug"
      :base-url="pagesStore.baseUrl"
      :is-initialized="pagesStore.isInitialized"
      :pages-settings-url="pagesSettingsUrl"
    />
  </div>
</template>
