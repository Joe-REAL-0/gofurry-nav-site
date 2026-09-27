<template>
  <div data-site-security-tls class="space-y-6">
    <section class="site-detail-surface site-detail-surface--secondary">
      <div class="flex flex-wrap items-baseline justify-between gap-2"><h3 class="site-overview-title">{{ t('siteSecurity.transport') }}</h3><p data-site-tls-state :data-state="transport.state">{{ transport.label }}</p></div>
      <p v-if="transport.stale" class="site-detail-note mt-1">{{ t('siteSecurity.stale') }}</p>
      <p v-if="transport.truncated" class="site-detail-note mt-1">{{ t('siteSecurity.truncated') }}</p>
      <SiteObservationFacts :items="transport.facts" class="mt-2" />
      <SiteObservationFacts v-if="transport.errors.length" :items="transport.errors" class="mt-2" />
    </section>
    <section data-site-certificate class="site-detail-surface">
      <h3 class="site-overview-title">{{ t('siteSecurity.certificate') }}</h3>
      <div class="mt-3 grid min-w-0 gap-6 md:grid-cols-2">
        <div>
          <h4 class="site-detail-label">{{ t('siteSecurity.verification') }}</h4>
          <p data-site-certificate-verification :data-state="certificate.verification.state" :data-tone="certificate.verification.tone" class="mt-2">{{ certificate.verification.label }}</p>
          <SiteObservationFacts v-if="certificate.errors.length" :items="certificate.errors" class="mt-2" />
        </div>
        <div>
          <h4 class="site-detail-label">{{ t('siteSecurity.validity') }}</h4>
          <p data-site-certificate-expiry :data-expiry="certificate.expiry.state" :data-tone="certificate.expiry.tone" class="mt-2">{{ certificate.expiry.value }} <span class="site-detail-note">· {{ certificate.expiry.label }}</span></p>
          <SiteObservationFacts :items="certificate.validity" class="mt-2" />
        </div>
      </div>
      <p class="site-detail-note mt-3">{{ t('siteSecurity.validityHint') }}</p>
    </section>
    <section data-site-certificate-identity class="site-detail-surface site-detail-surface--secondary">
      <h3 class="site-overview-title mb-3">{{ t('siteSecurity.identity') }}</h3>
      <div class="grid min-w-0 gap-5 md:grid-cols-2">
      <div class="min-w-0">
      <h3 class="site-overview-title">{{ t('siteSecurity.subject') }}</h3>
      <SiteObservationFacts :items="certificate.subject" class="mt-2" />
      <details data-site-certificate-san class="site-security-disclosure mt-3">
        <summary>{{ t('siteSecurity.san') }}</summary>
        <ul v-if="certificate.san.length" class="mt-2 space-y-1"><li v-for="(name, index) in certificate.san" :key="index" class="break-all">{{ name }}</li></ul>
        <p v-else class="site-detail-note mt-2">{{ t('siteSecurity.noEvidence') }}</p>
      </details>
      </div>
      <div class="min-w-0">
      <h3 class="site-overview-title">{{ t('siteSecurity.issuer') }}</h3>
      <SiteObservationFacts :items="certificate.issuer" class="mt-2" />
      <details data-site-certificate-chain class="site-security-disclosure mt-3">
        <summary>{{ t('siteSecurity.chain') }}</summary>
        <ol v-if="certificate.chain.length" class="mt-2 space-y-1"><li v-for="(name, index) in certificate.chain" :key="index" class="break-all">{{ name }}</li></ol>
        <p v-else class="site-detail-note mt-2">{{ t('siteSecurity.noEvidence') }}</p>
      </details>
      </div>
      </div>
    </section>
    <details data-site-certificate-crypto class="site-security-disclosure site-security-crypto">
      <summary>{{ t('siteSecurity.crypto') }}</summary>
      <SiteObservationFacts :items="certificate.crypto" break-all class="mt-2" />
      <p class="site-detail-note mt-2">{{ t('siteSecurity.cryptoHint') }}</p>
    </details>
  </div>
</template>
<script setup lang="ts">
import type { SiteSecurityPresentation } from '~/utils/siteSecurityPresentation'
import SiteObservationFacts from '../observation/SiteObservationFacts.vue'
defineProps<{ transport: SiteSecurityPresentation['transport']; certificate: SiteSecurityPresentation['certificate'] }>()
const { t } = useI18n()
</script>
