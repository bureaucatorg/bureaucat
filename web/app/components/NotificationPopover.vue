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
const activities = ref<UserActivityEntry[]>([]);
const loaded = ref(false);

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

async function loadActivity() {
  if (!user.value) return;
  loading.value = true;
  try {
    const response = await fetch(
      `/api/v1/users/${user.value.id}/activity?page=1&per_page=15`,
      { headers: getAuthHeader() }
    );
    if (response.ok) {
      const data = await response.json();
      activities.value = data.activities ?? [];
    }
  } catch {
    // silently fail
  } finally {
    loading.value = false;
    loaded.value = true;
  }
}

watch(open, (isOpen) => {
  if (isOpen) {
    loadActivity();
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
        <h3 class="text-sm font-semibold">Activity</h3>
      </div>

      <!-- Loading -->
      <div v-if="loading && !loaded" class="flex items-center justify-center py-8">
        <Loader2 class="size-5 animate-spin text-muted-foreground" />
      </div>

      <!-- Empty -->
      <div
        v-else-if="activities.length === 0"
        class="px-4 py-8 text-center text-sm text-muted-foreground"
      >
        No activity yet
      </div>

      <!-- Activity list -->
      <ScrollArea v-else class="h-[380px]">
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
      </ScrollArea>

      <!-- Footer -->
      <div v-if="user" class="border-t px-4 py-2 text-center">
        <NuxtLink
          :to="`/profile/${user.id}`"
          class="text-xs text-muted-foreground hover:text-foreground transition-colors"
          @click="open = false"
        >
          View all activity
        </NuxtLink>
      </div>
    </PopoverContent>
  </Popover>
</template>
