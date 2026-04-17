<script setup lang="ts">
import { LayoutDashboard, FolderKanban, Eye, MessageCircle, Settings } from "lucide-vue-next";

const { user } = useAuth();
const route = useRoute();
const { feedbackPublic, fetchFeedbackPublicSettings } = useSettings();

const profilePath = computed(() => (user.value ? `/profile/${user.value.id}` : "/"));

function isActive(path: string): boolean {
  if (path === profilePath.value) {
    return route.path === profilePath.value;
  }
  return route.path === path || route.path.startsWith(`${path}/`);
}

const showFeedback = ref(false);

onMounted(() => {
  fetchFeedbackPublicSettings();
});
</script>

<template>
  <aside
    class="fixed inset-y-0 left-0 z-50 flex w-12 flex-col items-center border-r border-border/60 bg-muted/40 py-3 backdrop-blur-xl"
  >
    <!-- Profile -->
    <NuxtLink
      :to="profilePath"
      :title="user ? `${user.first_name} ${user.last_name}` : 'Profile'"
      class="group flex size-9 items-center justify-center rounded-md outline-none transition-colors focus-visible:ring-2 focus-visible:ring-ring"
      :class="isActive(profilePath) ? 'bg-amber-500/15' : 'hover:bg-muted'"
    >
      <Avatar class="size-7">
        <AvatarImage
          v-if="user?.avatar_url"
          :src="user.avatar_url"
          :alt="`${user.first_name} ${user.last_name}`"
        />
        <AvatarFallback class="text-[11px]" :seed="user?.id">
          {{ user?.first_name?.[0] || "" }}{{ user?.last_name?.[0] || "" }}
        </AvatarFallback>
      </Avatar>
    </NuxtLink>

    <!-- Main nav -->
    <nav class="mt-4 flex flex-col items-center gap-1">
      <NuxtLink
        to="/dashboard"
        title="Dashboard"
        class="flex size-9 items-center justify-center rounded-md text-muted-foreground outline-none transition-colors hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
        :class="isActive('/dashboard') && 'bg-amber-500/15 text-amber-700 dark:text-amber-400'"
      >
        <LayoutDashboard class="size-4.5" />
      </NuxtLink>

      <NuxtLink
        to="/projects"
        title="Projects"
        class="flex size-9 items-center justify-center rounded-md text-muted-foreground outline-none transition-colors hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
        :class="isActive('/projects') && 'bg-amber-500/15 text-amber-700 dark:text-amber-400'"
      >
        <FolderKanban class="size-4.5" />
      </NuxtLink>

      <NuxtLink
        to="/views"
        title="Views"
        class="flex size-9 items-center justify-center rounded-md text-muted-foreground outline-none transition-colors hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
        :class="isActive('/views') && 'bg-amber-500/15 text-amber-700 dark:text-amber-400'"
      >
        <Eye class="size-4.5" />
      </NuxtLink>
    </nav>

    <!-- Feedback + settings at bottom -->
    <div class="mt-auto flex flex-col items-center gap-1">
      <button
        v-if="feedbackPublic.send_to_main_enabled"
        type="button"
        title="Feedback"
        aria-label="Send feedback"
        class="flex size-9 items-center justify-center rounded-md text-muted-foreground outline-none transition-colors hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
        @click="showFeedback = true"
      >
        <MessageCircle class="size-4.5" />
      </button>

      <NuxtLink
        to="/settings"
        title="Settings"
        class="flex size-9 items-center justify-center rounded-md text-muted-foreground outline-none transition-colors hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
        :class="isActive('/settings') && 'bg-amber-500/15 text-amber-700 dark:text-amber-400'"
      >
        <Settings class="size-4.5" />
      </NuxtLink>
    </div>

    <FeedbackDialog v-model:open="showFeedback" />
  </aside>
</template>
