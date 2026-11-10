---
status: draft
---

## Summary

- This repo's image can be produced on exactly one machine today: this Mac.
- Its sibling services can also build their image inside a Kubernetes cluster, unattended, from any committed ref.
- Add that second build path here as an arm beside the existing local one, leaving the local command and the tag it produces exactly as they are.
- The cluster arm builds a named commit as a Kubernetes Job and pushes the result to the same image repository the local arm pushes to, so producing this image no longer needs this machine.
- Two supporting edits make the arm reachable: the image must build from a fresh clone with no prepared working tree, and the repo must track the directories the pipeline writes into.

## Problem

`notification-discord` runs on nukedev and nukeprod, and its StatefulSet pulls `bborbe/notification-discord`, but nothing in this repo can produce that image without a human at this Mac: the build files carry a single local path, `Makefile.docker` runs `docker build` against the local daemon and has no notion of building anywhere else, so every release, every weekly rebuild and every future unattended pipeline needs this machine awake and logged in — for one image, while the sibling services build themselves. A naive copy of the sibling arm fails here at runtime rather than at review, for four reasons this repo does not share with the repos it would be copied from. It has no `Makefile.env`, so there is no branch guard to narrow — but `VERSION` is defaulted inside `Makefile.docker` itself, so a copied `$(origin VERSION)` guard either fires always or never depending on where it is placed. It defines no `IMAGE` variable and writes its tag as a literal `$(DOCKER_REGISTRY)/bborbe/$(SERVICE):$(VERSION)`, while the arm sources compose from a variable that already carries the namespace — pasted verbatim they render `docker.io/` and push to a repository nothing deploys. Its `build` target chains a preparation step that vendors modules into the working tree, which is meaningless for a build that clones its own ref. And its image build resolves Go dependencies from `vendor/`, which is not tracked, so a Job that clones the repo has no `vendor/` directory and fails before it starts.

## Goal

`make build` on this repo produces the image locally exactly as it does today, and `make build BUILD_TARGET=<cluster> REF=<full sha> VERSION=<tag>` produces it inside that cluster as a Kubernetes Job, pushing `docker.io/bborbe/notification-discord:<tag>` — the repository the deployment's mirror step consumes. Both arms work from the same checkout, and the cluster arm works from a fresh clone with no local preparation.

## Non-goals

- Changing the local arm's command or the image tag it produces. Both stay as they are; the tag is re-expressed through one shared variable and its rendered value is unchanged.
- Renaming the image or moving it to a private registry. The push stays on Docker Hub, which is what the `build` namespace's push credential covers; the private-registry copy the deployment pulls is produced by nuke's `mirror:` step, not here.
- Verifying that nukedev or nukeprod picks the new image up. That is the deploy path's concern, and a missed mirror there surfaces as `ImagePullBackOff`, not as a build failure.
- Building at release or tag time. This adds the capability; *when* a build is triggered is a separate change.
- Changing the workflow or the delivery setting in `.dark-factory.yaml`. This repo runs `workflow: worktree` with `pr: true`, which three of the four already-migrated siblings also run at `origin/master`, and one of them delivered its arm through it as a merged PR. What is missing here is the pipeline's own directories and a git-free make surface, not the workflow — and changing the workflow would cost the PR delivery that this repo deliberately adopted. The file's command overrides are likewise out of scope: this change makes the existing commands runnable in-container rather than replacing them.
- Extra build knobs beyond `BUILD_TARGET`, `VERSION` and `REF`. The copy source also carries a concurrency cap, a default-version opt-out and a configurable push-secret name; none is copied, because each either reintroduces an opt-out on a refusal this spec ships or widens what a credential-bearing build may be pointed at.
- Multi-arch images. `linux/amd64` only, matching the cluster nodes.

## Acceptance Criteria

