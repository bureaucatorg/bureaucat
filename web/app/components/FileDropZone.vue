<script setup lang="ts">
import { Paperclip, Loader2 } from "lucide-vue-next";

const props = withDefaults(
  defineProps<{
    disabled?: boolean;
    accept?: string;
    showButton?: boolean;
    uploading?: boolean;
  }>(),
  {
    disabled: false,
    accept: "*/*",
    showButton: true,
    uploading: false,
  }
);

const emit = defineEmits<{
  "files-dropped": [files: File[]];
}>();

const dragging = ref(false);
const fileInputRef = ref<HTMLInputElement | null>(null);
let dragCounter = 0;

function handleDragEnter(e: DragEvent) {
  if (props.disabled) return;
  e.preventDefault();
  dragCounter++;
  dragging.value = true;
}

function handleDragLeave(e: DragEvent) {
  if (props.disabled) return;
  e.preventDefault();
  dragCounter--;
  if (dragCounter === 0) {
    dragging.value = false;
  }
}

function handleDragOver(e: DragEvent) {
  if (props.disabled) return;
  e.preventDefault();
}

function handleDrop(e: DragEvent) {
  if (props.disabled) return;
  e.preventDefault();
  dragCounter = 0;
  dragging.value = false;

  const files = Array.from(e.dataTransfer?.files || []);
  if (files.length > 0) {
    emit("files-dropped", files);
  }
}

function handleFileInput(e: Event) {
  const input = e.target as HTMLInputElement;
  const files = Array.from(input.files || []);
  if (files.length > 0) {
    emit("files-dropped", files);
  }
  // Reset so the same file can be selected again
  input.value = "";
}

function openFilePicker() {
  fileInputRef.value?.click();
}

defineExpose({ openFilePicker });
</script>

<template>
  <div
    class="relative"
    @dragenter="handleDragEnter"
    @dragleave="handleDragLeave"
    @dragover="handleDragOver"
    @drop="handleDrop"
  >
    <slot />

    <!-- Drag overlay -->
    <div
      v-if="dragging"
      class="pointer-events-none absolute inset-0 z-10 flex items-center justify-center rounded-md border-2 border-dashed border-primary/50 bg-primary/5"
    >
      <p class="text-sm font-medium text-primary">Drop files to attach</p>
    </div>

    <!-- Hidden file input -->
    <input
      ref="fileInputRef"
      type="file"
      multiple
      class="hidden"
      :accept="accept"
      :disabled="disabled"
      @change="handleFileInput"
    />

    <!-- Attach button (optional) -->
    <slot name="button" :open-file-picker="openFilePicker" :uploading="uploading">
      <Button
        v-if="showButton"
        type="button"
        variant="ghost"
        size="sm"
        :disabled="disabled || uploading"
        aria-label="Attach file"
        @click="openFilePicker"
      >
        <Loader2 v-if="uploading" class="mr-1.5 size-3.5 animate-spin" />
        <Paperclip v-else class="mr-1.5 size-3.5" />
        Attach
      </Button>
    </slot>
  </div>
</template>
