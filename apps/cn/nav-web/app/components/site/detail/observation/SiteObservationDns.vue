<template>
  <div data-site-dns class="space-y-6">
    <section class="site-detail-surface">
      <h3 class="site-overview-title mb-3">{{ t('siteObservation.dnsSummary') }}</h3>
      <dl class="grid grid-cols-5 gap-2"><div v-for="item in presentation.facts.slice(0, 5)" :key="item.key" :data-site-evidence="item.key"><dt class="site-detail-note">{{ item.label }}</dt><dd>{{ item.value }}</dd></div></dl>
      <SiteObservationFacts :items="presentation.facts.slice(5)" class="mt-3" />
    </section>
    <section v-if="presentation.chains.length" data-site-dns-chain class="site-detail-surface site-detail-surface--secondary">
      <h3 class="site-overview-title">{{ t('siteObservation.resolution') }}</h3>
      <ol v-for="(chain, index) in presentation.chains" :key="index" class="site-observation-chain mt-3 flex flex-wrap gap-2">
        <li v-for="(value, step) in chain" :key="step" class="inline-flex min-w-0 items-center gap-2"><PhArrowRight v-if="step" class="site-detail-icon shrink-0" aria-hidden="true" /><span class="min-w-0 break-all">{{ value }}</span></li>
      </ol>
    </section>
    <section v-for="group in presentation.groups" :key="group.type" :data-site-dns-group="group.type">
      <h3 class="site-overview-title mb-2">{{ group.type }} <span class="site-detail-note">({{ group.records.length }})</span></h3>
      <div v-for="(record, index) in group.records" :key="index" class="site-observation-record">
        <div class="flex flex-wrap justify-between gap-2"><span class="min-w-0 break-all">{{ record.value }}</span><span class="site-detail-note">TTL {{ record.ttl }}</span></div>
        <details v-if="record.details.length" class="site-observation-disclosure mt-2">
          <summary>{{ t('siteObservation.recordDetails') }}</summary>
          <SiteObservationFacts :items="record.details" class="mt-2" />
        </details>
      </div>
    </section>
    <p v-if="!presentation.groups.length" class="site-detail-note">{{ t('siteObservation.noEvidence') }}</p>
    <section v-if="presentation.signals.length" data-site-dns-risks class="site-detail-surface site-detail-surface--secondary">
      <h3 class="site-overview-title">{{ t('siteObservation.dnsObservationSignals') }}</h3>
      <ul class="mt-3 space-y-2"><li v-for="signal in presentation.signals" :key="signal.code" :data-site-dns-signal="signal.code" class="flex min-w-0 items-start gap-2">
        <PhWarning v-if="signal.tone === 'warning'" class="site-detail-icon mt-1 shrink-0" data-tone="warning" aria-hidden="true" />
        <PhInfo v-else class="site-detail-icon mt-1 shrink-0" aria-hidden="true" />
        <div class="min-w-0"><p :data-tone="signal.tone" class="break-words">{{ signal.label }}</p><p class="site-detail-note break-all">{{ signal.kind }} · {{ signal.code }}</p></div>
      </li></ul>
    </section>
  </div>
</template>

<script setup lang="ts">
import type { SiteObservationPresentation } from '~/utils/siteObservationPresentation'
import { PhArrowRight, PhInfo, PhWarning } from '@phosphor-icons/vue'
import SiteObservationFacts from './SiteObservationFacts.vue'
defineProps<{ presentation: SiteObservationPresentation['dns'] }>()
const { t } = useI18n()
</script>
