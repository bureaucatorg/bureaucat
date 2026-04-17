<script setup lang="ts">
import { Sun, Moon, Monitor } from "lucide-vue-next";

const colorMode = useColorMode();

const modes = [
  { value: "light", label: "Light", icon: Sun },
  { value: "dark", label: "Dark", icon: Moon },
  { value: "system", label: "System", icon: Monitor },
] as const;

const currentIcon = computed(() => {
  const mode = modes.find((m) => m.value === colorMode.preference);
  return mode?.icon ?? Monitor;
});

function setMode(value: "light" | "dark" | "system") {
  colorMode.preference = value;
}
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <Button variant="ghost" size="icon" class="size-9" aria-label="Toggle theme">
        <component :is="currentIcon" class="size-4" />
        <span class="sr-only">Toggle theme</span>
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="end">
      <DropdownMenuItem
        v-for="mode in modes"
        :key="mode.value"
        @click="setMode(mode.value)"
      >
        <component :is="mode.icon" class="mr-2 size-4" />
        {{ mode.label }}
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
