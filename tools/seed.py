#!/usr/bin/env python3
"""Seed a local Bureaucat instance with rich, realistic demo data.

Creates: an admin (demo@gmail.com / password), a handful of users, a workspace,
a couple of projects, and — per project — sprints (cycles), epics (modules), and
meaningful tasks with sub-tasks, assignees, priorities and workflow states.

Idempotent: it RESETS all app data first (TRUNCATE via the postgres container),
then rebuilds, so `make seed` always yields the same clean dataset. The very first
account is created via /signup, which the app promotes to admin automatically.

Stdlib only. Configure via env: BUREAUCAT_URL, PG_CONTAINER, SEED, NUM_USERS,
NUM_PROJECTS, TASKS_PER_PROJECT.
"""
import json, os, random, subprocess, urllib.request, urllib.error
from datetime import date, timedelta

BASE = os.environ.get("BUREAUCAT_URL", "http://localhost:1341") + "/api/v1"
PG_CONTAINER = os.environ.get("PG_CONTAINER", "bureaucat-postgres-1")
PG_USER = os.environ.get("PG_USER", "bureaucat")
PG_DB = os.environ.get("PG_DB", "bureaucat")

STRONG_PW = "Passw0rd!"  # prod-safe default (upper + number + special char)

def _demo_password():
    """Single demo password: env override -> DEMO_USER_PASSWORD in .env -> strong
    default. One strong password works in both dev and prod (prod enforces a policy)."""
    v = os.environ.get("DEMO_USER_PASSWORD") or os.environ.get("DEMO_PASSWORD")
    if v:
        return v
    try:
        for line in open(os.environ.get("ENV_FILE", ".env")):
            s = line.strip()
            if s.startswith("DEMO_USER_PASSWORD="):
                return s.split("=", 1)[1].strip().strip('"').strip("'")
    except OSError:
        pass
    return STRONG_PW

ADMIN_EMAIL = os.environ.get("DEMO_EMAIL", "demo@gmail.com")
ADMIN_PW = _demo_password()
USER_PW = ADMIN_PW

random.seed(int(os.environ.get("SEED", "7")))
NUM_USERS = int(os.environ.get("NUM_USERS", "8"))
NUM_PROJECTS = int(os.environ.get("NUM_PROJECTS", "2"))
CYCLES_PER_PROJECT = int(os.environ.get("CYCLES_PER_PROJECT", "3"))
MODULES_PER_PROJECT = int(os.environ.get("MODULES_PER_PROJECT", "4"))
TASKS_PER_PROJECT = int(os.environ.get("TASKS_PER_PROJECT", "15"))

# ---------------------------------------------------------------- data pools
FIRST = ["Alice", "Bob", "Carla", "Diego", "Ellen", "Farah", "Grace", "Hiro",
         "Ivan", "Julia", "Kwame", "Lena", "Mateo", "Nina", "Omar", "Priya"]
LAST = ["Zhang", "Okafor", "Rossi", "Nguyen", "Schmidt", "Haddad", "Kim",
        "Silva", "Ivanov", "Brown", "Mensah", "Novak", "Garcia", "Patel"]
WORKSPACES = ["Northwind", "Acme", "Globex", "Umbrella", "Initech"]
PROJECTS = [("Web Platform", "WEB"), ("Mobile App", "MOBILE"),
            ("Data Pipeline", "DATA"), ("Billing Service", "BILLING")]
EPICS = ["Authentication & SSO", "Billing & Invoicing", "Search & Discovery",
         "Notifications", "Dashboard & Analytics", "Onboarding", "Public API",
         "Performance", "Admin Console", "Reporting"]
