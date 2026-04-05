<script setup lang="ts">
import { Key, Trash2, Loader2, Plus, Copy, Check, Eye, EyeOff, CalendarIcon, X } from "lucide-vue-next";
import { getLocalTimeZone, today } from "@internationalized/date";
import { cn } from "@/lib/utils";
import type { DateValue } from "reka-ui";

definePageMeta({
  middleware: ["auth"],
});

useSeoMeta({ title: "Settings" });

const { listTokens, createToken, deleteToken } = usePAT();

interface TokenInfo {
  id: string;
  name: string;
  token?: string;
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
const expiryDate = ref<DateValue>();
const expiryPopoverOpen = ref(false);
const createLoading = ref(false);

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

const formattedExpiry = computed(() => {
  if (!expiryDate.value) return "";
  const d = expiryDate.value;
  return `${String(d.day).padStart(2, "0")}-${String(d.month).padStart(2, "0")}-${d.year}`;
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

  // Convert DateValue to RFC3339 string for the API
  let expiresAt: string | undefined;
  if (expiryDate.value) {
    const d = expiryDate.value;
    expiresAt = `${d.year}-${String(d.month).padStart(2, "0")}-${String(d.day).padStart(2, "0")}T23:59:59Z`;
  }

  const result = await createToken(newTokenName.value.trim(), expiresAt);
  createLoading.value = false;

  if (result.success && result.data?.token) {
    createdToken.value = result.data.token;
    copied.value = false;
    tokenRevealed.value = true;
    showCreatedDialog.value = true;
    newTokenName.value = "";
    expiryDate.value = undefined;
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
  expiryPopoverOpen.value = false;
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
              Tokens can be used to authenticate with the API. They have the same permissions as your account.
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
              <form class="flex flex-col gap-4 sm:flex-row sm:items-end" @submit.prevent="handleCreate">
                <div class="flex-1 space-y-2">
                  <Label for="token-name">Name</Label>
                  <Input
                    id="token-name"
                    v-model="newTokenName"
                    placeholder="e.g. CI/CD pipeline"
                    :maxlength="100"
                  />
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
          <Card>
            <CardContent class="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead>Last Used</TableHead>
                    <TableHead>Expires</TableHead>
                    <TableHead class="w-[80px]">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow v-if="loading">
                    <TableCell colspan="5" class="py-8 text-center">
                      <Loader2 class="mx-auto size-6 animate-spin" />
                    </TableCell>
                  </TableRow>
                  <TableRow v-else-if="tokens.length === 0">
                    <TableCell colspan="5" class="py-8 text-center text-muted-foreground">
                      No tokens yet
                    </TableCell>
                  </TableRow>
                  <TableRow v-for="token in tokens" :key="token.id">
                    <TableCell class="font-medium">{{ token.name }}</TableCell>
                    <TableCell class="text-muted-foreground">{{ formatDate(token.created_at) }}</TableCell>
                    <TableCell class="text-muted-foreground">{{ formatDate(token.last_used_at) }}</TableCell>
                    <TableCell>
                      <span
                        v-if="token.expires_at"
                        :class="isExpired(token.expires_at) ? 'text-destructive' : 'text-muted-foreground'"
                      >
                        {{ formatDate(token.expires_at) }}
                        <span v-if="isExpired(token.expires_at)" class="text-xs">(expired)</span>
                      </span>
                      <span v-else class="text-muted-foreground">Never</span>
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="icon"
                        aria-label="Delete token"
                        class="text-destructive hover:text-destructive"
                        @click="confirmDelete(token)"
                      >
                        <Trash2 class="size-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </CardContent>
          </Card>
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
