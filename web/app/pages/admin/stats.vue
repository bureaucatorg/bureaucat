<script setup lang="ts">
import { BarChart3, Loader2, Users, LayoutGrid, ListTodo, GitBranch, FileText, Building2, Paperclip, HardDrive, Mail, Ban, Link2, Bot, UserCheck, MessageSquare } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { DateValue } from "@internationalized/date";
import type { AdminStats } from "~/composables/useAdmin";
import { formatBytes, statsRangeLabel, statsToday, titleCase, toIsoDay, type StatsBar } from "~/utils/stats";

definePageMeta({
  middleware: ["admin"],
});

useSeoMeta({ title: "Stats" });

const { getStats } = useAdmin();

const loading = ref(true);
const stats = ref<AdminStats | null>(null);

// Range selection (defaults to the trailing 30 days).
const toDate = ref<DateValue>(statsToday());
const fromDate = ref<DateValue>(statsToday().subtract({ days: 29 }));
const rangeLabel = computed(() => statsRangeLabel(fromDate.value, toDate.value));

async function loadStats() {
  loading.value = true;
  const result = await getStats(toIsoDay(fromDate.value), toIsoDay(toDate.value));
  loading.value = false;

  if (result.success && result.data) {
    stats.value = result.data;
  } else {
    toast.error(result.error || "Failed to load stats");
  }
}

// Reload the per-day series whenever the range changes.
watch([fromDate, toDate], loadStats);

onMounted(loadStats);

const totalCards = computed(() => {
  const t = stats.value?.totals;
  if (!t) return [];
  return [
    { label: "Workspaces", value: t.workspaces.toLocaleString(), icon: Building2, color: "text-violet-400" },
    { label: "Projects", value: t.projects.toLocaleString(), icon: LayoutGrid, color: "text-blue-400" },
    { label: "Tasks", value: t.tasks.toLocaleString(), icon: ListTodo, color: "text-emerald-400" },
    { label: "Subtasks", value: t.subtasks.toLocaleString(), icon: GitBranch, color: "text-teal-400" },
    { label: "Blocked tasks", value: t.blocked_tasks.toLocaleString(), icon: Ban, color: "text-red-400" },
    { label: "Blocking tasks", value: t.blocking_tasks.toLocaleString(), icon: Link2, color: "text-orange-400" },
    { label: "Pages", value: t.pages.toLocaleString(), icon: FileText, color: "text-amber-400" },
    { label: "Users", value: t.users.toLocaleString(), icon: Users, color: "text-rose-400" },
    {
      label: "Email notifications enabled",
      value: `${t.email_notification_users.toLocaleString()} / ${t.users.toLocaleString()}`,
      icon: Mail,
      color: "text-indigo-400",
    },
    { label: "Attachments", value: t.attachments.toLocaleString(), icon: Paperclip, color: "text-cyan-400" },
    { label: "Attachments size", value: formatBytes(t.attachments_bytes), icon: HardDrive, color: "text-sky-400" },
    { label: "AI conversations", value: t.ai_conversations.toLocaleString(), icon: Bot, color: "text-fuchsia-400" },
    {
      label: "AI via Mattermost",
      value: `${t.ai_mattermost_conversations.toLocaleString()} / ${t.ai_conversations.toLocaleString()}`,
      icon: MessageSquare,
      color: "text-purple-400",
    },
    {
      label: "AI users",
      value: `${t.ai_users.toLocaleString()} / ${t.users.toLocaleString()}`,
      icon: UserCheck,
      color: "text-lime-400",
    },
  ];
});

// --- Charts ---------------------------------------------------------------

const trendCharts = computed(() => {
  const s = stats.value?.series;
  return [
    { key: "tasks", title: "Tasks created", color: "#6EE7B7", data: s?.tasks ?? [] },
    { key: "subtasks", title: "Subtasks created", color: "#93C5FD", data: s?.subtasks ?? [] },
    { key: "comments", title: "Comments created", color: "#F9A8D4", data: s?.comments ?? [] },
    { key: "activity", title: "Activity created", color: "#A5B4FC", data: s?.activity ?? [] },
    { key: "attachments", title: "Attachments created", color: "#FDBA74", data: s?.attachments ?? [] },
    { key: "pages", title: "Pages created", color: "#C4B5FD", data: s?.pages ?? [] },
    { key: "cycles", title: "Cycles created", color: "#F0ABFC", data: s?.cycles ?? [] },
    { key: "modules", title: "Modules created", color: "#5EEAD4", data: s?.modules ?? [] },
  ];
});

