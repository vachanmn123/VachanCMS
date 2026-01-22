<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  PanelRightClose,
  PanelRightOpen,
  Copy,
  Check,
  Code,
  AlertCircle,
  ExternalLink,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

interface Props {
  ctSlug: string
  baseUrl: string
  isInitialized: boolean
  pagesSettingsUrl: string
}

const props = defineProps<Props>()

const STORAGE_KEY = 'api-info-panel-open'
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
      label: 'List entries (paginated)',
      description: 'Returns a paginated list of all entries for this content type.',
      usage: 'Replace {page} with page number starting from 1',
      url: `${base}data/${props.ctSlug}/index-{page}.json`,
      example: `${base}data/${props.ctSlug}/index-1.json`,
    },
    {
      label: 'Get single entry by slug',
      description: 'Returns a specific entry by its slug (preferred) or ID.',
      usage: 'Replace {slug} with the entry slug (e.g., my-blog-post)',
      url: `${base}data/${props.ctSlug}/{slug}.json`,
      example: `${base}data/${props.ctSlug}/my-blog-post.json`,
    },
    {
      label: 'Get single entry by ID',
      description: 'Returns a specific entry by its unique identifier.',
      usage: 'Replace {id} with the entry ID (UUID format)',
      url: `${base}data/${props.ctSlug}/{id}.json`,
      example: `${base}data/${props.ctSlug}/abc123-def456.json`,
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
      title="Show API Info"
    >
      <PanelRightOpen class="h-4 w-4 text-muted-foreground" />
      <span class="text-xs font-medium text-muted-foreground [writing-mode:vertical-lr]">API</span>
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
            <Code class="h-5 w-5 text-primary" />
            <div>
              <h3 class="font-semibold">API Endpoints</h3>
              <p class="text-xs text-muted-foreground">Access your content via GitHub Pages</p>
            </div>
          </div>
          <Button
            variant="ghost"
            size="icon"
            class="h-8 w-8"
            @click="togglePanel"
            title="Hide API Info"
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
                  Enable GitHub Pages to access your content via public API endpoints. Your data
                  will be available as static JSON files.
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
            <!-- Slug Tip -->
            <div class="rounded-lg border border-primary/20 bg-primary/5 p-3">
              <p class="text-xs text-primary">
                <span class="font-medium">💡 Tip:</span> Add a slug to your entries for
                human-readable URLs. Slugs are preferred over IDs when copying URLs.
              </p>
            </div>

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
                These endpoints return JSON data. Use them in your frontend application with fetch()
                or any HTTP client library. Prefer using slugs for cleaner, more memorable URLs.
              </p>
              <div class="mt-3">
                <code class="block rounded-md bg-muted px-3 py-2 font-mono text-xs">
                  fetch('{{ endpoints[0]?.example }}')<br />
                  &nbsp;&nbsp;.then(res => res.json())<br />
                  &nbsp;&nbsp;.then(data => console.log(data))
                </code>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
