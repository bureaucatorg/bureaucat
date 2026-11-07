<script setup lang="ts">
import { Bot, KeyRound, Plug, Workflow } from "lucide-vue-next";
import type { ActivityVia } from "~/types";

const props = defineProps<{ via?: ActivityVia; triggeredBy?: string }>();

const title = computed(() => {
  if (props.via === "assistant") return "Done via the AI assistant";
  if (props.via === "pat") return "Done via an API token";
  if (props.via === "mcp") return "Done via an MCP client";
  if (props.via === "workflow") {
    return props.triggeredBy ? `Done by a workflow rule, triggered by ${props.triggeredBy}` : "Done by a workflow rule";
  }
  return "";
});
</script>

<template>
  <span
    v-if="title"
    class="inline-flex translate-y-0.5 items-center text-muted-foreground"
    :title="title"
    :aria-label="title"
  >
    <Bot v-if="via === 'assistant'" class="size-3.5" />
    <Workflow v-else-if="via === 'workflow'" class="size-3.5" />
    <Plug v-else-if="via === 'mcp'" class="size-3.5" />
    <KeyRound v-else class="size-3.5" />
  </span>
</template>
