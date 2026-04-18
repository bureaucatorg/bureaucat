<script setup lang="ts">
import { Loader2, Calendar as CalendarIcon, X } from "lucide-vue-next";
import { toast } from "vue-sonner";
import { CalendarDate, type DateValue } from "@internationalized/date";
import type { Module, ModuleStatus, ProjectMember } from "~/types";

const props = defineProps<{
  projectKey: string;
  // If provided, dialog acts as an edit form for this module.
  module?: Module | null;
}>();

const open = defineModel<boolean>("open", { default: false });
const emit = defineEmits<{ saved: [Module] }>();

const { createModule, updateModule } = useModules();
const { members: projectMembers, listMembers } = useProjects();

const isEdit = computed(() => !!props.module);

const loading = ref(false);
const error = ref<string | null>(null);

interface FormShape {
  title: string;
  description: string;
  status: ModuleStatus;
  startDate: DateValue | undefined;
  endDate: DateValue | undefined;
  leadId: string | null;
  memberIds: string[];
}

function emptyForm(): FormShape {
  return {
    title: "",
    description: "",
    status: "backlog",
    startDate: undefined,
    endDate: undefined,
    leadId: null,
    memberIds: [],
  };
}

function isoToCalendar(s?: string): DateValue | undefined {
  if (!s) return undefined;
  const [y, m, d] = s.split("-").map(Number);
  if (!y || !m || !d) return undefined;
  return new CalendarDate(y, m, d);
}

function calendarToIso(d: DateValue | undefined): string | undefined {
  if (!d) return undefined;
  return `${d.year}-${String(d.month).padStart(2, "0")}-${String(d.day).padStart(2, "0")}`;
}

function formatDateValue(d: DateValue | undefined): string {
  return d ? calendarToIso(d)! : "Pick a date (optional)";
}

const form = ref<FormShape>(emptyForm());
const startOpen = ref(false);
const endOpen = ref(false);

function hydrateFrom(module: Module | null | undefined) {
  if (!module) {
    form.value = emptyForm();
    return;
  }
  form.value = {
    title: module.title,
    description: module.description ?? "",
    status: module.status,
    startDate: isoToCalendar(module.start_date),
    endDate: isoToCalendar(module.end_date),
    leadId: module.lead?.user_id ?? null,
    memberIds: (module.members ?? []).map((m) => m.user_id),
  };
}

watch(open, async (isOpen) => {
  if (!isOpen) return;
  error.value = null;
  hydrateFrom(props.module);
  if (!projectMembers.value.length) {
    await listMembers(props.projectKey);
  }
});

const canSubmit = computed(() => !!form.value.title.trim());

async function handleSubmit() {
  if (!canSubmit.value) return;
  error.value = null;

  const start = form.value.startDate as CalendarDate | undefined;
  const end = form.value.endDate as CalendarDate | undefined;
  if (start && end && end.compare(start) < 0) {
    error.value = "End date must be on or after start date";
    return;
  }

  loading.value = true;
  const leadID = form.value.leadId ?? undefined;
  const memberIds = form.value.memberIds;

  const payload = {
    title: form.value.title.trim(),
    description: form.value.description.trim() || undefined,
    status: form.value.status,
    start_date: calendarToIso(start),
    end_date: calendarToIso(end),
    lead_id: leadID,
    member_ids: memberIds,
  };

  const result = isEdit.value
    ? await updateModule(props.projectKey, props.module!.id, {
        ...payload,
        clear_start_date: !payload.start_date,
        clear_end_date: !payload.end_date,
        clear_lead: !payload.lead_id,
      })
    : await createModule(props.projectKey, payload);

  loading.value = false;
  if (result.success && result.data) {
    toast.success(isEdit.value ? "Module updated" : "Module created");
    open.value = false;
    emit("saved", result.data);
  } else {
    error.value = result.error || "Failed to save module";
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent class="sm:max-w-lg">
      <DialogHeader>
        <DialogTitle>{{ isEdit ? "Edit module" : "New module" }}</DialogTitle>
        <DialogDescription>
          A module groups tasks into a reusable sub-project you can track, duplicate, and assign.
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
          <Label for="module_title">Title</Label>
          <Input
            id="module_title"
            v-model="form.title"
            placeholder="Monthly newsletter"
            required
            :disabled="loading"
          />
        </div>

        <div class="space-y-2">
          <Label for="module_desc">Description</Label>
          <Textarea
            id="module_desc"
            v-model="form.description"
            placeholder="What's this module about?"
            rows="3"
            :disabled="loading"
          />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div class="space-y-2">
            <Label>Status</Label>
            <ModuleStatusSelector v-model="form.status" :disabled="loading" />
          </div>
          <MemberSelector
            v-model="form.leadId"
            label="Lead"
            add-label="Pick lead"
            empty-label="No lead"
            :members="(projectMembers as ProjectMember[])"
            :disabled="loading"
          />
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
                    :class="!form.startDate && 'text-muted-foreground'"
                    :disabled="loading"
                  >
                    <CalendarIcon class="size-4" />
                    <span class="truncate">{{ formatDateValue(form.startDate) }}</span>
                  </Button>
                </PopoverTrigger>
                <PopoverContent class="w-auto p-0" align="start">
                  <Calendar v-model="form.startDate" layout="month-and-year" />
                </PopoverContent>
              </Popover>
              <Button
                v-if="form.startDate"
                type="button"
                variant="ghost"
                size="icon"
                class="shrink-0"
                :disabled="loading"
                @click="form.startDate = undefined"
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
                    :class="!form.endDate && 'text-muted-foreground'"
                    :disabled="loading"
                  >
                    <CalendarIcon class="size-4" />
                    <span class="truncate">{{ formatDateValue(form.endDate) }}</span>
                  </Button>
                </PopoverTrigger>
                <PopoverContent class="w-auto p-0" align="start">
                  <Calendar
                    v-model="form.endDate"
                    layout="month-and-year"
                    :min-value="form.startDate"
                  />
                </PopoverContent>
              </Popover>
              <Button
                v-if="form.endDate"
                type="button"
                variant="ghost"
                size="icon"
                class="shrink-0"
                :disabled="loading"
                @click="form.endDate = undefined"
              >
                <X class="size-4" />
              </Button>
            </div>
          </div>
        </div>

        <MemberSelector
          v-model="form.memberIds"
          multi
          label="Initial members"
          add-label="Add member"
          empty-label="No members yet"
          :members="(projectMembers as ProjectMember[])"
          :disabled="loading"
        />

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
            {{ isEdit ? "Save changes" : "Create module" }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
