<script setup lang="ts">
import {
  Loader2,
  Calendar,
  Mail,
  AtSign,
  Plus,
  Edit2,
  Trash2,
  UserPlus,
  UserMinus,
  Tag,
  Tags,
  ArrowRight,
  MessageSquarePlus,
  MessageSquareDiff,
  MessageSquareX,
  Circle,
  ChevronLeft,
  ChevronRight,
} from "lucide-vue-next";
import type { ActivityType, UserActivityEntry, UserActivityDateCount } from "~/types";
import { ACTIVITY_TYPE_LABELS } from "~/types";

definePageMeta({
  middleware: ["auth"],
});

const route = useRoute();
const userId = computed(() => route.params.id as string);

const { getAuthHeader } = useAuth();

const loading = ref(true);
const error = ref<string | null>(null);
const user = ref<{
  id: string;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  user_type: string;
  avatar_url?: string;
  created_at: string;
} | null>(null);

// Activity state
const activities = ref<UserActivityEntry[]>([]);
const activitiesLoading = ref(false);
const activitiesPage = ref(1);
const activitiesTotal = ref(0);
const activitiesTotalPages = ref(0);

// Graph state
const graphData = ref<UserActivityDateCount[]>([]);
const graphLoading = ref(false);

useHead({
  title: computed(() => {
    if (user.value) return `${user.value.first_name} ${user.value.last_name}`;
    return "Profile";
  }),
});

async function loadUser() {
  loading.value = true;
  error.value = null;

  try {
    const response = await fetch(`/api/v1/users/${userId.value}`, {
      headers: getAuthHeader(),
    });

    if (!response.ok) {
      error.value = "User not found";
      return;
    }

    user.value = await response.json();
  } catch {
    error.value = "Failed to load user";
  } finally {
    loading.value = false;
  }
}

async function loadActivities(page = 1) {
  activitiesLoading.value = true;
  try {
    const response = await fetch(
      `/api/v1/users/${userId.value}/activity?page=${page}&per_page=20`,
      { headers: getAuthHeader() }
    );
    if (response.ok) {
      const data = await response.json();
      activities.value = data.activities ?? [];
      activitiesTotal.value = data.total ?? 0;
      activitiesPage.value = data.page ?? 1;
      activitiesTotalPages.value = data.total_pages ?? 0;
    } else {
      console.error("Failed to load activities:", response.status, await response.text());
    }
  } catch (e) {
    console.error("Failed to load activities:", e);
  } finally {
    activitiesLoading.value = false;
  }
}

async function loadGraph() {
  graphLoading.value = true;
  try {
    const response = await fetch(
      `/api/v1/users/${userId.value}/activity/graph`,
      { headers: getAuthHeader() }
    );
    if (response.ok) {
      graphData.value = await response.json();
    }
  } catch {
    // silently fail
  } finally {
    graphLoading.value = false;
  }
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

function formatRelativeDate(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return "just now";
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;

  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: date.getFullYear() !== now.getFullYear() ? "numeric" : undefined,
  });
}

const iconMap: Record<ActivityType, typeof Plus> = {
  task_created: Plus,
  task_updated: Edit2,
  task_deleted: Trash2,
  assignee_added: UserPlus,
  assignee_removed: UserMinus,
  label_added: Tag,
  label_removed: Tags,
  state_changed: ArrowRight,
  comment_created: MessageSquarePlus,
  comment_updated: MessageSquareDiff,
  comment_deleted: MessageSquareX,
};

// Contribution graph logic
const WEEKS = 52;
const DAYS_PER_WEEK = 7;

const graphCells = computed(() => {
  // Build a map of date -> count
  const countMap = new Map<string, number>();
  for (const entry of graphData.value) {
    countMap.set(entry.date, entry.count);
  }

  // Build grid: 52 weeks x 7 days, ending today
  const today = new Date();
  const cells: { date: string; count: number; level: number }[] = [];

  // Find the start date: go back to the Sunday of the week 52 weeks ago
  const startDate = new Date(today);
  startDate.setDate(startDate.getDate() - (WEEKS * DAYS_PER_WEEK) - startDate.getDay());

  // Find max for scaling
  let maxCount = 0;
  for (const entry of graphData.value) {
    if (entry.count > maxCount) maxCount = entry.count;
  }

  const totalDays = (WEEKS + 1) * DAYS_PER_WEEK;
  for (let i = 0; i <= totalDays; i++) {
    const d = new Date(startDate);
    d.setDate(startDate.getDate() + i);
    if (d > today) break;

    const dateStr = d.toISOString().split("T")[0];
    const count = countMap.get(dateStr) || 0;

    let level = 0;
    if (count > 0 && maxCount > 0) {
      const ratio = count / maxCount;
      if (ratio <= 0.25) level = 1;
      else if (ratio <= 0.5) level = 2;
      else if (ratio <= 0.75) level = 3;
      else level = 4;
    }

    cells.push({ date: dateStr, count, level });
  }

  return cells;
});

