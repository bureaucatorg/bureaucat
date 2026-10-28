<script setup lang="ts">
import {
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
  Link2,
  Link2Off,
} from "lucide-vue-next";
import type { ActivityType, NotificationEntry } from "~/types";
import { ACTIVITY_TYPE_LABELS, activityActorName } from "~/types";

const props = defineProps<{
  activity: NotificationEntry;
}>();

const emit = defineEmits<{
  click: [activity: NotificationEntry];
}>();

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
  blocker_added: Link2,
  blocker_removed: Link2Off,
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

const link = computed(() => {
  const base = `/projects/${props.activity.project_key}/tasks/${props.activity.task_number}`;
  // Deep-link to the specific comment so the task page scrolls to and highlights it.
  return props.activity.comment_id ? `${base}#comment-${props.activity.comment_id}` : base;
});
</script>

<template>
  <NuxtLink
    :to="link"
    class="flex gap-3 px-4 py-3 transition-colors hover:bg-muted/50"
    :class="!activity.is_read && 'bg-muted/30'"
    @click="emit('click', activity)"
  >
    <div class="flex size-6 shrink-0 items-center justify-center rounded-full border bg-background">
      <component
        :is="iconMap[activity.activity_type] || Circle"
        class="size-3 text-muted-foreground"
      />
    </div>
    <div class="min-w-0 flex-1">
      <p class="text-xs" :class="!activity.is_read && 'font-medium'">
        <span class="font-medium text-foreground">{{ activityActorName(activity) }}</span>
        <span :class="activity.is_read ? 'text-muted-foreground' : 'text-foreground'">
          {{ ' ' }}{{ ACTIVITY_TYPE_LABELS[activity.activity_type] || activity.activity_type }}
          <template v-if="activity.field_name">
            <span class="font-medium text-foreground">{{ activity.field_name }}</span>
          </template>
        </span>
        <span
          v-if="activity.event_count > 1"
          class="text-muted-foreground"
        >
          {{ ' ' }}· {{ activity.event_count }} updates
        </span>
      </p>
      <div class="mt-0.5 flex items-center gap-1.5">
        <span class="shrink-0 text-xs font-medium text-amber-600 dark:text-amber-500">
          {{ activity.project_key }}-{{ activity.task_number }}
        </span>
        <span class="truncate text-xs text-muted-foreground">
          {{ activity.task_title }}
        </span>
      </div>
      <p class="mt-0.5 text-[11px] text-muted-foreground/60">
        {{ formatRelativeDate(activity.updated_at || activity.created_at) }}
      </p>
    </div>
    <span
      v-if="!activity.is_read"
      class="mt-1 size-2 shrink-0 rounded-full bg-amber-500"
      aria-hidden="true"
    />
  </NuxtLink>
</template>
