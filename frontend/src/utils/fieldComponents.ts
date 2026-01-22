import type { Component } from 'vue'
import TextField from '@/components/fields/TextField.vue'
import TextareaField from '@/components/fields/TextareaField.vue'
import NumberField from '@/components/fields/NumberField.vue'
import BooleanField from '@/components/fields/BooleanField.vue'
import SelectField from '@/components/fields/SelectField.vue'
import MediaField from '@/components/fields/MediaField.vue'

export const fieldComponents: Record<string, Component> = {
  text: TextField,
  textarea: TextareaField,
  number: NumberField,
  boolean: BooleanField,
  select: SelectField,
  media: MediaField,
}
