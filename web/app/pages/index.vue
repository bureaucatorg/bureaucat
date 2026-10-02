<script setup lang="ts">
// The marketing page lives in /landing; signed-in users are bounced on to /dashboard by the guest middleware.
definePageMeta({
  middleware: () => navigateTo("/signin", { replace: true }),
});
</script>

<template>
  <div />
</template>
