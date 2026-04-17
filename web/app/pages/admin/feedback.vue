<script setup lang="ts">
import {
  MessageCircle,
  Loader2,
  ChevronLeft,
  ChevronRight,
  Trash2,
  Shield,
  Inbox,
  Send,
  Archive,
} from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { FeedbackItem } from "~/composables/useFeedback";
import type { FeedbackSettings } from "~/composables/useSettings";

definePageMeta({
  middleware: ["admin"],
});

useSeoMeta({ title: "Feedback" });

const { list, remove } = useFeedback();
const { fetchFeedbackSettings, updateFeedbackSettings } = useSettings();

const items = ref<FeedbackItem[]>([]);
const total = ref(0);
const page = ref(1);
const perPage = ref(50);
const totalPages = ref(1);
const loading = ref(true);
const settingsLoading = ref(true);
const savingReceive = ref(false);
const savingSend = ref(false);

const settings = ref<FeedbackSettings>({
  receive_enabled: false,
  send_to_main_enabled: false,
  store_sent_locally: true,
});
const savingStore = ref(false);

async function loadList(p = page.value) {
  loading.value = true;
  const res = await list(p, perPage.value);
  if (res.success && res.data) {
    items.value = res.data.items;
    total.value = res.data.total;
    page.value = res.data.page;
    totalPages.value = res.data.total_pages;
  } else if (res.error) {
    toast.error(res.error);
  }
  loading.value = false;
}

async function loadSettings() {
  settingsLoading.value = true;
  const res = await fetchFeedbackSettings();
  if (res.success && res.data) {
    settings.value = res.data;
  }
  settingsLoading.value = false;
}

async function toggleReceive(enabled: boolean) {
  savingReceive.value = true;
  const res = await updateFeedbackSettings({
    ...settings.value,
    receive_enabled: enabled,
  });
  savingReceive.value = false;
  if (res.success && res.data) {
    settings.value = res.data;
    toast.success(enabled ? "Accepting new feedback" : "Dropping new feedback");
  } else {
    toast.error(res.error || "Failed to update");
  }
}

async function toggleSend(enabled: boolean) {
  savingSend.value = true;
  const res = await updateFeedbackSettings({
    ...settings.value,
    send_to_main_enabled: enabled,
  });
  savingSend.value = false;
  if (res.success && res.data) {
    settings.value = res.data;
    toast.success(
      enabled
        ? "Users on this instance can now send feedback to bureaucat.org"
        : "Feedback button hidden for users on this instance"
    );
  } else {
    toast.error(res.error || "Failed to update");
  }
}

async function toggleStoreLocally(enabled: boolean) {
  savingStore.value = true;
  const res = await updateFeedbackSettings({
    ...settings.value,
    store_sent_locally: enabled,
  });
  savingStore.value = false;
  if (res.success && res.data) {
    settings.value = res.data;
    toast.success(
      enabled
        ? "Local copies of outbound feedback will be saved"
        : "Outbound feedback will no longer be mirrored locally"
    );
  } else {
    toast.error(res.error || "Failed to update");
  }
}

async function handleDelete(item: FeedbackItem) {
  if (!confirm("Delete this feedback entry? This cannot be undone.")) return;
  const res = await remove(item.id);
  if (res.success) {
    items.value = items.value.filter((i) => i.id !== item.id);
    total.value = Math.max(0, total.value - 1);
    toast.success("Deleted");
  } else {
    toast.error(res.error || "Failed to delete");
  }
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString();
}

function goToPage(p: number) {
  if (p < 1 || p > totalPages.value) return;
  loadList(p);
}

