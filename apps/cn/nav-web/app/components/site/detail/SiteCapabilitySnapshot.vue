<template>
  <section data-site-capability-snapshot :data-site-capabilities-state="state" class="site-detail-surface min-w-0" aria-labelledby="site-capability-title">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <h3 id="site-capability-title" class="site-overview-title">{{ t('siteOverview.capabilities') }}</h3>
      <NuxtLink :to="insightsTo" data-site-overview-insights class="site-intelligence-link inline-flex items-center gap-1">
        {{ t('siteOverview.fullEcosystem') }}<PhArrowRight class="site-detail-icon" aria-hidden="true" />
      </NuxtLink>
    </div>
    <p v-if="state === 'unavailable'" data-site-capabilities-unavailable class="site-overview-empty mt-3">{{ t('siteOverview.unavailable') }}</p>
    <button v-if="state === 'unavailable'" data-site-insights-retry type="button" :disabled="retrying" class="gf-button gf-button--ghost mt-2" @click="emit('retry')">{{ t(retrying ? 'siteIntelligence.loading' : 'siteIntelligence.retry') }}</button>
    <section v-for="group in groups" :key="group.key" :data-site-capability-group="group.key" class="site-capability-group mt-4">
      <h4 class="site-detail-label">{{ group.label }}</h4>
      <dl class="mt-1">
        <div v-for="item in group.items" :key="item.key" :data-site-capability="item.key" :data-site-capability-state="item.state" class="site-capability-row flex min-w-0 flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
          <dt class="min-w-0 break-words">{{ item.label }}</dt>
          <dd class="site-overview-state inline-flex items-baseline gap-2" :data-tone="item.tone">
            <span aria-hidden="true" class="site-overview-dot shrink-0" />{{ item.stateLabel }}
          </dd>
        </div>
      </dl>
    </section>
  </section>
</template>

<script setup lang="ts">
import type { SiteOverviewPresentation } from '~/utils/siteOverviewPresentation'
import type { RouteLocationRaw } from 'vue-router'
import { PhArrowRight } from '@phosphor-icons/vue'
defineProps<{ groups: SiteOverviewPresentation['capabilityGroups']; state: SiteOverviewPresentation['capabilityState']; insightsTo: RouteLocationRaw; retrying: boolean }>()
const emit = defineEmits<{ retry: [] }>()
const { t } = useI18n()
</script>
