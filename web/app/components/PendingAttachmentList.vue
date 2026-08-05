<script setup lang="ts">
import { Paperclip, X } from "lucide-vue-next";
import type { UploadedFile } from "~/composables/useFileAttach";

defineProps<{
  files: UploadedFile[];
  disabled?: boolean;
}>();

const emit = defineEmits<{ remove: [uploadId: string] }>();
</script>

<template>
  <div v-if="files.length" class="mt-2 flex flex-wrap gap-1.5">
    <span
      v-for="file in files"
      :key="file.uploadId"
      class="flex max-w-full items-center gap-1.5 rounded-md border bg-muted/50 py-1 pl-2 pr-1 text-xs"
    >
      <Paperclip class="size-3 shrink-0 text-muted-foreground" />
      <span class="truncate">{{ file.filename }}</span>
      <button
        type="button"
        :aria-label="`Remove ${file.filename}`"
        class="rounded p-0.5 text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
        :disabled="disabled"
        @click="emit('remove', file.uploadId)"
      >
        <X class="size-3" />
      </button>
    </span>
  </div>
</template>
