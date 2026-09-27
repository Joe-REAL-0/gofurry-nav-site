<template>
  <section data-site-recent-changes :data-site-changes-state="state" class="site-overview-change-section min-w-0" aria-labelledby="site-recent-changes-title">
    <div class="site-overview-section-head flex items-center gap-3">
      <h3 id="site-recent-changes-title" class="site-overview-heading shrink-0">{{ t('siteOverview.changes') }}</h3>
      <span class="site-overview-section-line min-w-0 flex-1" aria-hidden="true" />
    </div>
    <p v-if="state !== 'ready'" class="site-overview-empty mt-3">{{ t(state === 'unavailable' ? 'siteOverview.unavailable' : 'siteOverview.changesEmpty') }}</p>
    <ol v-else data-site-overview-change-grid class="site-overview-change-grid mt-3 grid min-w-0 gap-3 sm:grid-cols-2 xl:grid-cols-4">
      <li v-for="item in items" :key="item.key" data-site-change data-site-overview-change-card :data-site-change-type="item.type" :data-site-change-category="item.category" class="site-overview-change-card flex min-w-0 flex-col items-start gap-2">
        <SiteChangeCategory :category="item.category" :label="item.categoryLabel" class="site-overview-change-category" />
        <p class="site-overview-change-event break-words">{{ item.label }}</p>
        <time :datetime="item.dateTime" class="site-detail-note mt-auto block">{{ item.when }}{{ item.precise ? ' ' + t('siteOverview.utc') : '' }}</time>
      </li>
    </ol>
  </section>
</template>

<script setup lang="ts">
import type { SiteOverviewPresentation } from '~/utils/siteOverviewPresentation'
import SiteChangeCategory from './SiteChangeCategory.vue'
defineProps<{ items: SiteOverviewPresentation['recentChanges']; state: SiteOverviewPresentation['changesState'] }>()
const { t } = useI18n()
</script>
