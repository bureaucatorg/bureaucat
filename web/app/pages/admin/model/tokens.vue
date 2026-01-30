<script setup lang="ts">
import { Key, Trash2, Loader2, ChevronLeft, ChevronRight, Sparkles } from "lucide-vue-next";

definePageMeta({
  middleware: ["admin"],
});

useSeoMeta({ title: "Manage Tokens" });

const { listTokens, revokeToken, cleanupExpiredTokens } = useAdmin();

interface TokenInfo {
  id: string;
  user_id: string;
  username: string;
  email: string;
  created_at: string;
  expires_at: string;
}

// State
const tokens = ref<TokenInfo[]>([]);
const loading = ref(true);
const page = ref(1);
const perPage = ref(20);
const total = ref(0);
const totalPages = ref(0);
const error = ref<string | null>(null);

// Revoke dialog state
const showRevokeDialog = ref(false);
const revokeLoading = ref(false);
const tokenToRevoke = ref<TokenInfo | null>(null);

// Cleanup state
const cleanupLoading = ref(false);

async function fetchTokens() {
  loading.value = true;
  error.value = null;
  const result = await listTokens(page.value, perPage.value);
  if (result.success && result.data) {
    tokens.value = result.data.tokens || [];
    total.value = result.data.total;
    totalPages.value = result.data.total_pages;
  } else {
    error.value = result.error || "Failed to fetch tokens";
  }
  loading.value = false;
}

function confirmRevoke(token: TokenInfo) {
  tokenToRevoke.value = token;
  showRevokeDialog.value = true;
}

async function handleRevokeToken() {
  if (!tokenToRevoke.value) return;

  revokeLoading.value = true;
  const result = await revokeToken(tokenToRevoke.value.id);
  revokeLoading.value = false;

  if (result.success) {
    showRevokeDialog.value = false;
    tokenToRevoke.value = null;
    await fetchTokens();
  } else {
    error.value = result.error || "Failed to revoke token";
  }
}

async function handleCleanup() {
  cleanupLoading.value = true;
  error.value = null;
  const result = await cleanupExpiredTokens();
  cleanupLoading.value = false;

  if (result.success) {
    // Show success message (could use toast)
    await fetchTokens();
  } else {
    error.value = result.error || "Failed to cleanup tokens";
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

// Pagination
function prevPage() {
  if (page.value > 1) {
    page.value--;
    fetchTokens();
  }
}

function nextPage() {
  if (page.value < totalPages.value) {
    page.value++;
    fetchTokens();
  }
}

onMounted(() => {
  fetchTokens();
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main class="flex-1">
      <div class="mx-auto max-w-6xl px-6 py-12">
        <div class="mb-8 flex items-center justify-between">
          <div>
            <h1 class="flex items-center gap-2 text-3xl font-bold tracking-tight">
              <Key class="size-8" />
              Token Management
            </h1>
            <p class="mt-2 text-muted-foreground">
              Manage active refresh tokens in the system
            </p>
          </div>
          <Button variant="outline" :disabled="cleanupLoading" @click="handleCleanup">
            <Loader2 v-if="cleanupLoading" class="mr-2 size-4 animate-spin" />
            <Sparkles v-else class="mr-2 size-4" />
            Cleanup Expired
          </Button>
        </div>

        <div v-if="error" class="mb-4 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
          {{ error }}
        </div>

        <Card>
          <CardContent class="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>User</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>Expires</TableHead>
                  <TableHead class="w-[100px]">Actions</TableHead>
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
                    No active tokens found
                  </TableCell>
                </TableRow>
                <TableRow v-for="token in tokens" :key="token.id">
                  <TableCell class="font-medium">{{ token.username }}</TableCell>
                  <TableCell>{{ token.email }}</TableCell>
                  <TableCell>{{ formatDate(token.created_at) }}</TableCell>
                  <TableCell>{{ formatDate(token.expires_at) }}</TableCell>
                  <TableCell>
                    <Button
                      variant="ghost"
                      size="icon"
                      class="text-destructive hover:text-destructive"
                      @click="confirmRevoke(token)"
                    >
                      <Trash2 class="size-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </CardContent>
          <CardFooter class="flex items-center justify-between border-t px-6 py-4">
            <p class="text-sm text-muted-foreground">
              Showing {{ tokens.length }} of {{ total }} active tokens
            </p>
            <div class="flex items-center gap-2">
              <Button variant="outline" size="sm" :disabled="page === 1" @click="prevPage">
                <ChevronLeft class="size-4" />
              </Button>
              <span class="text-sm">Page {{ page }} of {{ totalPages || 1 }}</span>
              <Button variant="outline" size="sm" :disabled="page >= totalPages" @click="nextPage">
                <ChevronRight class="size-4" />
              </Button>
            </div>
          </CardFooter>
        </Card>

        <!-- Revoke Confirmation Dialog -->
        <Dialog v-model:open="showRevokeDialog">
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Revoke Token</DialogTitle>
              <DialogDescription>
                Are you sure you want to revoke the token for "{{ tokenToRevoke?.username }}"?
                This will log them out of that session.
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button variant="outline" @click="showRevokeDialog = false" :disabled="revokeLoading">
                Cancel
              </Button>
              <Button variant="destructive" :disabled="revokeLoading" @click="handleRevokeToken">
                <Loader2 v-if="revokeLoading" class="mr-2 size-4 animate-spin" />
                Revoke
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </main>
  </div>
</template>
