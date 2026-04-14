<script setup lang="ts">
import { Key, Trash2, Loader2, Plus, Copy, Check, Eye, EyeOff, CalendarIcon, X, Clock } from "lucide-vue-next";
import { getLocalTimeZone, today } from "@internationalized/date";
import { cn } from "@/lib/utils";
import type { DateValue } from "reka-ui";

definePageMeta({
  middleware: ["auth"],
});

useSeoMeta({ title: "Settings" });

const { listTokens, createToken, updateTokenScope, deleteToken } = usePAT();

type PATScope = "read_only" | "read_write";

interface TokenInfo {
  id: string;
  name: string;
  token?: string;
  scope: PATScope;
  expires_at: string | null;
  last_used_at: string | null;
  created_at: string;
}

// Token list state
const tokens = ref<TokenInfo[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);

// Create form state
const newTokenName = ref("");
const newTokenScope = ref<PATScope>("read_write");
const expiryDate = ref<DateValue>();
const expiryHour = ref("23");
const expiryMinute = ref("59");
const expiryPopoverOpen = ref(false);
const createLoading = ref(false);

// Per-row scope update state
const scopeUpdating = ref<Record<string, boolean>>({});

async function handleScopeChange(token: TokenInfo, scope: PATScope) {
  if (token.scope === scope) return;
  scopeUpdating.value[token.id] = true;
  const result = await updateTokenScope(token.id, scope);
  scopeUpdating.value[token.id] = false;
  if (result.success) {
    token.scope = scope;
  } else {
    error.value = result.error || "Failed to update scope";
  }
}

// Created token dialog
const showCreatedDialog = ref(false);
const createdToken = ref<string | null>(null);
const copied = ref(false);
const tokenRevealed = ref(true);

// Delete dialog state
const showDeleteDialog = ref(false);
const deleteLoading = ref(false);
const tokenToDelete = ref<TokenInfo | null>(null);

const minDate = today(getLocalTimeZone()).add({ days: 1 });

const hours = Array.from({ length: 24 }, (_, i) => String(i).padStart(2, "0"));
const minutes = Array.from({ length: 60 }, (_, i) => String(i).padStart(2, "0"));

const formattedExpiry = computed(() => {
  if (!expiryDate.value) return "";
  const d = expiryDate.value;
  return `${String(d.day).padStart(2, "0")}-${String(d.month).padStart(2, "0")}-${d.year} ${expiryHour.value}:${expiryMinute.value}`;
});

async function fetchTokens() {
  loading.value = true;
  error.value = null;
  const result = await listTokens();
  if (result.success && result.data) {
    tokens.value = result.data.tokens || [];
  } else {
    error.value = result.error || "Failed to fetch tokens";
  }
  loading.value = false;
}

async function handleCreate() {
  if (!newTokenName.value.trim()) return;

  createLoading.value = true;
  error.value = null;

  // Convert DateValue + time to RFC3339 string for the API
  let expiresAt: string | undefined;
  if (expiryDate.value) {
    const d = expiryDate.value;
    expiresAt = `${d.year}-${String(d.month).padStart(2, "0")}-${String(d.day).padStart(2, "0")}T${expiryHour.value}:${expiryMinute.value}:00Z`;
  }

  const result = await createToken(newTokenName.value.trim(), newTokenScope.value, expiresAt);
  createLoading.value = false;

  if (result.success && result.data?.token) {
    createdToken.value = result.data.token;
    copied.value = false;
    tokenRevealed.value = true;
    showCreatedDialog.value = true;
    newTokenName.value = "";
    newTokenScope.value = "read_write";
    expiryDate.value = undefined;
    expiryHour.value = "23";
    expiryMinute.value = "59";
    await fetchTokens();
  } else {
    error.value = result.error || "Failed to create token";
  }
}

function confirmDelete(token: TokenInfo) {
  tokenToDelete.value = token;
  showDeleteDialog.value = true;
}

async function handleDelete() {
  if (!tokenToDelete.value) return;

  deleteLoading.value = true;
  const result = await deleteToken(tokenToDelete.value.id);
  deleteLoading.value = false;

  if (result.success) {
    showDeleteDialog.value = false;
    tokenToDelete.value = null;
    await fetchTokens();
  } else {
    error.value = result.error || "Failed to delete token";
  }
}

async function copyToken() {
  if (!createdToken.value) return;
  try {
    await navigator.clipboard.writeText(createdToken.value);
    copied.value = true;
    setTimeout(() => (copied.value = false), 2000);
  } catch {
    // fallback: select the text
  }
}

