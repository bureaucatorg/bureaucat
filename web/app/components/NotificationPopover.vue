<script setup lang="ts">
import {
  Bell,
  Loader2,
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
} from "lucide-vue-next";
import type { ActivityType, UserActivityEntry } from "~/types";
import { ACTIVITY_TYPE_LABELS } from "~/types";

const { user, getAuthHeader } = useAuth();

const open = ref(false);
const loading = ref(false);
const loadingMore = ref(false);
const activities = ref<UserActivityEntry[]>([]);
const page = ref(1);
const hasMore = ref(true);
const perPage = 20;
const maxItems = 100;

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

async function loadActivity(pageNum: number) {
  if (!user.value) return;

  const isFirstPage = pageNum === 1;
  if (isFirstPage) {
    loading.value = true;
  } else {
    loadingMore.value = true;
  }

  try {
    const response = await fetch(
      `/api/v1/me/notifications?page=${pageNum}&per_page=${perPage}`,
      { headers: getAuthHeader() }
    );
    if (response.ok) {
      const data = await response.json();
      const items: UserActivityEntry[] = data.activities ?? [];
      if (isFirstPage) {
        activities.value = items;
      } else {
        activities.value = [...activities.value, ...items];
      }
      page.value = pageNum;
      hasMore.value = items.length === perPage && activities.value.length < maxItems;
    }
  } catch {
    // silently fail
  } finally {
    loading.value = false;
    loadingMore.value = false;
  }
}

function onScroll(e: Event) {
  if (loadingMore.value || !hasMore.value) return;
  const target = e.target as HTMLElement;
  if (!target) return;
  const nearBottom = target.scrollHeight - target.scrollTop - target.clientHeight < 80;
  if (nearBottom) {
    loadActivity(page.value + 1);
  }
}

watch(open, (isOpen) => {
  if (isOpen) {
    page.value = 1;
    hasMore.value = true;
    loadActivity(1);
  }
});
</script>

<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <Button variant="ghost" size="icon" class="size-9" aria-label="Notifications">
        <Bell class="size-4" />
      </Button>
    </PopoverTrigger>
    <PopoverContent align="end" class="w-80 p-0">
      <div class="border-b px-4 py-3">
        <h3 class="text-sm font-semibold">Notifications</h3>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-8">
        <Loader2 class="size-5 animate-spin text-muted-foreground" />
      </div>

      <!-- Empty -->
      <div
        v-else-if="activities.length === 0"
        class="px-4 py-8 text-center text-sm text-muted-foreground"
      >
        No notifications yet
      </div>

      <!-- Activity list -->
      <ScrollArea v-else class="h-[380px]" @scrollCapture="onScroll">
        <div class="divide-y">
          <NuxtLink
            v-for="activity in activities"
            :key="activity.id"
            :to="`/projects/${activity.project_key}/tasks/${activity.task_number}`"
            class="flex gap-3 px-4 py-3 transition-colors hover:bg-muted/50"
            @click="open = false"
          >
            <div class="flex size-6 shrink-0 items-center justify-center rounded-full border bg-background">
              <component
                :is="iconMap[activity.activity_type] || Circle"
                class="size-3 text-muted-foreground"
              />
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-xs">
                <span class="font-medium text-foreground">{{ activity.first_name }} {{ activity.last_name }}</span>
                <span class="text-muted-foreground">
                  {{ ' ' }}{{ ACTIVITY_TYPE_LABELS[activity.activity_type] || activity.activity_type }}
                  <template v-if="activity.field_name">
                    <span class="font-medium text-foreground">{{ activity.field_name }}</span>
                  </template>
                </span>
              </p>
              <div class="mt-0.5 flex items-center gap-1.5">
                <span class="text-xs font-medium text-amber-600 dark:text-amber-500">
                  {{ activity.project_key }}-{{ activity.task_number }}
                </span>
                <span class="truncate text-xs text-muted-foreground">
                  {{ activity.task_title }}
                </span>
              </div>
              <p class="mt-0.5 text-[11px] text-muted-foreground/60">
                {{ formatRelativeDate(activity.created_at) }}
              </p>
            </div>
          </NuxtLink>
        </div>
        <!-- Loading more indicator -->
        <div v-if="loadingMore" class="flex items-center justify-center py-3">
          <Loader2 class="size-4 animate-spin text-muted-foreground" />
        </div>
      </ScrollArea>
    </PopoverContent>
  </Popover>
</template>
