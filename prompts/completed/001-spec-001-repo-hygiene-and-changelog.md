---
status: completed
spec: [001-cluster-build-arm]
summary: 'Verified repo hygiene is already complete: .gitignore has all four entries, specs/ and prompts/ .keep files exist, and CHANGELOG has one Unreleased bullet naming the cluster build arm.'
execution_id: 50a80865-62f5-5a08-89f8-f15ea285d502-exec-001-spec-001-repo-hygiene-and-changelog
dark-factory-version: dev
created: "2026-11-10T11:08:00Z"
queued: "2026-11-10T11:20:22Z"
started: "2026-11-10T11:20:23Z"
completed: "2026-11-10T11:29:11Z"
branch: dark-factory/cluster-build-arm
---

<summary>
- The pipeline's own directories survive a fresh clone because empty `.keep` files make `specs/` and `prompts/` tracked.
- The pipeline's runtime logs, the tool's lock file and the build's vendor directory are ignored, so a local build no longer leaves the tree dirty.
- The existing log wildcard continues to cover the tool's own `.log` file, so no duplicate entry is added.
- The change is recorded once under `## Unreleased` in the changelog.
- No build logic, workflow or delivery setting changes in this prompt.
</summary>

<objective>
Ensure the repository is clone-safe and clean: `.gitignore` carries the four new literal entries, `specs/.keep` and `prompts/.keep` exist so those directories survive a fresh clone, and `CHANGELOG.md` records the cluster build arm under exactly one `## Unreleased` heading.
</objective>

<context>
Read these files before editing:
- `specs/in-progress/001-cluster-build-arm.md` — especially Acceptance Criterion 7 and the Constraints bullet about `.gitignore`.
- `.gitignore` — current contents are `data/`, `*.log`, `/vendor`, `/prompts/log`, `/specs/log`, `/.dark-factory.lock`.
- `CHANGELOG.md` — current contents already carry a `## Unreleased` section above `## v0.1.6`.
- `specs/` and `prompts/` — verify the `.keep` files on disk.
- `/home/node/.claude/plugins/marketplaces/coding/docs/changelog-guide.md` — the header is frozen, `## Unreleased` sits directly above the highest version, entries are a flat list with a conventional prefix (`feat:`, `fix:`, etc.).
- `/home/node/.claude/plugins/marketplaces/coding/docs/git-workflow.md` — do not commit; dark-factory handles git.

The current branch may already carry some or all of these changes (a prior commit tracked the pipeline directories). Treat this prompt as idempotent: verify each item and add only what is missing; never duplicate an entry or heading.
</context>

<requirements>
1. **`.gitignore`.** Ensure each of the following literal lines is present, exactly once, alongside the existing entries:
   ```
   /vendor
   /prompts/log
   /specs/log
   /.dark-factory.lock
   ```
   Keep the existing `data/` and `*.log` lines. Do NOT add an entry for `/.dark-factory.log` or any other `.log` file: the existing `*.log` wildcard already matches it. Do not remove or reorder existing entries.

2. **Tracked directories.** Ensure `specs/.keep` and `prompts/.keep` exist on disk as empty files. If the subdirectory `.keep` files (`specs/completed/.keep`, `specs/in-progress/.keep`, `prompts/completed/.keep`, `prompts/in-progress/.keep`) already exist, leave them in place. Do not add `.keep` files outside `specs/` and `prompts/`.

3. **`CHANGELOG.md`.** Ensure exactly one `## Unreleased` heading exists, placed immediately after the frozen preamble and directly above the highest `## vX.Y.Z` section. Under it there must be at least one bullet with a conventional prefix that names the cluster build arm, for example:
   ```
   - feat: add a cluster build arm so the image can be built in Kubernetes, not only on the Mac
   ```
   If an equivalent bullet is already present, leave it unchanged. Do NOT add a second `## Unreleased` heading, do not insert anything above the `# Changelog` title, and do not move or edit the SemVer preamble.

4. **Do not modify** `Makefile.docker`, `Makefile.variables`, `Makefile.precommit`, `Dockerfile`, `hack/`, or `.dark-factory.yaml` in this prompt. Those are covered by sibling prompts; the workflow and delivery setting are explicitly out of scope.
</requirements>

<constraints>
- `.gitignore` keeps its existing entries and adds exactly four: `/vendor`, `/prompts/log`, `/specs/log`, `/.dark-factory.lock`.
- The tool's log file needs no entry because the existing `*.log` wildcard already matches it; the two log directories need entries because their files are extensionless.
- `specs/` and `prompts/` must exist in a fresh clone, which is why their `.keep` files are tracked.
- `vendor/` must become ignored: it is untracked today but not ignored, so a local build currently leaves the tree dirty.
- `CHANGELOG.md` carries exactly one `## Unreleased` heading whose bullet names the cluster build arm.
- `.git` is masked inside this container: do NOT run any `git` command. The committed-tree checks (`git ls-tree`, `git check-ignore`) run on the operator rung, not here.
- Do NOT commit — dark-factory handles git.
- Existing tests must still pass.
- Do not modify `.dark-factory.yaml`.
</constraints>

<verification>
Run each command from the repo root and record the output. All must pass.

1. `grep -nF '/vendor' .gitignore`
2. `grep -nF '/prompts/log' .gitignore`
3. `grep -nF '/specs/log' .gitignore`
4. `grep -nF '/.dark-factory.lock' .gitignore`
   Each of the four greps must print exactly one matching line.
5. `ls -l specs/.keep prompts/.keep` succeeds and both files are empty (size 0).
6. `grep -c '^## Unreleased$' CHANGELOG.md` returns exactly `1`.
7. `grep -A2 '^## Unreleased$' CHANGELOG.md | grep -i 'cluster build arm'` prints the bullet.
8. `make -n build` succeeds (the make surface resolves `ROOTDIR` without a git checkout) and still prints the local `docker build` recipe.
9. `make precommit` — lint, format and vet clean.
10. `make test` — unit and integration suites pass.
</verification>
