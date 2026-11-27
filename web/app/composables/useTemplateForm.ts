import type { Ref } from "vue";
import type {
  ProjectLabel,
  ProjectMember,
  TaskTemplate,
  TemplateField,
  TemplateFieldValue,
} from "~/types";
import { mdToHtml } from "~/utils/markdown";

interface TemplateTargetForm {
  title: string;
  description: string;
  priority: number;
  labels: string[];
  assignees: string[];
}

function fieldError(field: TemplateField, value: TemplateFieldValue): string | null {
  const empty =
    value === undefined ||
    value === null ||
    (typeof value === "string" && value.trim() === "") ||
    (Array.isArray(value) && value.length === 0);
  if (empty) return field.required ? "Required" : null;
  if (field.type === "number") {
    const n = Number(value);
    if (!Number.isFinite(n)) return "Must be a number";
    if (field.min != null && n < field.min) return `Must be at least ${field.min}`;
    if (field.max != null && n > field.max) return `Must be at most ${field.max}`;
  }
  return null;
}

function cleanValue(field: TemplateField, value: TemplateFieldValue) {
  if (value === undefined || value === null) return undefined;
  if (Array.isArray(value)) return value.length ? value : undefined;
  if (field.type === "number") {
    if (typeof value === "string" && value.trim() === "") return undefined;
    return Number(value);
  }
  const s = String(value).trim();
  return s ? s : undefined;
}

export function useTemplateForm(opts: {
  selectedTemplateId: Ref<string>;
  templates: Ref<TaskTemplate[]>;
  labels: Ref<ProjectLabel[]>;
  members: Ref<ProjectMember[]>;
  form: Ref<TemplateTargetForm>;
}) {
  const fieldValues = ref<Record<string, TemplateFieldValue>>({});
  const fieldErrors = ref<Record<string, string>>({});

  const selectedTemplate = computed(
    () => opts.templates.value.find((t) => t.id === opts.selectedTemplateId.value) ?? null
  );
  const templateFields = computed(() => selectedTemplate.value?.fields ?? []);
  const hasFields = computed(() => templateFields.value.length > 0);

  watch(opts.selectedTemplateId, (id) => {
    fieldValues.value = {};
    fieldErrors.value = {};
    const tmpl = opts.templates.value.find((t) => t.id === id);
    if (!tmpl) return;
    const form = opts.form.value;
    form.title = tmpl.title;
    // Structured templates show their description as guidance, so the editor only holds extra details.
    form.description = tmpl.fields.length ? "" : mdToHtml(tmpl.description);
    if (tmpl.default_priority != null) form.priority = tmpl.default_priority;
    const labelIds = new Set(opts.labels.value.map((l) => l.id));
    const memberIds = new Set(opts.members.value.map((m) => m.user_id));
    form.labels = tmpl.default_labels.filter((id) => labelIds.has(id));
    form.assignees = tmpl.default_assignees.filter((id) => memberIds.has(id));
  });

  watch(
    fieldValues,
    (vals) => {
      for (const key of Object.keys(fieldErrors.value)) {
        const field = templateFields.value.find((f) => f.key === key);
        if (field && !fieldError(field, vals[key])) delete fieldErrors.value[key];
      }
    },
    { deep: true }
  );

  function validate(): boolean {
    const errors: Record<string, string> = {};
    for (const f of templateFields.value) {
      const err = fieldError(f, fieldValues.value[f.key]);
      if (err) errors[f.key] = err;
    }
    fieldErrors.value = errors;
    return Object.keys(errors).length === 0;
  }

  // Labels and assignees are always sent with a template so clearing a default sticks.
  function payload() {
    const tmpl = selectedTemplate.value;
    if (!tmpl) return {};
    const fields: Record<string, string | number | string[]> = {};
    for (const f of tmpl.fields) {
      const v = cleanValue(f, fieldValues.value[f.key]);
      if (v !== undefined) fields[f.key] = v;
    }
    return {
      template_id: tmpl.id,
      fields: tmpl.fields.length ? fields : undefined,
      labels: opts.form.value.labels,
      assignees: opts.form.value.assignees,
    };
  }

  return {
    selectedTemplate,
    templateFields,
    hasFields,
    fieldValues,
    fieldErrors,
    validate,
    payload,
  };
}
