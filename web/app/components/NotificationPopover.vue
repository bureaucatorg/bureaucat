<script setup lang="ts">
import { Bell, Loader2, CheckCheck } from "lucide-vue-next";
import type { NotificationEntry } from "~/types";

const { user } = useAuth();
const { unreadCount, listNotifications, refreshUnreadCount, markRead, markAllRead } =
  useNotifications();

const open = ref(false);
const loading = ref(false);
const loadingMore = ref(false);
const activities = ref<NotificationEntry[]>([]);
const page = ref(1);
const hasMore = ref(true);
const perPage = 20;
const maxItems = 100;

async function loadActivity(pageNum: number) {
  if (!user.value) return;

  const isFirstPage = pageNum === 1;
  if (isFirstPage) {
    loading.value = true;
  } else {
    loadingMore.value = true;
  }

  try {
    const { success, data } = await listNotifications(pageNum, perPage);
    if (success && data) {
      const items = data.activities ?? [];
      if (isFirstPage) {
        activities.value = items;
      } else {
        activities.value = [...activities.value, ...items];
      }
      page.value = pageNum;
      hasMore.value = items.length === perPage && activities.value.length < maxItems;
    }
  } finally {
    loading.value = false;
    loadingMore.value = false;
  }
}

function onScroll(e: Event) {
  if (loadingMore.value || !hasMore.value) return;
  const target = e.target as HTMLElement;
  if (!target) return;
  const nearBottom = target.scrollHeight - target.scrollTop - target.clientHeight < 80;
  if (nearBottom) {
    loadActivity(page.value + 1);
  }
}

function onActivityClick(activity: NotificationEntry) {
  if (!activity.is_read) {
    activity.is_read = true;
    markRead(activity.id);
  }
  open.value = false;
}

async function onMarkAllRead() {
  await markAllRead();
  activities.value = activities.value.map((a) => ({ ...a, is_read: true }));
}

const hasUnread = computed(() => activities.value.some((a) => !a.is_read));

watch(open, (isOpen) => {
  if (isOpen) {
    page.value = 1;
    hasMore.value = true;
    loadActivity(1);
  }
});

let pollTimer: ReturnType<typeof setInterval> | undefined;
onMounted(() => {
  refreshUnreadCount();
  // Light polling so the badge stays roughly fresh without a websocket.
  pollTimer = setInterval(refreshUnreadCount, 60_000);
});
onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer);
});
</script>

<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <Button
        variant="ghost"
        size="icon"
        class="relative size-9"
        aria-label="Notifications"
      >
        <Bell class="size-4" />
        <span
          v-if="unreadCount > 0"
          class="absolute -right-0.5 -top-0.5 flex min-w-4 items-center justify-center rounded-full bg-destructive px-1 text-[10px] font-semibold leading-4 text-white"
        >
          {{ unreadCount > 99 ? "99+" : unreadCount }}
        </span>
      </Button>
    </PopoverTrigger>
    <PopoverContent align="end" class="w-80 p-0">
      <div class="flex items-center justify-between border-b px-4 py-3">
        <h3 class="text-sm font-semibold">Notifications</h3>
        <button
          v-if="hasUnread"
          type="button"
          class="flex items-center gap-1 text-xs text-muted-foreground transition-colors hover:text-foreground"
          @click="onMarkAllRead"
        >
          <CheckCheck class="size-3.5" />
          Mark all read
        </button>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-8">
        <Loader2 class="size-5 animate-spin text-muted-foreground" />
      </div>

      <!-- Empty -->
      <div
        v-else-if="activities.length === 0"
        class="px-4 py-8 text-center text-sm text-muted-foreground"
      >
        No notifications yet
      </div>

      <!-- Activity list -->
      <ScrollArea v-else class="h-[380px]" @scrollCapture="onScroll">
        <div class="divide-y">
          <NotificationItem
            v-for="activity in activities"
            :key="activity.id"
            :activity="activity"
            @click="onActivityClick"
          />
        </div>
        <!-- Loading more indicator -->
        <div v-if="loadingMore" class="flex items-center justify-center py-3">
          <Loader2 class="size-4 animate-spin text-muted-foreground" />
        </div>
      </ScrollArea>

      <div class="border-t p-1">
        <Button
          variant="ghost"
          size="sm"
          class="w-full text-xs text-muted-foreground hover:text-foreground"
          as-child
        >
          <NuxtLink to="/notifications" @click="open = false">
            Show all notifications
          </NuxtLink>
        </Button>
      </div>
    </PopoverContent>
  </Popover>
</template>
