<template>
  <NuxtLink :to="localePath(`/games/${item.game_id}`)" class="game-collection-card flex h-full min-w-0 flex-col overflow-hidden">
    <SteamAssetImage :src="item.header_url" :alt="item.name" loading="lazy" class="game-collection-image game-collection-timeline__image" />
    <div class="flex flex-col gap-2 p-5">
      <p class="game-collection-time">{{ collectionTime(item, locale, t) }}</p>
      <h3 class="game-collection-title line-clamp-2">{{ item.name }}</h3>
      <p class="game-collection-copy line-clamp-3">{{ item.summary }}</p>
    </div>
  </NuxtLink>
</template>
<script setup lang="ts">
import type { GameCollectionTimelineItem } from '~/types/game'
import SteamAssetImage from '~/components/common/SteamAssetImage.vue'
import { collectionTime } from '~/utils/gameCollectionPresentation'
defineProps<{ item: GameCollectionTimelineItem }>()
const { t, locale } = useI18n()
const localePath = useLocalePath()
</script>
