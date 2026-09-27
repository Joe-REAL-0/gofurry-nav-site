<template>
  <header data-site-hero class="site-detail-hero grid min-w-0 grid-cols-[auto_minmax(0,1fr)] items-start gap-x-3 gap-y-2 sm:gap-x-5">
    <ManagedAssetImage :object-key="site?.icon || undefined" :alt="name" class="site-detail-hero__icon row-span-4 object-contain" />
    <div class="flex min-w-0 flex-wrap items-start justify-between gap-2">
      <h1 class="site-detail-hero__name min-w-0 break-words">{{ name }}</h1>
      <a v-if="visitUrl" :href="visitUrl" target="_blank" rel="noopener noreferrer" class="gf-button gf-button--primary gf-button--stationary shrink-0">
        {{ t('siteDetail.visit') }} <PhArrowSquareOut class="site-detail-icon" aria-hidden="true" />
      </a>
    </div>
    <p class="site-detail-hero__domain col-start-2 min-w-0 break-all">{{ domain }}</p>
    <div class="site-detail-hero__meta col-start-2 flex min-w-0 flex-wrap items-center justify-between gap-2">
      <div class="flex flex-wrap items-center gap-2">
        <span v-if="site?.country" class="site-detail-pill" :aria-label="t('siteDetail.country')">{{ site.country }}</span>
        <span class="site-detail-pill" :data-tone="site?.nsfw === '1' ? 'bad' : 'good'">{{ site?.nsfw === '1' ? 'NSFW' : 'SFW' }}</span>
        <span v-if="site?.welfare === '1'" class="site-detail-pill" data-tone="warning">{{ t('siteDetail.welfare') }}</span>
      </div>
      <span data-site-views class="inline-flex items-center gap-1" :aria-label="t('siteDetail.views') + ' ' + formattedViews"><PhEye class="site-detail-icon" aria-hidden="true" />{{ formattedViews }}</span>
    </div>
    <p v-if="site?.info" class="site-detail-hero__description col-start-2 min-w-0 line-clamp-3 break-words">{{ site.info }}</p>
  </header>
</template>

<script setup lang="ts">
import ManagedAssetImage from '@/components/common/ManagedAssetImage.vue'
import { computed } from 'vue'
import { PhArrowSquareOut, PhEye } from '@phosphor-icons/vue'
import type { SiteInfo } from '~/types/nav'
const props = defineProps<{ site: SiteInfo | null; name: string; domain: string; viewCount: number; visitUrl: string }>()
const { t, locale } = useI18n()
const formattedViews = computed(() => new Intl.NumberFormat(locale.value).format(props.viewCount))
</script>
