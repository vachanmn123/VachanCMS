import { ref } from 'vue'
import { defineStore } from 'pinia'

interface Repo {
  owner: string
  repo: string
}

interface Config {
  site_name: string
  content_types: ContentType[]
}

export interface ContentType {
  id: string
  name: string
  slug: string
  fields: ContentTypeField[]
}

export interface ContentTypeField {
  field_name: string
  field_type: string
  is_required: boolean
  options: string[]
}

export const useRepoStore = defineStore('repo', () => {
  const selected = ref<Repo | null>(null)
  const config = ref<Config | null>(null)

  function selectRepo(owner: string, repo: string) {
    selected.value = { owner, repo }
  }

  function setConfig(configData: Config) {
    config.value = configData
  }

  return { selected, config, selectRepo, setConfig }
})