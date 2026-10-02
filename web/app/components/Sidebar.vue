<script setup lang="ts">
import { LayoutDashboard, FolderKanban, Eye, Repeat, Layers, Timer, MessageCircle, Newspaper, BookOpen, Shield, Settings } from "lucide-vue-next";

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
const showReleases = ref(false);

const SEEN_VERSION_KEY = "bureaucat:last-seen-version";
const appVersion = ref("");
const hasUnseenRelease = ref(false);

async function checkUnseenRelease() {
  try {
    const res = await fetch("/api/v1/health");
    if (!res.ok) return;
    appVersion.value = (await res.json()).version || "";
    hasUnseenRelease.value = !!appVersion.value && localStorage.getItem(SEEN_VERSION_KEY) !== appVersion.value;
  } catch {}
}

function openReleases() {
  showReleases.value = true;
  hasUnseenRelease.value = false;
  if (appVersion.value) {
    try {
      localStorage.setItem(SEEN_VERSION_KEY, appVersion.value);
    } catch {}
  }
}

onMounted(() => {
  fetchFeedbackPublicSettings();
  checkUnseenRelease();
});
</script>

<template>
  <aside
    class="fixed inset-y-0 left-0 z-50 flex w-12 flex-col items-center border-r border-border/60 bg-muted/40 py-3 backdrop-blur-xl"
  >
    <!-- Workspace switcher -->
    <WorkspaceSwitcher class="mb-3" />

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

      <NuxtLink
        to="/cycles/active"
        title="Active Cycles"
        class="flex size-9 items-center justify-center rounded-md text-muted-foreground outline-none transition-colors hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
        :class="route.path === '/cycles/active' && 'bg-amber-500/15 text-amber-700 dark:text-amber-400'"
      >
        <Repeat class="size-4.5" />
      </NuxtLink>

      <NuxtLink
        to="/modules/active"
        title="Active Modules"
        class="flex size-9 items-center justify-center rounded-md text-muted-foreground outline-none transition-colors hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
        :class="route.path === '/modules/active' && 'bg-amber-500/15 text-amber-700 dark:text-amber-400'"
      >
        <Layers class="size-4.5" />
      </NuxtLink>
    </nav>

    <!-- Pomodoro + feedback + settings at bottom -->
    <div class="mt-auto flex flex-col items-center gap-1">
      <NuxtLink
        to="/pomodoro"
        title="Pomodoro"
        class="flex size-9 items-center justify-center rounded-md text-muted-foreground outline-none transition-colors hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
        :class="isActive('/pomodoro') && 'bg-amber-500/15 text-amber-700 dark:text-amber-400'"
      >
        <Timer class="size-4.5" />
      </NuxtLink>

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

      <button
        type="button"
        title="What's new"
        aria-label="What's new"
        class="relative flex size-9 items-center justify-center rounded-md text-muted-foreground outline-none transition-colors hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
        @click="openReleases"
      >
        <Newspaper class="size-4.5" />
        <span
          v-if="hasUnseenRelease"
          class="absolute right-2 top-2 size-1.5 rounded-full bg-red-500"
          aria-hidden="true"
        />
      </button>

      <NuxtLink
        to="/docs"
        title="API Docs"
        class="flex size-9 items-center justify-center rounded-md text-muted-foreground outline-none transition-colors hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
        :class="isActive('/docs') && 'bg-amber-500/15 text-amber-700 dark:text-amber-400'"
      >
        <BookOpen class="size-4.5" />
      </NuxtLink>

      <NuxtLink
        v-if="user?.user_type === 'admin'"
        to="/admin"
        title="Admin Dashboard"
        class="flex size-9 items-center justify-center rounded-md text-muted-foreground outline-none transition-colors hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
        :class="isActive('/admin') && 'bg-amber-500/15 text-amber-700 dark:text-amber-400'"
      >
        <Shield class="size-4.5" />
      </NuxtLink>

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
    <ReleasesDialog v-model:open="showReleases" />
  </aside>
</template>
