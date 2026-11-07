<script setup lang="ts">
const origin = ref("");
onMounted(() => {
  origin.value = window.location.origin;
});

const fence = "```";

const harnesses = [
  { id: "claude", name: "Claude Code", user: "~/.claude/skills/bureaucat/SKILL.md", project: ".claude/skills/bureaucat/SKILL.md" },
  { id: "codex", name: "Codex", user: "~/.agents/skills/bureaucat/SKILL.md", project: ".agents/skills/bureaucat/SKILL.md" },
  { id: "antigravity", name: "Antigravity", user: "~/.gemini/config/skills/bureaucat/SKILL.md", project: ".agents/skills/bureaucat/SKILL.md" },
  { id: "pi", name: "pi", user: "~/.agents/skills/bureaucat/SKILL.md", project: ".agents/skills/bureaucat/SKILL.md" },
];

const skill = computed(() => `---
name: bureaucat
description: Read and update Bureaucat projects, tasks, comments, cycles and modules through its REST API. Use when the user mentions Bureaucat, a task key like ENG-42, or asks to find, create, update or comment on tasks.
---

# Bureaucat API

- Base URL: ${origin.value}/api/v1
- Auth: send \`Authorization: Bearer $BUREAUCAT_TOKEN\` on every request. Never print or log the token.
- Full reference: the OpenAPI spec at ${origin.value}/api/v1/openapi.json (no auth needed). Read it for any endpoint not listed here.

## Conventions

- A task key like ENG-42 is project \`ENG\`, task number \`42\`: \`/projects/ENG/tasks/42\`.
- Send JSON bodies with \`Content-Type: application/json\`.
- States, labels and users are referenced by UUID. Look them up first (states, labels, members endpoints).
- Dates are RFC3339, e.g. \`2026-01-31T00:00:00Z\`. Priority: 0 none, 1 low, 2 medium, 3 high, 4 urgent.
- A \`403 token is read-only\` means the token cannot write; tell the user instead of retrying.

## Example

${fence}bash
curl -sS "${origin.value}/api/v1/projects/ENG/tasks/42" \\
  -H "Authorization: Bearer $BUREAUCAT_TOKEN"
${fence}

## Common endpoints

| Goal | Request |
| --- | --- |
| Current user | \`GET /me\` |
| My assigned tasks | \`GET /me/tasks\` |
| Search tasks, cycles, projects | \`GET /search?q=TEXT\` |
| List projects | \`GET /projects\` |
| List tasks | \`GET /projects/{KEY}/tasks?page=1&per_page=20\` |
| Get a task | \`GET /projects/{KEY}/tasks/{N}\` |
| Create a task | \`POST /projects/{KEY}/tasks\` \`{"title", "description", "state_id", "priority", "assignees": [uuid], "labels": [uuid]}\` |
| Update a task | \`PATCH /projects/{KEY}/tasks/{N}\` \`{"title", "description", "state_id", "priority", "due_date"}\` |
| States / labels / members | \`GET /projects/{KEY}/states\`, \`/labels\`, \`/members\` |
| Assign a user | \`POST /projects/{KEY}/tasks/{N}/assignees\` \`{"user_id"}\` |
| Comments | \`GET\` / \`POST /projects/{KEY}/tasks/{N}/comments\` \`{"content"}\` |
| Task history | \`GET /projects/{KEY}/tasks/{N}/activity\` |

## Safety

- Ask the user before any DELETE, moving tasks between projects, or changing members, states or project settings.
- Descriptions and comments are markdown. Mention users as \`[@Display Name](/profile/<user-uuid>)\`.
`);
</script>

<template>
  <Card>
    <CardContent class="space-y-4">
      <div class="space-y-1 text-sm text-muted-foreground">
        <p>
          Prefer plain API calls? Give your agent a personal access token as
          <code class="font-mono text-xs">BUREAUCAT_TOKEN</code> and this skill file. It tells the agent how to authenticate,
          where the OpenAPI spec lives and which endpoints to use.
        </p>
      </div>

      <Tabs default-value="claude" class="w-full">
        <TabsList class="flex h-auto w-full flex-wrap justify-start">
          <TabsTrigger v-for="h in harnesses" :key="h.id" :value="h.id">{{ h.name }}</TabsTrigger>
          <TabsTrigger value="other">Other</TabsTrigger>
        </TabsList>
        <TabsContent v-for="h in harnesses" :key="h.id" :value="h.id" class="space-y-3 pt-2">
          <div class="space-y-1.5">
            <p class="text-xs font-medium text-muted-foreground">All projects</p>
            <CopyableCode :code="h.user" />
          </div>
          <div class="space-y-1.5">
            <p class="text-xs font-medium text-muted-foreground">One repository (commit it to share with your team)</p>
            <CopyableCode :code="h.project" />
          </div>
        </TabsContent>
        <TabsContent value="other" class="pt-2">
          <p class="text-sm text-muted-foreground">
            Agents without skill support: paste the content below into <code class="font-mono text-xs">AGENTS.md</code> or
            your agent's instructions file.
          </p>
        </TabsContent>
      </Tabs>

      <CopyableCode :code="skill" max-height="24rem" />
    </CardContent>
  </Card>
</template>