- [ ] **The local arm's command and the tag it produces are unchanged** — with no `BUILD_TARGET` set, `make -n build` prints the `docker build` recipe, no reference to `hack/build-cluster.sh`, and the tag prefix `docker.io/bborbe/notification-discord:`. Evidence: those three readings from one `make -n build`, plus — on the operator rung, where `.git` is available — the same command run against the merge base, whose printed tag must be byte-identical to the branch's.
- [ ] **Post-Deploy (Rung-2):** the cluster arm builds and pushes from a fresh clone — `make build BUILD_TARGET=sun REF=<full 40-character sha> VERSION=<repo-unique scratch-tag>` exits 0, and `crane digest docker.io/bborbe/notification-discord:<scratch-tag>` resolves. Evidence: the exit code, the Job name, the build pod's `spec.nodeName`, the resolved digest, and `DOCKER_CONTEXT=orbstack docker image inspect` on the same reference failing with `No such image` — a stopped daemon or a malformed reference does not satisfy this. The pushed image's own `BUILD_GIT_COMMIT` must carry the short form of the `REF` that was passed, which is what distinguishes a build of that commit from a re-push of an older image.
  - `deploy_check:` `kubectlsun -n build get secret buildkit-push -o jsonpath='{.metadata.name}'`
  - `deploy_target:` `buildkit-push`
  - The `Rung-2` label is nominal here: the target is `sun`, the build cluster, not a dev stage, and `deploy_check` proves the push secret exists rather than that anything is running the new code. Read it as "the pre-flight a real build needs", never as "dev carries this change".
- [ ] **The cluster arm pushes to the repository the deployment consumes** — `make -n build BUILD_TARGET=sun REF=<sha> VERSION=probe` prints the composed reference `docker.io/bborbe/notification-discord:probe` on one line before it calls the runner. Evidence: that printed line, and the digest above resolving at exactly that repository.
- [ ] **The cluster arm refuses a tag it did not receive from the caller** — `make build BUILD_TARGET=sun REF=<sha>` with no `VERSION` on the command line exits non-zero and prints a message naming `VERSION`, before any Job is applied. Evidence: the exit code, the message, and `kubectlsun -n build get jobs` showing no Job for that run.
- [ ] **`hack/` holds the five runner files and the cluster branch invokes the runner** — `ls -l hack/` lists `build-cluster.sh`, `cluster-alias.sh`, `build-job.yaml`, `build-cluster-bootstrap.sh` and `dockerconfig.json.tpl`, each with a non-zero size, and `grep -cE '^[[:space:]]+[^#[:space:]].*hack/build-cluster\.sh' Makefile.docker` returns at least 1 — a recipe line, not a comment, and not the `.PHONY` or dependency declaration. Evidence: the five names with byte sizes from `ls -l`, and that count. Tracking in the committed tree is a separate check on the operator rung, because `.git` is masked inside the prompt container.
- [ ] **`upload`, `clean` and `buca` stay local, refused at parse time** — `make -n upload BUILD_TARGET=sun REF=<sha> VERSION=probe` and the same for `clean` print a skip line and no `docker push` / `docker rmi`; `make -n buca BUILD_TARGET=sun REF=<sha> VERSION=probe` exits non-zero and prints a message naming those targets as local-only. The refusal is a `$(error …)`, so it fires under `-n` where no recipe runs. Evidence: both outputs and all three exit codes.
- [ ] **The repo ignores the pipeline's own artifacts, tracks the directories it writes into, works without a git checkout, and documents the change** — `.gitignore` carries four new literal entries, `/vendor`, `/prompts/log`, `/specs/log` and `/.dark-factory.lock` (the two log directories need entries because their files are extensionless; `/.dark-factory.log` needs none, because the existing `*.log` wildcard already matches it); `specs/.keep` and `prompts/.keep` exist on disk so the directories survive a fresh clone; the make surface resolves its root directory without invoking git, so `make -n build` and `make test` succeed in a tree where `.git` is masked; and `CHANGELOG.md` carries exactly one `## Unreleased` heading whose bullet names the cluster build arm. Evidence: `grep -nF` over `.gitignore` for each of the four strings, `ls -l` on the `.keep` files, the exit code of `make -n build` in a tree with no `.git`, the heading count, and the bullet text. The committed-tree checks (`git ls-tree`, `git check-ignore`) run on the operator rung — `.git` is masked inside the prompt container, so they cannot run there.

