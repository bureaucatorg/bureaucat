<script setup lang="ts">
import {
  History,
  Loader2,
  ShieldCheck,
  ShieldAlert,
  Plus,
  Edit2,
  Trash2,
  UserPlus,
  UserMinus,
  Tag,
  Tags,
  ArrowRight,
  MessageSquare,
  Circle,
} from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { ActivityLogEntry, ActivityType, Comment } from "~/types";
import { ACTIVITY_TYPE_LABELS } from "~/types";

const props = defineProps<{
  activities: ActivityLogEntry[];
  comments: Comment[];
  projectKey: string;
  taskNum: number;
  activitiesLoading?: boolean;
  commentsLoading?: boolean;
  isMember: boolean;
}>();

const emit = defineEmits<{
  refreshComments: [];
  refreshActivity: [];
}>();

const { user } = useAuth();
const { verifyActivity } = useActivity();

const verifying = ref(false);
const verificationResult = ref<{ valid: boolean; message: string } | null>(null);

// Icon map for activity types
const iconMap: Record<ActivityType, typeof Plus> = {
  task_created: Plus,
  task_updated: Edit2,
  task_deleted: Trash2,
  assignee_added: UserPlus,
  assignee_removed: UserMinus,
  label_added: Tag,
  label_removed: Tags,
  state_changed: ArrowRight,
  comment_created: MessageSquare,
  comment_updated: Edit2,
  comment_deleted: Trash2,
};

type FeedItem =
  | { type: "activity"; data: ActivityLogEntry; timestamp: Date }
  | { type: "comment"; data: Comment; timestamp: Date };

// Merge and sort activities and comments by timestamp
const feedItems = computed<FeedItem[]>(() => {
  const items: FeedItem[] = [];

  // Add activities (except comment-related ones as we show actual comments)
  for (const activity of props.activities) {
    // Skip comment activities - we show the actual comments instead
    if (
      activity.activity_type === "comment_created" ||
      activity.activity_type === "comment_updated" ||
      activity.activity_type === "comment_deleted"
    ) {
      continue;
    }
    items.push({
      type: "activity",
      data: activity,
      timestamp: new Date(activity.created_at),
    });
  }

  // Add comments
  for (const comment of props.comments) {
    items.push({
      type: "comment",
      data: comment,
      timestamp: new Date(comment.created_at),
    });
  }

  // Sort by timestamp descending (newest first)
  items.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());

  return items;
});

const loading = computed(
  () => props.activitiesLoading || props.commentsLoading
);

function canEditComment(comment: Comment): boolean {
  return props.isMember && comment.created_by === user.value?.id;
}

async function handleVerify() {
  verifying.value = true;
  verificationResult.value = null;

  const result = await verifyActivity(props.projectKey, props.taskNum);

  verifying.value = false;

  if (result.success && result.data) {
    verificationResult.value = result.data;
    if (result.data.valid) {
      toast.success("Activity log verified successfully");
    } else {
      toast.error("Activity log integrity compromised");
    }
  } else {
    toast.error(result.error || "Failed to verify activity");
  }
}

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

function getActivityLabel(activity: ActivityLogEntry): string {
  return ACTIVITY_TYPE_LABELS[activity.activity_type] || activity.activity_type;
}

function getFieldLabel(fieldName?: string): string | null {
  if (!fieldName) return null;

  const labels: Record<string, string> = {
    title: "title",
    description: "description",
    state: "state",
    priority: "priority",
    assignees: "assignees",
    labels: "labels",
    content: "content",
  };
  return labels[fieldName] || fieldName;
}

// Parse activity value to extract meaningful details
function parseActivityValue(value: unknown): Record<string, unknown> | null {
  if (!value) return null;

  // If it's a string (JSON), try to parse it
  if (typeof value === "string") {
    try {
      return JSON.parse(value);
    } catch {
      return null;
    }
  }

  // If it's already an object
  if (typeof value === "object") {
    return value as Record<string, unknown>;
  }

  return null;
}

