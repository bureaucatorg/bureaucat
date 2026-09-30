<script setup lang="ts">
import { Handle, Position } from "@vue-flow/core";
import { ListTodo, Mail } from "lucide-vue-next";
import { Avatar, AvatarFallback, AvatarImage } from "~/components/ui/avatar";
import { Tooltip, TooltipContent, TooltipTrigger } from "~/components/ui/tooltip";
import type { GraphUser } from "~/composables/useAdmin";

defineProps<{ data: GraphUser & { task_count: number } }>();

const centerHandle = { top: "50%", left: "50%", transform: "translate(-50%, -50%)", opacity: 0 };
</script>

<template>
  <Tooltip>
    <TooltipTrigger as-child>
      <div class="cursor-pointer">
        <Avatar class="size-12 shadow-md ring-2 ring-background">
          <AvatarImage v-if="data.avatar_url" :src="data.avatar_url" />
          <AvatarFallback class="text-sm font-medium" :seed="data.id">
            {{ data.first_name[0] }}{{ data.last_name[0] }}
          </AvatarFallback>
        </Avatar>
        <Handle type="source" :position="Position.Top" :connectable="false" :style="centerHandle" />
      </div>
    </TooltipTrigger>
    <TooltipContent
      side="top"
      :side-offset="10"
      class="w-72 rounded-lg border bg-popover p-0 text-sm text-popover-foreground shadow-lg [&_.rotate-45]:hidden"
    >
      <div class="flex items-center gap-3 p-3">
        <Avatar class="size-10">
          <AvatarImage v-if="data.avatar_url" :src="data.avatar_url" />
          <AvatarFallback class="text-sm font-medium" :seed="data.id">
            {{ data.first_name[0] }}{{ data.last_name[0] }}
          </AvatarFallback>
        </Avatar>
        <div class="min-w-0">
          <p class="truncate font-semibold">{{ data.first_name }} {{ data.last_name }}</p>
          <p class="truncate text-xs text-muted-foreground">@{{ data.username }}</p>
        </div>
      </div>
      <div class="space-y-1.5 border-t px-3 py-2.5 text-xs text-muted-foreground">
        <p class="flex items-center gap-2">
          <Mail class="size-3.5 shrink-0" />
          <span class="truncate">{{ data.email }}</span>
        </p>
        <p class="flex items-center gap-2">
          <ListTodo class="size-3.5 shrink-0" />
          {{ data.task_count }} {{ data.task_count === 1 ? "task" : "tasks" }} in view
        </p>
      </div>
    </TooltipContent>
  </Tooltip>
</template>
