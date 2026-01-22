<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import axios from 'axios'
import { toast } from 'vue-sonner'
import { Image as ImageIcon, File, Check, Loader2, Upload, X, FolderOpen } from 'lucide-vue-next'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { AspectRatio } from '@/components/ui/aspect-ratio'
import { ScrollArea } from '@/components/ui/scroll-area'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import type { AcceptableValue } from 'reka-ui'

interface MediaFile {
  id: string
  file_name: string
  file_type: string
}

interface Props {
  open: boolean
  multiple?: boolean
  selected?: MediaFile[]
}

const props = withDefaults(defineProps<Props>(), {
  multiple: false,
  selected: () => [],
})

const emit = defineEmits<{
  'update:open': [value: boolean]
  select: [media: MediaFile[]]
}>()

const route = useRoute()
const owner = computed(() => String(route.params.owner))
const repo = computed(() => String(route.params.repo))

const media = ref<MediaFile[]>([])
const loading = ref(false)
const localSelection = ref<MediaFile[]>([])

// View state
const activeView = ref('browse')

// Upload state
const uploading = ref(false)
const dragOver = ref(false)
const selectedFile = ref<File | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)
const MAX_FILE_SIZE = 100 * 1024 * 1024 // 100MB

const isFileTooLarge = computed(() => {
  if (!selectedFile.value) return false
  return selectedFile.value.size > MAX_FILE_SIZE
})

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      fetchMedia()
      localSelection.value = [...props.selected]
      activeView.value = 'browse'
      selectedFile.value = null
    }
  },
)

async function fetchMedia() {
  loading.value = true
  try {
    const response = await axios.get(`/api/${owner.value}/${repo.value}/media`)
    media.value = response.data.media || []
  } catch {
    media.value = []
  } finally {
    loading.value = false
  }
}

function isSelected(id: string): boolean {
  return localSelection.value.some((m) => m.id === id)
}

function toggleSelection(mediaFile: MediaFile) {
  if (props.multiple) {
    if (isSelected(mediaFile.id)) {
      localSelection.value = localSelection.value.filter((m) => m.id !== mediaFile.id)
    } else {
      // Create a new array to ensure reactivity
      localSelection.value = [...localSelection.value, mediaFile]
    }
  } else {
    localSelection.value = [mediaFile]
  }
}

function confirm() {
  emit('select', localSelection.value)
  emit('update:open', false)
}

function close() {
  emit('update:open', false)
}

function isImage(type: string): boolean {
  return type.startsWith('image/')
}

function getMediaUrl(id: string): string {
  return `/api/${owner.value}/${repo.value}/media/${id}`
}

function setActiveView(value: AcceptableValue) {
  if (value === 'browse' || value === 'upload') {
    activeView.value = value
  }
}

// Upload handlers
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
  if (files && files.length > 0 && files[0]) {
    selectedFile.value = files[0]
  }
}

function handleFileSelect(e: Event) {
  const target = e.target as HTMLInputElement
  const files = target.files
  if (files && files.length > 0 && files[0]) {
    selectedFile.value = files[0]
  }
}

function triggerFileInput() {
  fileInputRef.value?.click()
}

function clearSelectedFile() {
  selectedFile.value = null
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
}

function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

