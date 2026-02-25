<script setup lang="ts">
import { Users, Key, Shield, ArrowRight, Loader2, Upload, CheckCircle2 } from "lucide-vue-next";
import { toast } from "vue-sonner";

definePageMeta({
  middleware: ["admin"],
});

useSeoMeta({ title: "Admin" });

const { branding, updateBranding } = useSettings();
const { getAuthHeader } = useAuth();

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

// Plane import state
interface ImportSummary {
  users_created: number;
  users_skipped: number;
  projects_created: number;
  states_created: number;
  labels_created: number;
  tasks_created: number;
  assignees_linked: number;
  labels_assigned: number;
  comments_created: number;
}

const selectedFile = ref<File | null>(null);
const importing = ref(false);
const importResult = ref<ImportSummary | null>(null);
const fileInputRef = ref<HTMLInputElement | null>(null);

function handleFileSelect(event: Event) {
  const target = event.target as HTMLInputElement;
  selectedFile.value = target.files?.[0] ?? null;
  importResult.value = null;
}

async function handlePlaneImport() {
  if (!selectedFile.value) return;
  importing.value = true;
  importResult.value = null;

  const formData = new FormData();
  formData.append("file", selectedFile.value);

  try {
    const response = await fetch("/api/v1/admin/import/plane", {
      method: "POST",
      headers: {
        ...getAuthHeader(),
      },
      credentials: "include",
      body: formData,
    });

    const data = await response.json();

    if (!response.ok) {
      toast.error(data.message || "Import failed");
      return;
    }

    importResult.value = data.summary;
    toast.success(data.message || "Import completed");
  } catch {
    toast.error("Network error during import");
  } finally {
    importing.value = false;
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

        <!-- Data Import -->
        <div class="mt-12">
          <div class="mb-4">
            <h2 class="text-xl font-semibold">Data Import</h2>
            <p class="text-sm text-muted-foreground">
              Import data from external project management tools
            </p>
          </div>

          <Card>
            <CardHeader>
              <div class="flex items-center gap-3">
                <div class="flex size-10 items-center justify-center rounded-lg bg-emerald-500/10">
                  <Upload class="size-5 text-emerald-500" />
                </div>
                <div>
                  <CardTitle>Import from Plane.so</CardTitle>
                  <CardDescription>
                    Upload a PostgreSQL dump file to import projects, tasks, users, and comments
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div class="space-y-4">
                <div class="flex items-center gap-4">
                  <input
                    ref="fileInputRef"
                    type="file"
                    accept=".sql"
                    class="block w-full text-sm text-muted-foreground file:mr-4 file:rounded-md file:border-0 file:bg-primary file:px-4 file:py-2 file:text-sm file:font-medium file:text-primary-foreground hover:file:bg-primary/90 file:cursor-pointer"
                    @change="handleFileSelect"
                  />
                </div>

                <div v-if="selectedFile" class="text-sm text-muted-foreground">
                  Selected: {{ selectedFile.name }} ({{ (selectedFile.size / 1024 / 1024).toFixed(1) }} MB)
                </div>

                <Button
                  @click="handlePlaneImport"
                  :disabled="!selectedFile || importing"
                >
                  <Loader2 v-if="importing" class="mr-2 size-4 animate-spin" />
                  <Upload v-else class="mr-2 size-4" />
                  {{ importing ? 'Importing...' : 'Start Import' }}
                </Button>

                <!-- Import Results -->
                <div v-if="importResult" class="mt-4 rounded-lg border bg-muted/50 p-4">
                  <div class="flex items-center gap-2 mb-3">
                    <CheckCircle2 class="size-5 text-emerald-500" />
                    <p class="font-medium">Import Complete</p>
                  </div>
                  <div class="grid grid-cols-2 gap-3 sm:grid-cols-3">
                    <div class="text-sm">
                      <span class="text-muted-foreground">Users created:</span>
                      <span class="ml-1 font-medium">{{ importResult.users_created }}</span>
                    </div>
                    <div class="text-sm">
                      <span class="text-muted-foreground">Users skipped:</span>
                      <span class="ml-1 font-medium">{{ importResult.users_skipped }}</span>
                    </div>
                    <div class="text-sm">
                      <span class="text-muted-foreground">Projects:</span>
                      <span class="ml-1 font-medium">{{ importResult.projects_created }}</span>
                    </div>
                    <div class="text-sm">
                      <span class="text-muted-foreground">States:</span>
                      <span class="ml-1 font-medium">{{ importResult.states_created }}</span>
                    </div>
                    <div class="text-sm">
                      <span class="text-muted-foreground">Labels:</span>
                      <span class="ml-1 font-medium">{{ importResult.labels_created }}</span>
                    </div>
                    <div class="text-sm">
                      <span class="text-muted-foreground">Tasks:</span>
                      <span class="ml-1 font-medium">{{ importResult.tasks_created }}</span>
                    </div>
                    <div class="text-sm">
                      <span class="text-muted-foreground">Assignees:</span>
                      <span class="ml-1 font-medium">{{ importResult.assignees_linked }}</span>
                    </div>
                    <div class="text-sm">
                      <span class="text-muted-foreground">Labels linked:</span>
                      <span class="ml-1 font-medium">{{ importResult.labels_assigned }}</span>
                    </div>
                    <div class="text-sm">
                      <span class="text-muted-foreground">Comments:</span>
                      <span class="ml-1 font-medium">{{ importResult.comments_created }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
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
