<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import axios from 'axios'
import { File, X, Plus } from 'lucide-vue-next'

import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { AspectRatio } from '@/components/ui/aspect-ratio'
import MediaPicker from '@/components/MediaPicker.vue'

interface MediaFile {
  id: string
  file_name: string
  file_type: string
}

interface Props {
  label: string
  required?: boolean
  modelValue: string | string[]
  options?: string[]
}

const props = withDefaults(defineProps<Props>(), {
  required: false,
  options: () => [],
})

const emit = defineEmits<{
  'update:modelValue': [value: string | string[]]
}>()

const route = useRoute()
const owner = computed(() => String(route.params.owner))
const repo = computed(() => String(route.params.repo))

const showPicker = ref(false)
const isMultiple = computed(() => {
  if (!props.options || !Array.isArray(props.options)) {
    return false
  }
  return props.options.includes('multiple')
})
const selectedMedia = ref<MediaFile[]>([])

// Fetch media metadata for the current value on mount and when modelValue changes
watch(
  () => props.modelValue,
  async (newValue) => {
    await loadSelectedMedia(newValue)
  },
  { immediate: true },
)

onMounted(async () => {
  await loadSelectedMedia(props.modelValue)
})

async function loadSelectedMedia(value: string | string[]) {
  if (!value || (Array.isArray(value) && value.length === 0) || value === '') {
    selectedMedia.value = []
    return
  }

  const ids = Array.isArray(value) ? value : [value]
  if (ids.length === 0 || ids.every((id) => id === '')) {
    selectedMedia.value = []
    return
  }

  try {
    // Fetch media list to get metadata
    const response = await axios.get(`/api/${owner.value}/${repo.value}/media`)
    const allMedia: MediaFile[] = response.data.media || []

    // Filter to only selected IDs
    selectedMedia.value = allMedia.filter((m) => ids.includes(m.id))
  } catch {
    // If we can't fetch metadata, create placeholder objects
    selectedMedia.value = ids
      .filter((id) => id !== '')
      .map((id) => ({
        id,
        file_name: 'Unknown',
        file_type: 'application/octet-stream',
      }))
  }
}

function openPicker() {
  showPicker.value = true
}

function handleSelect(media: MediaFile[]) {
  selectedMedia.value = media
  if (isMultiple.value) {
    emit(
      'update:modelValue',
      media.map((m) => m.id),
    )
  } else {
    emit('update:modelValue', media[0]?.id || '')
  }
  showPicker.value = false
}

function removeMedia(id: string) {
  selectedMedia.value = selectedMedia.value.filter((m) => m.id !== id)
  if (isMultiple.value) {
    emit(
      'update:modelValue',
      selectedMedia.value.map((m) => m.id),
    )
  } else {
    emit('update:modelValue', '')
  }
}

function isImage(type: string): boolean {
  return type.startsWith('image/')
}

function getMediaUrl(id: string): string {
  return `/api/${owner.value}/${repo.value}/media/${id}`
}
</script>

<template>
  <div class="space-y-2">
    <Label>
      {{ label }}
      <span v-if="required" class="text-destructive">*</span>
    </Label>

    <!-- Selected Media Preview -->
    <div v-if="selectedMedia.length > 0" class="flex flex-wrap gap-2">
      <div
        v-for="media in selectedMedia"
        :key="media.id"
        class="group relative overflow-hidden rounded-lg border"
      >
        <div class="flex items-center gap-2 p-2">
          <!-- Thumbnail -->
          <div class="h-10 w-10 flex-shrink-0 overflow-hidden rounded">
            <AspectRatio :ratio="1">
              <img
                v-if="isImage(media.file_type)"
                :src="getMediaUrl(media.id)"
                :alt="media.file_name"
                class="h-full w-full object-cover"
              />
              <div v-else class="flex h-full items-center justify-center bg-muted">
                <File class="h-4 w-4 text-muted-foreground" />
              </div>
            </AspectRatio>
          </div>

          <!-- File name -->
          <span class="max-w-32 truncate text-sm" :title="media.file_name">
            {{ media.file_name }}
          </span>

          <!-- Remove button -->
          <Button
            type="button"
            variant="ghost"
            size="sm"
            class="ml-1 h-6 w-6 p-0"
            @click="removeMedia(media.id)"
          >
            <X class="h-3 w-3" />
          </Button>
        </div>
      </div>
    </div>

    <!-- Empty state / Add button -->
    <Button type="button" variant="outline" size="sm" @click="openPicker">
      <Plus class="mr-2 h-4 w-4" />
      {{ selectedMedia.length > 0 ? (isMultiple ? 'Add More' : 'Change') : 'Select Media' }}
    </Button>

    <!-- Media Picker Dialog -->
    <MediaPicker
      v-model:open="showPicker"
      :multiple="isMultiple"
      :selected="selectedMedia"
      @select="handleSelect"
    />
  </div>
</template>
