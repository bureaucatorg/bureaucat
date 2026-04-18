<script setup lang="ts">
interface StateBucket {
  state_id: string;
  state_name: string;
  state_color: string;
  state_type: string;
  count: number;
}

defineProps<{
  buckets: StateBucket[];
}>();
</script>

<template>
  <section v-if="buckets.length" class="rounded-lg border p-4">
    <h3 class="mb-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
      By state
    </h3>
    <ul class="space-y-1.5">
      <li
        v-for="b in buckets"
        :key="b.state_id"
        class="flex items-center justify-between gap-2 text-sm"
      >
        <span class="flex items-center gap-2 truncate">
          <span
            class="size-2 rounded-full"
            :style="{ backgroundColor: b.state_color || '#6B7280' }"
          />
          <span class="truncate">{{ b.state_name }}</span>
        </span>
        <span class="font-medium tabular-nums text-muted-foreground">
          {{ b.count }}
        </span>
      </li>
    </ul>
  </section>
</template>