STORIES = [
    "Add password reset flow", "Implement OAuth login", "Design settings page",
    "Optimize task list query", "Add CSV export", "Fix timezone handling in reports",
    "Build notification preferences", "Add dark mode", "Implement bulk actions",
    "Paginate the activity feed", "Add rate limiting", "Write E2E tests for checkout",
    "Migrate to the new table component", "Add keyboard shortcuts", "Improve empty states",
    "Add an audit-log viewer", "Support markdown in comments", "Add file attachments",
    "Implement webhooks", "Cache dashboard metrics", "Add SSO group mapping",
    "Redesign the onboarding wizard", "Add per-project roles", "Ship the mobile nav",
]
SUBTASKS = ["Write unit tests", "Update API docs", "Add DB migration",
            "Address code-review feedback", "Handle error states", "Add loading skeleton",
            "Wire up the endpoint", "Emit analytics event", "QA on staging",
            "Accessibility pass", "Add feature flag", "Backfill existing rows"]
# (state name, relative weight) over the project's auto-created default states
STATE_WEIGHTS = [("Backlog", 3), ("Todo", 3), ("Approval Pending", 1),
                 ("In Progress", 4), ("Blocked", 1), ("Testing", 2),
                 ("Done", 4), ("Cancelled", 1)]

# ---------------------------------------------------------------- http
TOKEN = [None]

def api(method, path, body=None, auth=True, ignore=()):
    data = json.dumps(body).encode() if body is not None else None
    r = urllib.request.Request(BASE + path, data=data, method=method)
    if body is not None:
        r.add_header("Content-Type", "application/json")
    if auth and TOKEN[0]:
        r.add_header("Authorization", "Bearer " + TOKEN[0])
    try:
        with urllib.request.urlopen(r) as resp:
            b = resp.read()
            return json.loads(b) if b else {}
    except urllib.error.HTTPError as e:
        if e.code == 401 and auth:
            signin()
            return api(method, path, body, auth, ignore)
        if e.code in ignore:
            return {"_status": e.code}
        raise RuntimeError(f"{method} {path} -> {e.code}: {e.read().decode('utf-8','replace')}")

def signin():
    TOKEN[0] = api("POST", "/signin",
                   {"identifier": ADMIN_EMAIL, "password": ADMIN_PW}, auth=False)["access_token"]

def reset():
    sql = "TRUNCATE users, workspaces RESTART IDENTITY CASCADE;"
    p = subprocess.run(["docker", "exec", PG_CONTAINER, "psql", "-U", PG_USER, "-d", PG_DB, "-c", sql],
                       capture_output=True, text=True)
    if p.returncode != 0:
        raise SystemExit(f"reset failed (is the stack up & migrated?):\n{p.stderr.strip()}")

