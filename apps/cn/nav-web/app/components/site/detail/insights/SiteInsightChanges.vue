<template>
  <section data-site-insight-changes :data-site-changes-state="state" class="site-detail-surface" aria-labelledby="site-insight-changes-title">
    <h3 id="site-insight-changes-title" class="site-overview-title">{{ t('siteIntelligence.recentChanges') }}</h3>
    <p v-if="state !== 'ready'" class="site-detail-note mt-3">{{ t(state === 'unavailable' ? 'siteIntelligence.unavailable' : 'siteIntelligence.noChanges') }}</p>
    <ol v-else class="mt-2">
      <li v-for="item in items" :key="item.key" data-site-insight-change :data-site-change-category="item.category" class="site-intelligence-change flex min-w-0 flex-col gap-2 sm:flex-row sm:items-baseline sm:justify-between sm:gap-5">
        <div class="min-w-0"><SiteChangeCategory :category="item.category" :label="item.categoryLabel" /><p class="break-words">{{ item.label }}</p></div>
        <time :datetime="item.dateTime" :data-precise="item.precise" class="site-detail-note shrink-0">{{ item.when }}</time>
      </li>
    </ol>
  </section>
</template>
<script setup lang="ts">
import type { SiteInsightsPresentation } from '~/utils/siteInsightsPresentation'
import SiteChangeCategory from '../SiteChangeCategory.vue'
defineProps<{ items: SiteInsightsPresentation['changes']; state: SiteInsightsPresentation['changesState'] }>()
const { t } = useI18n()
</script>
