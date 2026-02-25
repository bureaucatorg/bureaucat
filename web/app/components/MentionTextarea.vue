<script setup lang="ts">
import { Search } from "lucide-vue-next";
import type { ProjectMember } from "~/types";

const props = defineProps<{
  modelValue: string;
  placeholder?: string;
  rows?: number;
  disabled?: boolean;
  members: ProjectMember[];
}>();

const emit = defineEmits<{
  "update:modelValue": [value: string];
  keydown: [event: KeyboardEvent];
}>();

const textareaRef = ref<{ $el?: HTMLTextAreaElement } | null>(null);
const showMentions = ref(false);
const mentionQuery = ref("");
const mentionStartPos = ref(0);
const popoverPos = ref({ top: 0, left: 0 });

const filteredMembers = computed(() => {
  const q = mentionQuery.value.toLowerCase();
  if (!q) return props.members;
  return props.members.filter(
    (m) =>
      m.first_name.toLowerCase().includes(q) ||
      m.last_name.toLowerCase().includes(q) ||
      m.username.toLowerCase().includes(q)
  );
});

function getTextarea(): HTMLTextAreaElement | null {
  const el = textareaRef.value;
  if (!el) return null;
  // shadcn Textarea wraps a native textarea
  if (el.$el && el.$el.tagName === "TEXTAREA") return el.$el;
  const native = (el as any)?.$el?.querySelector?.("textarea");
  if (native) return native;
  return el as any;
}

function handleInput(event: Event) {
  const target = event.target as HTMLTextAreaElement;
  const value = target.value;
  emit("update:modelValue", value);

  const cursorPos = target.selectionStart;
  // Check if we just typed '@' or are mid-mention
  const textBefore = value.slice(0, cursorPos);
  const atIdx = textBefore.lastIndexOf("@");

  if (atIdx >= 0) {
    // Check char before @ is whitespace or start of string
    const charBefore = atIdx > 0 ? textBefore[atIdx - 1] : " ";
    if (charBefore === " " || charBefore === "\n" || atIdx === 0) {
      const query = textBefore.slice(atIdx + 1);
      // Only show if no space in query (still typing a mention)
      if (!query.includes(" ") && query.length <= 30) {
        mentionQuery.value = query;
        mentionStartPos.value = atIdx;
        showMentions.value = true;
        return;
      }
    }
  }
  showMentions.value = false;
}

function selectMember(member: ProjectMember) {
  const textarea = getTextarea();
  if (!textarea) return;

  const value = props.modelValue;
  const name = `${member.first_name} ${member.last_name}`;
  const link = `[@${name}](/profile/${member.user_id})`;
  // Replace @query with the link
  const before = value.slice(0, mentionStartPos.value);
  const after = value.slice(mentionStartPos.value + 1 + mentionQuery.value.length);
  const newValue = before + link + " " + after;

  emit("update:modelValue", newValue);
  showMentions.value = false;

  // Restore focus and cursor position
  nextTick(() => {
    textarea.focus();
    const newPos = before.length + link.length + 1;
    textarea.setSelectionRange(newPos, newPos);
  });
}

function handleKeydown(event: KeyboardEvent) {
  if (showMentions.value && event.key === "Escape") {
    showMentions.value = false;
    event.stopPropagation();
    return;
  }
  emit("keydown", event);
}
</script>

<template>
  <div class="relative">
    <Textarea
      ref="textareaRef"
      :model-value="modelValue"
      :placeholder="placeholder"
      :rows="rows"
      :disabled="disabled"
      @input="handleInput"
      @keydown="handleKeydown"
    />

    <!-- Mention autocomplete dropdown -->
    <div
      v-if="showMentions && filteredMembers.length > 0"
      class="absolute bottom-full left-0 z-50 mb-1 w-56 rounded-md border bg-popover shadow-md"
    >
      <div class="max-h-48 overflow-y-auto py-1">
        <button
          v-for="member in filteredMembers"
          :key="member.user_id"
          type="button"
          class="flex w-full items-center gap-2 px-3 py-1.5 text-sm hover:bg-accent"
          @mousedown.prevent="selectMember(member)"
        >
          <Avatar class="size-6">
            <AvatarFallback class="text-xs">
              {{ member.first_name[0] }}{{ member.last_name[0] }}
            </AvatarFallback>
          </Avatar>
          {{ member.first_name }} {{ member.last_name }}
        </button>
      </div>
    </div>
  </div>
</template>
