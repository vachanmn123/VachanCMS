<script setup lang="ts">
import { ref, watch } from 'vue'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'

interface Props {
  label: string
  required?: boolean
  modelValue: boolean
}

const props = withDefaults(defineProps<Props>(), {
  required: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const localValue = ref(props.modelValue ?? false)

watch(
  () => props.modelValue,
  (newVal) => {
    localValue.value = newVal ?? false
  },
)

function emitUpdate(value: boolean) {
  localValue.value = value
  emit('update:modelValue', value)
}
</script>

<template>
  <div class="flex items-center space-x-3 rounded-md border p-3">
    <Checkbox :id="label" :checked="localValue" :required="required" @update:checked="emitUpdate" />
    <Label :for="label" class="cursor-pointer font-normal">
      {{ label }}
      <span v-if="required" class="text-destructive">*</span>
    </Label>
  </div>
</template>
