<script setup lang="ts">
import { Loader2, Calendar as CalendarIcon, X } from "lucide-vue-next";
import { toast } from "vue-sonner";
import { CalendarDate, type DateValue } from "@internationalized/date";
import type { Module, ModuleTask } from "~/types";

const props = defineProps<{
  projectKey: string;
  source: Module;
  // Tasks currently linked to the source module. Rendered as a checklist.
  sourceTasks: ModuleTask[];
}>();

const open = defineModel<boolean>("open", { default: false });
const emit = defineEmits<{ duplicated: [Module] }>();

const { duplicateModule } = useModules();

const loading = ref(false);
const error = ref<string | null>(null);
const title = ref("");
const startDate = ref<DateValue | undefined>(undefined);
const endDate = ref<DateValue | undefined>(undefined);
const selectedTaskIds = ref<Set<string>>(new Set());
const startOpen = ref(false);
const endOpen = ref(false);

function calendarToIso(d: DateValue | undefined): string | undefined {
  if (!d) return undefined;
  return `${d.year}-${String(d.month).padStart(2, "0")}-${String(d.day).padStart(2, "0")}`;
}

function formatDateValue(d: DateValue | undefined): string {
  return d ? calendarToIso(d)! : "Pick a date (optional)";
}

watch(open, (isOpen) => {
  if (!isOpen) return;
  error.value = null;
  title.value = `${props.source.title} — Copy`;
  startDate.value = undefined;
  endDate.value = undefined;
  selectedTaskIds.value = new Set();
});

function toggleTask(id: string) {
  if (selectedTaskIds.value.has(id)) selectedTaskIds.value.delete(id);
  else selectedTaskIds.value.add(id);
  selectedTaskIds.value = new Set(selectedTaskIds.value);
}

function selectAll() {
  selectedTaskIds.value = new Set(props.sourceTasks.map((t) => t.id));
}
function clearAll() {
  selectedTaskIds.value = new Set();
}

const canSubmit = computed(() => !!title.value.trim());

async function handleSubmit() {
  if (!canSubmit.value) return;
  error.value = null;

  const start = startDate.value as CalendarDate | undefined;
  const end = endDate.value as CalendarDate | undefined;
  if (start && end && end.compare(start) < 0) {
    error.value = "End date must be on or after start date";
    return;
  }

  loading.value = true;
  const result = await duplicateModule(props.projectKey, props.source.id, {
    title: title.value.trim(),
    start_date: calendarToIso(start),
    end_date: calendarToIso(end),
    task_ids: Array.from(selectedTaskIds.value),
  });
  loading.value = false;

  if (result.success && result.data) {
    toast.success("Module duplicated");
    open.value = false;
    emit("duplicated", result.data);
  } else {
    error.value = result.error || "Failed to duplicate module";
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent class="sm:max-w-lg">
      <DialogHeader>
        <DialogTitle>Duplicate module</DialogTitle>
        <DialogDescription>
          The new module will start in <span class="font-medium">backlog</span>.
          Members and lead are copied; pick which tasks to carry over.
        </DialogDescription>
      </DialogHeader>

      <form class="space-y-4" @submit.prevent="handleSubmit">
        <div
          v-if="error"
          class="rounded-md bg-destructive/10 p-3 text-sm text-destructive"
        >
          {{ error }}
        </div>

        <div class="space-y-2">
          <Label for="dup_title">Title</Label>
          <Input id="dup_title" v-model="title" :disabled="loading" />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div class="space-y-2">
            <Label>Start date</Label>
            <div class="flex gap-1">
              <Popover v-model:open="startOpen">
                <PopoverTrigger as-child>
                  <Button
                    type="button"
                    variant="outline"
                    class="flex-1 justify-start gap-2 font-normal"
                    :class="!startDate && 'text-muted-foreground'"
                    :disabled="loading"
                  >
                    <CalendarIcon class="size-4" />
                    <span class="truncate">{{ formatDateValue(startDate) }}</span>
                  </Button>
                </PopoverTrigger>
                <PopoverContent class="w-auto p-0" align="start">
                  <Calendar v-model="startDate" layout="month-and-year" />
                </PopoverContent>
              </Popover>
              <Button
                v-if="startDate"
                type="button"
                variant="ghost"
                size="icon"
                :disabled="loading"
                @click="startDate = undefined"
              >
                <X class="size-4" />
              </Button>
            </div>
          </div>
          <div class="space-y-2">
            <Label>End date</Label>
            <div class="flex gap-1">
              <Popover v-model:open="endOpen">
                <PopoverTrigger as-child>
                  <Button
                    type="button"
                    variant="outline"
                    class="flex-1 justify-start gap-2 font-normal"
                    :class="!endDate && 'text-muted-foreground'"
                    :disabled="loading"
                  >
                    <CalendarIcon class="size-4" />
                    <span class="truncate">{{ formatDateValue(endDate) }}</span>
                  </Button>
                </PopoverTrigger>
                <PopoverContent class="w-auto p-0" align="start">
                  <Calendar v-model="endDate" layout="month-and-year" :min-value="startDate" />
                </PopoverContent>
              </Popover>
              <Button
                v-if="endDate"
                type="button"
                variant="ghost"
                size="icon"
                :disabled="loading"
                @click="endDate = undefined"
              >
                <X class="size-4" />
              </Button>
            </div>
          </div>
        </div>

        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <Label>Tasks to carry over ({{ selectedTaskIds.size }} / {{ sourceTasks.length }})</Label>
            <div class="flex gap-1 text-xs">
              <button type="button" class="text-primary hover:underline" @click="selectAll">
                Select all
              </button>
              <span class="text-muted-foreground">·</span>
              <button type="button" class="text-primary hover:underline" @click="clearAll">
                Clear
              </button>
            </div>
          </div>
          <div class="max-h-60 overflow-y-auto rounded-md border">
            <label
              v-for="task in sourceTasks"
              :key="task.id"
              class="flex cursor-pointer items-center gap-2 border-b border-border/40 px-3 py-1.5 text-sm last:border-0 hover:bg-muted/40"
            >
              <Checkbox
                :model-value="selectedTaskIds.has(task.id)"
                @update:model-value="toggleTask(task.id)"
              />
              <span class="font-mono text-[11px] text-muted-foreground">{{ task.task_id }}</span>
              <span class="min-w-0 truncate">{{ task.title }}</span>
            </label>
            <div v-if="!sourceTasks.length" class="px-3 py-3 text-center text-xs text-muted-foreground">
              No tasks linked to the source module.
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            :disabled="loading"
            @click="open = false"
          >
            Cancel
          </Button>
          <Button type="submit" :disabled="loading || !canSubmit">
            <Loader2 v-if="loading" class="mr-2 size-4 animate-spin" />
            Duplicate
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
