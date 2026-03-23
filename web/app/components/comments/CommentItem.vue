<script setup lang="ts">
import { MoreHorizontal, Pencil, Trash2, Loader2, Check, X, ChevronDown, ChevronUp, History } from "lucide-vue-next";
import { marked } from "marked";
import { toast } from "vue-sonner";
import type { Comment } from "~/types";

const renderer = new marked.Renderer();
renderer.link = ({ href, title, text }) => {
  const titleAttr = title ? ` title="${title}"` : "";
  return `<a href="${href}"${titleAttr} target="_blank" rel="noopener noreferrer">${text}</a>`;
};
marked.setOptions({ breaks: true, gfm: true, renderer });

interface CommentVersion {
  content: string;
  version: number;
  editedAt: string;
  editedBy: string;
}

const props = withDefaults(
  defineProps<{
    comment: Comment;
    projectKey: string;
    taskNum: number;
    canEdit: boolean;
    compact?: boolean;
    editHistory?: CommentVersion[];
  }>(),
  {
    compact: false,
    editHistory: () => [],
  }
);

const emit = defineEmits<{
  deleted: [];
  updated: [];
}>();

const { updateComment, deleteComment } = useComments();

const editing = ref(false);
const editContent = ref("");
const loading = ref(false);
const showHistory = ref(false);

const renderedContent = computed(() => {
  return marked(props.comment.content) as string;
});

function startEdit() {
  editing.value = true;
  editContent.value = props.comment.content;
}

function cancelEdit() {
  editing.value = false;
  editContent.value = "";
}

async function handleUpdate() {
  if (!editContent.value.trim()) return;

  loading.value = true;
  const result = await updateComment(
    props.projectKey,
    props.taskNum,
    props.comment.id,
    { content: editContent.value }
  );
  loading.value = false;

  if (result.success) {
    toast.success("Comment updated");
    editing.value = false;
    emit("updated");
  } else {
    toast.error(result.error || "Failed to update comment");
  }
}

async function handleDelete() {
  loading.value = true;
  const result = await deleteComment(props.projectKey, props.taskNum, props.comment.id);
  loading.value = false;

  if (result.success) {
    toast.success("Comment deleted");
    emit("deleted");
  } else {
    toast.error(result.error || "Failed to delete comment");
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
</script>

<template>
  <div class="group flex gap-3">
    <Avatar v-if="!compact" class="size-8">
      <AvatarFallback class="text-xs">
        {{ comment.first_name[0] }}{{ comment.last_name[0] }}
      </AvatarFallback>
    </Avatar>

    <div class="min-w-0 flex-1 space-y-1">
      <div class="flex items-center gap-2">
        <span class="text-sm font-medium">
          {{ comment.first_name }} {{ comment.last_name }}
        </span>
        <span class="text-xs text-muted-foreground">
          {{ formatDate(comment.created_at) }}
        </span>
        <span
          v-if="comment.version > 1"
          class="text-xs text-muted-foreground"
        >
          (edited)
        </span>
        <span class="flex-1" />
        <DropdownMenu v-if="canEdit && !editing">
          <DropdownMenuTrigger as-child>
            <Button
              variant="ghost"
              size="sm"
              class="h-6 w-6 p-0 opacity-0 transition-opacity group-hover:opacity-100"
            >
              <MoreHorizontal class="size-3.5" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem @click="startEdit">
              <Pencil class="mr-2 size-3.5" />
              Edit
            </DropdownMenuItem>
            <DropdownMenuItem
              class="text-destructive focus:text-destructive"
              @click="handleDelete"
            >
              <Trash2 class="mr-2 size-3.5" />
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      <!-- Editing -->
      <div v-if="editing" class="space-y-2">
        <Textarea
          v-model="editContent"
          rows="3"
          :disabled="loading"
        />
        <div class="flex gap-2">
          <Button size="sm" :disabled="loading || !editContent.trim()" @click="handleUpdate">
            <Loader2 v-if="loading" class="mr-1.5 size-3 animate-spin" />
            <Check v-else class="mr-1.5 size-3" />
            Save
          </Button>
          <Button size="sm" variant="outline" :disabled="loading" @click="cancelEdit">
            <X class="mr-1.5 size-3" />
            Cancel
          </Button>
        </div>
      </div>

      <!-- Display -->
      <template v-else>
        <div
          class="prose prose-sm max-w-none overflow-hidden break-words dark:prose-invert"
          v-html="renderedContent"
        />

        <!-- Edit history toggle -->
        <button
          v-if="editHistory.length > 0"
          type="button"
          class="mt-1 flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
          @click="showHistory = !showHistory"
        >
          <History class="size-3" />
          <span>{{ editHistory.length }} previous version{{ editHistory.length > 1 ? 's' : '' }}</span>
          <ChevronDown v-if="!showHistory" class="size-3" />
          <ChevronUp v-else class="size-3" />
        </button>

        <!-- Edit history -->
        <div
          v-if="showHistory && editHistory.length > 0"
          class="mt-2 space-y-2 border-l-2 border-muted pl-3"
        >
          <div
            v-for="version in editHistory"
            :key="version.version"
            class="text-sm"
          >
            <p class="text-xs text-muted-foreground">
              v{{ version.version }} - {{ formatDate(version.editedAt) }}
            </p>
            <div
              class="prose prose-sm max-w-none break-words text-muted-foreground dark:prose-invert"
              v-html="marked(version.content) as string"
            />
          </div>
        </div>

      </template>
    </div>
  </div>
</template>
