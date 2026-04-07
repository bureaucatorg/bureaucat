<script setup lang="ts">
import { FileText, Image as ImageIcon, X, Loader2, Download } from "lucide-vue-next";
import type { Attachment } from "~/composables/useAttachments";

const props = withDefaults(
  defineProps<{
    attachments: Attachment[];
    canDelete?: boolean;
    loading?: boolean;
  }>(),
  {
    canDelete: false,
    loading: false,
  }
);

const emit = defineEmits<{
  delete: [attachmentId: string];
}>();

const lightboxOpen = ref(false);
const lightboxSrc = ref("");
const lightboxAlt = ref("");

function isImage(mimeType: string): boolean {
  return mimeType.startsWith("image/");
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function handleClick(attachment: Attachment) {
  if (isImage(attachment.mime_type)) {
    lightboxSrc.value = attachment.url;
    lightboxAlt.value = attachment.filename;
    lightboxOpen.value = true;
  } else {
    window.open(attachment.url, "_blank");
  }
}
</script>

<template>
  <div v-if="loading" class="flex items-center gap-2 py-2">
    <Loader2 class="size-4 animate-spin text-muted-foreground" />
    <span class="text-xs text-muted-foreground">Loading attachments...</span>
  </div>

  <div v-else-if="attachments.length > 0" class="flex flex-wrap gap-1.5 pt-2">
    <div
      v-for="attachment in attachments"
      :key="attachment.id"
      class="group flex items-center gap-1.5 rounded-full border bg-muted/30 py-1 pl-2 pr-1.5 text-xs transition-colors hover:bg-muted/60"
    >
      <button
        type="button"
        class="flex items-center gap-1.5"
        @click="handleClick(attachment)"
      >
        <ImageIcon v-if="isImage(attachment.mime_type)" class="size-3.5 shrink-0 text-muted-foreground" />
        <FileText v-else class="size-3.5 shrink-0 text-muted-foreground" />
        <span class="max-w-[150px] truncate">{{ attachment.filename }}</span>
        <span class="text-muted-foreground/60">{{ formatSize(attachment.size_bytes) }}</span>
      </button>

      <a
        :href="attachment.url"
        :download="attachment.filename"
        class="rounded-full p-0.5 text-muted-foreground/50 hover:text-foreground"
        aria-label="Download"
        @click.stop
      >
        <Download class="size-3" />
      </a>

      <button
        v-if="canDelete"
        type="button"
        class="rounded-full p-0.5 text-muted-foreground/50 hover:bg-destructive/10 hover:text-destructive"
        aria-label="Remove attachment"
        @click.stop="emit('delete', attachment.id)"
      >
        <X class="size-3" />
      </button>
    </div>
  </div>

  <ImageLightbox v-model:open="lightboxOpen" :src="lightboxSrc" :alt="lightboxAlt" />
</template>
