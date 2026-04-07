<script setup lang="ts">
const props = defineProps<{
  open: boolean;
  src: string;
  alt?: string;
}>();

const emit = defineEmits<{
  "update:open": [value: boolean];
}>();
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="max-w-[90vw] max-h-[90vh] border-none bg-transparent p-0 shadow-none">
      <DialogTitle class="sr-only">{{ alt || 'Image preview' }}</DialogTitle>
      <img
        :src="src"
        :alt="alt || 'Image preview'"
        class="max-h-[85vh] w-full rounded-lg object-contain"
      />
    </DialogContent>
  </Dialog>
</template>
