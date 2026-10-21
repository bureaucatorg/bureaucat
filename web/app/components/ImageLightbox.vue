<script setup lang="ts">
import { ChevronLeft, ChevronRight } from "lucide-vue-next";

const props = defineProps<{
  open: boolean;
  images: { src: string; alt?: string }[];
  index: number;
}>();

const emit = defineEmits<{
  "update:open": [value: boolean];
  "update:index": [value: number];
}>();

const current = computed(() => props.images[props.index]);
const hasMany = computed(() => props.images.length > 1);

function step(delta: number) {
  if (!hasMany.value) return;
  const n = props.images.length;
  emit("update:index", (props.index + delta + n) % n);
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === "ArrowLeft") {
    e.preventDefault();
    step(-1);
  } else if (e.key === "ArrowRight") {
    e.preventDefault();
    step(1);
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent
      class="w-auto max-w-[95vw] sm:max-w-[95vw] max-h-[95vh] border-none bg-transparent p-0 shadow-none"
      @keydown="onKeydown"
    >
      <DialogTitle class="sr-only">{{ current?.alt || 'Image preview' }}</DialogTitle>
      <div v-if="current" class="flex flex-col items-center gap-2">
        <div class="flex items-center gap-2 sm:gap-3">
          <button
            v-if="hasMany"
            type="button"
            class="shrink-0 rounded-full bg-background/80 p-1.5 text-muted-foreground shadow-sm backdrop-blur hover:text-foreground"
            aria-label="Previous image"
            @click="step(-1)"
          >
            <ChevronLeft class="size-5" />
          </button>
          <img
            :src="current.src"
            :alt="current.alt || 'Image preview'"
            class="h-auto w-auto rounded-lg object-contain"
            :class="hasMany ? 'max-h-[calc(90vh-2.5rem)] max-w-[calc(95vw-6rem)]' : 'max-h-[90vh] max-w-[95vw]'"
          />
          <button
            v-if="hasMany"
            type="button"
            class="shrink-0 rounded-full bg-background/80 p-1.5 text-muted-foreground shadow-sm backdrop-blur hover:text-foreground"
            aria-label="Next image"
            @click="step(1)"
          >
            <ChevronRight class="size-5" />
          </button>
        </div>
        <div
          v-if="hasMany"
          class="rounded-full bg-background/80 px-2 py-0.5 text-xs text-muted-foreground shadow-sm backdrop-blur"
        >
          {{ index + 1 }} / {{ images.length }}
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>
