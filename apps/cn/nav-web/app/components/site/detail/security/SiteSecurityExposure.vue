<template>
  <div data-site-security-exposure class="space-y-8">
    <section data-site-port-check :data-state="ports.state">
      <div class="flex flex-wrap items-baseline justify-between gap-2"><h3 class="site-overview-title">{{ t('siteSecurity.portCheck') }}</h3><p>{{ ports.label }}</p></div>
      <p class="site-detail-note mt-1">{{ t('siteSecurity.portHint') }}</p>
      <p v-if="ports.stale" class="site-detail-note mt-1">{{ t('siteSecurity.stale') }}</p>
      <p v-if="ports.truncated" data-site-port-truncated class="site-detail-note mt-1">{{ t('siteSecurity.truncated') }}</p>
      <SiteObservationFacts :items="ports.summary" class="mt-3" />
      <SiteObservationFacts v-if="ports.errors.length" :items="ports.errors" class="mt-2" />
      <div v-for="item in ports.results" :key="item.key" data-site-port-result :data-state="item.status" :data-tone="item.tone" class="site-security-row">
        <dl class="grid min-w-0 gap-3 md:grid-cols-4"><div v-for="field in item.facts" :key="field.key" class="min-w-0"><dt class="site-detail-note">{{ field.label }}</dt><dd class="break-all">{{ field.value }}</dd></div></dl>
        <SiteObservationFacts v-if="item.errors.length" :items="item.errors" class="mt-2" />
      </div>
      <details class="site-security-disclosure mt-3" data-site-port-metadata><summary>{{ t('siteSecurity.probeMetadata') }}</summary><SiteObservationFacts :items="ports.metadata" class="mt-2" /><SiteObservationFacts :items="ports.facts" /></details>
    </section>
    <section data-site-waf-canary :data-state="waf.state">
      <h3 class="site-overview-title">{{ t('siteSecurity.wafCanary') }}</h3>
      <p class="mt-2">{{ waf.label }}</p><p class="site-detail-note mt-1">{{ t('siteSecurity.wafHint') }}</p>
      <p v-if="waf.stale" class="site-detail-note mt-1">{{ t('siteSecurity.stale') }}</p>
      <p v-if="waf.truncated" data-site-waf-truncated class="site-security-attention mt-3">{{ t('siteSecurity.truncated') }}</p>
      <SiteObservationFacts :items="waf.summary" class="mt-3" />
      <SiteObservationFacts v-if="waf.errors.length" :items="waf.errors" class="mt-2" />
      <details class="site-security-disclosure mt-3" data-site-waf-cases>
        <summary>{{ t('siteSecurity.cases') }}</summary>
        <div v-for="item in waf.cases" :key="item.key" data-site-waf-case class="site-security-row">
          <dl class="grid min-w-0 gap-3 md:grid-cols-3"><div v-for="field in item.facts" :key="field.key" class="min-w-0"><dt class="site-detail-note">{{ field.label }}</dt><dd class="break-all">{{ field.value }}</dd></div></dl>
          <SiteObservationFacts v-if="item.errors.length" :items="item.errors" class="mt-2" />
        </div>
        <p v-if="!waf.cases.length" class="site-detail-note mt-2">{{ t('siteSecurity.noEvidence') }}</p>
        <SiteObservationFacts :items="waf.metadata" class="mt-2" /><SiteObservationFacts :items="waf.facts" />
      </details>
    </section>
  </div>
</template>
<script setup lang="ts">
import type { SiteSecurityPresentation } from '~/utils/siteSecurityPresentation'
import SiteObservationFacts from '../observation/SiteObservationFacts.vue'
defineProps<{ ports: SiteSecurityPresentation['portCheck']; waf: SiteSecurityPresentation['wafCanary'] }>()
const { t } = useI18n()
</script>
