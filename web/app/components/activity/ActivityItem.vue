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
} from "lucide-vue-next";
import type { ActivityLogEntry, ActivityType } from "~/types";
import { ACTIVITY_TYPE_LABELS } from "~/types";

const props = defineProps<{
  activity: ActivityLogEntry;
  isLast: boolean;
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
};

const Icon = computed(() => iconMap[props.activity.activity_type] || Circle);
const label = computed(() => ACTIVITY_TYPE_LABELS[props.activity.activity_type] || props.activity.activity_type);

function formatDate(dateStr: string): string {
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

// Parse value changes
const fieldLabel = computed(() => {
  const field = props.activity.field_name;
  if (!field) return null;

  const labels: Record<string, string> = {
    title: "title",
    description: "description",
    state: "state",
    priority: "priority",
    assignees: "assignees",
    labels: "labels",
    content: "content",
  };
  return labels[field] || field;
});
</script>

<template>
  <div class="relative flex gap-3 pb-4">
    <!-- Timeline line -->
    <div
      v-if="!isLast"
      class="absolute left-[11px] top-6 h-full w-px bg-border"
    />

    <!-- Icon -->
    <div
      class="relative z-10 flex size-6 shrink-0 items-center justify-center rounded-full border bg-background"
    >
      <component :is="Icon" class="size-3 text-muted-foreground" />
    </div>

    <!-- Content -->
    <div class="min-w-0 flex-1 pt-0.5">
      <p class="text-sm">
        <span class="font-medium">
          {{ activity.first_name }} {{ activity.last_name }}
        </span>
        <span class="text-muted-foreground">
          {{ " " }}{{ label }}
          <template v-if="fieldLabel">
            <span class="font-medium">{{ fieldLabel }}</span>
          </template>
        </span>
      </p>
      <p class="mt-0.5 text-xs text-muted-foreground">
        {{ formatDate(activity.created_at) }}
      </p>
    </div>
  </div>
</template>
