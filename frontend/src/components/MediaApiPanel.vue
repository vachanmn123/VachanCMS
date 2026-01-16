<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  PanelRightClose,
  PanelRightOpen,
  Copy,
  Check,
  AlertCircle,
  ExternalLink,
  Image as ImageIcon,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

interface Props {
  baseUrl: string
  isInitialized: boolean
  pagesSettingsUrl: string
}

const props = defineProps<Props>()

const STORAGE_KEY = 'media-api-panel-open'
const isOpen = ref(true)
const copiedIndex = ref<number | null>(null)

onMounted(() => {
  const stored = localStorage.getItem(STORAGE_KEY)
  if (stored !== null) {
    isOpen.value = stored === 'true'
  }
})

function togglePanel() {
  isOpen.value = !isOpen.value
  localStorage.setItem(STORAGE_KEY, String(isOpen.value))
}

const endpoints = computed(() => {
  if (!props.isInitialized || !props.baseUrl) return []

  const base = props.baseUrl.endsWith('/') ? props.baseUrl : `${props.baseUrl}/`

  return [
    {
      label: 'Access media file',
      description: 'Download or embed media files directly from your GitHub Pages site.',
      usage: 'Replace {id} with the media file ID (shown in preview dialog)',
      url: `${base}media/{id}`,
      example: `${base}media/abc123`,
    },
  ]
})

async function copyToClipboard(url: string, index: number) {
  try {
    await navigator.clipboard.writeText(url)
    copiedIndex.value = index
    setTimeout(() => {
      copiedIndex.value = null
    }, 2000)
  } catch {
    console.error('Failed to copy to clipboard')
  }
}
</script>

<template>
  <div class="relative flex shrink-0">
    <!-- Toggle Button (visible when collapsed) -->
    <button
      v-if="!isOpen"
      @click="togglePanel"
      class="flex h-full flex-col items-center justify-center gap-2 border-l bg-muted/30 px-2 transition-colors hover:bg-muted"
      title="Show Media API Info"
    >
      <PanelRightOpen class="h-4 w-4 text-muted-foreground" />
      <span class="text-xs font-medium text-muted-foreground [writing-mode:vertical-lr]"
        >MEDIA</span
      >
    </button>

    <!-- Panel Content -->
    <div
      :class="
        cn(
          'overflow-hidden border-l bg-background transition-all duration-300 ease-in-out',
          isOpen ? 'w-96 opacity-100' : 'w-0 opacity-0',
        )
      "
    >
      <div class="flex h-full w-96 flex-col">
        <!-- Panel Header -->
        <div class="flex items-center justify-between border-b px-4 py-3">
          <div class="flex items-center gap-2">
            <ImageIcon class="h-5 w-5 text-primary" />
            <div>
              <h3 class="font-semibold">Media URLs</h3>
              <p class="text-xs text-muted-foreground">Access your media files</p>
            </div>
          </div>
          <Button
            variant="ghost"
            size="icon"
            class="h-8 w-8"
            @click="togglePanel"
            title="Hide Media Info"
          >
            <PanelRightClose class="h-4 w-4" />
          </Button>
        </div>

        <!-- Panel Body -->
        <div class="flex-1 overflow-auto p-4">
          <!-- Not Initialized State -->
          <div v-if="!isInitialized">
            <Card class="border-dashed">
              <CardContent class="flex flex-col items-center py-8 text-center">
                <div class="rounded-full bg-muted p-4">
                  <AlertCircle class="h-8 w-8 text-muted-foreground" />
                </div>
                <CardTitle class="mt-4 text-base">GitHub Pages not enabled</CardTitle>
                <CardDescription class="mt-2 max-w-xs">
                  Enable GitHub Pages to access your media files via public URLs. Your uploaded
                  files will be available for download or embedding.
                </CardDescription>
                <Button
                  as="a"
                  :href="pagesSettingsUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                  variant="default"
                  size="sm"
                  class="mt-6"
                >
                  <ExternalLink class="mr-2 h-4 w-4" />
                  Configure GitHub Pages
                </Button>
              </CardContent>
            </Card>
          </div>

          <!-- Endpoints List -->
          <div v-else class="space-y-4">
            <Card v-for="(endpoint, index) in endpoints" :key="index">
              <CardHeader class="pb-2">
                <div class="flex items-center justify-between">
                  <CardTitle class="text-sm font-medium">{{ endpoint.label }}</CardTitle>
                  <Badge variant="secondary" class="font-mono text-xs">GET</Badge>
                </div>
                <CardDescription class="text-xs">
                  {{ endpoint.description }}
                </CardDescription>
              </CardHeader>
              <CardContent class="space-y-3">
                <!-- URL Pattern -->
                <div>
                  <div class="mb-1 flex items-center justify-between">
                    <span class="text-xs font-medium text-muted-foreground">URL Pattern</span>
                    <Button
                      variant="ghost"
                      size="sm"
                      class="h-6 gap-1 px-2 text-xs"
                      @click="copyToClipboard(endpoint.url, index)"
                    >
                      <Check v-if="copiedIndex === index" class="h-3 w-3 text-green-500" />
                      <Copy v-else class="h-3 w-3" />
                      {{ copiedIndex === index ? 'Copied!' : 'Copy' }}
                    </Button>
                  </div>
                  <code
                    class="block w-full break-all rounded-md bg-muted px-3 py-2 font-mono text-xs"
                  >
                    {{ endpoint.url }}
                  </code>
                </div>

                <!-- Usage Note -->
                <div class="rounded-md bg-muted/50 px-3 py-2">
                  <p class="text-xs text-muted-foreground">
                    <span class="font-medium text-foreground">Usage:</span>
                    {{ endpoint.usage }}
                  </p>
                </div>

                <!-- Example -->
                <div>
                  <span class="mb-1 block text-xs font-medium text-muted-foreground">Example</span>
                  <code
                    class="block w-full break-all rounded-md border bg-background px-3 py-2 font-mono text-xs text-muted-foreground"
                  >
                    {{ endpoint.example }}
                  </code>
                </div>
              </CardContent>
            </Card>

            <!-- Help Section -->
            <div class="rounded-lg border border-dashed p-4">
              <h4 class="text-sm font-medium">Need Help?</h4>
              <p class="mt-1 text-xs text-muted-foreground">
                Use these URLs to embed images, download files, or reference media in your
                applications. Media files are served directly from your GitHub Pages site.
              </p>
              <div class="mt-3">
                <code class="block rounded-md bg-muted px-3 py-2 font-mono text-xs">
                  &lt;img src="{{ endpoints[0]?.example }}" alt="My Image" /&gt;
                </code>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
