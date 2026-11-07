<script setup lang="ts">
import { Check, Copy } from "lucide-vue-next";

interface Snippet {
  label: string;
  hint: string;
  code: string;
}

interface Client {
  id: string;
  name: string;
  snippets: Snippet[];
}

const url = ref("/mcp");
onMounted(() => {
  url.value = `${window.location.origin}/mcp`;
});

const clients = computed<Client[]>(() => [
  {
    id: "claude",
    name: "Claude Code",
    snippets: [
      {
        label: "Terminal",
        hint: "Add --scope user to use it in every project.",
        code: `claude mcp add --transport http bureaucat ${url.value} \\\n  --header "Authorization: Bearer $BUREAUCAT_TOKEN"`,
      },
    ],
  },
  {
    id: "codex",
    name: "Codex",
    snippets: [
      {
        label: "Terminal",
        hint: "Codex reads the token from BUREAUCAT_TOKEN when it starts.",
        code: `codex mcp add bureaucat --url ${url.value} --bearer-token-env-var BUREAUCAT_TOKEN`,
      },
      {
        label: "Or ~/.codex/config.toml",
        hint: "",
        code: `[mcp_servers.bureaucat]\nurl = "${url.value}"\nbearer_token_env_var = "BUREAUCAT_TOKEN"`,
      },
    ],
  },
  {
    id: "antigravity",
    name: "Antigravity",
    snippets: [
      {
        label: "mcp_config.json",
        hint: "Agent panel → … → MCP Servers → Manage MCP Servers → View raw config.",
        code: JSON.stringify(
          { mcpServers: { bureaucat: { serverUrl: url.value, headers: { Authorization: "Bearer <YOUR_TOKEN>" } } } },
          null,
          2,
        ),
      },
    ],
  },
  {
    id: "pi",
    name: "pi",
    snippets: [
      {
        label: "Terminal",
        hint: "Run pi mcp list to check the connection.",
        code: `pi mcp add bureaucat --url ${url.value} --bearer-token-env-var BUREAUCAT_TOKEN`,
      },
      {
        label: "Or ~/.pi/agent/mcp.json",
        hint: "",
        code: JSON.stringify(
          { mcpServers: { bureaucat: { url: url.value, headers: { Authorization: "Bearer ${BUREAUCAT_TOKEN}" } } } },
          null,
          2,
        ),
      },
    ],
  },
  {
    id: "generic",
    name: "Other",
    snippets: [
      {
        label: "Any MCP client",
        hint: "Transport: Streamable HTTP. Auth: Authorization header with a personal access token.",
        code: JSON.stringify(
          { mcpServers: { bureaucat: { type: "http", url: url.value, headers: { Authorization: "Bearer <YOUR_TOKEN>" } } } },
          null,
          2,
        ),
      },
    ],
  },
]);

const copiedKey = ref<string | null>(null);

async function copy(key: string, text: string) {
  try {
    await navigator.clipboard.writeText(text);
    copiedKey.value = key;
    setTimeout(() => (copiedKey.value = null), 2000);
  } catch {
    // clipboard unavailable; the snippet is still selectable
  }
}
</script>

<template>
  <Card>
    <CardContent class="space-y-4 pt-6">
      <div class="space-y-1 text-sm">
        <p>
          Server URL:
          <code class="break-all rounded bg-muted px-1.5 py-0.5 font-mono text-xs">{{ url }}</code>
        </p>
        <p class="text-muted-foreground">
          Create a personal access token above and set it as <code class="font-mono text-xs">BUREAUCAT_TOKEN</code> (or paste it
          in place of <code class="font-mono text-xs">&lt;YOUR_TOKEN&gt;</code>). A read-only token exposes only read tools.
        </p>
      </div>

      <Tabs default-value="claude" class="w-full">
        <TabsList class="flex h-auto w-full flex-wrap justify-start">
          <TabsTrigger v-for="c in clients" :key="c.id" :value="c.id">{{ c.name }}</TabsTrigger>
        </TabsList>
        <TabsContent v-for="c in clients" :key="c.id" :value="c.id" class="space-y-4 pt-2">
          <div v-for="(s, i) in c.snippets" :key="i" class="space-y-1.5">
            <p class="text-xs font-medium text-muted-foreground">{{ s.label }}</p>
            <div class="relative">
              <pre class="overflow-x-auto rounded-md border bg-muted/50 p-3 pr-12 font-mono text-xs leading-relaxed">{{ s.code }}</pre>
              <Button
                variant="ghost"
                size="icon"
                class="absolute right-1.5 top-1.5 size-7"
                :aria-label="copiedKey === `${c.id}-${i}` ? 'Copied' : 'Copy snippet'"
                @click="copy(`${c.id}-${i}`, s.code)"
              >
                <Check v-if="copiedKey === `${c.id}-${i}`" class="size-3.5 text-green-500" />
                <Copy v-else class="size-3.5" />
              </Button>
            </div>
            <p v-if="s.hint" class="text-xs text-muted-foreground">{{ s.hint }}</p>
          </div>
        </TabsContent>
      </Tabs>
    </CardContent>
  </Card>
</template>
