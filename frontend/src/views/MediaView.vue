<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { usePagesStore } from '@/stores/pages'
import axios from 'axios'
import { toast } from 'vue-sonner'
import type { AcceptableValue } from 'reka-ui'
import {
  Upload,
  LayoutGrid,
  List,
  File,
  FileText,
  FileCode,
  Video,
  Music,
  Image as ImageIcon,
  Download,
  X,
  Loader2,
  Copy,
  Check,
} from 'lucide-vue-next'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardTitle } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import { AspectRatio } from '@/components/ui/aspect-ratio'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import MediaApiPanel from '@/components/MediaApiPanel.vue'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area'

interface MediaFile {
  id: string
  file_name: string
  file_type: string
}

interface MediaResponse {
  page: number
  media: MediaFile[]
}

const route = useRoute()
const pagesStore = usePagesStore()

const owner = computed(() => String(route.params.owner))
const repo = computed(() => String(route.params.repo))
const pagesSettingsUrl = computed(
  () => `https://github.com/${owner.value}/${repo.value}/settings/pages`,
)

// State
const media = ref<MediaFile[]>([])
const loading = ref(true)
const uploading = ref(false)
const currentPage = ref(1)
const totalPages = ref(1)
const viewMode = ref<'grid' | 'list'>('grid')
const showUploadDialog = ref(false)
const showPreviewDialog = ref(false)
const selectedMedia = ref<MediaFile | null>(null)
const dragOver = ref(false)
const selectedFile = ref<File | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)
const copiedUrlIndex = ref<number | null>(null)

const MAX_FILE_SIZE = 100 * 1024 * 1024 // 100MB

// Computed
const isFileTooLarge = computed(() => {
  if (!selectedFile.value) return false
  return selectedFile.value.size > MAX_FILE_SIZE
})

const mediaUrl = computed(() => {
  if (!selectedMedia.value) return ''
  return `/api/${owner.value}/${repo.value}/media/${selectedMedia.value.id}`
})

// Helpers
function isImage(type: string): boolean {
  return type.startsWith('image/')
}

function getFileIcon(type: string) {
  if (type.startsWith('image/')) return ImageIcon
  if (type.startsWith('video/')) return Video
  if (type.startsWith('audio/')) return Music
  if (type === 'application/pdf') return FileText
  if (type.startsWith('text/')) return FileCode
  return File
}

