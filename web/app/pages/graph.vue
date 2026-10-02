<script setup lang="ts">
definePageMeta({ middleware: ["auth", "graph-defaults"] });

useSeoMeta({ title: "Graph View" });
</script>

<template>
  <TaskGraphView api-base="/api/v1/graph" />
</template>
