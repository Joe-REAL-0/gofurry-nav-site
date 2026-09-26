<template>
  <div class="space-y-6">
    <section data-site-security-headers>
      <h3 class="site-overview-title">{{ t('siteSecurity.headers') }}</h3>
      <p class="site-detail-note mt-1">{{ t('siteSecurity.headerHint') }}</p>
      <p class="site-detail-note mt-1">{{ headers.observed }}<span v-if="headers.stale"> · {{ t('siteSecurity.stale') }}</span></p>
      <p v-if="headers.truncated" class="site-detail-note mt-1">{{ t('siteSecurity.truncated') }}</p>
      <div v-for="item in headers.rows" :key="item.key" :data-site-security-header="item.key" :data-state="item.state" class="site-security-row grid min-w-0 gap-2 md:grid-cols-[minmax(0,1fr)_7rem_minmax(0,2fr)] md:gap-4">
        <h4 class="site-detail-label break-words">{{ item.name }}</h4><p>{{ item.label }}</p>
        <div class="min-w-0"><p class="whitespace-pre-wrap break-all">{{ item.value }}</p>
          <details v-if="item.details.length" class="site-security-disclosure mt-2"><summary>{{ t('siteSecurity.collectorSummary') }}</summary><SiteObservationFacts :items="item.details" class="mt-2" /></details>
        </div>
      </div>
      <NuxtLink data-site-security-raw-headers :to="rawHeadersTo" class="site-security-link mt-3 inline-block">{{ t('siteSecurity.rawHeaders') }}</NuxtLink>
    </section>
    <section data-site-security-txt :data-state="securityTxt.state">
      <div class="flex flex-wrap items-baseline justify-between gap-2"><h3 class="site-overview-title">security.txt</h3><p>{{ securityTxt.label }}</p></div>
      <p v-if="securityTxt.stale" class="site-detail-note mt-1">{{ t('siteSecurity.stale') }}</p>
      <p v-if="securityTxt.truncated" class="site-detail-note mt-1">{{ t('siteSecurity.truncated') }}</p>
      <ul v-if="securityTxt.validation.length" data-site-security-txt-validation class="site-security-attention mt-3 space-y-1"><li v-for="error in securityTxt.validation" :key="error" class="break-all">{{ error }}</li></ul>
      <SiteObservationFacts :items="securityTxt.facts" class="mt-3" />
      <SiteObservationFacts v-if="securityTxt.errors.length" :items="securityTxt.errors" />
      <SiteObservationFacts :items="securityTxt.metaFacts" class="mt-2" />
    </section>
  </div>
</template>
<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'
import type { SiteSecurityPresentation } from '~/utils/siteSecurityPresentation'
import SiteObservationFacts from '../observation/SiteObservationFacts.vue'
defineProps<{ headers: SiteSecurityPresentation['headers']; securityTxt: SiteSecurityPresentation['securityTxt']; rawHeadersTo: RouteLocationRaw }>()
const { t } = useI18n()
</script>
