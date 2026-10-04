<template>
  <section v-if="item" ref="root" class="game-home-showcase mb-6 overflow-hidden lg:mb-7" :aria-label="t('gameShowcase.label')">
    <div class="flex flex-col lg:h-80 lg:flex-row">
      <div class="game-home-showcase__artwork-frame relative aspect-video min-w-0 overflow-hidden sm:aspect-[2/1] lg:aspect-auto lg:w-[64%] lg:shrink-0">
        <GameShowcaseArtwork :key="snapshot.snapshot_id" :items="snapshot.items" :active-key="item.key" @display="artworkState = $event" />
        <component
          :is="artworkLink ? NuxtLink : 'div'" v-bind="artworkLink" class="game-home-showcase__artwork absolute inset-0 z-20 block"
          :aria-label="artworkLink ? displayedArtwork?.title : undefined" @click="artworkLink && click('artwork', displayedArtwork)"
        />
      </div>

      <div class="game-home-showcase__content flex min-h-0 min-w-0 flex-1 flex-col lg:w-[36%]">
        <div :key="item.key" class="game-home-showcase__copy min-h-0 min-w-0 flex-1 overflow-hidden" :class="{ 'game-home-showcase__copy--switch': interacted }">
          <p class="game-home-showcase__context">{{ t(showcaseContextKey(item)) }}</p>
          <p v-if="releaseDate" class="game-home-showcase__meta truncate">
            {{ t(item.release?.availability === 'upcoming' ? 'gameShowcase.release.expected' : 'gameShowcase.release.released', { date: releaseDate }) }}
          </p>
          <h2 class="game-home-showcase__title mt-2 line-clamp-2">
            <component :is="primary ? NuxtLink : 'span'" v-bind="primary" @click="primary && click('title', item)">{{ item.title }}</component>
          </h2>
          <p class="game-home-showcase__summary mt-2 line-clamp-2 sm:line-clamp-3">{{ item.summary }}</p>
          <ul v-if="item.tags.length" class="game-home-showcase__tags mt-2 flex min-w-0 gap-1.5">
            <li v-for="(tag, tagIndex) in item.tags.slice(0, 3)" :key="tagIndex" class="truncate" :class="tagIndex === 2 ? 'hidden sm:block' : ''">{{ tag }}</li>
          </ul>
          <p v-if="item.reason === 'editorial' && !item.sponsored && item.editorial_note" class="game-home-showcase__note mt-2 hidden sm:line-clamp-2">{{ item.editorial_note }}</p>
        </div>

        <div class="mt-3 flex shrink-0 flex-col items-stretch justify-between gap-2 sm:flex-row sm:items-end">
          <div class="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
            <NuxtLink v-if="primary" v-bind="primary" class="gf-button gf-button--primary gf-button--stationary" @click="click('primary', item)">{{ t(`gameShowcase.action.${item.primary_action.type}`) }}</NuxtLink>
            <NuxtLink v-if="secondary" v-bind="secondary" class="game-home-showcase__secondary" @click="click('secondary', item)">
              {{ t(`gameShowcase.secondary.${item.secondary_action!.type}`) }} <span aria-hidden="true">↗</span>
            </NuxtLink>
          </div>
          <div v-if="snapshot.items.length > 1" class="game-home-showcase__controls flex shrink-0 items-center justify-end">
            <span class="game-home-showcase__count whitespace-nowrap">{{ count }}</span>
            <button type="button" :aria-label="t('gameShowcase.control.previous')" :disabled="index === 0" @click="move(-1)"><span aria-hidden="true">‹</span></button>
            <button type="button" :aria-label="t('gameShowcase.control.next')" :disabled="index === snapshot.items.length - 1" @click="move(1)"><span aria-hidden="true">›</span></button>
          </div>
        </div>
      </div>
    </div>
    <p class="sr-only" aria-live="polite" aria-atomic="true">{{ announcement }}</p>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { NuxtLink } from '#components'
import GameShowcaseArtwork from './GameShowcaseArtwork.vue'
import type { GameShowcaseAction, GameShowcaseSecondaryAction, GameShowcaseSnapshot } from '~/types/game'
import { showcaseContextKey, showcaseDestination, showcaseReleaseDate } from '~/utils/gameShowcasePresentation'
import { useGameShowcaseTracking } from '~/composables/useGameShowcaseTracking'

const props = defineProps<{ snapshot: GameShowcaseSnapshot }>()
const { t, locale } = useI18n()
const localePath = useLocalePath()
const root = ref<HTMLElement | null>(null)
const index = ref(0)
const announcement = ref('')
const interacted = ref(false)
const artworkState = ref({ key: props.snapshot.items[0]?.key ?? '', ready: false })
const item = computed(() => props.snapshot.items[index.value])
const count = computed(() => `${String(index.value + 1).padStart(2, '0')} / ${String(props.snapshot.items.length).padStart(2, '0')}`)
const releaseDate = computed(() => showcaseReleaseDate(item.value?.release, locale.value))
const displayedArtwork = computed(() => props.snapshot.items.find(entry => entry.key === artworkState.value.key))
const artworkLink = computed(() => link(displayedArtwork.value?.primary_action))
function link(action: GameShowcaseAction | GameShowcaseSecondaryAction | undefined) {
  const destination = showcaseDestination(action)
  if (!destination) return undefined
  return { to: destination.external ? destination.path : localePath(destination.path), external: destination.external,
    target: destination.external ? '_blank' : undefined, rel: destination.external ? 'noopener noreferrer' : undefined, prefetch: false }
}
const primary = computed(() => link(item.value?.primary_action))
const secondary = computed(() => link(item.value?.secondary_action))
// During a slow media handoff only the displayed artwork keeps its own link;
// text/actions follow the new index immediately. Never attribute the old image
// to the new campaign or count a pending/hidden frame as an impression.
const impressionItem = computed(() => artworkState.value.ready && artworkState.value.key === item.value?.key ? item.value : undefined)
const { click } = useGameShowcaseTracking(root, () => props.snapshot.snapshot_id, impressionItem)

watch(() => props.snapshot.snapshot_id, () => {
  index.value = 0; announcement.value = ''; interacted.value = false
  artworkState.value = { key: props.snapshot.items[0]?.key ?? '', ready: false }
}, { flush: 'sync' })
function move(delta: number) {
  const next = Math.max(0, Math.min(props.snapshot.items.length - 1, index.value + delta))
  if (next === index.value) return
  interacted.value = true
  index.value = next
  announcement.value = t('gameShowcase.live.slide', { current: index.value + 1, total: props.snapshot.items.length, title: item.value?.title })
}
</script>
