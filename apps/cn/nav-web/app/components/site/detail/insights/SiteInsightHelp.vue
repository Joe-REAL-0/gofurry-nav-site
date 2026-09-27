<template>
  <span class="inline-flex items-center gap-1">{{ label }}
    <span class="site-detail-help inline-flex" tabindex="0" :title="help" :aria-label="label + ': ' + help"><PhInfo class="site-detail-icon" aria-hidden="true" /></span>
  </span>
</template>
<script setup lang="ts">
import { PhInfo } from '@phosphor-icons/vue'
defineProps<{ label: string; help: string }>()
</script>