function formatDate(dateStr: string | null) {
  if (!dateStr) return "-";
  const d = new Date(dateStr);
  const day = String(d.getDate()).padStart(2, "0");
  const month = String(d.getMonth() + 1).padStart(2, "0");
  const year = d.getFullYear();
  const hours = String(d.getHours()).padStart(2, "0");
  const minutes = String(d.getMinutes()).padStart(2, "0");
  return `${day}-${month}-${year} ${hours}:${minutes}`;
}

function isExpired(dateStr: string | null) {
  if (!dateStr) return false;
  return new Date(dateStr) < new Date();
}

function onExpirySelect(date: DateValue) {
  expiryDate.value = date;
}

onMounted(() => {
  fetchTokens();
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main id="main-content" class="flex-1">
      <div class="mx-auto max-w-4xl px-6 py-12">
        <div class="mb-8">
          <h1 class="text-3xl font-bold tracking-tight">Settings</h1>
          <p class="mt-2 text-muted-foreground">
            Manage your account settings
          </p>
        </div>

        <!-- Personal Access Tokens Section -->
        <div>
          <div class="mb-4">
            <h2 class="flex items-center gap-2 text-lg font-semibold">
              <Key class="size-5" />
              Personal Access Tokens
            </h2>
            <p class="mt-1 text-sm text-muted-foreground">
              Tokens authenticate API requests. Choose <span class="font-medium">Read only</span> to restrict a token to safe methods (GET), or <span class="font-medium">Read &amp; write</span> for full access. Scope can be changed anytime.
            </p>
          </div>

          <div v-if="error" role="alert" class="mb-4 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
            {{ error }}
          </div>

          <!-- Create Token Form -->
          <Card class="mb-6">
            <CardHeader>
              <CardTitle class="text-base">Create a new token</CardTitle>
            </CardHeader>
            <CardContent>
              <form class="flex flex-col gap-4 sm:flex-row sm:items-end sm:flex-wrap" @submit.prevent="handleCreate">
                <div class="flex-1 space-y-2 sm:min-w-[14rem]">
                  <Label for="token-name">Name</Label>
                  <Input
                    id="token-name"
                    v-model="newTokenName"
                    placeholder="e.g. CI/CD pipeline"
                    :maxlength="100"
                  />
                </div>
                <div class="w-full space-y-2 sm:w-44">
                  <Label for="token-scope">Permissions</Label>
                  <Select v-model="newTokenScope">
                    <SelectTrigger id="token-scope" class="w-full">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="read_write">Read &amp; write</SelectItem>
                      <SelectItem value="read_only">Read only</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div class="w-full space-y-2 sm:w-56">
                  <Label>Expiry (optional)</Label>
                  <Popover v-model:open="expiryPopoverOpen">
                    <PopoverTrigger as-child>
                      <Button
                        variant="outline"
                        :class="cn('w-full justify-start text-left font-normal', !expiryDate && 'text-muted-foreground')"
                      >
                        <CalendarIcon class="mr-2 size-4" />
                        <span>{{ formattedExpiry || 'Pick a date' }}</span>
                        <button
                          v-if="expiryDate"
                          type="button"
                          class="ml-auto text-muted-foreground hover:text-foreground"
                          @click.stop="expiryDate = undefined"
                        >
                          <X class="size-3.5" />
                        </button>
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent class="w-auto p-0" align="start">
                      <Calendar
                        :model-value="expiryDate"
                        :min-value="minDate"
                        @update:model-value="onExpirySelect"
                      />
                      <div class="border-t px-3 py-3">
                        <div class="flex items-center gap-2">
                          <Clock class="size-4 text-muted-foreground" />
                          <Select v-model="expiryHour">
                            <SelectTrigger class="h-8 w-[4.5rem]">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem v-for="h in hours" :key="h" :value="h">{{ h }}</SelectItem>
                            </SelectContent>
                          </Select>
                          <span class="text-sm font-medium text-muted-foreground">:</span>
                          <Select v-model="expiryMinute">
                            <SelectTrigger class="h-8 w-[4.5rem]">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem v-for="m in minutes" :key="m" :value="m">{{ m }}</SelectItem>
                            </SelectContent>
                          </Select>
                        </div>
                      </div>
                    </PopoverContent>
                  </Popover>
                </div>
                <Button type="submit" :disabled="createLoading || !newTokenName.trim()" class="shrink-0">
                  <Loader2 v-if="createLoading" class="mr-2 size-4 animate-spin" />
                  <Plus v-else class="mr-2 size-4" />
                  Create Token
                </Button>
              </form>
            </CardContent>
          </Card>

          <!-- Tokens List -->
          <div v-if="loading" class="flex items-center justify-center py-12">
            <Loader2 class="size-6 animate-spin text-muted-foreground" />
          </div>

          <div
            v-else-if="tokens.length === 0"
            class="flex flex-col items-center justify-center rounded-lg border border-dashed py-12"
          >
            <div class="flex size-12 items-center justify-center rounded-full bg-muted">
              <Key class="size-6 text-muted-foreground" />
            </div>
            <p class="mt-3 text-sm text-muted-foreground">No tokens yet</p>
            <p class="mt-1 text-xs text-muted-foreground/70">Create a token above to get started</p>
          </div>

          <div v-else class="space-y-3">
            <div
              v-for="token in tokens"
              :key="token.id"
              class="group flex items-center justify-between rounded-lg border bg-card px-4 py-3 transition-colors hover:bg-muted/50"
            >
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2">
                  <Key class="size-3.5 shrink-0 text-muted-foreground" />
                  <span class="truncate text-sm font-medium">{{ token.name }}</span>
                  <span
                    v-if="token.expires_at && isExpired(token.expires_at)"
                    class="rounded-full bg-destructive/10 px-2 py-0.5 text-[10px] font-medium text-destructive"
                  >
                    Expired
                  </span>
                  <Select
                    :model-value="token.scope"
                    :disabled="scopeUpdating[token.id]"
                    @update:model-value="(v) => handleScopeChange(token, v as PATScope)"
                  >
                    <SelectTrigger
                      :class="[
                        'h-6 w-auto gap-1 rounded-full border-0 px-2 py-0 text-[10px] font-medium shadow-none focus:ring-0',
                        token.scope === 'read_only'
                          ? 'bg-muted text-muted-foreground hover:bg-muted/80'
                          : 'bg-amber-500/15 text-amber-700 hover:bg-amber-500/25 dark:text-amber-400',
                      ]"
                    >
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="read_write">Read &amp; write</SelectItem>
                      <SelectItem value="read_only">Read only</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div class="mt-1.5 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground">
                  <span class="flex items-center gap-1">
                    <CalendarIcon class="size-3" />
                    Created {{ formatDate(token.created_at) }}
                  </span>
                  <span v-if="token.last_used_at" class="flex items-center gap-1">
                    <Clock class="size-3" />
                    Last used {{ formatDate(token.last_used_at) }}
                  </span>
                  <span v-else class="flex items-center gap-1">
                    <Clock class="size-3" />
                    Never used
                  </span>
                  <span v-if="token.expires_at && !isExpired(token.expires_at)" class="flex items-center gap-1">
                    Expires {{ formatDate(token.expires_at) }}
                  </span>
                  <span v-else-if="!token.expires_at" class="flex items-center gap-1">
                    No expiry
                  </span>
                </div>
              </div>
              <Button
                variant="ghost"
                size="icon"
                aria-label="Delete token"
                class="shrink-0 text-muted-foreground opacity-0 transition-opacity hover:text-destructive group-hover:opacity-100"
                @click="confirmDelete(token)"
              >
                <Trash2 class="size-4" />
              </Button>
            </div>
          </div>
        </div>

        <!-- Token Created Dialog -->
        <Dialog v-model:open="showCreatedDialog">
          <DialogContent class="sm:max-w-lg">
            <DialogHeader>
              <DialogTitle>Token Created</DialogTitle>
              <DialogDescription>
                Copy your token now. You won't be able to see it again.
              </DialogDescription>
            </DialogHeader>
            <div class="space-y-3">
              <div class="flex items-center gap-2">
                <div class="relative flex-1">
                  <Input
                    :model-value="tokenRevealed ? createdToken || '' : '***************'"
                    readonly
                    class="pr-10 font-mono text-sm"
                  />
                  <button
                    type="button"
                    class="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    @click="tokenRevealed = !tokenRevealed"
                  >
                    <EyeOff v-if="tokenRevealed" class="size-4" />
                    <Eye v-else class="size-4" />
                  </button>
                </div>
                <Button variant="outline" size="icon" @click="copyToken" :title="copied ? 'Copied!' : 'Copy to clipboard'">
                  <Check v-if="copied" class="size-4 text-green-500" />
                  <Copy v-else class="size-4" />
                </Button>
              </div>
              <p class="text-xs text-amber-600 dark:text-amber-400">
                Make sure to copy the token. It will not be shown again.
              </p>
            </div>
            <DialogFooter>
              <Button @click="showCreatedDialog = false">Done</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <!-- Delete Confirmation Dialog -->
        <Dialog v-model:open="showDeleteDialog">
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Delete Token</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete "{{ tokenToDelete?.name }}"?
                Any applications using this token will no longer be able to authenticate.
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button variant="outline" :disabled="deleteLoading" @click="showDeleteDialog = false">
                Cancel
              </Button>
              <Button variant="destructive" :disabled="deleteLoading" @click="handleDelete">
                <Loader2 v-if="deleteLoading" class="mr-2 size-4 animate-spin" />
                Delete
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </main>
  </div>
</template>
