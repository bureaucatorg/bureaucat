<script setup lang="ts">
import { Bell, Loader2, ChevronLeft, ChevronRight, CheckCheck } from "lucide-vue-next";
import type { NotificationEntry } from "~/types";

definePageMeta({
  middleware: ["auth"],
});

useSeoMeta({ title: "Notifications" });

const route = useRoute();
const router = useRouter();
const { listNotifications, markRead, markAllRead } = useNotifications();

const perPage = 100;
const activities = ref<NotificationEntry[]>([]);
const total = ref(0);
const totalPages = ref(1);
const loading = ref(true);

const page = computed(() => {
  const p = Number(route.query.page);
  return Number.isInteger(p) && p > 0 ? p : 1;
});

async function load() {
  loading.value = true;
  const { success, data } = await listNotifications(page.value, perPage);
  if (success && data) {
    activities.value = data.activities ?? [];
    total.value = data.total;
    totalPages.value = Math.max(1, data.total_pages);
  }
  loading.value = false;
}

function goToPage(p: number) {
  if (p < 1 || p > totalPages.value) return;
  router.push({ query: { ...route.query, page: p === 1 ? undefined : String(p) } });
  window.scrollTo({ top: 0 });
}

function onActivityClick(activity: NotificationEntry) {
  if (!activity.is_read) {
    activity.is_read = true;
    markRead(activity.id);
  }
}

async function onMarkAllRead() {
  await markAllRead();
  activities.value = activities.value.map((a) => ({ ...a, is_read: true }));
}

const hasUnread = computed(() => activities.value.some((a) => !a.is_read));

watch(page, load, { immediate: true });
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main id="main-content" class="flex-1">
      <div class="mx-auto max-w-3xl px-4 py-8 sm:px-6">
        <div class="mb-6 flex items-center justify-between gap-4">
          <h1 class="flex items-center gap-3 text-2xl font-bold tracking-tight sm:text-3xl">
            <Bell class="size-6 sm:size-7" />
            Notifications
          </h1>
          <Button
            v-if="hasUnread"
            variant="ghost"
            size="sm"
            class="text-muted-foreground"
            @click="onMarkAllRead"
          >
            <CheckCheck class="mr-1.5 size-4" />
            Mark all read
          </Button>
        </div>

        <div v-if="loading" class="flex items-center justify-center py-12">
          <Loader2 class="size-8 animate-spin text-muted-foreground" />
        </div>

        <div
          v-else-if="activities.length === 0"
          class="flex flex-col items-center justify-center rounded-lg border border-dashed py-16"
        >
          <Bell class="size-8 text-muted-foreground" />
          <p class="mt-4 text-sm text-muted-foreground">No notifications yet</p>
        </div>

        <template v-else>
          <div class="divide-y overflow-hidden rounded-lg border">
            <NotificationItem
              v-for="activity in activities"
              :key="activity.id"
              :activity="activity"
              @click="onActivityClick"
            />
          </div>

          <div
            v-if="totalPages > 1"
            class="mt-6 flex flex-col items-center gap-3 sm:flex-row sm:justify-between"
          >
            <p class="text-sm text-muted-foreground">
              {{ total }} notification{{ total === 1 ? '' : 's' }}
            </p>
            <div class="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                :disabled="page <= 1"
                @click="goToPage(page - 1)"
              >
                <ChevronLeft class="mr-1 size-4" />
                Prev
              </Button>
              <span class="text-sm text-muted-foreground">
                Page {{ page }} of {{ totalPages }}
              </span>
              <Button
                variant="outline"
                size="sm"
                :disabled="page >= totalPages"
                @click="goToPage(page + 1)"
              >
                Next
                <ChevronRight class="ml-1 size-4" />
              </Button>
            </div>
          </div>
        </template>
      </div>
    </main>
  </div>
</template>
