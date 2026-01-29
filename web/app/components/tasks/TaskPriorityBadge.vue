<script setup lang="ts">
import { PRIORITY_LABELS } from "~/types";

const props = defineProps<{
  priority: number;
  showLabel?: boolean;
}>();

const info = computed(() => PRIORITY_LABELS[props.priority] || PRIORITY_LABELS[0]);
</script>

<template>
  <div class="flex items-center gap-1.5">
    <div
      class="size-2.5 rounded-full"
      :style="{ backgroundColor: info.color }"
    />
    <span v-if="showLabel" class="text-sm">{{ info.label }}</span>
  </div>
</template>
