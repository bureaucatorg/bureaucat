<script setup lang="ts">
definePageMeta({ middleware: ["admin", "graph-defaults"] });

useSeoMeta({ title: "Graph View" });
</script>

<template>
  <TaskGraphView api-base="/api/v1/admin/graph" />
</template>
