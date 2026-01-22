<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useRepoStore, type ContentType, type ContentTypeField } from '@/stores/repo'
import { usePagesStore } from '@/stores/pages'
import axios from 'axios'
import { toast } from 'vue-sonner'
import { Plus, Pencil, FileText, Loader2, Copy, Check } from 'lucide-vue-next'

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
  await fetchValues()
  loading.value = false
  dataFetched.value = true
}

// Watch for route changes to refetch data
watch(
  () => route.params.ctSlug,
  async (newCtSlug, oldSlug) => {
    if (newCtSlug && newCtSlug !== oldSlug) {
      dataFetched.value = false
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
  }
}

async function fetchValues(page = 1) {
  const { owner, repo, ctSlug } = route.params
  try {
    const response = await axios.get(
      `/api/${String(owner)}/${String(repo)}/${String(ctSlug)}?page=${page}`,
    )
    values.value = response.data.items || []
  } catch (error: unknown) {
    console.error('Failed to fetch values', error)
    toast.error('Failed to load content entries')
    values.value = []
  }
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
    await fetchValues()
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
              {{ values.length }} {{ values.length === 1 ? 'entry' : 'entries' }}
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
      <Card v-else-if="values.length === 0">
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
                <TableHead class="w-25">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="(item, index) in values" :key="item.id">
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
                    <Button @click="editValue(item)" variant="ghost" size="sm">
                      <Pencil class="mr-1 h-3 w-3" />
                      Edit
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
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
          <ScrollBar orientation="horizontal" />
        </ScrollArea>
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
