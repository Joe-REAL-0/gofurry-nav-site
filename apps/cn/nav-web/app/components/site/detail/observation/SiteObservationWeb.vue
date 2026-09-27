<template>
  <div data-site-web class="space-y-6">
    <section data-site-web-metadata class="site-detail-surface site-detail-surface--secondary">
      <h3 class="site-overview-title mb-2">{{ t('siteObservation.sections.metadata') }}</h3>
      <SiteObservationFacts :items="presentation.metadata" />
    </section>
    <div class="grid min-w-0 items-start gap-4 lg:grid-cols-2">
    <section v-for="probe in presentation.probes" :key="probe.protocol" :data-site-web-probe="probe.protocol" class="site-detail-surface site-detail-surface--secondary min-w-0">
      <div class="mb-3 flex flex-wrap items-baseline justify-between gap-2">
        <h3 class="site-overview-title">{{ t('siteObservation.probes.' + probe.protocol) }}</h3>
        <p class="site-detail-note">{{ probe.statusLabel }}</p>
      </div>
      <SiteObservationFacts :items="probe.summaryFacts" />
      <details class="site-observation-disclosure mt-3">
        <summary>{{ t('siteObservation.probeDetails') }}</summary>
        <p class="site-detail-note mt-2">{{ probe.duration }} · {{ probe.observed }}</p>
        <section v-for="section in probe.details" :key="section.key" class="mt-3">
          <h4 class="site-detail-label mb-1">{{ section.title }}</h4><SiteObservationFacts :items="section.items" />
        </section>
      </details>
    </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { SiteObservationPresentation } from '~/utils/siteObservationPresentation'
import SiteObservationFacts from './SiteObservationFacts.vue'
defineProps<{ presentation: SiteObservationPresentation['web'] }>()
const { t } = useI18n()
</script>