// Views-created series, split by visibility (stacked private + shared).
const viewsTrendSeries = [
  { key: "private", label: "Private", color: "#FCD34D" }, // amber-300
  { key: "shared", label: "Shared", color: "#93C5FD" }, // blue-300
];

const viewsSeries = computed(() => stats.value?.series.views ?? []);

const aiTrendSeries = [
  { key: "web", label: "Web", color: "#F0ABFC" }, // fuchsia-300
  { key: "mattermost", label: "Mattermost", color: "#C4B5FD" }, // violet-300
];

const aiSeries = computed(() => stats.value?.series.ai_conversations ?? []);

// Light/pastel palette (Tailwind -300 shades) matching the app's soft chips.
const STATE_LIGHT: Record<string, string> = {
  backlog: "#D1D5DB", // gray-300
  unstarted: "#93C5FD", // blue-300
  started: "#6EE7B7", // emerald-300
  completed: "#86EFAC", // green-300
  cancelled: "#E5E7EB", // gray-200
};

const PRIORITY_LIGHT: Record<string, string> = {
  "No priority": "#D1D5DB", // gray-300
  "Low": "#93C5FD", // blue-300
  "Medium": "#FCD34D", // amber-300
  "High": "#FDBA74", // orange-300
  "Urgent": "#FCA5A5", // red-300
};

const stateBars = computed<StatsBar[]>(() =>
  (stats.value?.tasks_by_state ?? []).map((s) => ({
    label: titleCase(s.label),
    count: s.count,
    fill: STATE_LIGHT[s.label] ?? "#D1D5DB",
  }))
);

const priorityBars = computed<StatsBar[]>(() =>
  (stats.value?.tasks_by_priority ?? []).map((p) => ({
    label: p.label,
    count: p.count,
    fill: PRIORITY_LIGHT[p.label] ?? "#D1D5DB",
  }))
);

const maxWorkspaceProjects = computed(() =>
  Math.max(1, ...(stats.value?.projects_per_workspace ?? []).map((w) => w.project_count))
);