onMounted(() => {
  loadSettings();
  loadList(1);
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main id="main-content" class="flex-1">
      <div class="mx-auto max-w-6xl px-6 py-12">
        <nav class="mb-4 flex items-center gap-2 text-sm text-muted-foreground">
          <NuxtLink to="/admin" class="flex items-center gap-1 hover:text-foreground">
            <ChevronLeft class="size-4" />
            Admin
          </NuxtLink>
          <span>/</span>
          <span class="font-semibold text-foreground">Feedback</span>
        </nav>

        <div class="mb-8 flex items-center gap-3">
          <div class="flex size-10 items-center justify-center rounded-lg bg-foreground">
            <MessageCircle class="size-5 text-background" />
          </div>
          <div>
            <h1 class="text-3xl font-bold tracking-tight">Feedback</h1>
            <p class="text-muted-foreground">
              Anonymous messages received by this instance.
            </p>
          </div>
        </div>

        <!-- Settings -->
        <div class="mb-8 space-y-4">
          <Card>
            <CardContent class="pt-6">
              <div class="flex items-center justify-between gap-4">
                <div class="flex items-start gap-3">
                  <Inbox class="mt-0.5 size-5 text-muted-foreground" />
                  <div>
                    <p class="font-medium">Accept incoming feedback</p>
                    <p class="text-sm text-muted-foreground">
                      When off, submissions to this instance's
                      <span class="font-mono">/api/v1/feedback</span> are rejected
                      with 503. Incoming POSTs are rate-limited to 5 per IP per
                      hour regardless of this setting.
                    </p>
                  </div>
                </div>
                <Switch
                  :checked="settings.receive_enabled"
                  :disabled="settingsLoading || savingReceive"
                  @update:checked="toggleReceive"
                />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent class="pt-6">
              <div class="flex items-center justify-between gap-4">
                <div class="flex items-start gap-3">
                  <Archive class="mt-0.5 size-5 text-muted-foreground" />
                  <div>
                    <p class="font-medium">
                      Store a local copy of outbound feedback
                    </p>
                    <p class="text-sm text-muted-foreground">
                      On by default. When users here send feedback to
                      bureaucat.org, a copy is also saved to this instance —
                      tagged with a <span class="font-mono">local:</span> origin
                      so you can tell them apart in the list below.
                    </p>
                  </div>
                </div>
                <Switch
                  :checked="settings.store_sent_locally"
                  :disabled="settingsLoading || savingStore"
                  @update:checked="toggleStoreLocally"
                />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent class="pt-6">
              <div class="flex items-center justify-between gap-4">
                <div class="flex items-start gap-3">
                  <Send class="mt-0.5 size-5 text-muted-foreground" />
                  <div>
                    <p class="font-medium">
                      Let users on this instance send feedback to bureaucat.org
                    </p>
                    <p class="text-sm text-muted-foreground">
                      Off by default. When on, a feedback button appears in the
                      sidebar for every signed-in user. Submissions go straight
                      to bureaucat.org — nothing is stored here.
                    </p>
                  </div>
                </div>
                <Switch
                  :checked="settings.send_to_main_enabled"
                  :disabled="settingsLoading || savingSend"
                  @update:checked="toggleSend"
                />
              </div>
            </CardContent>
          </Card>
        </div>

        <!-- List -->
        <div class="mb-4 flex items-center justify-between">
          <h2 class="text-xl font-semibold">
            Received messages
            <span v-if="!loading" class="ml-2 text-sm font-normal text-muted-foreground">
              ({{ total }})
            </span>
          </h2>
        </div>

        <div v-if="loading" class="flex items-center justify-center py-12">
          <Loader2 class="size-8 animate-spin text-muted-foreground" />
        </div>

        <div
          v-else-if="items.length === 0"
          class="flex flex-col items-center justify-center rounded-lg border border-dashed py-16"
        >
          <Shield class="size-8 text-muted-foreground" />
          <h3 class="mt-4 font-semibold">No feedback yet</h3>
          <p class="mt-1 max-w-sm text-center text-sm text-muted-foreground">
            When anyone on any Bureaucat instance sends feedback to this server,
            it'll show up here.
          </p>
        </div>

        <div v-else class="space-y-3">
          <Card v-for="item in items" :key="item.id">
            <CardContent class="pt-6">
              <div class="flex items-start justify-between gap-4">
                <div class="min-w-0 flex-1">
                  <p class="whitespace-pre-wrap break-words text-sm">
                    {{ item.message }}
                  </p>
                  <div class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground">
                    <span>{{ formatDate(item.created_at) }}</span>
                    <span v-if="item.source_origin" class="font-mono">
                      {{ item.source_origin }}
                    </span>
                    <span
                      v-if="item.user_agent"
                      class="max-w-[24rem] truncate font-mono"
                      :title="item.user_agent"
                    >
                      {{ item.user_agent }}
                    </span>
                  </div>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label="Delete feedback"
                  class="text-muted-foreground hover:text-destructive"
                  @click="handleDelete(item)"
                >
                  <Trash2 class="size-4" />
                </Button>
              </div>
            </CardContent>
          </Card>

          <!-- Pagination -->
          <div
            v-if="totalPages > 1"
            class="mt-6 flex items-center justify-between"
          >
            <p class="text-sm text-muted-foreground">
              Page {{ page }} of {{ totalPages }}
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
        </div>
      </div>
    </main>
  </div>
</template>
