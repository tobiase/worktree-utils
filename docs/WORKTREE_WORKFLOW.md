# Worktree Workflow

This document captures the canonical workflow for managing backlog tasks with dedicated Git worktrees. Follow these steps every session so `main` stays clean and task context is easy to recover.

## 1. Before You Touch Code

1. **Read the Backlog task**: Confirm scope, acceptance criteria, and the latest notes in `backlog/tasks/task-<id>*.md`.
2. **Log ownership**: Add a Backlog note in the task file (`[YYYY-MM-DD HH:MM] Worktree task-<id>-<slug> @ <path> (base <sha>)`). This is the traceable hand-off.
3. **Sync `main`**:
   - `git fetch --all --prune`
   - `git checkout main && git pull --rebase`
4. **Check cleanliness**: `git status -sb` on `main` must be clean (only intentional backlog commits allowed). If dirty:
   - Commit generated backlog files immediately.
   - Move unrelated edits into the correct worktree (capture diff, apply elsewhere, then `git restore --source HEAD -- path`).
   - Only stash as a last resort; document it if you do.

## 2. Creating the Task Worktree

1. **Commit backlog stub**: After creating a task on `main`, stage and commit the generated `backlog/tasks/task-<id>*.md` file(s). This keeps `main` ready for new worktrees.
2. **Name consistently**: Branch and worktree names must be `task-<id>-<slug>` (e.g., `task-37-search-bug`).
3. **Create from clean `main`**:
   ```bash
   git worktree add ../repo-worktrees/task-<id>-<slug> -b task-<id>-<slug>
   # or: wt new task-<id>-<slug>
   ```
4. **Record the base commit** in Backlog notes so reviewers know where you branched.

## 3. Daily Safety Checks

- **Always know where you are**:
  ```bash
  git status -sb
  git rev-parse --abbrev-ref HEAD
  ```
  If either command reports `main`, stop immediately, switch to the task worktree, and log the correction in Backlog notes.
- **Session boundaries**: At the start *and* end of each session, open the primary `main` checkout and run `git status -sb`. Clean up strays before you leave.
- **Frequent rebases**: `git fetch && git rebase origin/main` from the task branch whenever `main` moves.
- **Commit often**: Make WIP commits before stepping away so hand-offs never rely on an uncommitted working tree.
- **One-step cleanup**: Prefer `wt rm <branch> --branch` when you're ready to delete a worktree—this keeps `main` clean and deletes the branch only if it's fully merged (use `--force` only when you intentionally mirror `git branch -D`).

## 4. Handling Mistakes

If you ever edit inside `main`:
1. `git diff > /tmp/main-slip.patch`
2. Switch to the correct worktree.
3. Apply the diff (`git apply /tmp/main-slip.patch` or redo changes manually).
4. Clean `main` (`git restore --source HEAD -- path`).
5. Document the slip and fix in Backlog notes before proceeding.

## 5. Session Wrap-Up Checklist

1. Run required tests/linters for the task and capture the result in the Backlog notes.
2. Update Acceptance Criteria and Implementation Notes in the task file.
3. Push the branch if remote collaboration is needed.
4. Record which branch/worktree you are leaving checked out.

## 6. Integration & Cleanup

When the user says “integrate into main” (or similar wording) you can now run:

```bash
wt integrate task-<id>-<slug>
```

`wt integrate` does the documented checklist for you: it fetches, rebases the task branch onto the latest `main`, fast-forward merges from the primary checkout, and then removes both the worktree and branch.

If you need to do things manually (or the command reports a conflict):

1. From the task worktree: `git fetch && git rebase origin/main`.
2. From the primary checkout:
   ```bash
   git checkout main
   git merge --ff-only task-<id>-<slug>
   ```
3. Immediately clean up:
   ```bash
   git worktree remove ../repo-worktrees/task-<id>-<slug>
   git branch -d task-<id>-<slug>
   ```
4. Confirm `main` is clean with `git status -sb` and note the integration in Backlog.

## 7. Backlog Discipline

- Keep the Backlog thread as the source of truth: add notes whenever you pause for >1h, change scope, or finish a session.
- When wrapping up, explicitly state in the hand-off which branch/worktree is active so the next person can verify quickly.
- Re-run this checklist whenever you onboard a new task to avoid stale worktrees lingering in the repo.
