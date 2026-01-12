<script setup lang="ts">
// This view shows when no ctSlug is selected - serves as a landing/empty state
</script>

<template>
  <div class="flex h-full flex-col items-center justify-center py-16">
    <div class="flex flex-col items-center text-center">
      <div class="rounded-full bg-muted p-4">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="h-12 w-12 text-muted-foreground"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="1.5"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z"
          />
        </svg>
      </div>
      <h2 class="mt-6 text-xl font-semibold tracking-tight">Select a Content Type</h2>
      <p class="mt-2 max-w-sm text-sm text-muted-foreground">
        Choose a content type from the sidebar to view and manage its content entries, or create a
        new content type to get started.
      </p>
    </div>
  </div>
</template>