## Verification

### Container-executable (runs inside the YOLO container at prompt time)

⚠️ The prompt container has `.git` masked under this repo's workflow, so **no item here may call `git`**. Everything git-shaped lives on the operator rung below.

- `make precommit` — lint, format and vet clean
- `make test` — unit and integration suites pass
- `make -n build` succeeds at all — proof that the make surface resolves `ROOTDIR` without a git checkout, which is the precondition for the two items above
- `bash -n hack/build-cluster.sh && bash -n hack/cluster-alias.sh && bash -n hack/build-cluster-bootstrap.sh` — the copied runners parse
- `grep -c '^BUILD_TARGET' Makefile.docker` returns at least 1
- the runner appears on a recipe line in `Makefile.docker`, not only in a comment
- `ls -l hack/` lists the five expected files, each non-empty
- `grep -nF` over `.gitignore` finds each of `/vendor`, `/prompts/log`, `/specs/log` and `/.dark-factory.lock`
- `ls -l specs/.keep prompts/.keep` succeeds
- `grep -c '^## Unreleased$' CHANGELOG.md` returns 1, and its bullet names the cluster build arm
- `make -n build` prints a `docker build` recipe, no `hack/build-cluster.sh`, and the tag prefix `docker.io/bborbe/notification-discord:` — the tag *value* is resolved by git, so it is checked on the operator rung, not here
- `make -n build BUILD_TARGET=sun REF=<sha> VERSION=probe` prints `hack/build-cluster.sh` and the composed `docker.io/bborbe/notification-discord:probe` line
- `make -n upload BUILD_TARGET=sun REF=<sha> VERSION=probe` prints a skip line and no `docker push`
- `make -n buca BUILD_TARGET=sun REF=<sha> VERSION=probe` exits non-zero naming `upload`, `clean` and `buca` as local-only
- a `grep` sweep for `BUILD_TARGET` across the repo's tracked files — every reference names the new invocation; no stale build documentation is left behind

### Operator-executable (runs on the host after PR merge, spec verification ladder)

These carry every check the container cannot make, plus the real build.

- `git ls-tree --name-only HEAD hack/` lists the five runner files
- `git ls-tree -r --name-only HEAD specs/ prompts/` lists every `.keep`
- `git check-ignore -v --no-index specs/probe.md` and the same for `prompts/probe.md` both exit 1
- `git check-ignore -v --no-index vendor/x.go` exits 0
- `make build BUILD_TARGET=sun REF=<sha> VERSION=<repo-unique scratch-tag>` — one real cluster build, with Job name, node and digest recorded
- `crane digest docker.io/bborbe/notification-discord:<scratch-tag>` — the push landed
- read the pushed image's `BUILD_GIT_COMMIT` from its config — it carries the commit that was built
- `DOCKER_CONTEXT=orbstack docker image inspect` on that same reference — fails with `No such image`
- `DOCKER_CONTEXT=orbstack make build`, then inspect the latest tag — the local arm still works

## Desired Behavior

