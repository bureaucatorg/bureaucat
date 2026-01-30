<script setup lang="ts">
import { Users, Key, Shield, ArrowRight, Loader2 } from "lucide-vue-next";
import { toast } from "vue-sonner";

definePageMeta({
  middleware: ["admin"],
});

useSeoMeta({ title: "Admin" });

const { branding, updateBranding } = useSettings();

const brandingForm = ref({
  enabled: branding.value.enabled,
  app_name: branding.value.app_name,
});

const savingBranding = ref(false);

// Sync form with branding when it changes
watch(branding, (newBranding) => {
  brandingForm.value.enabled = newBranding.enabled;
  brandingForm.value.app_name = newBranding.app_name;
}, { immediate: true });

async function handleSaveBranding() {
  savingBranding.value = true;
  const result = await updateBranding({
    enabled: brandingForm.value.enabled,
    app_name: brandingForm.value.app_name || "Bureaucat",
  });
  savingBranding.value = false;

  if (result.success) {
    toast.success("Branding settings saved");
  } else {
    toast.error(result.error || "Failed to save branding settings");
  }
}

const adminModels = [
  {
    title: "Users",
    description: "Manage user accounts, create new users, and control access levels",
    icon: Users,
    href: "/admin/model/users",
    color: "text-blue-500",
    bgColor: "bg-blue-500/10",
  },
  {
    title: "Tokens",
    description: "Monitor active sessions, revoke tokens, and cleanup expired sessions",
    icon: Key,
    href: "/admin/model/tokens",
    color: "text-amber-500",
    bgColor: "bg-amber-500/10",
  },
];
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main class="flex-1">
      <div class="mx-auto max-w-6xl px-6 py-12">
        <div class="mb-8">
          <div class="flex items-center gap-3 mb-2">
            <div class="flex size-10 items-center justify-center rounded-lg bg-foreground">
              <Shield class="size-5 text-background" />
            </div>
            <h1 class="text-3xl font-bold tracking-tight">Admin Dashboard</h1>
          </div>
          <p class="text-muted-foreground">
            Manage your application's data and settings
          </p>
        </div>

        <div class="grid gap-4 sm:grid-cols-2">
          <NuxtLink
            v-for="model in adminModels"
            :key="model.title"
            :to="model.href"
            class="group"
          >
            <Card class="h-full transition-all hover:border-foreground/20 hover:shadow-lg">
              <CardHeader>
                <div class="flex items-center justify-between">
                  <div :class="[model.bgColor, 'flex size-12 items-center justify-center rounded-lg']">
                    <component :is="model.icon" :class="['size-6', model.color]" />
                  </div>
                  <ArrowRight class="size-5 text-muted-foreground transition-transform group-hover:translate-x-1" />
                </div>
                <CardTitle class="mt-4">{{ model.title }}</CardTitle>
                <CardDescription>{{ model.description }}</CardDescription>
              </CardHeader>
            </Card>
          </NuxtLink>
        </div>

        <!-- Branding Settings -->
        <div class="mt-12">
          <div class="mb-4">
            <h2 class="text-xl font-semibold">Branding</h2>
            <p class="text-sm text-muted-foreground">
              Customize the application name and appearance
            </p>
          </div>

          <Card>
            <CardContent class="pt-6">
              <div class="space-y-6">
                <!-- Toggle -->
                <div class="flex items-center justify-between">
                  <div>
                    <p class="font-medium">Hide from the bureaucrats 😾</p>
                    <p class="text-sm text-muted-foreground">
                      Replace "Bureaucat" with a custom name
                    </p>
                  </div>
                  <Switch
                    :checked="brandingForm.enabled"
                    @update:checked="brandingForm.enabled = $event"
                  />
                </div>

                <!-- Custom name input - always visible when toggle is on -->
                <div v-if="brandingForm.enabled" class="space-y-2">
                  <Label for="app-name">Custom Application Name</Label>
                  <Input
                    id="app-name"
                    v-model="brandingForm.app_name"
                    placeholder="Enter a custom name"
                    :disabled="savingBranding"
                  />
                </div>

                <!-- Save button -->
                <div class="flex justify-end pt-2">
                  <Button @click="handleSaveBranding">
                    <Loader2 v-if="savingBranding" class="mr-2 size-4 animate-spin" />
                    Save Changes
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </main>
  </div>
</template>
