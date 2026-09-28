<script setup lang="ts">
import { PhArrowLeft, PhArrowRight } from '@phosphor-icons/vue'
import UpdateMetadata from '~/components/updates/UpdateMetadata.vue'
import UpdateMarkdown from '~/components/updates/UpdateMarkdown.vue'
import { getNavUpdateDetail } from '~/services/nav'
import { authoritativePageStatus } from '~/utils/authoritativePageError'
import { updateCopy, updateDetailSeo } from '~/utils/updatePresentation'

definePageMeta({ key: route => route.path })
const route = useRoute()
const { locale } = useI18n()
const localePath = useLocalePath()
const config = useRuntimeConfig()
const lang = computed(() => locale.value === 'en' ? 'en' : 'zh')
const copy = computed(() => updateCopy[lang.value])
const id = String(route.params.id)
if (!/^[1-9]\d*$/.test(id)) throw createError({ statusCode: 404, statusMessage: 'Release note not found' })
const { data, error } = await useAsyncData(`update-detail:${id}:${lang.value}`, () => getNavUpdateDetail(id, lang.value))
if (error.value || !data.value?.item) {
  const statusCode = authoritativePageStatus(error.value, 'update')
  throw createError({ statusCode, statusMessage: statusCode === 404 ? 'Release note not found' : 'Release notes temporarily unavailable', cause: error.value })
}
const release = computed(() => data.value!.item)
const seo = computed(() => updateDetailSeo(release.value, lang.value))
useSeoMeta({
  title: () => seo.value.title, description: () => seo.value.description,
  ogTitle: () => seo.value.title, ogDescription: () => seo.value.description,
  twitterTitle: () => seo.value.title, twitterDescription: () => seo.value.description,
  ogType: 'article', ogUrl: () => String(config.public.siteUrl).replace(/\/$/, '') + localePath(`/updates/${id}`),
  articlePublishedTime: () => release.value.published_at,
})
</script>

<template>
  <article class="updates-page" data-update-detail>
    <div class="updates-article mx-auto w-full max-w-[1000px] px-5 py-10 md:px-10 md:py-14">
      <NuxtLink :to="localePath('/updates')" class="updates-action inline-flex items-center gap-2"><PhArrowLeft aria-hidden="true" />{{ copy.all }}</NuxtLink>
      <header class="updates-article-header mt-10 md:mt-14">
        <div class="flex flex-wrap items-center gap-4">
          <p class="updates-eyebrow">{{ copy.article }}</p>
          <span v-if="release.version" class="updates-version" data-update-version>{{ release.version }}</span>
        </div>
        <h1 class="updates-title mt-5">{{ release.title }}</h1>
        <p v-if="release.summary" class="updates-summary mt-5" data-update-summary>{{ release.summary }}</p>
        <UpdateMetadata class="mt-6" :published-at="release.published_at" :commit-sha="release.commit_sha" />
      </header>
      <div class="mx-auto mt-10 w-full max-w-[800px] md:mt-14">
        <UpdateMarkdown :source="release.body" />
        <nav v-if="data?.previous || data?.next" class="updates-neighbors mt-14 grid grid-cols-1 gap-4 sm:grid-cols-2" :aria-label="copy.history">
          <NuxtLink v-if="data.previous" :to="localePath(`/updates/${data.previous.id}`)" class="updates-neighbor p-5" data-update-older>
            <span class="updates-note inline-flex items-center gap-2"><PhArrowLeft aria-hidden="true" />{{ copy.older }}</span>
            <span v-if="data.previous.version" class="updates-version mt-3 block">{{ data.previous.version }}</span>
            <span class="updates-neighbor-title mt-2 block">{{ data.previous.title }}</span>
          </NuxtLink>
          <NuxtLink v-if="data.next" :to="localePath(`/updates/${data.next.id}`)" class="updates-neighbor p-5 sm:col-start-2" data-update-newer>
            <span class="updates-note inline-flex items-center gap-2">{{ copy.newer }}<PhArrowRight aria-hidden="true" /></span>
            <span v-if="data.next.version" class="updates-version mt-3 block">{{ data.next.version }}</span>
            <span class="updates-neighbor-title mt-2 block">{{ data.next.title }}</span>
          </NuxtLink>
        </nav>
      </div>
    </div>
  </article>
</template>