1. **A build target selects the arm.** `BUILD_TARGET` defaults to `local`. With `local`, `make build` behaves exactly as it does today: it vendors the module tree, it runs `docker build` on this machine, and it tags the reference the repo already tags.
2. **Any other value routes to a cluster Job, and the local-only targets refuse it.** With `BUILD_TARGET` set to a cluster alias, `make build` prints the composed image reference, then runs the vendored runner, which renders and applies the Job, waits for it and reports success or failure with the Job's own log lines. The runner receives its inputs as environment variables exported on the invocation line, because a Make variable is not exported to a recipe's subshell. `upload` and `clean` become no-ops that say so, and `buca` — which chains both — is refused at parse time rather than half-running.
3. **The cluster arm needs a caller-supplied tag.** The cluster arm refuses to run unless `VERSION` came from the command line or the environment — in particular when it resolved from the file's own default — and the refusal names `VERSION` and happens before any Job exists. The local arm keeps its default: a bare `make build` still picks the latest tag.
4. **The cluster arm prepares nothing.** The vendoring step runs on the local arm only; a build that clones its own ref must not vendor the caller's working tree.
5. **Both arms publish to one repository.** The image reference is defined once and rendered identically by both arms, so the cluster arm cannot drift into a path the deployment does not pull.
6. **The cluster build works from a clone.** The image build resolves its Go dependencies in module mode, so a Job that clones the repo with no `vendor/` directory builds the same way a developer's checkout does, and the local arm is unaffected in the image it produces.
7. **The pipeline's own directories are versioned, and the make surface works without a git checkout.** `specs/` and `prompts/` are tracked, with `.keep` files so they exist in a fresh clone, while their runtime logs and the tool's own lock file are ignored. The make surface resolves its root directory without invoking git, so a prompt container with `.git` masked can still run the build targets and the test suite. The CHANGELOG records the change under `## Unreleased`.

## Constraints

- The local arm's command and the reference it tags must not change. `docker.io/bborbe/notification-discord` at the latest tag must be produced by the same recipe, with the same arguments, as today.
- `VERSION` is defaulted inside `Makefile.docker` through an empty-test. A value the caller passed keeps origin `command line` (or `environment`); only the defaulted value becomes origin `file`. The cluster guard therefore tests `$(origin VERSION)` for `file` — the same predicate the reference implementation uses — and must not test for emptiness, because an empty caller-supplied value is not a default.
- The local-only refusal splits in two, because one mechanism cannot serve both. `upload` and `clean` select a skip-line recipe branch under the cluster arm — they must print and exit 0, which a parse-time error cannot do. `buca` is a parse-time `$(error …)` keyed on the arm **and** on `$(MAKECMDGOALS)`, because an arm-keyed error with no goal filter fires for `make build BUILD_TARGET=…` too and would break the cluster arm entirely, while a recipe-level `exit 1` is invisible to `make -n` and the criterion could not observe it.
- The vendoring step stays a prerequisite of the local arm only. Leaving it chained to both would make the cluster arm run it in the caller's tree before doing nothing with the result.
- The five `hack/` files are copied from `agent-task-executor` at `origin/master` — the runner, the cluster alias table and the Job template — and from `go-skeleton` at `origin/master` — the bootstrap script and the dockerconfig template. The Job template from `agent-task-executor` is the one whose Job carries `automountServiceAccountToken: false`; the other source's does not, which is why the files are split rather than taken from one repository. **The copy cannot be a local `cp`:** the prompt container mounts only this worktree, so no sibling checkout exists inside it. Both sources are public, so the route is an anonymous fetch of each file at a pinned commit from `raw.githubusercontent.com`, or a faithful hand-authoring from the source; either way the content must match the source byte for byte, since these files are meant to be re-syncable rather than owned here.
- The runner derives its Job name from the tag alone, so a scratch tag must be unique across the whole repository, not merely unique to this run. Two builds sharing a tag collide on the second apply and the runner refuses rather than overwriting.
- The image build resolves dependencies in module mode, stated explicitly in the Dockerfile rather than inferred. Deleting a vendor-mode flag is not enough: with a `vendor/modules.txt` present, Go selects vendor mode automatically, so the local arm would build from `vendor/` while the cluster arm built from modules — same tag, different dependency resolution. The explicit flag is what keeps the two arms producing the same image. `vendor/` must **become** untracked and ignored: it is untracked today but not ignored, so a local build currently leaves the tree dirty with a directory the pipeline's own commit can sweep in.
- The push secret is a Docker Hub credential. The arm pushes `docker.io/bborbe/notification-discord`; a private-registry push needs a different secret and is out of scope.
- `.gitignore` keeps its existing entries and adds four: the vendor directory, the prompt log, the spec log and the tool's lock file. The tool's log file needs no entry — the existing `*.log` wildcard already matches it — and the two log directories need entries because their files are extensionless.
- `ROOTDIR` currently resolves through `git rev-parse --show-toplevel` in two make files, and `Makefile.precommit` then includes `$(ROOTDIR)/tools.env`. With `.git` masked in the prompt container that resolves empty and the include fails at parse time, so the build and test targets are unreachable there. The fallback belongs in the file included **first** — `Makefile.variables` — because `Makefile.precommit`'s own `?=` then never fires; and it must yield a **non-empty** value (the current directory is the right one, since the make file is at the repo root), because `?=` with an empty shell result still counts as defined and `include /tools.env` would fail exactly as before. A normal checkout's behaviour must stay byte-identical, so the fallback applies only when git cannot answer.
- The runner reads its inputs from the environment, not from arguments: the build target, repository URL, ref, tag, image, dockerfile, push flag, push-secret name, registry and timeout. A missing one fails at the runner's own check, so every one of them must be exported on the invocation line — a Make variable is not exported to a recipe's subshell, and a value that is silently empty fails only at runtime, after the Job path has begun.
- The copied alias table accepts `sun`, `nukedev`, `nukeprod` and literal contexts. It is copied wholesale rather than pruned to `sun`, because a pruned copy diverges from the source it is meant to track; the cluster is caller-selected exactly as the reference allows, and the credential-bearing surface it widens is bounded by the repository and secret being literals.
- `.dark-factory.yaml` is not modified.