# ---------------------------------------------------------------- seed
def main():
    global ADMIN_PW, USER_PW
    print("→ resetting app data")
    reset()

    print(f"→ creating admin {ADMIN_EMAIL} (first signup is promoted to admin)")
    signup = {"username": "demo", "email": ADMIN_EMAIL, "first_name": "Demo", "last_name": "Admin"}
    try:
        api("POST", "/signup", {**signup, "password": ADMIN_PW}, auth=False)
    except RuntimeError as e:
        if "requirements" not in str(e):
            raise
        ADMIN_PW = USER_PW = STRONG_PW  # prod enforces strong passwords
        api("POST", "/signup", {**signup, "password": STRONG_PW}, auth=False)
        print(f"   (prod password policy → using '{STRONG_PW}')")
    signin()

    print(f"→ creating {NUM_USERS} users")
    users = []                       # list of {id, name}
    used = set()
    for _ in range(NUM_USERS):
        while True:
            fn, ln = random.choice(FIRST), random.choice(LAST)
            uname = (fn + ln).lower()
            if uname not in used:
                used.add(uname); break
        r = api("POST", "/admin/users", {"username": uname, "email": f"{uname}@demo.local",
                "password": USER_PW, "first_name": fn, "last_name": ln, "user_type": "user"},
                ignore=(409,))
        if r.get("_status") == 409:
            continue
        users.append({"id": r["id"], "name": f"{fn} {ln}"})
    user_ids = [u["id"] for u in users]

    ws_key = random.choice(WORKSPACES).upper()[:10]
    print(f"→ workspace {ws_key}")
    ws = api("POST", "/workspaces", {"workspace_key": ws_key, "name": random.choice(WORKSPACES) + " HQ",
                                     "description": "Demo workspace"})
    for uid in user_ids:
        api("POST", f"/workspaces/{ws_key}/members", {"user_id": uid}, ignore=(409,))

    total = {"projects": 0, "cycles": 0, "modules": 0, "tasks": 0, "subtasks": 0}
    for name, key in PROJECTS[:NUM_PROJECTS]:
        print(f"→ project {key} ({name})")
        total["projects"] += 1
        api("POST", "/projects", {"project_key": key, "name": name, "workspace_id": ws["id"],
                                  "description": f"{name} — demo project"})
        for uid in user_ids:
            api("POST", f"/projects/{key}/members", {"user_id": uid, "role": "member"}, ignore=(409,))

        states = {s["name"]: s["id"] for s in api("GET", f"/projects/{key}/states")}
        state_pool = [states[n] for n, w in STATE_WEIGHTS for _ in range(w) if n in states]

        # sprints (cycles): consecutive, non-overlapping, centred on today
        cycles = []
        base = date.today() - timedelta(days=14 * (CYCLES_PER_PROJECT - 2))
        for i in range(CYCLES_PER_PROJECT):
            s = base + timedelta(days=14 * i)
            c = api("POST", f"/projects/{key}/cycles", {"title": f"Sprint {i + 1}",
                    "start_date": s.isoformat(), "end_date": (s + timedelta(days=13)).isoformat()})
            cycles.append(c["id"]); total["cycles"] += 1

        # epics (modules)
        modules = []
        for title in random.sample(EPICS, MODULES_PER_PROJECT):
            m = api("POST", f"/projects/{key}/modules",
                    {"title": title, "status": "in_progress", "description": f"Epic: {title}"})
            modules.append(m["id"]); total["modules"] += 1

        # tasks (+ sub-tasks)
        cycle_task_ids = {c: [] for c in cycles}
        module_task_ids = {m: [] for m in modules}
        for title in random.sample(STORIES, min(TASKS_PER_PROJECT, len(STORIES))):
            assignees = random.sample(user_ids, k=random.choice([0, 1, 1, 2]))
            t = api("POST", f"/projects/{key}/tasks", {"title": title, "description": f"{title}.",
                    "state_id": random.choice(state_pool), "priority": random.choice([0, 1, 2, 2, 3, 4]),
                    "assignees": assignees})
            total["tasks"] += 1
            cyc = random.choice(cycles); cycle_task_ids[cyc].append(t["id"])
            if random.random() < 0.7:
                mod = random.choice(modules); module_task_ids[mod].append(t["id"])
            for st in random.sample(SUBTASKS, k=random.choice([0, 1, 2, 2, 3])):
                sub = api("POST", f"/projects/{key}/tasks", {"title": st,
                          "state_id": random.choice(state_pool), "priority": random.choice([0, 1, 2]),
                          "assignees": random.sample(user_ids, k=random.choice([0, 1])),
                          "parent_task_number": t["task_number"]})
                total["subtasks"] += 1
                cycle_task_ids[cyc].append(sub["id"])

        for cid, ids in cycle_task_ids.items():
            if ids:
                api("POST", f"/projects/{key}/cycles/{cid}/tasks", {"task_ids": ids}, ignore=(409,))
        for mid, ids in module_task_ids.items():
            if ids:
                api("POST", f"/projects/{key}/modules/{mid}/tasks", {"task_ids": ids}, ignore=(409,))

    print("\n✅ Seed complete")
    print(f"   workspace: {ws_key}   projects: {total['projects']}   "
          f"cycles: {total['cycles']}   modules: {total['modules']}   "
          f"tasks: {total['tasks']}   sub-tasks: {total['subtasks']}   users: {len(users)}")
    print(f"   admin login: {ADMIN_EMAIL} / {ADMIN_PW}")
    print(f"   users (log in to see per-person views): password '{USER_PW}', e.g. "
          + ", ".join(u["name"] for u in users[:3]) + ", …")

if __name__ == "__main__":
    main()
