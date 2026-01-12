<script setup lang="ts">
import { ref, watch } from 'vue'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

interface Props {
  label: string
  required?: boolean
  modelValue: string
}

const props = withDefaults(defineProps<Props>(), {
  required: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const localValue = ref(props.modelValue ?? '')

watch(
  () => props.modelValue,
  (newVal) => {
    localValue.value = newVal ?? ''
  },
)

function emitUpdate() {
  emit('update:modelValue', localValue.value)
}
</script>

<template>
  <div class="space-y-2">
    <Label :for="label">
      {{ label }}
      <span v-if="required" class="text-destructive">*</span>
    </Label>
    <Input :id="label" v-model="localValue" :required="required" @input="emitUpdate" />
  </div>
</template>
