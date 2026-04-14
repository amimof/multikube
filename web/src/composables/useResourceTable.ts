import { ref, computed, type Ref } from 'vue'

interface ResourceWithMeta {
  meta?: {
    name?: string
  }
}

export function useResourceTable<T extends ResourceWithMeta>(items: Ref<T[]>) {
  const nameFilter = ref('')

  const displayItems = computed(() => {
    const filter = nameFilter.value.trim().toLowerCase()
    if (!filter) return items.value

    return items.value.filter((item) => {
      const name = (item.meta?.name ?? '').toLowerCase()
      return name.includes(filter)
    })
  })

  return {
    nameFilter,
    displayItems,
  }
}
