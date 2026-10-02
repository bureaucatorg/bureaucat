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
    <DialogContent class="w-auto max-w-[95vw] sm:max-w-[95vw] max-h-[95vh] border-none bg-transparent p-0 shadow-none">
      <DialogTitle class="sr-only">{{ alt || 'Image preview' }}</DialogTitle>
      <img
        :src="src"
        :alt="alt || 'Image preview'"
        class="h-auto max-h-[90vh] w-auto max-w-[95vw] rounded-lg object-contain"
      />
    </DialogContent>
  </Dialog>
</template>
