<script setup lang="ts">
import { MessageSquare, Loader2 } from "lucide-vue-next";
import type { Comment } from "~/types";

const props = defineProps<{
  comments: Comment[];
  projectKey: string;
  taskNum: number;
  loading?: boolean;
  isMember: boolean;
}>();

const emit = defineEmits<{
  refresh: [];
}>();

const { user } = useAuth();

function canEdit(comment: Comment): boolean {
  return props.isMember && comment.created_by === user.value?.id;
}
</script>

<template>
  <div class="space-y-4">
    <h3 class="flex items-center gap-2 font-semibold">
      <MessageSquare class="size-4" />
      Comments
      <span class="text-sm font-normal text-muted-foreground">
        ({{ comments.length }})
      </span>
    </h3>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-8">
      <Loader2 class="size-5 animate-spin text-muted-foreground" />
    </div>

    <!-- Comments -->
    <div v-else-if="comments.length > 0" class="space-y-4">
      <CommentItem
        v-for="comment in comments"
        :key="comment.id"
        :comment="comment"
        :project-key="projectKey"
        :task-num="taskNum"
        :can-edit="canEdit(comment)"
        @deleted="emit('refresh')"
        @updated="emit('refresh')"
      />
    </div>

    <!-- Empty state -->
    <div
      v-else
      class="rounded-lg border border-dashed py-8 text-center text-sm text-muted-foreground"
    >
      No comments yet
    </div>

    <!-- Add comment form -->
    <CommentForm
      v-if="isMember"
      :project-key="projectKey"
      :task-num="taskNum"
      @created="emit('refresh')"
    />
  </div>
</template>
