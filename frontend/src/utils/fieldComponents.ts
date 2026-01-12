import type { Component } from 'vue'
import TextField from '@/components/fields/TextField.vue'
import NumberField from '@/components/fields/NumberField.vue'
import BooleanField from '@/components/fields/BooleanField.vue'
import SelectField from '@/components/fields/SelectField.vue'

export const fieldComponents: Record<string, Component> = {
  text: TextField,
  number: NumberField,
  boolean: BooleanField,
  select: SelectField,
}
