# BureauCat: Open-Source Task Manager with Approval Workflows

BureauCat is an open-source, self-hosted task manager. Requests, sign-offs and to-dos are turned into tasks that move, with approval workflows, kanban boards and a tamper-evident audit log.

- Try it: [__APP_URL__/signup](__APP_URL__/signup)
- Source: [github.com/bureaucatorg/bureaucat](https://github.com/bureaucatorg/bureaucat) (AGPL-3.0)
- API reference: [__APP_URL__/docs](__APP_URL__/docs)

## How a task moves

1. **Intake.** Requests, sign-offs and to-dos pile up. BureauCat files every one of them as a task, so nothing stays loose on someone's desk.
2. **Structure.** Workspaces hold projects, and projects hold tasks. Each task gets a numbered key like DEVOPS-780, not "that thing from Tuesday".
3. **Lifecycle.** Tasks travel through states you define. The defaults include Approval Pending, so a sign-off is a column, not an email thread. Priority, assignees and followers ride along.
4. **Planning.** Break a task into subtasks, move cards across the kanban board, and slot work into cycles and modules to see the bigger picture.
5. **Graph.** The graph view maps who is on what and how tasks nest, across every project you are a member of.
6. **Notices.** Comments with @mentions and attachments. Updates within 15 minutes fold into one notification: in-app, as an opt-in email digest, or a Mattermost DM.
7. **Record.** Every change lands in an append-only audit log, each entry hash-chained to the one before. One click verifies nothing was rewritten.

## Also included

- **Single sign-on:** Google and Zitadel, alongside passwords.
- **Tokens & CLI:** Read-only or read-write API tokens, plus a CLI.
- **API docs:** Every endpoint, documented.
- **Saved views:** Filter, group and sort. Keep it private or share it.
- **Pages:** Rich-text docs that live next to the project.
- **Search & quick create:** Ctrl K to find anything. Shift C to file a task.
- **Your branding:** Rename the app for your organisation.
- **Pomodoro:** A focus timer.
- **Admin tools:** Merge users, restore projects, instance stats.

## FAQ

### What is BureauCat?
BureauCat is an open-source, self-hosted task manager built around approval workflows. Requests and sign-offs become numbered tasks that move through states such as Approval Pending, and every change is recorded in a tamper-evident audit log.

### Is BureauCat free?
Yes. BureauCat is free and open source under the GNU Affero General Public License v3.0 (AGPL-3.0). You host it yourself.

### How do I self-host BureauCat?
Run the single `bureaucat` binary or the `codingcoffee/bureaucat` Docker image with a PostgreSQL database. Set `DATABASE_URL` and `JWT_SECRET`, then start it with `bureaucat serve --migrate`. The web app and API are served from one port.

### How do approvals work?
Approval is a task state. The default workflow includes Approval Pending, so a request waiting for sign-off sits in its own column on the board, with assignees, followers and a full history, instead of in an email thread.

### What makes the audit log tamper-evident?
The audit log is append-only and each entry is hash-chained (SHA-256) to the one before it. Rewriting any past entry breaks the chain, and one click verifies the whole log.

### Does BureauCat have an API?
Yes. Every endpoint is documented with an OpenAPI spec at `/api/v1/openapi.json` and interactive docs at `/docs`. API tokens can be read-only or read-write, and the `bureaucat` binary doubles as a CLI.
