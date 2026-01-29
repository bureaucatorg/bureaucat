<script setup lang="ts">
import { MoreHorizontal, Pencil, Trash2, Loader2, Check, X } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { Comment } from "~/types";

const props = defineProps<{
  comment: Comment;
  projectKey: string;
  taskNum: number;
  canEdit: boolean;
}>();

const emit = defineEmits<{
  deleted: [];
  updated: [];
}>();

const { updateComment, deleteComment } = useComments();

const editing = ref(false);
const editContent = ref("");
const loading = ref(false);

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
    <Avatar class="size-8">
      <AvatarFallback class="text-xs">
        {{ comment.first_name[0] }}{{ comment.last_name[0] }}
      </AvatarFallback>
    </Avatar>

    <div class="flex-1 space-y-1">
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
        <p class="whitespace-pre-wrap text-sm">{{ comment.content }}</p>

        <!-- Actions -->
        <div
          v-if="canEdit"
          class="opacity-0 transition-opacity group-hover:opacity-100"
        >
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <Button variant="ghost" size="sm" class="h-7 px-2">
                <MoreHorizontal class="size-3.5" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start">
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
      </template>
    </div>
  </div>
</template>
