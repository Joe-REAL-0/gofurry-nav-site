<template>
  <section data-site-capability-matrix aria-labelledby="site-matrix-title">
    <h3 id="site-matrix-title" class="site-overview-title">{{ t('siteIntelligence.matrix') }}</h3>
    <div aria-hidden="true" class="site-intelligence-matrix-heading mt-3 hidden gap-4 lg:grid lg:grid-cols-[minmax(0,1.3fr)_repeat(3,minmax(0,1fr))_minmax(0,1.1fr)]">
      <span>{{ t('siteIntelligence.capability') }}</span><span>{{ t('siteIntelligence.siteState') }}</span><span>{{ t('siteIntelligence.adoption') }}</span><span>{{ t('siteIntelligence.coverage') }}</span><span>{{ t('siteIntelligence.factDate') }}</span>
    </div>
    <button v-for="row in rows" :key="row.key" type="button" :data-site-capability="row.key" :data-site-capability-state="row.state"
      :data-site-capability-selected="selected === row.key" :aria-pressed="selected === row.key" :aria-label="row.label + ': ' + row.stateLabel"
      class="site-intelligence-row grid w-full min-w-0 grid-cols-2 gap-x-4 gap-y-3 text-left lg:grid-cols-[minmax(0,1.3fr)_repeat(3,minmax(0,1fr))_minmax(0,1.1fr)]" @click="emit('select', row.key)">
      <span class="col-span-2 min-w-0 lg:col-span-1"><span class="site-intelligence-capability block break-words">{{ row.label }}</span><span class="site-detail-note block">{{ row.categoryLabel }}</span></span>
      <span class="min-w-0"><span class="site-detail-label mb-1 block lg:hidden">{{ t('siteIntelligence.siteState') }}</span><span class="site-overview-state break-words" :data-tone="row.tone">{{ row.stateLabel }}</span></span>
      <span class="min-w-0"><span class="site-detail-label mb-1 block lg:hidden">{{ t('siteIntelligence.adoption') }}</span><span data-site-capability-adoption>{{ row.adoption }}</span></span>
      <span class="min-w-0"><span class="site-detail-label mb-1 block lg:hidden">{{ t('siteIntelligence.coverage') }}</span><span data-site-capability-coverage>{{ row.coverage }}</span></span>
      <span class="min-w-0"><span class="site-detail-label mb-1 block lg:hidden">{{ t('siteIntelligence.factDate') }}</span><time data-site-capability-date :datetime="row.dateTime" class="break-words">{{ row.date }}</time></span>
    </button>
  </section>
</template>
<script setup lang="ts">
import type { SiteInsightsPresentation } from '~/utils/siteInsightsPresentation'
import type { SiteInsightMetric } from '~/utils/siteDetailRouteState'
defineProps<{ rows: SiteInsightsPresentation['rows']; selected: SiteInsightMetric }>()
const emit = defineEmits<{ select: [metric: SiteInsightMetric] }>()
const { t } = useI18n()
</script>
