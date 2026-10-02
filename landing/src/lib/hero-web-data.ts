export type WebNode = { id: string; x: number; y: number; user?: boolean };

export const nodes: WebNode[] = [
  { id: "AR", x: 260, y: 150, user: true },
  { id: "KM", x: 920, y: 150, user: true },
  { id: "SN", x: 1330, y: 340, user: true },
  { id: "LT", x: 280, y: 710, user: true },
  { id: "RM", x: 1000, y: 820, user: true },
  { id: "DEVOPS-780", x: 110, y: 330 },
  { id: "QA-133", x: 370, y: 240 },
  { id: "WEB-310", x: 430, y: 95 },
  { id: "API-112", x: 660, y: 190 },
  { id: "AUTH-19", x: 700, y: 85 },
  { id: "MOB-245", x: 1130, y: 90 },
  { id: "HR-12", x: 980, y: 250 },
  { id: "UI-402", x: 1230, y: 200 },
  { id: "DATA-77", x: 1380, y: 120 },
  { id: "BILL-88", x: 1380, y: 500 },
  { id: "PAY-260", x: 1370, y: 630 },
  { id: "SEC-91", x: 1250, y: 760 },
  { id: "LEGAL-6", x: 1290, y: 860 },
  { id: "NET-73", x: 1110, y: 690 },
  { id: "DOCS-18", x: 820, y: 720 },
  { id: "DESIGN-55", x: 850, y: 640 },
  { id: "SALES-140", x: 620, y: 700 },
  { id: "INFRA-204", x: 500, y: 820 },
  { id: "CI-301", x: 130, y: 850 },
  { id: "OPS-57", x: 80, y: 560 },
];

export const assigned: [string, string][] = [
  ["AR", "DEVOPS-780"], ["AR", "QA-133"], ["AR", "WEB-310"], ["AR", "API-112"],
  ["AR", "AUTH-19"], ["AR", "OPS-57"],
  ["KM", "AUTH-19"], ["KM", "API-112"], ["KM", "MOB-245"], ["KM", "HR-12"],
  ["KM", "UI-402"], ["KM", "WEB-310"], ["KM", "DESIGN-55"],
  ["SN", "DATA-77"], ["SN", "UI-402"], ["SN", "BILL-88"], ["SN", "PAY-260"],
  ["SN", "MOB-245"], ["SN", "SEC-91"], ["SN", "HR-12"],
  ["LT", "OPS-57"], ["LT", "DEVOPS-780"], ["LT", "CI-301"], ["LT", "INFRA-204"],
  ["LT", "SALES-140"], ["LT", "QA-133"],
  ["RM", "DOCS-18"], ["RM", "NET-73"], ["RM", "LEGAL-6"], ["RM", "SEC-91"],
  ["RM", "DESIGN-55"], ["RM", "SALES-140"], ["RM", "PAY-260"],
];

export const linked: [string, string][] = [
  ["DEVOPS-780", "QA-133"], ["WEB-310", "API-112"], ["API-112", "AUTH-19"],
  ["MOB-245", "UI-402"], ["BILL-88", "PAY-260"], ["INFRA-204", "CI-301"],
  ["DOCS-18", "DESIGN-55"], ["NET-73", "SEC-91"],
];

export const STATES = {
  backlog: { name: "Backlog", color: "#6B7280" },
  todo: { name: "Todo", color: "#3B82F6" },
  approval: { name: "Approval Pending", color: "#F59E0B" },
  progress: { name: "In Progress", color: "#10B981" },
  blocked: { name: "Blocked", color: "#EF4444" },
  testing: { name: "Testing", color: "#8B5CF6" },
  done: { name: "Done", color: "#22C55E" },
} as const;

export const PRIORITIES = {
  low: { name: "Low", color: "#3B82F6", bars: 1 },
  medium: { name: "Medium", color: "#EAB308", bars: 2 },
  high: { name: "High", color: "#F97316", bars: 3 },
  urgent: { name: "Urgent", color: "#EF4444", bars: 4 },
} as const;

type TaskMeta = { title: string; state: keyof typeof STATES; priority: keyof typeof PRIORITIES; due: string; labels: string[] };

export const tasks: Record<string, TaskMeta> = {
  "DEVOPS-780": { title: "Host Qwen3.8:27B on the GPU cluster", state: "progress", priority: "high", due: "Oct 14", labels: ["infra", "ml"] },
  "QA-133": { title: "Regression suite for checkout", state: "todo", priority: "medium", due: "Oct 18", labels: ["testing"] },
  "WEB-310": { title: "Revamp the pricing page", state: "progress", priority: "medium", due: "Oct 21", labels: ["marketing"] },
  "API-112": { title: "Rate-limit public endpoints", state: "approval", priority: "high", due: "Oct 09", labels: ["security"] },
  "AUTH-19": { title: "Rotate OAuth client secrets", state: "approval", priority: "urgent", due: "Oct 06", labels: ["security"] },
  "MOB-245": { title: "Offline mode for task lists", state: "backlog", priority: "low", due: "Nov 12", labels: ["mobile"] },
  "HR-12": { title: "Onboarding checklist for Q4 hires", state: "todo", priority: "low", due: "Oct 30", labels: ["people"] },
  "UI-402": { title: "Dark mode polish", state: "testing", priority: "medium", due: "Oct 11", labels: ["design"] },
  "DATA-77": { title: "Nightly warehouse sync", state: "blocked", priority: "high", due: "Oct 08", labels: ["data"] },
  "BILL-88": { title: "Invoice PDF redesign", state: "todo", priority: "medium", due: "Oct 25", labels: ["finance"] },
  "PAY-260": { title: "Approve refund policy changes", state: "approval", priority: "high", due: "Oct 10", labels: ["finance"] },
  "SEC-91": { title: "Quarterly access review", state: "progress", priority: "urgent", due: "Oct 07", labels: ["compliance"] },
  "LEGAL-6": { title: "Vendor DPA renewal", state: "approval", priority: "medium", due: "Oct 31", labels: ["legal"] },
  "NET-73": { title: "Replace the office VPN", state: "testing", priority: "high", due: "Oct 16", labels: ["infra"] },
  "DOCS-18": { title: "API reference examples", state: "done", priority: "low", due: "Oct 02", labels: ["docs"] },
  "DESIGN-55": { title: "Icon set refresh", state: "progress", priority: "low", due: "Oct 28", labels: ["design"] },
  "SALES-140": { title: "Q4 pricing deck", state: "todo", priority: "medium", due: "Oct 20", labels: ["sales"] },
  "INFRA-204": { title: "Upgrade Postgres to 18", state: "testing", priority: "high", due: "Oct 13", labels: ["infra", "db"] },
  "CI-301": { title: "Cache Docker build layers", state: "done", priority: "medium", due: "Oct 01", labels: ["ci"] },
  "OPS-57": { title: "On-call runbook cleanup", state: "backlog", priority: "low", due: "Nov 04", labels: ["ops"] },
};

export const users: Record<string, { name: string; role: string; tone: string }> = {
  AR: { name: "Aarav Rao", role: "Platform lead", tone: "bg-sky-100 text-sky-800" },
  KM: { name: "Kavya Menon", role: "Product manager", tone: "bg-violet-100 text-violet-800" },
  SN: { name: "Sameer Nair", role: "Finance ops", tone: "bg-emerald-100 text-emerald-800" },
  LT: { name: "Leela Thomas", role: "Engineering manager", tone: "bg-rose-100 text-rose-800" },
  RM: { name: "Rohan Mehta", role: "IT & operations", tone: "bg-amber-100 text-amber-800" },
};