// Get display text for activity details (assignee name, label name, etc.)
function getActivityDetail(activity: ActivityLogEntry): string | null {
  const type = activity.activity_type;

  if (type === "assignee_added") {
    const data = parseActivityValue(activity.new_value);
    if (data?.first_name && data?.last_name) {
      return `${data.first_name} ${data.last_name}`;
    }
  }

  if (type === "assignee_removed") {
    const data = parseActivityValue(activity.old_value);
    if (data?.first_name && data?.last_name) {
      return `${data.first_name} ${data.last_name}`;
    }
  }

  if (type === "label_added") {
    const data = parseActivityValue(activity.new_value);
    if (data?.name) {
      return data.name as string;
    }
  }

  if (type === "label_removed") {
    const data = parseActivityValue(activity.old_value);
    if (data?.name) {
      return data.name as string;
    }
  }

  return null;
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h3 class="flex items-center gap-2 font-semibold">
        <History class="size-4" />
        Activity
      </h3>
      <Button
        variant="outline"
        size="sm"
        :disabled="verifying"
        @click="handleVerify"
      >
        <Loader2 v-if="verifying" class="mr-1.5 size-3.5 animate-spin" />
        <ShieldCheck v-else class="mr-1.5 size-3.5" />
        Verify Log
      </Button>
    </div>

    <!-- Verification result -->
    <Alert
      v-if="verificationResult"
      :variant="verificationResult.valid ? 'success' : 'destructive'"
    >
      <component
        :is="verificationResult.valid ? ShieldCheck : ShieldAlert"
        class="size-4"
      />
      <AlertTitle>
        {{ verificationResult.valid ? "Verified" : "Integrity Issue" }}
      </AlertTitle>
      <AlertDescription>
        {{ verificationResult.message }}
      </AlertDescription>
    </Alert>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-8">
      <Loader2 class="size-5 animate-spin text-muted-foreground" />
    </div>

    <!-- Feed -->
    <div v-else-if="feedItems.length > 0" class="space-y-0">
      <div
        v-for="(item, index) in feedItems"
        :key="item.type === 'comment' ? `c-${item.data.id}` : `a-${item.data.id}`"
        class="relative flex gap-3 pb-4"
      >
        <!-- Timeline line -->
        <div
          v-if="index < feedItems.length - 1"
          class="absolute left-[11px] top-6 h-full w-px bg-border"
        />

        <!-- Activity item -->
        <template v-if="item.type === 'activity'">
          <div
            class="relative z-10 flex size-6 shrink-0 items-center justify-center rounded-full border bg-background"
          >
            <component
              :is="iconMap[item.data.activity_type] || Circle"
              class="size-3 text-muted-foreground"
            />
          </div>
          <div class="min-w-0 flex-1 pt-0.5">
            <p class="text-sm">
              <span class="font-medium">
                {{ item.data.first_name }} {{ item.data.last_name }}
              </span>
              <span class="text-muted-foreground">
                {{ " " }}{{ getActivityLabel(item.data) }}
                <template v-if="getActivityDetail(item.data)">
                  <span class="font-medium">{{ getActivityDetail(item.data) }}</span>
                </template>
                <template v-else-if="getFieldLabel(item.data.field_name)">
                  <span class="font-medium">{{
                    getFieldLabel(item.data.field_name)
                  }}</span>
                </template>
              </span>
            </p>
            <p class="mt-0.5 text-xs text-muted-foreground">
              {{ formatDate(item.data.created_at) }}
            </p>
          </div>
        </template>

        <!-- Comment item -->
        <template v-else>
          <div
            class="relative z-10 flex size-6 shrink-0 items-center justify-center rounded-full border bg-primary text-primary-foreground"
          >
            <MessageSquare class="size-3" />
          </div>
          <div class="min-w-0 flex-1">
            <CommentItem
              :comment="item.data"
              :project-key="projectKey"
              :task-num="taskNum"
              :can-edit="canEditComment(item.data)"
              compact
              @deleted="emit('refreshComments')"
              @updated="emit('refreshComments')"
            />
          </div>
        </template>
      </div>
    </div>

    <!-- Empty state -->
    <div
      v-else
      class="rounded-lg border border-dashed py-8 text-center text-sm text-muted-foreground"
    >
      No activity yet
    </div>
  </div>
</template>
