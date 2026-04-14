<script setup lang="ts">
import { useEditor, EditorContent } from "@tiptap/vue-3";
import StarterKit from "@tiptap/starter-kit";
import Link from "@tiptap/extension-link";
import type { Editor } from "@tiptap/vue-3";
import type { ProjectMember } from "~/types";
import {
  Bold,
  Italic,
  Strikethrough,
  Code,
  Heading1,
  Heading2,
  Heading3,
  List,
  ListOrdered,
  Quote,
  CodeSquare,
  Minus,
  Undo,
  Redo,
  Paperclip,
  Loader2,
} from "lucide-vue-next";

const props = defineProps<{
  modelValue: string;
  disabled?: boolean;
  uploading?: boolean;
  compact?: boolean;
  members?: ProjectMember[];
}>();

const emit = defineEmits<{
  "update:modelValue": [value: string];
  "files-dropped": [files: File[]];
}>();

const fileInputRef = ref<HTMLInputElement | null>(null);
const wrapperRef = ref<HTMLElement | null>(null);

// --- @mention state ---
const showMentions = ref(false);
const mentionQuery = ref("");
const mentionFrom = ref(0); // ProseMirror position of the "@"
const mentionTo = ref(0); // ProseMirror position just after the query
const mentionCoords = ref<{ top: number; left: number }>({ top: 0, left: 0 });
const highlightedMentionIndex = ref(0);

const filteredMembers = computed(() => {
  const list = props.members ?? [];
  if (!list.length) return [];
  const q = mentionQuery.value.toLowerCase();
  if (!q) return list;
  return list.filter(
    (m) =>
      m.first_name.toLowerCase().includes(q) ||
      m.last_name.toLowerCase().includes(q) ||
      m.username.toLowerCase().includes(q)
  );
});

function updateMentionState(editorInstance: Editor) {
  if (!props.members?.length) {
    showMentions.value = false;
    return;
  }
  const { from, empty } = editorInstance.state.selection;
  if (!empty) {
    showMentions.value = false;
    return;
  }
  // Look back up to 50 chars from cursor to find an in-progress "@query".
  const lookback = 50;
  const start = Math.max(0, from - lookback);
  const before = editorInstance.state.doc.textBetween(start, from, "\n", "\n");
  const match = before.match(/(?:^|\s)@([\p{L}\p{N}_.-]{0,30})$/u);
  if (!match) {
    showMentions.value = false;
    return;
  }
  const query = match[1] ?? "";
  // Compute ProseMirror pos of the "@" — account for whether match started with whitespace.
  const matchedLen = match[0].length;
  const leadingWs = match[0].startsWith("@") ? 0 : 1;
  const atDocPos = from - (matchedLen - leadingWs);
  mentionFrom.value = atDocPos;
  mentionTo.value = from;
  mentionQuery.value = query;
  highlightedMentionIndex.value = 0;
  showMentions.value = true;

  // Position the popup just below the "@".
  try {
    const coords = editorInstance.view.coordsAtPos(atDocPos);
    const wrap = wrapperRef.value?.getBoundingClientRect();
    if (wrap) {
      mentionCoords.value = {
        top: coords.bottom - wrap.top + 4,
        left: coords.left - wrap.left,
      };
    }
  } catch {
    // ignore coord errors (can happen mid-transaction)
  }
}

function insertMention(member: ProjectMember) {
  const ed = editor.value;
  if (!ed) return;
  const displayName = `${member.first_name} ${member.last_name}`;
  ed.chain()
    .focus()
    .deleteRange({ from: mentionFrom.value, to: mentionTo.value })
    .insertContent([
      {
        type: "text",
        text: `@${displayName}`,
        marks: [
          {
            type: "link",
            attrs: { href: `/profile/${member.user_id}` },
          },
        ],
      },
      { type: "text", text: " " },
    ])
    .run();
  showMentions.value = false;
}

function handleMentionKeydown(event: KeyboardEvent): boolean {
  if (!showMentions.value || filteredMembers.value.length === 0) return false;
  if (event.key === "ArrowDown") {
    event.preventDefault();
    highlightedMentionIndex.value =
      (highlightedMentionIndex.value + 1) % filteredMembers.value.length;
    return true;
  }
  if (event.key === "ArrowUp") {
    event.preventDefault();
    highlightedMentionIndex.value =
      (highlightedMentionIndex.value - 1 + filteredMembers.value.length) %
      filteredMembers.value.length;
    return true;
  }
  if (event.key === "Enter" || event.key === "Tab") {
    event.preventDefault();
    const m = filteredMembers.value[highlightedMentionIndex.value];
    if (m) insertMention(m);
    return true;
  }
  if (event.key === "Escape") {
    event.preventDefault();
    showMentions.value = false;
    return true;
  }
  return false;
}

