import { z } from 'zod'

interface Field {
  field_name: string
  field_type: string
  is_required: boolean
  options: string[]
}

export function buildFieldSchema(field: Field) {
  let schema: z.ZodTypeAny
  switch (field.field_type) {
    case 'text':
      schema = z.string()
      break
    case 'number':
      schema = z.number()
      break
    case 'boolean':
      schema = z.boolean()
      break
    case 'select':
      schema = z.string().refine((val: string) => field.options.includes(val), {
        message: `Must be one of: ${field.options.join(', ')}`
      })
      break
    default:
      schema = z.any()
  }
  if (field.is_required) {
    schema = schema.refine((val) => val !== '' && val !== null && val !== undefined, {
      message: `${field.field_name} is required`
    })
  }
  return schema
}

export function buildFormSchema(fields: Field[]) {
  const shape: Record<string, z.ZodTypeAny> = {}
  fields.forEach(field => {
    shape[field.field_name] = buildFieldSchema(field)
  })
  return z.object(shape)
}