async function uploadFile() {
  if (!selectedFile.value || isFileTooLarge.value) return

  uploading.value = true
  const formData = new FormData()
  formData.append('file', selectedFile.value)

  try {
    const response = await axios.post(`/api/${owner.value}/${repo.value}/media`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })

    toast.success('File uploaded', {
      description: `"${selectedFile.value.name}" has been uploaded successfully.`,
    })

    // Add the newly uploaded file to the selection
    const newMedia: MediaFile = response.data
    if (props.multiple) {
      // Create a new array to ensure reactivity
      localSelection.value = [...localSelection.value, newMedia]
    } else {
      localSelection.value = [newMedia]
    }

    // Refresh media list and switch to browse tab
    await fetchMedia()
    selectedFile.value = null
    activeView.value = 'browse'
  } catch {
    toast.error('Failed to upload file', {
      description: 'Please try again or check your connection.',
    })
  } finally {
    uploading.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="max-h-[80vh] max-w-3xl">
      <DialogHeader>
        <DialogTitle>Select Media</DialogTitle>
        <DialogDescription>
          {{ multiple ? 'Select one or more media files' : 'Select a media file' }}
        </DialogDescription>
      </DialogHeader>

      <!-- View Toggle -->
      <div class="flex justify-center">
        <ToggleGroup type="single" :model-value="activeView" @update:model-value="setActiveView">
          <ToggleGroupItem value="browse" aria-label="Browse library">
            <FolderOpen class="mr-2 h-4 w-4" />
            Browse Library
          </ToggleGroupItem>
          <ToggleGroupItem value="upload" aria-label="Upload new">
            <Upload class="mr-2 h-4 w-4" />
            Upload New
          </ToggleGroupItem>
        </ToggleGroup>
      </div>

      <!-- Browse View -->
      <div v-if="activeView === 'browse'">
        <!-- Loading State -->
        <div v-if="loading" class="flex items-center justify-center py-12">
          <Loader2 class="h-8 w-8 animate-spin text-muted-foreground" />
        </div>

        <!-- Empty State -->
        <div v-else-if="media.length === 0" class="py-12 text-center">
          <ImageIcon class="mx-auto h-12 w-12 text-muted-foreground" />
          <p class="mt-4 text-sm text-muted-foreground">No media files available</p>
          <Button variant="outline" class="mt-4" @click="activeView = 'upload'">
            <Upload class="mr-2 h-4 w-4" />
            Upload your first file
          </Button>
        </div>

        <!-- Media Grid -->
        <ScrollArea v-else class="h-80">
          <div class="grid grid-cols-3 gap-3 p-1 sm:grid-cols-4">
            <div
              v-for="item in media"
              :key="item.id"
              class="relative cursor-pointer rounded-lg border-2 p-1 transition-colors"
              :class="
                isSelected(item.id)
                  ? 'border-primary bg-primary/10'
                  : 'border-transparent hover:border-muted-foreground/30'
              "
              @click="toggleSelection(item)"
            >
              <AspectRatio :ratio="1">
                <img
                  v-if="isImage(item.file_type)"
                  :src="getMediaUrl(item.id)"
                  :alt="item.file_name"
                  class="h-full w-full rounded object-cover"
                />
                <div v-else class="flex h-full items-center justify-center rounded bg-muted">
                  <File class="h-8 w-8 text-muted-foreground" />
                </div>
              </AspectRatio>
              <p class="mt-1 truncate text-xs" :title="item.file_name">{{ item.file_name }}</p>

              <!-- Selection indicator -->
              <div
                v-if="isSelected(item.id)"
                class="absolute right-2 top-2 flex h-5 w-5 items-center justify-center rounded-full bg-primary"
              >
                <Check class="h-3 w-3 text-primary-foreground" />
              </div>
            </div>
          </div>
        </ScrollArea>
      </div>

      <!-- Upload View -->
      <div v-else-if="activeView === 'upload'" class="space-y-4">
        <!-- Drop Zone -->
        <div
          class="relative flex min-h-48 cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed p-6 transition-colors"
          :class="
            dragOver
              ? 'border-primary bg-primary/5'
              : 'border-muted-foreground/25 hover:border-muted-foreground/50'
          "
          @dragover="handleDragOver"
          @dragleave="handleDragLeave"
          @drop="handleDrop"
          @click="triggerFileInput"
        >
          <input
            ref="fileInputRef"
            type="file"
            class="hidden"
            @change="handleFileSelect"
            accept="image/*,video/*,audio/*,application/pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx"
          />

          <template v-if="!selectedFile">
            <Upload class="h-10 w-10 text-muted-foreground" />
            <p class="mt-4 text-sm font-medium">Drag and drop a file here, or click to browse</p>
            <p class="mt-1 text-xs text-muted-foreground">Maximum file size: 100MB</p>
          </template>

          <!-- Selected File Preview -->
          <template v-else>
            <div class="flex w-full items-center gap-4 rounded-lg border bg-muted/50 p-4">
              <div
                class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded bg-muted"
              >
                <ImageIcon
                  v-if="selectedFile.type.startsWith('image/')"
                  class="h-6 w-6 text-muted-foreground"
                />
                <File v-else class="h-6 w-6 text-muted-foreground" />
              </div>
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm font-medium">{{ selectedFile.name }}</p>
                <p
                  class="text-xs"
                  :class="isFileTooLarge ? 'text-destructive' : 'text-muted-foreground'"
                >
                  {{ formatFileSize(selectedFile.size) }}
                  <span v-if="isFileTooLarge"> - File too large</span>
                </p>
              </div>
              <Button type="button" variant="ghost" size="sm" @click.stop="clearSelectedFile">
                <X class="h-4 w-4" />
              </Button>
            </div>
          </template>
        </div>

        <!-- Upload Button -->
        <div class="flex justify-end">
          <Button @click="uploadFile" :disabled="!selectedFile || isFileTooLarge || uploading">
            <Loader2 v-if="uploading" class="mr-2 h-4 w-4 animate-spin" />
            <Upload v-else class="mr-2 h-4 w-4" />
            {{ uploading ? 'Uploading...' : 'Upload File' }}
          </Button>
        </div>
      </div>

      <DialogFooter class="gap-2 gap-0">
        <Button variant="outline" @click="close">Cancel</Button>
        <Button @click="confirm" :disabled="localSelection.length === 0">
          Select{{ localSelection.length > 0 ? ` (${localSelection.length})` : '' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