const maxProjectTasks = computed(() =>
  Math.max(1, ...(stats.value?.top_projects ?? []).map((p) => p.task_count))
);
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main id="main-content" class="flex-1">
      <div class="mx-auto max-w-6xl px-6 py-8">
        <!-- Header -->
        <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
          <div>
            <div class="mb-1 flex items-center gap-2.5">
              <div class="flex size-8 items-center justify-center rounded-lg bg-foreground">
                <BarChart3 class="size-4 text-background" />
              </div>
              <h1 class="text-2xl font-bold tracking-tight">Stats</h1>
            </div>
            <p class="text-sm text-muted-foreground">
              System-wide metrics and activity trends
            </p>
          </div>

          <StatsRangePicker v-model:from="fromDate" v-model:to="toDate" />
        </div>

        <!-- Loading -->
        <div v-if="loading && !stats" class="flex items-center justify-center py-24">
          <Loader2 class="size-6 animate-spin text-muted-foreground" />
        </div>

        <template v-else-if="stats">
          <!-- Totals -->
          <div class="grid gap-3 sm:grid-cols-3 lg:grid-cols-4">
            <Card v-for="card in totalCards" :key="card.label" class="gap-0 py-4">
              <CardHeader class="flex flex-row items-center justify-between space-y-0 px-4 pb-1.5">
                <CardDescription class="text-xs">{{ card.label }}</CardDescription>
                <component :is="card.icon" :class="['size-4', card.color]" />
              </CardHeader>
              <CardContent class="px-4">
                <div class="text-2xl font-bold tracking-tight">{{ card.value }}</div>
              </CardContent>
            </Card>
          </div>

          <!-- Trends -->
          <div class="mt-8">
            <div class="mb-3">
              <h2 class="text-lg font-semibold">Activity trends</h2>
              <p class="text-sm text-muted-foreground">
                Items created per day · {{ rangeLabel }}
              </p>
            </div>

            <div class="grid gap-3 lg:grid-cols-3">
              <StatsTrendChart
                v-for="(chart, i) in trendCharts"
                :key="chart.key"
                :title="chart.title"
                :series="[{ key: 'count', label: 'Created', color: chart.color }]"
                :data="chart.data"
                :style="{ order: i < 2 ? i + 1 : i + 2 }"
              />

              <!-- Views created, stacked by visibility (placed as the third trend) -->
              <StatsTrendChart
                title="Views created"
                :series="viewsTrendSeries"
                :data="viewsSeries"
                empty-text="No views yet"
                :style="{ order: 3 }"
              />

              <StatsTrendChart
                title="AI conversations"
                :series="aiTrendSeries"
                :data="aiSeries"
                empty-text="No AI conversations yet"
                :style="{ order: trendCharts.length + 2 }"
              />
            </div>
          </div>

          <!-- Breakdowns -->
          <div class="mt-8">
            <div class="mb-3">
              <h2 class="text-lg font-semibold">Task breakdowns</h2>
              <p class="text-sm text-muted-foreground">
                Distribution of top-level tasks by state and priority
              </p>
            </div>

            <div class="grid gap-3 lg:grid-cols-2">
              <StatsBarChart title="By state" :bars="stateBars" empty-text="No tasks yet" />
              <StatsBarChart title="By priority" :bars="priorityBars" empty-text="No tasks yet" />
            </div>
          </div>

          <!-- Projects -->
          <div class="mt-8">
            <div class="mb-3">
              <h2 class="text-lg font-semibold">Projects</h2>
              <p class="text-sm text-muted-foreground">
                Most active projects and distribution across workspaces
              </p>
            </div>

            <div class="grid gap-3 lg:grid-cols-2">
              <Card class="gap-3 py-4">
                <CardHeader class="px-4 pb-1">
                  <CardTitle class="text-base">Top projects by tasks</CardTitle>
                </CardHeader>
                <CardContent class="px-4">
                  <div v-if="stats.top_projects.length" class="space-y-3">
                    <div v-for="p in stats.top_projects" :key="p.project_key" class="space-y-1">
                      <div class="flex items-center justify-between text-sm">
                        <span class="truncate">
                          <span class="text-muted-foreground">{{ p.project_key }}</span>
                          <span class="ml-2 font-medium">{{ p.name }}</span>
                        </span>
                        <span class="ml-2 font-medium tabular-nums">{{ p.task_count }}</span>
                      </div>
                      <div class="h-1.5 w-full rounded-full bg-muted">
                        <div
                          class="h-1.5 rounded-full bg-blue-300"
                          :style="{ width: `${(p.task_count / maxProjectTasks) * 100}%` }"
                        />
                      </div>
                    </div>
                  </div>
                  <p v-else class="py-12 text-center text-sm text-muted-foreground">No projects yet</p>
                </CardContent>
              </Card>

              <Card class="gap-3 py-4">
                <CardHeader class="px-4 pb-1">
                  <CardTitle class="text-base">Projects per workspace</CardTitle>
                </CardHeader>
                <CardContent class="px-4">
                  <div v-if="stats.projects_per_workspace.length" class="space-y-3">
                    <div v-for="w in stats.projects_per_workspace" :key="w.workspace_key" class="space-y-1">
                      <div class="flex items-center justify-between text-sm">
                        <span class="truncate">
                          <span class="text-muted-foreground">{{ w.workspace_key }}</span>
                          <span class="ml-2 font-medium">{{ w.name }}</span>
                        </span>
                        <span class="ml-2 font-medium tabular-nums">{{ w.project_count }}</span>
                      </div>
                      <div class="h-1.5 w-full rounded-full bg-muted">
                        <div
                          class="h-1.5 rounded-full bg-blue-300"
                          :style="{ width: `${(w.project_count / maxWorkspaceProjects) * 100}%` }"
                        />
                      </div>
                    </div>
                  </div>
                  <p v-else class="py-12 text-center text-sm text-muted-foreground">No workspaces yet</p>
                </CardContent>
              </Card>
            </div>
          </div>
        </template>
      </div>
    </main>
  </div>
</template>
