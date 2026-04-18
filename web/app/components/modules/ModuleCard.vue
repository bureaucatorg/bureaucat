<script setup lang="ts">
import { Layers, CalendarDays, User as UserIcon } from "lucide-vue-next";
import type { Module } from "~/types";

const props = defineProps<{
  module: Module;
  showProject?: boolean;
  to?: string;
}>();

const destination = computed(
  () =>
    props.to ??
    (props.module.project_key
      ? `/projects/${props.module.project_key}/modules/${props.module.id}`
      : `#`)
);

const statusStyles: Record<string, string> = {
  backlog:     "border-muted-foreground/30 bg-muted text-muted-foreground",
  planned:     "border-sky-500/30 bg-sky-500/10 text-sky-700 dark:text-sky-300",
  in_progress: "border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-500",
  paused:      "border-orange-500/30 bg-orange-500/10 text-orange-700 dark:text-orange-300",
  completed:   "border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300",
  cancelled:   "border-rose-500/30 bg-rose-500/10 text-rose-700 dark:text-rose-300",
};

const progressPct = computed(() => {
  if (props.module.total_tasks === 0) return 0;
  return Math.round((props.module.completed_tasks / props.module.total_tasks) * 100);
});

const visibleMembers = computed(() => (props.module.members ?? []).slice(0, 3));
const extraMembers = computed(
  () => Math.max(0, (props.module.members?.length ?? 0) - visibleMembers.value.length)
);

function formatDate(d?: string): string {
  if (!d) return "";
  const dt = new Date(d + "T00:00:00");
  return dt.toLocaleDateString("en-US", { month: "short", day: "numeric" });
}

function formatRange(a?: string, b?: string): string {
  if (!a && !b) return "No dates";
  if (a && !b) return `From ${formatDate(a)}`;
  if (!a && b) return `Until ${formatDate(b)}`;
  const da = new Date(a! + "T00:00:00");
  const db = new Date(b! + "T00:00:00");
  if (da.getFullYear() !== db.getFullYear()) {
    return `${da.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" })} \u2192 ${db.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" })}`;
  }
  return `${formatDate(a)} \u2192 ${formatDate(b)}, ${db.getFullYear()}`;
}

function initials(first: string, last: string, username: string): string {
  const f = (first || "").trim()[0] || "";
  const l = (last || "").trim()[0] || "";
  if (f || l) return (f + l).toUpperCase();
  return (username || "?").slice(0, 2).toUpperCase();
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
            <Layers
              class="size-5 text-muted-foreground transition-colors group-hover:text-amber-600 dark:group-hover:text-amber-500"
            />
          </div>
          <span
            :class="[
              'rounded-md border px-2 py-0.5 text-[11px] font-medium uppercase tracking-wide',
              statusStyles[module.status] || statusStyles.backlog,
            ]"
          >
            {{ module.status.replace("_", " ") }}
          </span>
        </div>
        <CardTitle class="mt-3 line-clamp-2 text-base font-semibold">
          {{ module.title }}
        </CardTitle>
        <p
          v-if="showProject && module.project_name"
          class="font-mono text-xs text-muted-foreground"
        >
          {{ module.project_key }} · {{ module.project_name }}
        </p>
        <div class="mt-1 flex items-center gap-1.5 text-xs text-muted-foreground">
          <CalendarDays class="size-3.5" />
          <span>{{ formatRange(module.start_date, module.end_date) }}</span>
        </div>
      </CardHeader>
      <CardContent class="pt-0">
        <div class="mb-3 flex items-center justify-between gap-2">
          <div v-if="module.lead" class="flex min-w-0 items-center gap-1.5 text-xs">
            <UserIcon class="size-3.5 text-muted-foreground" />
            <span class="truncate">
              Lead:
              <span class="font-medium">
                {{ module.lead.first_name || module.lead.username }}
              </span>
            </span>
          </div>
          <div v-else class="text-xs text-muted-foreground">No lead</div>
          <div class="flex -space-x-1.5">
            <Avatar
              v-for="m in visibleMembers"
              :key="m.user_id"
              class="size-6 border border-background"
              :title="`${m.first_name} ${m.last_name}`"
            >
              <AvatarImage v-if="m.avatar_url" :src="m.avatar_url" />
              <AvatarFallback class="text-[10px]" :seed="m.user_id">
                {{ initials(m.first_name, m.last_name, m.username) }}
              </AvatarFallback>
            </Avatar>
            <div
              v-if="extraMembers > 0"
              class="flex size-6 items-center justify-center rounded-full border border-background bg-muted text-[10px] font-semibold text-muted-foreground"
            >
              +{{ extraMembers }}
            </div>
          </div>
        </div>
        <div class="mb-2 flex items-center justify-between text-xs">
          <span class="text-muted-foreground">
            {{ module.completed_tasks }} / {{ module.total_tasks }} done
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