function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`
}

function getMediaThumbnailUrl(item: MediaFile): string {
  return `/api/${owner.value}/${repo.value}/media/${item.id}`
}

// API Methods
async function fetchMedia(page: number = 1) {
  loading.value = true
  try {
    const response = await axios.get<MediaResponse>(
      `/api/${owner.value}/${repo.value}/media?page=${page}`,
    )
    media.value = response.data.media || []
    currentPage.value = response.data.page
    totalPages.value = response.data.total_pages
  } catch (error: unknown) {
    const axiosError = error as { response?: { status?: number } }
    if (axiosError.response?.status === 404 || axiosError.response?.status === 500) {
      // No media yet or config doesn't exist
      media.value = []
      totalPages.value = 1
    } else {
      toast.error('Failed to load media')
    }
  } finally {
    loading.value = false
  }
}

async function uploadFile() {
  if (!selectedFile.value || isFileTooLarge.value) return

  uploading.value = true
  const formData = new FormData()
  formData.append('file', selectedFile.value)

  try {
    await axios.post(`/api/${owner.value}/${repo.value}/media`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    toast.success('File uploaded', {
      description: `"${selectedFile.value.name}" has been uploaded successfully.`,
    })
    closeUploadDialog()
    await fetchMedia(currentPage.value)
  } catch {
    toast.error('Failed to upload file', {
      description: 'Please try again or check your connection.',
    })
  } finally {
    uploading.value = false
  }
}

// Dialog handlers
function openUploadDialog() {
  selectedFile.value = null
  showUploadDialog.value = true
}

function closeUploadDialog() {
  showUploadDialog.value = false
  selectedFile.value = null
  dragOver.value = false
}

function openPreview(item: MediaFile) {
  selectedMedia.value = item
  showPreviewDialog.value = true
}

function closePreview() {
  showPreviewDialog.value = false
  selectedMedia.value = null
}

function downloadFile(item: MediaFile) {
  const url = getMediaThumbnailUrl(item)
  const link = document.createElement('a')
  link.href = url
  link.download = item.file_name
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

// Drag and drop handlers
function handleDragOver(e: DragEvent) {
  e.preventDefault()
  dragOver.value = true
}

function handleDragLeave(e: DragEvent) {
  e.preventDefault()
  dragOver.value = false
}

function handleDrop(e: DragEvent) {
  e.preventDefault()
  dragOver.value = false
  const files = e.dataTransfer?.files
  if (files && files.length > 0) {
    selectedFile.value = files[0] as File
  }
}

function handleFileSelect(e: Event) {
  const target = e.target as HTMLInputElement
  const files = target.files
  if (files && files.length > 0) {
    selectedFile.value = files[0] as File
  }
}

function triggerFileInput() {
  fileInputRef.value?.click()
}

function getMediaUrl(mediaId: string): string {
  if (!pagesStore.baseUrl) return ''
  const base = pagesStore.baseUrl.endsWith('/') ? pagesStore.baseUrl : `${pagesStore.baseUrl}/`
  return `${base}media/${mediaId}`
}

async function copyMediaUrl(mediaId: string, index: number) {
  const url = getMediaUrl(mediaId)
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

// View mode persistence
function setViewMode(mode: AcceptableValue | AcceptableValue[]) {
  const modeStr = Array.isArray(mode) ? mode[0] : mode
  if (modeStr === 'grid' || modeStr === 'list') {
    viewMode.value = modeStr
    localStorage.setItem('media-view-mode', modeStr)
  }
}

// Pagination
function goToPage(page: number) {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
    fetchMedia(page)
  }
}

// Lifecycle
onMounted(() => {
  const savedMode = localStorage.getItem('media-view-mode')
  if (savedMode === 'grid' || savedMode === 'list') {
    viewMode.value = savedMode
  }
  fetchMedia(1)
})
</script>

<template>
  <div class="flex gap-5 h-full">
    <!-- Main Content -->
    <div class="flex-1 space-y-6 overflow-auto">
      <!-- Header -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">Media Library</h1>
          <p class="text-sm text-muted-foreground">
            Upload and manage your media files
            <Badge v-if="!loading" variant="secondary" class="ml-2">
              {{ media.length }} {{ media.length === 1 ? 'file' : 'files' }}
            </Badge>
          </p>
        </div>
        <div class="flex items-center gap-2">
          <ToggleGroup type="single" :model-value="viewMode" @update:model-value="setViewMode">
            <ToggleGroupItem value="grid" aria-label="Grid view">
              <LayoutGrid class="h-4 w-4" />
            </ToggleGroupItem>
            <ToggleGroupItem value="list" aria-label="List view">
              <List class="h-4 w-4" />
            </ToggleGroupItem>
          </ToggleGroup>
          <Button @click="openUploadDialog" :disabled="loading">
            <Upload class="mr-2 h-4 w-4" />
            Upload
          </Button>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading">
        <div
          v-if="viewMode === 'grid'"
          class="grid gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4"
        >
          <Skeleton v-for="i in 8" :key="i" class="aspect-square rounded-lg" />
        </div>
        <Card v-else>
          <CardContent class="p-4">
            <div class="space-y-3">
              <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- Empty State -->
      <Card v-else-if="media.length === 0">
        <CardContent class="flex flex-col items-center justify-center py-16">
          <div class="rounded-full bg-muted p-4">
            <ImageIcon class="h-8 w-8 text-muted-foreground" />
          </div>
          <CardTitle class="mt-4 text-lg">No media files yet</CardTitle>
          <CardDescription class="mt-2 max-w-sm text-center">
            Upload images, documents, and other files to your media library. Files are stored in
            your GitHub repository.
          </CardDescription>
          <Button class="mt-6" @click="openUploadDialog">
            <Upload class="mr-2 h-4 w-4" />
            Upload your first file
          </Button>
        </CardContent>
      </Card>

      <!-- Grid View -->
      <div
        v-else-if="viewMode === 'grid'"
        class="grid gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4"
      >
        <Card
          v-for="(item, index) in media"
          :key="item.id"
          class="group cursor-pointer overflow-hidden transition-colors hover:bg-accent"
          @click="openPreview(item)"
        >
          <AspectRatio :ratio="1" class="bg-muted">
            <img
              v-if="isImage(item.file_type)"
              :src="getMediaThumbnailUrl(item)"
              :alt="item.file_name"
              class="h-full w-full object-cover"
            />
            <div v-else class="flex h-full w-full items-center justify-center">
              <component
                :is="getFileIcon(item.file_type)"
                class="h-12 w-12 text-muted-foreground"
              />
            </div>
          </AspectRatio>
          <CardContent class="p-3">
            <p class="truncate text-sm font-medium">{{ item.file_name }}</p>
            <p class="truncate text-xs text-muted-foreground">{{ item.file_type }}</p>
            <div class="mt-2 flex items-center justify-between">
              <div class="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="sm"
                  class="h-6 w-6 p-0"
                  @click.stop="openPreview(item)"
                >
                  <Download class="h-3 w-3" />
                </Button>
                <Button
                  v-if="pagesStore.isInitialized"
                  variant="ghost"
                  size="sm"
                  class="h-6 w-6 p-0"
                  @click.stop="copyMediaUrl(item.id, index)"
                  :title="copiedUrlIndex === index ? 'Copied!' : 'Copy URL'"
                >
                  <Check v-if="copiedUrlIndex === index" class="h-3 w-3 text-green-500" />
                  <Copy v-else class="h-3 w-3" />
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- List View -->
      <Card v-else>
        <ScrollArea class="w-full">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead class="w-12.5">Type</TableHead>
                <TableHead>Name</TableHead>
                <TableHead class="hidden sm:table-cell">MIME Type</TableHead>
                <TableHead class="w-37.5">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow
                v-for="(item, index) in media"
                :key="item.id"
                class="cursor-pointer"
                @click="openPreview(item)"
              >
                <TableCell>
                  <component
                    :is="getFileIcon(item.file_type)"
                    class="h-5 w-5 text-muted-foreground"
                  />
                </TableCell>
                <TableCell class="max-w-50 truncate font-medium">
                  {{ item.file_name }}
                </TableCell>
                <TableCell class="hidden text-muted-foreground sm:table-cell">
                  {{ item.file_type }}
                </TableCell>
                <TableCell>
                  <div class="flex items-center gap-1">
                    <Button variant="ghost" size="sm" @click.stop="openPreview(item)">
                      View
                    </Button>
                    <Button variant="ghost" size="sm" @click.stop="downloadFile(item)">
                      <Download class="h-4 w-4" />
                    </Button>
                    <Button
                      v-if="pagesStore.isInitialized"
                      variant="ghost"
                      size="sm"
                      @click.stop="copyMediaUrl(item.id, index)"
                      :title="copiedUrlIndex === index ? 'Copied!' : 'Copy URL'"
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

      <!-- Pagination -->
      <div v-if="!loading && media.length > 0 && totalPages > 1" class="flex justify-center">
        <div class="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            :disabled="currentPage <= 1"
            @click="goToPage(currentPage - 1)"
          >
            Previous
          </Button>
          <span class="px-4 text-sm text-muted-foreground">
            Page {{ currentPage }} of {{ totalPages }}
          </span>
          <Button
            variant="outline"
            size="sm"
            :disabled="currentPage >= totalPages"
            @click="goToPage(currentPage + 1)"
          >
            Next
          </Button>
        </div>
      </div>
    </div>

    <!-- Upload Dialog -->
    <Dialog v-model:open="showUploadDialog" @update:open="(open) => !open && closeUploadDialog()">
      <DialogContent class="sm:max-w-125">
        <DialogHeader>
          <DialogTitle>Upload Media</DialogTitle>
          <DialogDescription>
            Drag and drop a file or click to browse. Maximum file size: 100MB.
          </DialogDescription>
        </DialogHeader>

        <div class="space-y-4 py-4">
          <!-- Drop Zone -->
          <div
            class="relative flex min-h-50 cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed transition-colors"
            :class="[
              dragOver
                ? 'border-primary bg-primary/5'
                : 'border-muted-foreground/25 hover:border-primary/50',
            ]"
            @dragover="handleDragOver"
            @dragleave="handleDragLeave"
            @drop="handleDrop"
            @click="triggerFileInput"
          >
            <input ref="fileInputRef" type="file" class="hidden" @change="handleFileSelect" />
            <Upload class="mb-4 h-10 w-10 text-muted-foreground" />
            <p class="text-sm font-medium">Drag & drop a file here</p>
            <p class="mt-1 text-xs text-muted-foreground">or click to browse</p>
          </div>

          <!-- Selected File Info -->
          <div v-if="selectedFile" class="rounded-lg border p-3">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3 overflow-hidden">
                <component
                  :is="getFileIcon(selectedFile.type || 'application/octet-stream')"
                  class="h-8 w-8 shrink-0 text-muted-foreground"
                />
                <div class="min-w-0">
                  <p class="truncate text-sm font-medium">{{ selectedFile.name }}</p>
                  <p class="text-xs text-muted-foreground">
                    {{ formatFileSize(selectedFile.size) }}
                  </p>
                </div>
              </div>
              <Button variant="ghost" size="sm" @click.stop="selectedFile = null">
                <X class="h-4 w-4" />
              </Button>
            </div>
            <p v-if="isFileTooLarge" class="mt-2 text-sm text-destructive">
              File size exceeds 100MB limit. Please choose a smaller file.
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" @click="closeUploadDialog" :disabled="uploading">
            Cancel
          </Button>
          <Button @click="uploadFile" :disabled="!selectedFile || isFileTooLarge || uploading">
            <Loader2 v-if="uploading" class="mr-2 h-4 w-4 animate-spin" />
            Upload
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Preview Dialog -->
    <Dialog v-model:open="showPreviewDialog" @update:open="(open) => !open && closePreview()">
      <DialogContent class="sm:max-w-150">
        <DialogHeader>
          <DialogTitle class="truncate pr-8">{{ selectedMedia?.file_name }}</DialogTitle>
          <DialogDescription>{{ selectedMedia?.file_type }}</DialogDescription>
        </DialogHeader>

        <div class="py-4">
          <!-- Image Preview -->
          <div
            v-if="selectedMedia && isImage(selectedMedia.file_type)"
            class="overflow-hidden rounded-lg bg-muted"
          >
            <img
              :src="mediaUrl"
              :alt="selectedMedia.file_name"
              class="mx-auto max-h-100 object-contain"
            />
          </div>

          <!-- Non-Image Preview -->
          <div
            v-else-if="selectedMedia"
            class="flex flex-col items-center justify-center rounded-lg bg-muted py-16"
          >
            <component
              :is="getFileIcon(selectedMedia.file_type)"
              class="h-16 w-16 text-muted-foreground"
            />
            <p class="mt-4 text-sm text-muted-foreground">Preview not available</p>
          </div>

          <!-- File Details -->
          <div v-if="selectedMedia" class="mt-4 space-y-2 text-sm">
            <div class="flex justify-between">
              <span class="text-muted-foreground">Type:</span>
              <span>{{ selectedMedia.file_type }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-muted-foreground">ID:</span>
              <span class="font-mono text-xs">{{ selectedMedia.id }}</span>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" @click="closePreview()">Close</Button>
          <Button v-if="selectedMedia" @click="downloadFile(selectedMedia)">
            <Download class="mr-2 h-4 w-4" />
            Download
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Media API Panel -->
    <MediaApiPanel
      :base-url="pagesStore.baseUrl"
      :is-initialized="pagesStore.isInitialized"
      :pages-settings-url="pagesSettingsUrl"
    />
  </div>
</template>
