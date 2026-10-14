<script setup lang="ts">
import { Loader2, ListTodo, GitBranch, Ban, Link2, Users, FileText, Repeat, Layers, Paperclip, HardDrive } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { DateValue } from "@internationalized/date";
import type { ProjectStats, ProjectStatsDay } from "~/composables/useProjects";
import { formatBytes, statsRangeLabel, statsToday, toIsoDay, type StatsBar } from "~/utils/stats";

const props = defineProps<{ projectKey: string }>();

const { getProjectStats } = useProjects();

const loading = ref(true);
const stats = ref<ProjectStats | null>(null);
const toDate = ref<DateValue>(statsToday());
const fromDate = ref<DateValue>(statsToday().subtract({ days: 29 }));
const rangeLabel = computed(() => statsRangeLabel(fromDate.value, toDate.value));
let requestId = 0;

async function loadStats() {
  const id = ++requestId;
  loading.value = true;
  const result = await getProjectStats(props.projectKey, toIsoDay(fromDate.value), toIsoDay(toDate.value));
  if (id !== requestId) return;
  loading.value = false;
  if (result.success && result.data) {
    stats.value = result.data;
  } else {
    toast.error(result.error || "Failed to load project stats");
  }
}

watch([fromDate, toDate, () => props.projectKey], loadStats);
onMounted(loadStats);

const totalCards = computed(() => {
  const t = stats.value?.totals;
  if (!t) return [];
  return [
    { label: "Tasks", value: t.tasks.toLocaleString(), icon: ListTodo, color: "text-emerald-400" },
    { label: "Subtasks", value: t.subtasks.toLocaleString(), icon: GitBranch, color: "text-teal-400" },
    { label: "Blocked tasks", value: t.blocked_tasks.toLocaleString(), icon: Ban, color: "text-red-400" },
    { label: "Blocking tasks", value: t.blocking_tasks.toLocaleString(), icon: Link2, color: "text-orange-400" },
    { label: "Members", value: t.members.toLocaleString(), icon: Users, color: "text-rose-400" },
    { label: "Pages", value: t.pages.toLocaleString(), icon: FileText, color: "text-amber-400" },
    { label: "Cycles", value: t.cycles.toLocaleString(), icon: Repeat, color: "text-fuchsia-400" },
    { label: "Modules", value: t.modules.toLocaleString(), icon: Layers, color: "text-violet-400" },
    { label: "Attachments", value: t.attachments.toLocaleString(), icon: Paperclip, color: "text-cyan-400" },
    { label: "Attachments size", value: formatBytes(t.attachments_bytes), icon: HardDrive, color: "text-sky-400" },
  ];
});

const days = computed<ProjectStatsDay[]>(() => stats.value?.series.days ?? []);

const tasksSeries = [
  { key: "tasks", label: "Created", color: "#6EE7B7" }, // emerald-300
  { key: "completed", label: "Completed", color: "#93C5FD" }, // blue-300
];

const trendCharts: { key: keyof ProjectStatsDay; title: string; color: string }[] = [
  { key: "subtasks", title: "Subtasks created", color: "#93C5FD" },
  { key: "comments", title: "Comments created", color: "#F9A8D4" },
  { key: "activity", title: "Activity created", color: "#A5B4FC" },
  { key: "attachments", title: "Attachments created", color: "#FDBA74" },
  { key: "pages", title: "Pages created", color: "#C4B5FD" },
  { key: "cycles", title: "Cycles created", color: "#F0ABFC" },
  { key: "modules", title: "Modules created", color: "#5EEAD4" },
];

const PRIORITY_LIGHT: Record<string, string> = {
  "No priority": "#D1D5DB", // gray-300
  "Low": "#93C5FD", // blue-300
  "Medium": "#FCD34D", // amber-300
  "High": "#FDBA74", // orange-300
  "Urgent": "#FCA5A5", // red-300
};

const stateBars = computed<StatsBar[]>(() =>
  (stats.value?.tasks_by_state ?? []).map((s) => ({ label: s.name, count: s.count, fill: s.color || "#D1D5DB" })),
);

const priorityBars = computed<StatsBar[]>(() =>
  (stats.value?.tasks_by_priority ?? []).map((p) => ({
    label: p.label,
    count: p.count,
    fill: PRIORITY_LIGHT[p.label] ?? "#D1D5DB",
  })),
);

