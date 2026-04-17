<script setup lang="ts">
import { Repeat, CalendarDays } from "lucide-vue-next";
import type { Cycle } from "~/types";

const props = defineProps<{
  cycle: Cycle;
  showProject?: boolean;
  to?: string;
}>();

const destination = computed(
  () =>
    props.to ??
    (props.cycle.project_key
      ? `/projects/${props.cycle.project_key}/cycles/${props.cycle.id}`
      : `#`)
);

const statusStyles: Record<string, string> = {
  upcoming:
    "border-sky-500/30 bg-sky-500/10 text-sky-700 dark:text-sky-300",
  active:
    "border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300",
  completed:
    "border-muted-foreground/30 bg-muted text-muted-foreground",
};

const progressPct = computed(() => {
  if (props.cycle.total_tasks === 0) return 0;
  return Math.round((props.cycle.completed_tasks / props.cycle.total_tasks) * 100);
});

function formatDate(d: string): string {
  if (!d) return "";
  const dt = new Date(d + "T00:00:00");
  return dt.toLocaleDateString("en-US", { month: "short", day: "numeric" });
}

function formatRange(a: string, b: string): string {
  if (!a || !b) return "";
  const da = new Date(a + "T00:00:00");
  const db = new Date(b + "T00:00:00");
  if (da.getFullYear() !== db.getFullYear()) {
    return `${da.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" })} \u2192 ${db.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" })}`;
  }
  return `${formatDate(a)} \u2192 ${formatDate(b)}, ${db.getFullYear()}`;
}
</script>

<template>
  <NuxtLink :to="destination">
    <Card
      class="group h-full cursor-pointer border-border/50 bg-background/50 transition-all hover:border-amber-500/30 hover:shadow-lg hover:shadow-amber-500/5"
    >
      <CardHeader class="pb-3">
        <div class="flex items-start justify-between gap-2">
          <div
            class="flex size-10 items-center justify-center rounded-lg bg-muted transition-colors group-hover:bg-amber-500/10"
          >
            <Repeat
              class="size-5 text-muted-foreground transition-colors group-hover:text-amber-600 dark:group-hover:text-amber-500"
            />
          </div>
          <span
            :class="[
              'rounded-md border px-2 py-0.5 text-[11px] font-medium uppercase tracking-wide',
              statusStyles[cycle.status] || statusStyles.upcoming,
            ]"
          >
            {{ cycle.status }}
          </span>
        </div>
        <CardTitle class="mt-3 line-clamp-2 text-base font-semibold">
          {{ cycle.title }}
        </CardTitle>
        <p
          v-if="showProject && cycle.project_name"
          class="font-mono text-xs text-muted-foreground"
        >
          {{ cycle.project_key }} · {{ cycle.project_name }}
        </p>
        <div class="mt-1 flex items-center gap-1.5 text-xs text-muted-foreground">
          <CalendarDays class="size-3.5" />
          <span>{{ formatRange(cycle.start_date, cycle.end_date) }}</span>
        </div>
      </CardHeader>
      <CardContent class="pt-0">
        <div class="mb-2 flex items-center justify-between text-xs">
          <span class="text-muted-foreground">
            {{ cycle.completed_tasks }} / {{ cycle.total_tasks }} done
          </span>
          <span class="font-semibold tabular-nums">{{ progressPct }}%</span>
        </div>
        <div class="h-1.5 w-full overflow-hidden rounded-full bg-muted">
          <div
            class="h-full rounded-full bg-amber-500 transition-all"
            :style="{ width: progressPct + '%' }"
          />
        </div>
      </CardContent>
    </Card>
  </NuxtLink>
</template>
