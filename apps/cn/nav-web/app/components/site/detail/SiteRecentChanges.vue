<template>
  <section data-site-recent-changes :data-site-changes-state="state" class="site-detail-surface min-w-0" aria-labelledby="site-recent-changes-title">
    <h3 id="site-recent-changes-title" class="site-overview-title">{{ t('siteOverview.changes') }}</h3>
    <p v-if="state !== 'ready'" class="site-overview-empty mt-3">{{ t(state === 'unavailable' ? 'siteOverview.unavailable' : 'siteOverview.changesEmpty') }}</p>
    <ol v-else class="mt-3">
      <li v-for="item in items" :key="item.key" data-site-change :data-site-change-type="item.type" :data-site-change-category="item.category" class="site-recent-change">
        <SiteChangeCategory :category="item.category" :label="item.categoryLabel" />
        <p class="break-words">{{ item.label }}</p>
        <time :datetime="item.dateTime" class="site-detail-note mt-1 block sm:text-right">{{ item.when }}{{ item.precise ? ' ' + t('siteOverview.utc') : '' }}</time>
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
