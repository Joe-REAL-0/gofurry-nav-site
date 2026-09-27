<template>
  <div data-site-observation-overview class="space-y-6">
    <section class="site-detail-surface">
    <div class="flex flex-wrap items-baseline justify-between gap-2">
      <h3 class="site-overview-title">{{ t('siteObservation.currentObservation') }}</h3>
      <p data-site-observation-status>{{ presentation.statusLabel }}</p>
    </div>
    <dl class="site-observation-protocols mt-3 grid min-w-0 gap-4 md:grid-cols-3">
      <div v-for="item in presentation.protocols" :key="item.protocol" :data-site-observation-protocol="item.protocol" class="grid min-w-0 gap-1">
        <dt class="site-overview-title">{{ item.protocol.toUpperCase() }}</dt>
        <dd :data-tone="item.tone">{{ item.statusLabel }}</dd>
        <dd><span class="site-detail-note block">{{ t('siteObservation.fields.duration') }}</span>{{ item.duration }}</dd>
        <dd class="site-detail-note break-words"><span class="sr-only">{{ t('siteObservation.fields.observed') }}</span>{{ item.observed }}</dd>
        <dd data-site-protocol-freshness><span class="site-detail-note block">{{ t('siteObservation.freshness') }}</span>{{ item.freshness }}</dd>
      </div>
    </dl>
    </section>
    <section v-if="presentation.endpoint.length" class="site-detail-surface site-detail-surface--secondary" aria-labelledby="site-endpoint-title">
      <h3 id="site-endpoint-title" class="site-overview-title mb-2">{{ t('siteObservation.endpoint') }}</h3>
      <SiteObservationFacts :items="presentation.endpoint" class="site-detail-technical" />
    </section>
    <section v-if="presentation.risks.length" data-site-observation-risks class="site-overview-attention">
      <h3 class="site-overview-title flex items-center gap-2"><PhWarning class="site-detail-icon" data-tone="warning" aria-hidden="true" />{{ t('siteObservation.risks') }}</h3>
      <ul class="mt-2 space-y-2"><li v-for="message in presentation.risks" :key="message" class="break-words">{{ message }}</li></ul>
    </section>
  </div>
</template>

<script setup lang="ts">
import type { SiteObservationPresentation } from '~/utils/siteObservationPresentation'
import { PhWarning } from '@phosphor-icons/vue'
import SiteObservationFacts from './SiteObservationFacts.vue'
defineProps<{ presentation: SiteObservationPresentation }>()
const { t } = useI18n()
</script>
