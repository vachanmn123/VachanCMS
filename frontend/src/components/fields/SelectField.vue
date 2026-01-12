<script setup lang="ts">
import { ref, watch } from 'vue'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { AcceptableValue } from 'reka-ui'

interface Props {
  label: string
  required?: boolean
  modelValue: string
  options: string[]
}

const props = withDefaults(defineProps<Props>(), {
  required: false,
  options: () => [],
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const localValue = ref(props.modelValue || '')

watch(
  () => props.modelValue,
  (newVal) => {
    localValue.value = newVal || ''
  },
)

function emitUpdate(value: AcceptableValue) {
  const stringValue = String(value || '')
  localValue.value = stringValue
  emit('update:modelValue', stringValue)
}
</script>

<template>
  <div class="space-y-2">
    <Label :for="label">
      {{ label }}
      <span v-if="required" class="text-destructive">*</span>
    </Label>
    <Select :model-value="localValue" @update:model-value="emitUpdate">
      <SelectTrigger :id="label">
        <SelectValue placeholder="Select an option..." />
      </SelectTrigger>
      <SelectContent>
        <SelectItem v-for="option in options" :key="option" :value="option">
          {{ option }}
        </SelectItem>
      </SelectContent>
    </Select>
  </div>
</template>
