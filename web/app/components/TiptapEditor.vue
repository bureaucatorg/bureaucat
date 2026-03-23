<script setup lang="ts">
import { useEditor, EditorContent } from "@tiptap/vue-3";
import StarterKit from "@tiptap/starter-kit";
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
} from "lucide-vue-next";

const props = defineProps<{
  modelValue: string;
  disabled?: boolean;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

const editor = useEditor({
  content: props.modelValue,
  editable: !props.disabled,
  extensions: [
    StarterKit.configure({
      heading: { levels: [1, 2, 3] },
    }),
  ],
  editorProps: {
    attributes: {
      class: "prose prose-sm max-w-none dark:prose-invert focus:outline-none min-h-[200px] px-3 py-2",
    },
  },
  onUpdate: ({ editor }) => {
    emit("update:modelValue", editor.getHTML());
  },
});

watch(
  () => props.disabled,
  (val) => {
    editor.value?.setEditable(!val);
  }
);

onBeforeUnmount(() => {
  editor.value?.destroy();
});

function isActive(name: string, attrs?: Record<string, unknown>) {
  return editor.value?.isActive(name, attrs) ?? false;
}
</script>

<template>
  <div class="tiptap-editor rounded-md border border-input bg-background">
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
