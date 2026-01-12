<script setup lang="ts">
import { ref, watch } from 'vue'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

interface Props {
  label: string
  required?: boolean
  modelValue: number
}

const props = withDefaults(defineProps<Props>(), {
  required: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: number]
}>()

const localValue = ref(props.modelValue ?? 0)

watch(
  () => props.modelValue,
  (newVal) => {
    localValue.value = newVal ?? 0
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
    <Input
      :id="label"
      v-model.number="localValue"
      type="number"
      :required="required"
      @input="emitUpdate"
    />
  </div>
</template>
