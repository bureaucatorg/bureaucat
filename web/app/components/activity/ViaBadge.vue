<script setup lang="ts">
import { Bot, KeyRound } from "lucide-vue-next";

defineProps<{ via?: "web" | "assistant" | "pat" }>();
</script>

<template>
  <span
    v-if="via === 'assistant' || via === 'pat'"
    class="inline-flex translate-y-0.5 items-center text-muted-foreground"
    :title="via === 'assistant' ? 'Done via the AI assistant' : 'Done via an API token'"
    :aria-label="via === 'assistant' ? 'via AI assistant' : 'via API token'"
  >
    <Bot v-if="via === 'assistant'" class="size-3.5" />
    <KeyRound v-else class="size-3.5" />
  </span>
</template>
