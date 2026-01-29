<script setup lang="ts">
import { User, LogOut, LayoutDashboard, FolderKanban, Shield } from "lucide-vue-next";

const { user, isAuthenticated, logout } = useAuth();

async function handleLogout() {
  await logout();
  await navigateTo("/");
}
</script>

<template>
  <header class="sticky top-0 z-40 border-b border-border/50 bg-background/80 backdrop-blur-xl">
    <div class="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
      <NuxtLink to="/" class="flex items-center gap-3">
        <div class="flex size-8 items-center justify-center rounded-md bg-foreground">
          <span class="font-mono text-sm font-bold text-background">B</span>
        </div>
        <span class="font-display text-lg font-semibold tracking-tight">Bureaucat</span>
      </NuxtLink>

      <div class="flex items-center gap-4">
        <nav class="hidden items-center gap-6 md:flex">
          <template v-if="!isAuthenticated">
            <a href="#features" class="text-sm text-muted-foreground transition-colors hover:text-foreground">Features</a>
            <a href="#how-it-works" class="text-sm text-muted-foreground transition-colors hover:text-foreground">How it works</a>
          </template>
          <template v-else>
            <NuxtLink to="/dashboard" class="text-sm text-muted-foreground transition-colors hover:text-foreground">Dashboard</NuxtLink>
            <NuxtLink to="/projects" class="text-sm text-muted-foreground transition-colors hover:text-foreground">Projects</NuxtLink>
            <template v-if="user?.user_type === 'admin'">
              <NuxtLink to="/admin" class="text-sm text-muted-foreground transition-colors hover:text-foreground">Admin</NuxtLink>
            </template>
          </template>
        </nav>

        <ThemeToggle />

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
                <div class="flex flex-col">
                  <span>{{ user?.first_name }} {{ user?.last_name }}</span>
                  <span class="text-xs font-normal text-muted-foreground">{{ user?.email }}</span>
                </div>
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
