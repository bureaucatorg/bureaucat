<script setup lang="ts">
import { Search } from "lucide-vue-next";
import type { ProjectMember } from "~/types";

interface MentionEntry {
  displayText: string;
  userId: string;
}

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
const mentions = ref<MentionEntry[]>([]);

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
  const textBefore = value.slice(0, cursorPos);
  const atIdx = textBefore.lastIndexOf("@");

  if (atIdx >= 0) {
    const charBefore = atIdx > 0 ? textBefore[atIdx - 1] : " ";
    if (charBefore === " " || charBefore === "\n" || atIdx === 0) {
      const query = textBefore.slice(atIdx + 1);
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
  const displayName = `${member.first_name} ${member.last_name}`;
  const displayText = `@${displayName}`;

  // Track this mention for markdown conversion later
  const existing = mentions.value.find((m) => m.userId === member.user_id);
  if (!existing) {
    mentions.value.push({ displayText: displayName, userId: member.user_id });
  }

  // Replace @query with readable @Name
  const before = value.slice(0, mentionStartPos.value);
  const after = value.slice(mentionStartPos.value + 1 + mentionQuery.value.length);
  const newValue = before + displayText + " " + after;

  emit("update:modelValue", newValue);
  showMentions.value = false;

  nextTick(() => {
    textarea.focus();
    const newPos = before.length + displayText.length + 1;
    textarea.setSelectionRange(newPos, newPos);
  });
}

/** Convert display text to markdown with mention links */
function getMarkdownContent(): string {
  let content = props.modelValue;
  for (const mention of mentions.value) {
    const displayText = `@${mention.displayText}`;
    const markdownLink = `[@${mention.displayText}](/profile/${mention.userId})`;
    content = content.replaceAll(displayText, markdownLink);
  }
  return content;
}

/** Clear tracked mentions (call after successful submit) */
function clearMentions() {
  mentions.value = [];
}

defineExpose({ getMarkdownContent, clearMentions });

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
