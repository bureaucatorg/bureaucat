<script setup lang="ts">
import type { AvatarFallbackProps } from "reka-ui"
import type { HTMLAttributes } from "vue"
import { computed } from "vue"
import { reactiveOmit } from "@vueuse/core"
import { AvatarFallback } from "reka-ui"
import { cn } from "@/lib/utils"
import { getAvatarColor } from "@/utils/avatarColor"

const props = defineProps<
  AvatarFallbackProps & {
    class?: HTMLAttributes["class"];
    seed?: string | number | null;
  }
>()

const delegatedProps = reactiveOmit(props, "class", "seed")

const seedClass = computed(() =>
  props.seed ? getAvatarColor(props.seed) : "bg-muted"
)
</script>

<template>
  <AvatarFallback
    data-slot="avatar-fallback"
    v-bind="delegatedProps"
    :class="cn(seedClass, 'flex size-full items-center justify-center rounded-full', props.class)"
  >
    <slot />
  </AvatarFallback>
</template>
