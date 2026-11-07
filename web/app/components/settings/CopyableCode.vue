<script setup lang="ts">
import { Check, Copy } from "lucide-vue-next";

const props = defineProps<{ code: string; maxHeight?: string }>();

const copied = ref(false);

async function copy() {
  try {
    await navigator.clipboard.writeText(props.code);
    copied.value = true;
    setTimeout(() => (copied.value = false), 2000);
  } catch {
    // clipboard unavailable; the snippet is still selectable
  }
}
</script>

<template>
  <div class="relative">
    <pre
      class="overflow-auto rounded-md border bg-muted/50 p-3 pr-12 font-mono text-xs leading-relaxed"
      :style="maxHeight ? { maxHeight } : undefined"
    >{{ code }}</pre>
    <Button
      variant="ghost"
      size="icon"
      class="absolute right-1.5 top-1.5 size-7"
      :aria-label="copied ? 'Copied' : 'Copy snippet'"
      @click="copy"
    >
      <Check v-if="copied" class="size-3.5 text-green-500" />
      <Copy v-else class="size-3.5" />
    </Button>
  </div>
</template>
