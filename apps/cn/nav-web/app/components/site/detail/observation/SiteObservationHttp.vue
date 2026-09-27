<template>
  <div data-site-http class="space-y-6">
    <section class="site-detail-surface"><h3 class="site-overview-title mb-2">{{ t('siteObservation.responseSummary') }}</h3><SiteObservationFacts :items="presentation.facts" /></section>
    <section v-if="presentation.redirects.length" data-site-http-redirects class="site-detail-surface site-detail-surface--secondary">
      <h3 class="site-overview-title">{{ t('siteObservation.redirects') }}</h3>
      <ol class="site-observation-chain mt-3 space-y-2">
        <li v-for="(url, index) in presentation.redirects" :key="index" class="min-w-0 break-all"><PhArrowDown v-if="index" class="site-detail-icon mb-2" aria-hidden="true" />{{ url }}</li>
      </ol>
    </section>
    <section data-site-http-headers class="site-detail-surface">
      <h3 class="site-overview-title mb-2">{{ t('siteObservation.commonHeaders') }}</h3>
      <SiteObservationFacts v-if="presentation.commonHeaders.length" :items="presentation.commonHeaders" />
      <p v-else class="site-detail-note">{{ t('siteObservation.noEvidence') }}</p>
      <details v-if="presentation.headers.length" data-site-http-all-headers class="site-observation-disclosure mt-4">
        <summary>{{ t('siteObservation.allHeaders') }} ({{ presentation.headers.length }})</summary>
        <SiteObservationFacts :items="presentation.headers" class="mt-3" />
      </details>
    </section>
  </div>
</template>

<script setup lang="ts">
import type { SiteObservationPresentation } from '~/utils/siteObservationPresentation'
import { PhArrowDown } from '@phosphor-icons/vue'
import SiteObservationFacts from './SiteObservationFacts.vue'
defineProps<{ presentation: SiteObservationPresentation['http'] }>()
const { t } = useI18n()
</script>