const maxWorkload = computed(() =>
  Math.max(1, stats.value?.totals.unassigned_open ?? 0, ...(stats.value?.workload ?? []).map((w) => w.open_tasks)),
);
</script>

<template>
  <div>
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <p class="text-sm text-muted-foreground">Project metrics and activity trends</p>
      <StatsRangePicker v-model:from="fromDate" v-model:to="toDate" />
    </div>

    <div v-if="loading && !stats" class="flex items-center justify-center py-24">
      <Loader2 class="size-6 animate-spin text-muted-foreground" />
    </div>

    <template v-else-if="stats">
      <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
        <Card v-for="card in totalCards" :key="card.label" class="gap-0 py-4">
          <CardHeader class="flex flex-row items-center justify-between space-y-0 px-4 pb-1.5">
            <CardDescription class="text-xs">{{ card.label }}</CardDescription>
            <component :is="card.icon" :class="['size-4 shrink-0', card.color]" />
          </CardHeader>
          <CardContent class="px-4">
            <div class="text-2xl font-bold tracking-tight">{{ card.value }}</div>
          </CardContent>
        </Card>
      </div>

      <div class="mt-8">
        <div class="mb-3">
          <h2 class="text-lg font-semibold">Activity trends</h2>
          <p class="text-sm text-muted-foreground">Items per day · {{ rangeLabel }}</p>
        </div>
        <div class="grid gap-3 lg:grid-cols-3">
          <StatsTrendChart title="Tasks created vs completed" :series="tasksSeries" :data="days" />
          <StatsTrendChart
            v-for="chart in trendCharts"
            :key="chart.key"
            :title="chart.title"
            :series="[{ key: chart.key, label: 'Created', color: chart.color }]"
            :data="days"
          />
        </div>
      </div>

      <div class="mt-8">
        <div class="mb-3">
          <h2 class="text-lg font-semibold">Task breakdowns</h2>
          <p class="text-sm text-muted-foreground">Top-level tasks by state and priority</p>
        </div>
        <div class="grid gap-3 lg:grid-cols-2">
          <StatsBarChart title="By state" :bars="stateBars" empty-text="No states yet" />
          <StatsBarChart title="By priority" :bars="priorityBars" empty-text="No tasks yet" />
        </div>
      </div>

      <div class="mt-8">
        <div class="mb-3">
          <h2 class="text-lg font-semibold">Workload</h2>
          <p class="text-sm text-muted-foreground">Open tasks and subtasks per assignee</p>
        </div>
        <Card class="gap-3 py-4">
          <CardContent class="px-4">
            <div v-if="stats.workload.length || stats.totals.unassigned_open" class="space-y-3">
              <div v-for="w in stats.workload" :key="w.user_id" class="space-y-1">
                <div class="flex items-center justify-between gap-2 text-sm">
                  <span class="truncate">
                    <span class="font-medium">{{ w.first_name }} {{ w.last_name }}</span>
                    <span class="ml-2 text-muted-foreground">@{{ w.username }}</span>
                  </span>
                  <span class="font-medium tabular-nums">{{ w.open_tasks }}</span>
                </div>
                <div class="h-1.5 w-full rounded-full bg-muted">
                  <div class="h-1.5 rounded-full bg-blue-300" :style="{ width: `${(w.open_tasks / maxWorkload) * 100}%` }" />
                </div>
              </div>
              <div v-if="stats.totals.unassigned_open" class="space-y-1">
                <div class="flex items-center justify-between gap-2 text-sm">
                  <span class="italic text-muted-foreground">Unassigned</span>
                  <span class="font-medium tabular-nums">{{ stats.totals.unassigned_open }}</span>
                </div>
                <div class="h-1.5 w-full rounded-full bg-muted">
                  <div
                    class="h-1.5 rounded-full bg-muted-foreground/40"
                    :style="{ width: `${(stats.totals.unassigned_open / maxWorkload) * 100}%` }"
                  />
                </div>
              </div>
            </div>
            <p v-else class="py-12 text-center text-sm text-muted-foreground">No open tasks</p>
          </CardContent>
        </Card>
      </div>
    </template>
  </div>
</template>
