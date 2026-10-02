<script setup lang="ts">
import { Handle, Position } from "@vue-flow/core";
import { Avatar, AvatarFallback, AvatarImage } from "~/components/ui/avatar";
import type { GraphUser } from "~/composables/useTaskGraph";

defineProps<{ data: GraphUser }>();

const centerHandle = { top: "50%", left: "50%", transform: "translate(-50%, -50%)", opacity: 0 };
</script>

<template>
  <div class="cursor-pointer">
    <Avatar class="size-12 shadow-md ring-2 ring-background">
      <AvatarImage v-if="data.avatar_url" :src="data.avatar_url" />
      <AvatarFallback class="text-sm font-medium" :seed="data.id">
        {{ data.first_name[0] }}{{ data.last_name[0] }}
      </AvatarFallback>
    </Avatar>
    <Handle type="source" :position="Position.Top" :connectable="false" :style="centerHandle" />
  </div>
</template>
