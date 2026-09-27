<template>
  <section data-site-capability-detail :data-site-selected-metric="capability.key" aria-labelledby="site-capability-detail-title">
    <h3 id="site-capability-detail-title" class="site-overview-title">{{ capability.label }}</h3>
    <dl class="site-intelligence-facts mt-3 grid min-w-0 grid-cols-2 gap-4 lg:grid-cols-4">
      <div class="min-w-0"><dt class="site-detail-label">{{ t('siteIntelligence.siteFact') }}</dt><dd data-site-selected-fact :data-site-capability-state="capability.state" class="mt-1 break-words">{{ capability.stateLabel }}</dd></div>
      <div class="min-w-0"><dt class="site-detail-label">{{ t('siteIntelligence.factDate') }}</dt><dd class="mt-1 break-words"><time :datetime="capability.dateTime">{{ capability.date }}</time></dd></div>
      <div class="min-w-0"><dt class="site-detail-label">{{ t('siteIntelligence.sameDayAdoption') }}</dt><dd data-site-selected-adoption class="mt-1">{{ capability.adoption }}</dd></div>
      <div class="min-w-0"><dt class="site-detail-label">{{ t('siteIntelligence.coverage') }}</dt><dd data-site-selected-coverage class="mt-1">{{ capability.coverage }}</dd></div>
    </dl>
    <p class="site-detail-note mt-2">{{ t('siteIntelligence.coverageHint') }}</p>
    <slot />
  </section>
</template>
<script setup lang="ts">
import type { SiteInsightsPresentation } from '~/utils/siteInsightsPresentation'
defineProps<{ capability: SiteInsightsPresentation['selected'] }>()
const { t } = useI18n()
</script>
