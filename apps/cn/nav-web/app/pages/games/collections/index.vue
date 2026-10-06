<template>
  <section class="games-page game-collections-page mx-auto w-full max-w-[1440px] px-6 py-10">
    <header class="mb-8 max-w-3xl">
      <h1 class="game-collections-heading">{{ t('game.collections.heading') }}</h1>
      <p class="game-collection-copy mt-3">{{ t('game.collections.intro') }}</p>
    </header>
    <div v-if="snapshot" class="game-collections-grid grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-3" :aria-busy="pending">
      <GameCollectionCard v-for="collection in snapshot.items" :key="collection.code" :collection="collection" />
    </div>
    <p v-if="snapshot && !snapshot.items.length" class="game-collection-note py-8">{{ t('game.collections.empty') }}</p>
    <div v-if="error" class="game-collections-error mt-6 flex flex-wrap items-center gap-4" role="alert">
      <p>{{ t(snapshot ? 'game.collections.refreshFailed' : 'game.collections.unavailable') }}</p>
      <button class="gf-button gf-button--surface" :disabled="pending" @click="refresh">{{ t('game.collections.retry') }}</button>
    </div>
    <div v-if="snapshot?.has_more && !error" class="game-collections-load mt-8 flex flex-col items-center gap-3">
      <p v-if="loadError" role="alert">{{ t('game.collections.loadFailed') }}</p>
      <button class="gf-button gf-button--surface" :disabled="loadingMore || pending || snapshotMode !== mode" @click="loadMore">
        {{ t(loadingMore ? 'game.collections.loading' : loadError ? 'game.collections.retry' : 'game.collections.loadMore') }}
      </button>
    </div>
  </section>
</template>
<script setup lang="ts">
import { setResponseStatus } from 'h3'
import GameCollectionCard from '~/components/game/collections/GameCollectionCard.vue'
import { getGameCollections } from '~/services/game'
import { buildGameCollectionSeo } from '~/utils/seo'
import { useGameCollectionModeRefresh } from '~/composables/useGameCollectionModeRefresh'

definePageMeta({ key: route => route.path })
const { t, locale } = useI18n()
const lang = locale.value === 'en' ? 'en' : 'zh'
const initial = await useAsyncData(`game-collections:${lang}`, () => getGameCollections(lang, 'sfw'))
if (import.meta.server && !initial.data.value) setResponseStatus(useRequestEvent()!, 503)
const { snapshot, mode, snapshotMode, revision, pending, error, refresh } = useGameCollectionModeRefresh({
  initial: initial.data.value ?? null, initialError: initial.error.value,
  load: mode => getGameCollections(lang, mode),
})
const loadingMore = ref(false)
const loadError = ref(false)
watch(revision, () => { loadError.value = false; loadingMore.value = false }, { flush: 'sync' })
async function loadMore() {
  if (!snapshot.value?.has_more || loadingMore.value || pending.value || mode.value !== snapshotMode.value) return
  const generation = revision.value
  const current = snapshot.value
  loadingMore.value = true
  loadError.value = false
  try {
    const next = await getGameCollections(lang, snapshotMode.value, current.page + 1, current.page_size)
    if (generation !== revision.value) return
    const seen = new Set(current.items.map(item => item.code))
    snapshot.value = { ...next, items: [...current.items, ...next.items.filter(item => !seen.has(item.code))] }
  } catch {
    if (generation === revision.value) loadError.value = true
  } finally { if (generation === revision.value) loadingMore.value = false }
}
const seo = computed(() => buildGameCollectionSeo(locale.value, t('game.collections.intro')))
useSeoMeta({ title: () => seo.value.title, description: () => seo.value.description,
  ogTitle: () => seo.value.title, ogDescription: () => seo.value.description })
</script>
