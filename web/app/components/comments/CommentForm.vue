<script setup lang="ts">
import { Send, Loader2, Paperclip, X, FileText } from "lucide-vue-next";
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
const { uploadFiles, uploading } = useFileAttach();
const { attachFile } = useAttachments();
const { user } = useAuth();

const content = ref("");
const loading = ref(false);
const dropZoneRef = ref<InstanceType<typeof FileDropZone> | null>(null);

// Tiptap emits `<p></p>` (and variants) for a visually-empty editor.
// Strip tags and whitespace to test for genuine emptiness.
const isContentEmpty = computed(() => {
  const text = content.value.replace(/<[^>]*>/g, "").replace(/&nbsp;/g, " ").trim();
  return text.length === 0;
});

// Pending uploads (not yet attached to a comment)
const pendingUploads = ref<{ uploadId: string; filename: string; mimeType: string }[]>([]);

async function handleFilesDropped(files: File[]) {
  const results = await uploadFiles(files);
  for (const r of results) {
    pendingUploads.value.push({
      uploadId: r.uploadId,
      filename: r.filename,
      mimeType: r.mimeType,
    });
  }
  if (results.length > 0) {
    toast.success(`${results.length} file${results.length > 1 ? "s" : ""} uploaded`);
  }
}

function removePendingUpload(index: number) {
  pendingUploads.value.splice(index, 1);
}

function handlePaste(event: ClipboardEvent) {
  const files = Array.from(event.clipboardData?.files || []);
  if (files.length > 0) {
    event.preventDefault();
    handleFilesDropped(files);
  }
}

function isImage(mimeType: string): boolean {
  return mimeType.startsWith("image/");
}

async function handleSubmit() {
  if (isContentEmpty.value && pendingUploads.value.length === 0) return;

  loading.value = true;
  const result = await createComment(props.projectKey, props.taskNum, {
    content: content.value || "(attachment)",
  });

  if (result.success && result.data) {
    // Attach pending files to the new comment
    for (const upload of pendingUploads.value) {
      await attachFile(
        props.projectKey,
        props.taskNum,
        "comment",
        upload.uploadId,
        result.data.id
      );
    }

    content.value = "";
    pendingUploads.value = [];
    emit("created");
  } else {
    toast.error(result.error || "Failed to add comment");
  }
  loading.value = false;
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
      <AvatarFallback class="text-xs" :seed="user?.id">
        {{ user?.first_name?.[0] }}{{ user?.last_name?.[0] }}
      </AvatarFallback>
    </Avatar>

    <FileDropZone
      ref="dropZoneRef"
      class="flex-1"
      :disabled="loading"
      :uploading="uploading"
      :show-button="false"
      accept="*/*"
      @files-dropped="handleFilesDropped"
    >
      <form class="space-y-2" @submit.prevent="handleSubmit" @paste="handlePaste" @keydown="handleKeyDown">
        <TiptapEditor
          v-model="content"
          :disabled="loading || uploading"
          :uploading="uploading"
          :members="members"
          compact
          @files-dropped="handleFilesDropped"
        />

        <!-- Pending uploads -->
        <div v-if="pendingUploads.length > 0" class="flex flex-wrap gap-1.5">
          <div
            v-for="(upload, index) in pendingUploads"
            :key="upload.uploadId"
            class="flex items-center gap-1.5 rounded-md border bg-muted/30 px-2 py-1 text-xs"
          >
            <FileText v-if="!isImage(upload.mimeType)" class="size-3 text-muted-foreground" />
            <span class="max-w-[120px] truncate">{{ upload.filename }}</span>
            <button
              type="button"
              class="text-muted-foreground hover:text-destructive"
              @click="removePendingUpload(index)"
            >
              <X class="size-3" />
            </button>
          </div>
        </div>

        <div class="flex items-center justify-between">
          <p class="text-xs text-muted-foreground">
            <kbd class="rounded border px-1 py-0.5 text-[10px]">
              {{ navigator?.platform?.includes("Mac") ? "⌘" : "Ctrl" }}
            </kbd>
            +
            <kbd class="rounded border px-1 py-0.5 text-[10px]">Enter</kbd>
            to submit
          </p>
          <div class="flex items-center gap-1">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              :disabled="loading || uploading"
              aria-label="Attach file"
              @click="dropZoneRef?.openFilePicker()"
            >
              <Loader2 v-if="uploading" class="size-3.5 animate-spin" />
              <Paperclip v-else class="size-3.5" />
            </Button>
            <Button type="submit" size="sm" :disabled="loading || uploading || (isContentEmpty && pendingUploads.length === 0)">
              <Loader2 v-if="loading" class="mr-1.5 size-3.5 animate-spin" />
              <Send v-else class="mr-1.5 size-3.5" />
              Comment
            </Button>
          </div>
        </div>
      </form>
    </FileDropZone>
  </div>
</template>
