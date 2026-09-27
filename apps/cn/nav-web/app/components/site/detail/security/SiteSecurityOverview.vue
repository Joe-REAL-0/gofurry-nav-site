<template>
  <div data-site-security-overview class="space-y-6">
    <dl data-site-security-summary-plane class="site-detail-surface grid min-w-0 gap-x-6 gap-y-2 sm:grid-cols-2 lg:grid-cols-3">
      <div v-for="item in presentation.sections" :key="item.key" :data-site-security-summary="item.key" class="site-security-row min-w-0">
        <dt class="site-overview-title">{{ item.title }}</dt>
        <dd class="mt-2 whitespace-pre-wrap break-words" :data-tone="item.tone">{{ item.value }}</dd>
        <dd v-if="item.detail" class="site-detail-note mt-1 break-words">{{ item.detail }}</dd>
      </div>
    </dl>
    <p data-site-security-scope class="site-detail-note flex items-start gap-2"><PhInfo class="site-detail-icon mt-0.5 shrink-0" aria-hidden="true" />{{ t('siteSecurity.scope') }}</p>
    <section v-if="presentation.attention.length" data-site-security-attention class="site-security-attention">
      <h3 class="site-overview-title flex items-center gap-2"><PhWarning class="site-detail-icon" data-tone="warning" aria-hidden="true" />{{ t('siteSecurity.attentionTitle') }}</h3>
      <ul class="mt-2 space-y-3">
        <li v-for="item in presentation.attention" :key="item.key" :data-site-security-attention-item="item.key">
          <p>{{ item.message }}</p><p v-if="item.detail" class="site-detail-note break-all">{{ item.detail }}</p>
        </li>
      </ul>
    </section>
  </div>
</template>
<script setup lang="ts">
import type { SiteSecurityPresentation } from '~/utils/siteSecurityPresentation'
import { PhInfo, PhWarning } from '@phosphor-icons/vue'
defineProps<{ presentation: SiteSecurityPresentation['overview'] }>()
const { t } = useI18n()
</script>