## Assumptions

These must hold before the cluster criteria can pass; none is created by this spec.

- The `build` namespace exists on the target cluster and holds a push secret of type `kubernetes.io/dockerconfigjson` covering Docker Hub.
- The bootstrap script ships as a tool for an operator to run once, not something the pipeline executes: it creates the namespace and the push secret from a credential an operator supplies, and the prompt container has neither that credential nor cluster write access. It is shipped because it documents how the preconditions are met, and its presence is verified, not its execution.
- The repository is public, so the Job's anonymous clone succeeds — the copied Job carries no token-minting init container and no git credential.
- `crane` is installed on the host running the verification commands.
- `REF` exists in the local repository **and on `origin`**: the runner derives the build clock from the local repository, and the Job's init container fetches the ref from the remote. A commit that exists only locally fails at the clone.
- The build pod can reach the Go module proxy.
- `kubectl` is configured for the target cluster through the alias table the runner resolves.

## Failure Modes

| Trigger | Expected behavior | Recovery |
|---|---|---|
| `BUILD_TARGET` names an unknown cluster | The runner fails before applying anything and lists the accepted values | Re-run with a listed target; no Job exists to clean up |
| `VERSION` omitted on the cluster arm | Refusal naming `VERSION`, before the Job is created | Re-run passing a tag |
| `REF` is not a commit in the local repository | The runner dies before rendering: it cannot derive the build clock | `git fetch origin`, then re-run with the full 40-character SHA |
| `REF` is a short SHA | The runner resolves the clock locally (a short SHA works there) and the Job's clone fails with `couldn't find remote ref` — so the failure appears after the Job is created, not at the make invocation | Re-run with the full 40-character SHA; the failed Job can be deleted |
| `REF` exists locally but was never pushed | The clone init container fails: the ref does not exist on `origin` | Push the branch or commit, then re-run |
| The ref's epoch differs between the cluster and the local repository | The build container refuses before BuildKit runs and names both epochs — the ref moved under the local checkout | `git fetch origin`, then re-run |
| A Job with the derived name already exists | The runner refuses and names the existing Job rather than overwriting it | Delete the finished Job, then re-run with a unique tag |
| The runner is interrupted after apply | The Job keeps running while the local invocation exits | `kubectlsun -n build get jobs` to find it, `kubectlsun -n build logs job/<name> --all-containers` to read it, then `kubectlsun -n build delete job <name>` once finished |
| The build pod cannot reach the module proxy | The Job fails; the runner prints the last log lines and the Job name | Inspect the Job logs; the local arm is unaffected |
| The build pod is OOM-killed | The Job reports failed; the pod status carries the kill reason | Raise the template's limits in the same change, or build locally for that image |
| The push is rejected with `401 Unauthorized` | The build succeeds and the push fails | Check that the build namespace secret covers the registry being pushed to; this is a credential finding, not a code one |
| A caller passes an existing release tag as `VERSION` | The guard passes — it was supplied on the command line — and the push **overwrites that released image irreversibly** | No recovery: re-push the image from the released tag, then cut a new release. Treat every cluster build as a publish; use a repo-unique scratch tag |
| The local arm regresses | `make build` fails or tags differently | Revert the arm; the local path is the regression gate |
| A cluster build is attempted on a checkout that predates the arm | The target does not exist, so the invocation is rejected rather than silently building locally | Build from a checkout that carries the arm |

## Security / Abuse Cases

- The runner substitutes caller-supplied values into a manifest and into embedded shell, so every value is checked against a character whitelist and rejected for embedded newlines before rendering. No value path may bypass that check.
- The repository URL and the push-secret name are literals in the dispatch, not configurable inputs. A configurable clone source on a build that mounts a push credential would let a caller point the credential-bearing Job at another repository.
- The build Job mounts a push credential from a build-namespace secret and runs with `automountServiceAccountToken: false`. The credential is scoped to pushing an image and grants nothing else.
- The build container runs with relaxed sandbox profiles because rootless BuildKit requires it. That relaxation must not be copied to the clone init container.
- The tag is caller-controlled and becomes both an image tag and a Job name. The tag is whitelisted and the Job name sanitised to a valid DNS name, truncated, and stripped of trailing separators.
- The registry defaults to Docker Hub in both the template and the Dockerfile, so the base images resolve from the public registry inside the cluster rather than from a private mirror. Do not change that back.

## Suggested Decomposition

Prompts should be generated in this order — each row is a single prompt with a clear scope.

| # | Prompt focus | Covers DBs | Covers ACs | Depends on |
|---|---|---|---|---|
| 1 | The build-file dispatch: the target split, one shared image reference, the composed-reference echo, the caller-supplied-tag refusal, local-only upload / clean / buca, **and the git-free `ROOTDIR` fallback** — it comes first because every later prompt's verification begins with `make precommit` / `make test`, both of which die at `include $(ROOTDIR)/tools.env` while `.git` is masked | 1, 2, 3, 4, 5 | 1, 3, 4, 6, 7 (the `make -n build` half) | — |
| 2 | The build path itself: the five runner files and the module-mode image build | 2, 6 | 2, 5 | prompt 1 (the dispatch calls the runner) |
| 3 | Repo hygiene and docs: the tracked spec and prompt directories with `.keep` files, the ignore layout, the CHANGELOG bullet | 7 | 7 | — |

Rationale: prompt 1 establishes the dispatch and the refusals; prompt 2 supplies the runner the dispatch calls and makes the image buildable from a clone; prompt 3 is independent and touches no build logic. Row 1 leaves the dispatch naming a runner that row 2 has not yet created — harmless, because every dispatch criterion is a `make -n` that renders without executing, and no intermediate state runs a build.

## Do-Nothing Option

Leaving this repo without a cluster arm keeps one service image tied to this Mac. Every release cut while the operator is away from this machine, every weekly rebuild and every future unattended pipeline needs a human at the keyboard for this one image — while its siblings build themselves. The cost is not the one build; it is that the automation has a hole shaped exactly like the repo nobody is watching.
