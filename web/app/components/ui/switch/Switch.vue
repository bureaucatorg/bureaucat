<script setup lang="ts">
const props = defineProps<{
  checked?: boolean;
  disabled?: boolean;
}>();

const emit = defineEmits<{
  "update:checked": [value: boolean];
}>();

function toggle() {
  if (!props.disabled) {
    emit("update:checked", !props.checked);
  }
}
</script>

<template>
  <button
    type="button"
    role="switch"
    :aria-checked="checked"
    :disabled="disabled"
    class="relative inline-flex h-6 w-11 shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:cursor-not-allowed disabled:opacity-50"
    :class="checked ? 'bg-primary' : 'bg-input'"
    @click="toggle"
  >
    <span
      class="pointer-events-none block size-5 rounded-full bg-background shadow-lg ring-0 transition-transform"
      :class="checked ? 'translate-x-5' : 'translate-x-0'"
    />
  </button>
</template>
