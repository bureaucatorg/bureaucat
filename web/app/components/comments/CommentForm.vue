<script setup lang="ts">
import { Send, Loader2 } from "lucide-vue-next";
import { toast } from "vue-sonner";

const props = defineProps<{
  projectKey: string;
  taskNum: number;
}>();

const emit = defineEmits<{
  created: [];
}>();

const { createComment } = useComments();
const { user } = useAuth();

const content = ref("");
const loading = ref(false);

async function handleSubmit() {
  if (!content.value.trim()) return;

  loading.value = true;
  const result = await createComment(props.projectKey, props.taskNum, {
    content: content.value,
  });
  loading.value = false;

  if (result.success) {
    content.value = "";
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
      <AvatarFallback class="text-xs">
        {{ user?.first_name?.[0] }}{{ user?.last_name?.[0] }}
      </AvatarFallback>
    </Avatar>

    <form class="flex-1 space-y-2" @submit.prevent="handleSubmit">
      <Textarea
        v-model="content"
        placeholder="Add a comment..."
        rows="2"
        :disabled="loading"
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