const editor = useEditor({
  content: props.modelValue,
  editable: !props.disabled,
  extensions: [
    StarterKit.configure({
      heading: { levels: [1, 2, 3] },
    }),
    Link.configure({
      openOnClick: false,
      autolink: false,
      HTMLAttributes: {
        class: "text-primary underline underline-offset-2",
      },
    }),
  ],
  editorProps: {
    attributes: {
      class: `prose prose-sm max-w-none dark:prose-invert focus:outline-none px-3 py-2 ${props.compact ? "min-h-[72px]" : "min-h-[200px]"}`,
    },
    handleKeyDown: (_view, event) => {
      return handleMentionKeydown(event);
    },
    handleDrop: (_view, event, _slice, moved) => {
      if (moved || !event.dataTransfer?.files.length) return false;
      event.preventDefault();
      emit("files-dropped", Array.from(event.dataTransfer.files));
      return true;
    },
    handlePaste: (_view, event) => {
      const files = Array.from(event.clipboardData?.files || []);
      if (files.length > 0) {
        event.preventDefault();
        emit("files-dropped", files);
        return true;
      }
      return false;
    },
  },
  onUpdate: ({ editor }) => {
    emit("update:modelValue", editor.getHTML());
    updateMentionState(editor);
  },
  onSelectionUpdate: ({ editor }) => {
    updateMentionState(editor);
  },
  onBlur: () => {
    // Delay so button mousedown can fire first.
    setTimeout(() => {
      showMentions.value = false;
    }, 150);
  },
});

watch(
  () => props.disabled,
  (val) => {
    editor.value?.setEditable(!val);
  }
);

// Sync external modelValue changes (e.g., parent resetting content after submit)
// into the editor. Skip when the value already matches to avoid ping-pong with onUpdate.
watch(
  () => props.modelValue,
  (val) => {
    if (!editor.value) return;
    if (editor.value.getHTML() === val) return;
    editor.value.commands.setContent(val || "", false);
  }
);

onBeforeUnmount(() => {
  editor.value?.destroy();
});

function isActive(name: string, attrs?: Record<string, unknown>) {
  return editor.value?.isActive(name, attrs) ?? false;
}

function openFilePicker() {
  fileInputRef.value?.click();
}

function handleFileInput(e: Event) {
  const input = e.target as HTMLInputElement;
  if (input.files?.length) {
    emit("files-dropped", Array.from(input.files));
  }
  input.value = "";
}
</script>

