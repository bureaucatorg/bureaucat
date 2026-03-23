<script setup lang="ts">
import { User, LogOut, LayoutDashboard, FolderKanban, Shield, Star } from "lucide-vue-next";

const { user, isAuthenticated, logout } = useAuth();
const { appName } = useSettings();
const route = useRoute();

const isLandingPage = computed(() => route.path === "/");

const appVersion = ref("");
onMounted(async () => {
  try {
    const res = await fetch("/api/v1/health");
    if (res.ok) {
      const data = await res.json();
      appVersion.value = data.version || "";
    }
  } catch {}
});

async function handleLogout() {
  await logout();
  await navigateTo("/");
}

</script>

<template>
  <header class="sticky top-0 z-40 border-b border-border/50 bg-background/80 backdrop-blur-xl">
    <div class="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
      <NuxtLink to="/" class="flex items-center gap-2.5">
        <BurecatLogo :size="28" />
        <span class="font-display text-lg font-semibold tracking-tight">{{ appName }}</span>
      </NuxtLink>

      <div class="flex items-center gap-4">
        <a
          v-if="isLandingPage"
          href="https://github.com/bureaucatorg/bureaucat"
          target="_blank"
          rel="noopener noreferrer"
          aria-label="Star on GitHub"
          class="inline-flex items-center gap-1.5 rounded-md border border-border/60 px-2.5 py-1 text-xs text-muted-foreground transition-colors hover:border-amber-500/40 hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
        >
          <Star class="size-3.5" />
          <span class="hidden sm:inline">Star on GitHub</span>
        </a>

        <ThemeToggle />

        <NotificationPopover v-if="isAuthenticated" />

        <template v-if="!isAuthenticated">
          <NuxtLink to="/signin">
            <Button variant="ghost" size="sm">Sign In</Button>
          </NuxtLink>
          <NuxtLink to="/signup">
            <Button size="sm">Sign Up</Button>
          </NuxtLink>
        </template>

        <template v-else>
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <Button variant="ghost" size="sm" class="gap-2">
                <User class="size-4" />
                <span class="hidden sm:inline">{{ user?.first_name }}</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" class="w-48">
              <DropdownMenuLabel>
                <NuxtLink :to="`/profile/${user?.id}`" class="flex flex-col hover:opacity-80 transition-opacity">
                  <span>{{ user?.first_name }} {{ user?.last_name }}</span>
                  <span class="text-xs font-normal text-muted-foreground">{{ user?.email }}</span>
                </NuxtLink>
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem as-child>
                <NuxtLink to="/dashboard" class="flex cursor-pointer items-center gap-2">
                  <LayoutDashboard class="size-4" />
                  <span>Dashboard</span>
                </NuxtLink>
              </DropdownMenuItem>
              <DropdownMenuItem as-child>
                <NuxtLink to="/projects" class="flex cursor-pointer items-center gap-2">
                  <FolderKanban class="size-4" />
                  <span>Projects</span>
                </NuxtLink>
              </DropdownMenuItem>
              <template v-if="user?.user_type === 'admin'">
                <DropdownMenuSeparator />
                <DropdownMenuItem as-child>
                  <NuxtLink to="/admin" class="flex cursor-pointer items-center gap-2">
                    <Shield class="size-4" />
                    <span>Admin Dashboard</span>
                  </NuxtLink>
                </DropdownMenuItem>
              </template>
              <DropdownMenuSeparator />
              <DropdownMenuItem class="cursor-pointer text-destructive focus:text-destructive" @click="handleLogout">
                <LogOut class="mr-2 size-4" />
                <span>Log out</span>
              </DropdownMenuItem>
              <template v-if="appVersion">
                <DropdownMenuSeparator />
                <div class="px-2 leading-none py-px text-center">
                  <span class="font-mono text-[10px] text-muted-foreground/60">v{{ appVersion }}</span>
                </div>
              </template>
            </DropdownMenuContent>
          </DropdownMenu>
        </template>
      </div>
    </div>
  </header>
</template>

<style scoped>
.font-display {
  font-family: 'DM Sans', system-ui, sans-serif;
}
</style>
