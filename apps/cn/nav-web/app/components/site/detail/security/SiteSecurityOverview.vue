<template>
  <div data-site-security-overview class="space-y-6">
    <dl class="grid min-w-0 gap-x-8 md:grid-cols-2">
      <div v-for="item in presentation.sections" :key="item.key" :data-site-security-summary="item.key" class="site-security-row min-w-0">
        <dt class="site-overview-title">{{ item.title }}</dt>
        <dd class="mt-2 whitespace-pre-wrap break-words">{{ item.value }}</dd>
        <dd v-if="item.detail" class="site-detail-note mt-1 break-words">{{ item.detail }}</dd>
      </div>
    </dl>
    <section v-if="presentation.attention.length" data-site-security-attention class="site-security-attention">
      <h3 class="site-overview-title">{{ t('siteSecurity.attentionTitle') }}</h3>
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
defineProps<{ presentation: SiteSecurityPresentation['overview'] }>()
const { t } = useI18n()
</script>
