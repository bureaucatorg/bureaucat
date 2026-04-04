<script setup lang="ts">
import { Send, Loader2 } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { ProjectMember } from "~/types";

const props = defineProps<{
  projectKey: string;
  taskNum: number;
  members: ProjectMember[];
}>();

const emit = defineEmits<{
  created: [];
}>();

const { createComment } = useComments();
const { user } = useAuth();

const content = ref("");
const loading = ref(false);
const mentionTextareaRef = ref<InstanceType<typeof MentionTextarea> | null>(null);

async function handleSubmit() {
  if (!content.value.trim()) return;

  // Convert @Name display text to markdown links before sending
  const markdownContent = mentionTextareaRef.value?.getMarkdownContent() ?? content.value;

  loading.value = true;
  const result = await createComment(props.projectKey, props.taskNum, {
    content: markdownContent,
  });
  loading.value = false;

  if (result.success) {
    content.value = "";
    mentionTextareaRef.value?.clearMentions();
    emit("created");
  } else {
    toast.error(result.error || "Failed to add comment");
  }
}

function handleKeyDown(event: KeyboardEvent) {
  if (event.key === "Enter" && (event.metaKey || event.ctrlKey)) {
    event.preventDefault();
    handleSubmit();
  }
}
</script>

<template>
  <div class="flex gap-3">
    <Avatar class="size-8">
      <AvatarImage
        v-if="user?.avatar_url"
        :src="user.avatar_url"
      />
      <AvatarFallback class="text-xs">
        {{ user?.first_name?.[0] }}{{ user?.last_name?.[0] }}
      </AvatarFallback>
    </Avatar>

    <form class="flex-1 space-y-2" @submit.prevent="handleSubmit">
      <MentionTextarea
        ref="mentionTextareaRef"
        v-model="content"
        placeholder="Add a comment..."
        :rows="2"
        :disabled="loading"
        :members="members"
        @keydown="handleKeyDown"
      />
      <div class="flex items-center justify-between">
        <p class="text-xs text-muted-foreground">
          <kbd class="rounded border px-1 py-0.5 text-[10px]">
            {{ navigator?.platform?.includes("Mac") ? "⌘" : "Ctrl" }}
          </kbd>
          +
          <kbd class="rounded border px-1 py-0.5 text-[10px]">Enter</kbd>
          to submit
        </p>
        <Button type="submit" size="sm" :disabled="loading || !content.trim()">
          <Loader2 v-if="loading" class="mr-1.5 size-3.5 animate-spin" />
          <Send v-else class="mr-1.5 size-3.5" />
          Comment
        </Button>
      </div>
    </form>
  </div>
</template>