const graphWeeks = computed(() => {
  const weeks: typeof graphCells.value[] = [];
  let currentWeek: typeof graphCells.value = [];

  for (const cell of graphCells.value) {
    const dayOfWeek = new Date(cell.date + "T00:00:00").getDay();
    if (dayOfWeek === 0 && currentWeek.length > 0) {
      weeks.push(currentWeek);
      currentWeek = [];
    }
    currentWeek.push(cell);
  }
  if (currentWeek.length > 0) {
    weeks.push(currentWeek);
  }

  return weeks;
});

const totalActivitiesThisYear = computed(() => {
  return graphData.value.reduce((sum, d) => sum + d.count, 0);
});

const monthLabels = computed(() => {
  const labels: { label: string; col: number }[] = [];
  let lastMonth = -1;

  for (let w = 0; w < graphWeeks.value.length; w++) {
    const week = graphWeeks.value[w];
    if (week.length === 0) continue;
    const firstDay = new Date(week[0].date + "T00:00:00");
    const month = firstDay.getMonth();
    if (month !== lastMonth) {
      labels.push({
        label: firstDay.toLocaleDateString("en-US", { month: "short" }),
        col: w,
      });
      lastMonth = month;
    }
  }

  return labels;
});

onMounted(async () => {
  await loadUser();
  if (user.value) {
    loadGraph();
    loadActivities();
  }
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main id="main-content" class="flex-1">
      <div class="mx-auto max-w-4xl px-6 py-8">
        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center py-20">
          <Loader2 class="size-8 animate-spin text-muted-foreground" />
        </div>

        <!-- Error -->
        <div
          v-else-if="error"
          class="flex flex-col items-center justify-center py-20"
        >
          <p class="text-lg text-destructive">{{ error }}</p>
          <Button class="mt-4" variant="outline" as-child>
            <NuxtLink to="/projects">Back to Projects</NuxtLink>
          </Button>
        </div>

        <!-- Profile -->
        <template v-else-if="user">
          <div class="flex items-start gap-6">
            <Avatar class="size-20">
              <AvatarImage
                v-if="user.avatar_url"
                :src="user.avatar_url"
                :alt="`${user.first_name} ${user.last_name}`"
              />
              <AvatarFallback class="text-2xl">
                {{ user.first_name?.[0] }}{{ user.last_name?.[0] }}
              </AvatarFallback>
            </Avatar>

            <div class="flex-1">
              <h1 class="text-2xl font-bold">
                {{ user.first_name }} {{ user.last_name }}
              </h1>

              <div class="mt-3 space-y-2 text-sm text-muted-foreground">
                <div class="flex items-center gap-2">
                  <AtSign class="size-4" />
                  <span>{{ user.username }}</span>
                </div>
                <div class="flex items-center gap-2">
                  <Mail class="size-4" />
                  <span>{{ user.email }}</span>
                </div>
                <div class="flex items-center gap-2">
                  <Calendar class="size-4" />
                  <span>Member since {{ formatDate(user.created_at) }}</span>
                </div>
              </div>

              <Badge class="mt-3 capitalize" variant="secondary">
                {{ user.user_type }}
              </Badge>
            </div>
          </div>

          <!-- Contribution Graph -->
          <div class="mt-10">
            <h2 class="mb-4 text-lg font-semibold">
              {{ totalActivitiesThisYear }} contributions in the last year
            </h2>

            <Card>
              <CardContent class="pt-6">
                <div v-if="graphLoading" class="flex items-center justify-center py-8">
                  <Loader2 class="size-6 animate-spin text-muted-foreground" />
                </div>
                <div v-else class="overflow-x-auto">
                  <div class="inline-flex flex-col">
                    <!-- Month labels row -->
                    <div class="flex">
                      <div class="w-7 shrink-0" />
                      <div class="flex gap-[3px]">
                        <div
                          v-for="(week, wi) in graphWeeks"
                          :key="'m' + wi"
                          class="w-[11px] shrink-0"
                        >
                          <span
                            v-if="monthLabels.some(m => m.col === wi)"
                            class="text-[10px] text-muted-foreground"
                          >{{ monthLabels.find(m => m.col === wi)!.label }}</span>
                        </div>
                      </div>
                    </div>
                    <!-- Grid with day labels -->
                    <div class="mt-1 flex">
                      <!-- Day labels -->
                      <div class="flex w-7 shrink-0 flex-col gap-[3px] text-[10px] text-muted-foreground">
                        <span class="h-[11px]" />
                        <span class="flex h-[11px] items-center">Mon</span>
                        <span class="h-[11px]" />
                        <span class="flex h-[11px] items-center">Wed</span>
                        <span class="h-[11px]" />
                        <span class="flex h-[11px] items-center">Fri</span>
                        <span class="h-[11px]" />
                      </div>
                      <!-- Cells -->
                      <div class="flex gap-[3px]">
                        <div
                          v-for="(week, wi) in graphWeeks"
                          :key="wi"
                          class="flex flex-col gap-[3px]"
                        >
                          <div
                            v-for="cell in week"
                            :key="cell.date"
                            class="size-[11px] rounded-[2px]"
                            :class="{
                              'bg-muted dark:bg-muted/50': cell.level === 0,
                              'bg-amber-200 dark:bg-amber-900': cell.level === 1,
                              'bg-amber-400 dark:bg-amber-700': cell.level === 2,
                              'bg-amber-500 dark:bg-amber-500': cell.level === 3,
                              'bg-amber-700 dark:bg-amber-400': cell.level === 4,
                            }"
                            :title="`${cell.count} ${cell.count === 1 ? 'activity' : 'activities'} on ${cell.date}`"
                          />
                        </div>
                      </div>
                    </div>
                  </div>
                  <!-- Legend -->
                  <div class="mt-3 flex items-center justify-end gap-1 text-[10px] text-muted-foreground">
                    <span>Less</span>
                    <div class="size-[11px] rounded-[2px] bg-muted dark:bg-muted/50" />
                    <div class="size-[11px] rounded-[2px] bg-amber-200 dark:bg-amber-900" />
                    <div class="size-[11px] rounded-[2px] bg-amber-400 dark:bg-amber-700" />
                    <div class="size-[11px] rounded-[2px] bg-amber-500 dark:bg-amber-500" />
                    <div class="size-[11px] rounded-[2px] bg-amber-700 dark:bg-amber-400" />
                    <span>More</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          <!-- Activity Feed -->
          <div class="mt-10">
            <h2 class="mb-4 text-lg font-semibold">Activity</h2>

            <div v-if="activitiesLoading && activities.length === 0" class="flex items-center justify-center py-12">
              <Loader2 class="size-6 animate-spin text-muted-foreground" />
            </div>

            <div
              v-else-if="activities.length === 0"
              class="rounded-lg border border-dashed py-12 text-center text-sm text-muted-foreground"
            >
              No activity yet
            </div>

            <div v-else class="space-y-0">
              <div
                v-for="(activity, index) in activities"
                :key="activity.id"
                class="relative flex gap-3 pb-4"
              >
                <!-- Timeline line -->
                <div
                  v-if="index < activities.length - 1"
                  class="absolute left-[11px] top-6 h-full w-px bg-border"
                />

                <!-- Icon -->
                <div class="relative z-10 flex size-6 shrink-0 items-center justify-center rounded-full border bg-background">
                  <component
                    :is="iconMap[activity.activity_type] || Circle"
                    class="size-3 text-muted-foreground"
                  />
                </div>

                <!-- Content -->
                <div class="min-w-0 flex-1 pt-0.5">
                  <p class="text-sm">
                    <span class="text-muted-foreground">
                      {{ ACTIVITY_TYPE_LABELS[activity.activity_type] || activity.activity_type }}
                      <template v-if="activity.field_name">
                        <span class="font-medium text-foreground">{{ activity.field_name }}</span>
                      </template>
                    </span>
                  </p>
                  <div class="mt-1 flex flex-wrap items-center gap-x-2 gap-y-0.5">
                    <NuxtLink
                      :to="`/projects/${activity.project_key}/tasks/${activity.task_number}`"
                      class="text-xs font-medium text-amber-600 hover:text-amber-700 dark:text-amber-500 dark:hover:text-amber-400"
                    >
                      {{ activity.project_key }}-{{ activity.task_number }}
                    </NuxtLink>
                    <span class="truncate text-xs text-muted-foreground">
                      {{ activity.task_title }}
                    </span>
                    <span class="text-xs text-muted-foreground/60">
                      {{ formatRelativeDate(activity.created_at) }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- Pagination -->
              <div
                v-if="activitiesTotalPages > 1"
                class="flex items-center justify-between border-t pt-4"
              >
                <p class="text-sm text-muted-foreground">
                  {{ activitiesTotal }} total
                </p>
                <div class="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    aria-label="Previous page"
                    :disabled="activitiesPage <= 1"
                    @click="loadActivities(activitiesPage - 1)"
                  >
                    <ChevronLeft class="size-4" />
                  </Button>
                  <span class="text-sm text-muted-foreground">
                    Page {{ activitiesPage }} of {{ activitiesTotalPages }}
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    aria-label="Next page"
                    :disabled="activitiesPage >= activitiesTotalPages"
                    @click="loadActivities(activitiesPage + 1)"
                  >
                    <ChevronRight class="size-4" />
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </template>
      </div>
    </main>
  </div>
</template>