<template>
  <div ref="wrapperRef" class="tiptap-editor relative rounded-md border border-input bg-background">
    <!-- Toolbar -->
    <div
      v-if="editor"
      class="flex flex-wrap items-center gap-0.5 border-b border-input px-1.5 py-1"
    >
      <button
        type="button"
        aria-label="Bold"
        class="inline-flex size-7 items-center justify-center rounded-md hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
        :class="{ 'bg-muted text-foreground': isActive('bold') }"
        @click="editor!.chain().focus().toggleBold().run()"
      >
        <Bold class="size-3.5" />
      </button>
      <button
        type="button"
        aria-label="Italic"
        class="inline-flex size-7 items-center justify-center rounded-md hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
        :class="{ 'bg-muted text-foreground': isActive('italic') }"
        @click="editor!.chain().focus().toggleItalic().run()"
      >
        <Italic class="size-3.5" />
      </button>
      <button
        type="button"
        aria-label="Strikethrough"
        class="inline-flex size-7 items-center justify-center rounded-md hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
        :class="{ 'bg-muted text-foreground': isActive('strike') }"
        @click="editor!.chain().focus().toggleStrike().run()"
      >
        <Strikethrough class="size-3.5" />
      </button>
      <button
        type="button"
        aria-label="Inline code"
        class="inline-flex size-7 items-center justify-center rounded-md hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
        :class="{ 'bg-muted text-foreground': isActive('code') }"
        @click="editor!.chain().focus().toggleCode().run()"
      >
        <Code class="size-3.5" />
      </button>

      <div class="mx-1 h-4 w-px bg-border" role="separator" />

      <button
        type="button"
        aria-label="Heading 1"
        class="inline-flex size-7 items-center justify-center rounded-md hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
        :class="{ 'bg-muted text-foreground': isActive('heading', { level: 1 }) }"
        @click="editor!.chain().focus().toggleHeading({ level: 1 }).run()"
      >
        <Heading1 class="size-3.5" />
      </button>
      <button
        type="button"
        aria-label="Heading 2"
        class="inline-flex size-7 items-center justify-center rounded-md hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
        :class="{ 'bg-muted text-foreground': isActive('heading', { level: 2 }) }"
        @click="editor!.chain().focus().toggleHeading({ level: 2 }).run()"
      >
        <Heading2 class="size-3.5" />
      </button>
      <button
        type="button"
        aria-label="Heading 3"
        class="inline-flex size-7 items-center justify-center rounded-md hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
        :class="{ 'bg-muted text-foreground': isActive('heading', { level: 3 }) }"
        @click="editor!.chain().focus().toggleHeading({ level: 3 }).run()"
      >
        <Heading3 class="size-3.5" />
      </button>

      <div class="mx-1 h-4 w-px bg-border" role="separator" />

      <button
        type="button"
        aria-label="Bullet list"
        class="inline-flex size-7 items-center justify-center rounded-md hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
        :class="{ 'bg-muted text-foreground': isActive('bulletList') }"
        @click="editor!.chain().focus().toggleBulletList().run()"
      >
        <List class="size-3.5" />
      </button>
      <button
        type="button"
        aria-label="Ordered list"
        class="inline-flex size-7 items-center justify-center rounded-md hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
        :class="{ 'bg-muted text-foreground': isActive('orderedList') }"
        @click="editor!.chain().focus().toggleOrderedList().run()"
      >
        <ListOrdered class="size-3.5" />
      </button>
      <button
        type="button"
        aria-label="Blockquote"
        class="inline-flex size-7 items-center justify-center rounded-md hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
        :class="{ 'bg-muted text-foreground': isActive('blockquote') }"
        @click="editor!.chain().focus().toggleBlockquote().run()"
      >
        <Quote class="size-3.5" />
      </button>
      <button
        type="button"
        aria-label="Code block"
        class="inline-flex size-7 items-center justify-center rounded-md hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
        :class="{ 'bg-muted text-foreground': isActive('codeBlock') }"
        @click="editor!.chain().focus().toggleCodeBlock().run()"
      >
        <CodeSquare class="size-3.5" />
      </button>

      <div class="mx-1 h-4 w-px bg-border" role="separator" />

      <button
        type="button"
        aria-label="Horizontal rule"
        class="inline-flex size-7 items-center justify-center rounded-md hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
        @click="editor!.chain().focus().setHorizontalRule().run()"
      >
        <Minus class="size-3.5" />
      </button>

      <div class="mx-1 h-4 w-px bg-border" role="separator" />

      <button
        type="button"
        aria-label="Attach file"
        class="inline-flex size-7 items-center justify-center rounded-md hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
        :disabled="uploading"
        @click="openFilePicker"
      >
        <Loader2 v-if="uploading" class="size-3.5 animate-spin" />
        <Paperclip v-else class="size-3.5" />
      </button>

      <div class="mx-1 h-4 w-px bg-border" role="separator" />

      <button
        type="button"
        aria-label="Undo"
        class="inline-flex size-7 items-center justify-center rounded-md hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none disabled:opacity-40"
        :disabled="!editor!.can().undo()"
        @click="editor!.chain().focus().undo().run()"
      >
        <Undo class="size-3.5" />
      </button>
      <button
        type="button"
        aria-label="Redo"
        class="inline-flex size-7 items-center justify-center rounded-md hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none disabled:opacity-40"
        :disabled="!editor!.can().redo()"
        @click="editor!.chain().focus().redo().run()"
      >
        <Redo class="size-3.5" />
      </button>
    </div>

    <!-- Editor content -->
    <EditorContent :editor="editor" />

    <!-- Hidden file input -->
    <input
      ref="fileInputRef"
      type="file"
      multiple
      accept="*/*"
      class="hidden"
      @change="handleFileInput"
    />

    <!-- @mention dropdown -->
    <div
      v-if="showMentions && filteredMembers.length > 0"
      class="absolute z-50 w-60 rounded-md border bg-popover shadow-md"
      :style="{ top: `${mentionCoords.top}px`, left: `${mentionCoords.left}px` }"
    >
      <div class="max-h-48 overflow-y-auto py-1">
        <button
          v-for="(member, idx) in filteredMembers"
          :key="member.user_id"
          type="button"
          class="flex w-full items-center gap-2 px-3 py-1.5 text-sm transition-colors"
          :class="idx === highlightedMentionIndex ? 'bg-accent text-accent-foreground' : 'hover:bg-accent hover:text-accent-foreground'"
          @mousedown.prevent="insertMention(member)"
          @mouseenter="highlightedMentionIndex = idx"
        >
          <Avatar class="size-6">
            <AvatarFallback class="text-xs" :seed="member.user_id">
              {{ member.first_name[0] }}{{ member.last_name[0] }}
            </AvatarFallback>
          </Avatar>
          <span class="truncate">{{ member.first_name }} {{ member.last_name }}</span>
          <span class="ml-auto truncate text-xs text-muted-foreground">@{{ member.username }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style>
.tiptap-editor .tiptap p.is-editor-empty:first-child::before {
  content: "Add a description...";
  float: left;
  color: var(--muted-foreground);
  pointer-events: none;
  height: 0;
}
</style>
