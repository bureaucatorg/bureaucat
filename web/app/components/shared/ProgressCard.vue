<script setup lang="ts">
interface ProgressMetrics {
  total: number;
  completed: number;
  in_progress: number;
  todo: number;
  cancelled: number;
}

const props = defineProps<{
  metrics: ProgressMetrics | null;
}>();

const progressPct = computed(() => {
  const m = props.metrics;
  if (!m || m.total === 0) return 0;
  return Math.round((m.completed / m.total) * 100);
});
</script>

<template>
  <section class="rounded-lg border p-4">
    <h3 class="mb-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
      Progress
    </h3>
    <div class="flex items-baseline gap-2">
      <span class="text-3xl font-bold tabular-nums">{{ progressPct }}%</span>
      <span class="text-sm text-muted-foreground">
        {{ metrics?.completed ?? 0 }} / {{ metrics?.total ?? 0 }} done
      </span>
    </div>
    <div class="mt-3 h-2 w-full overflow-hidden rounded-full bg-muted">
      <div
        class="h-full rounded-full bg-amber-500 transition-all"
        :style="{ width: progressPct + '%' }"
      />
    </div>
    <dl class="mt-4 grid grid-cols-2 gap-2 text-xs">
      <div class="flex justify-between">
        <dt class="text-muted-foreground">Todo</dt>
        <dd class="font-medium tabular-nums">{{ metrics?.todo ?? 0 }}</dd>
      </div>
      <div class="flex justify-between">
        <dt class="text-muted-foreground">In progress</dt>
        <dd class="font-medium tabular-nums">{{ metrics?.in_progress ?? 0 }}</dd>
      </div>
      <div class="flex justify-between">
        <dt class="text-muted-foreground">Done</dt>
        <dd class="font-medium tabular-nums">{{ metrics?.completed ?? 0 }}</dd>
      </div>
      <div class="flex justify-between">
        <dt class="text-muted-foreground">Cancelled</dt>
        <dd class="font-medium tabular-nums">{{ metrics?.cancelled ?? 0 }}</dd>
      </div>
    </dl>
  </section>
</template>
