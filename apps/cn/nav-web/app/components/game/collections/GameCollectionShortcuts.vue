<template>
  <nav :aria-label="t('game.collections.heading')" class="game-collection-shortcuts grid gap-2" :class="compact ? 'grid-cols-2' : 'grid-cols-3'">
    <NuxtLink v-for="entry in entries" :key="entry.collection.code" :to="localePath(`/games/collections/${entry.collection.code}`)" class="game-collection-shortcut flex min-w-0 items-center justify-center text-center">
      <span class="line-clamp-2">{{ entry.collection.name }}</span>
    </NuxtLink>
    <NuxtLink :to="localePath('/games/collections')" class="game-collection-shortcut flex items-center justify-center text-center">{{ t('game.collections.all') }}</NuxtLink>
  </nav>
</template>
<script setup lang="ts">
import type { GameCollectionHome } from '~/types/game'
const props = defineProps<{ collections: GameCollectionHome; compact?: boolean }>()
const { t } = useI18n()
const localePath = useLocalePath()
const entries = computed(() => [...props.collections.slots].sort((a, b) => a.slot - b.slot))
</script>
