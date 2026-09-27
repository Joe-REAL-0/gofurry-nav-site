<template>
  <div data-site-http class="space-y-5">
    <section data-site-http-response class="site-observation-composite">
      <h3 class="site-observation-heading">{{ t('siteObservation.responseSummary') }}</h3>
      <dl class="site-observation-response-strip mt-4 grid min-w-0 grid-cols-2 items-end gap-4 sm:grid-cols-3">
        <div v-for="item in presentation.headline" :key="item.key" :data-site-evidence="item.key" class="min-w-0" :class="{ 'col-span-2 sm:col-span-1': item.key === 'status' }">
          <dt class="site-observation-caption">{{ item.label }}</dt><dd class="site-observation-response-value mt-1 break-words" :data-tone="item.tone">{{ item.value }}</dd>
        </div>
      </dl>
      <SiteObservationEvidence :items="presentation.evidence" class="site-observation-response-evidence mt-4" />
    </section>
    <section v-if="presentation.redirects.length" data-site-http-redirects class="site-observation-flow">
      <h3 class="site-observation-heading">{{ t('siteObservation.redirects') }}</h3>
      <ol class="mt-3 space-y-2">
        <li v-for="(url, index) in presentation.redirects" :key="index" class="site-observation-technical min-w-0 break-all"><PhArrowDown v-if="index" class="site-detail-icon mb-2" aria-hidden="true" />{{ url }}</li>
      </ol>
    </section>
    <section data-site-http-headers class="site-observation-evidence relative">
      <h3 class="site-observation-heading mb-4" :class="{ 'max-w-[50%]': presentation.headers.length }">{{ t('siteObservation.commonHeaders') }}</h3>
      <SiteObservationEvidence v-if="presentation.commonHeaders.length" :items="presentation.commonHeaders" />
      <p v-else class="site-observation-caption">{{ t('siteObservation.noEvidence') }}</p>
      <details v-if="presentation.headers.length" data-site-http-all-headers class="site-observation-disclosure site-observation-header-disclosure">
        <summary class="absolute right-4 top-4 max-w-[45%] text-right">{{ t('siteObservation.allHeaders') }} ({{ presentation.headers.length }})</summary>
        <SiteObservationEvidence :items="presentation.headers" class="site-observation-raw-headers mt-4" />
      </details>
    </section>
  </div>
</template>

<script setup lang="ts">
import type { SiteObservationPresentation } from '~/utils/siteObservationPresentation'
import { PhArrowDown } from '@phosphor-icons/vue'
import SiteObservationEvidence from './SiteObservationEvidence.vue'
defineProps<{ presentation: SiteObservationPresentation['http'] }>()
const { t } = useI18n()
</script>
