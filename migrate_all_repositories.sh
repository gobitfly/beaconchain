#!/usr/bin/env bash
set -euo pipefail

# ============================
# CONFIG (edit to your taste)
# ============================

# Destination monorepo remote
MONOREPO_REMOTE="${MONOREPO_REMOTE:-git@github.com:gobitfly/beaconchain-monorepo.git}"

# Workspace dir for mirrors (kept so re-runs are incremental)
WORKDIR="${WORKDIR:-./_monorepo_migration}"

# List of sources: "<ns> <remote_url> <subdir_name>"
# - ns is used to prefix refs in the monorepo: heads/<ns>/*, tags/<ns>/*
# - subdir_name is the folder under which history will be placed.
SOURCES=(
  "v1 git@github.com:gobitfly/eth2-beaconchain-explorer.git v1"
  "v2 git@github.com:gobitfly/beaconchain.git v2"
  "v3 git@github.com:gobitfly/beaconchain-backend.git v3"
  "docs git@github.com:gobitfly/eth2-knowledge-base.git docs"
  "app git@github.com:gobitfly/eth2-beaconchain-explorer-app.git app"
)

# Push only branches updated within the last N days (set 0 to push ALL)
ACTIVE_DAYS="${ACTIVE_DAYS:-0}"   # 0 = all branches

# Also merge selected mains into monorepo's main?
MERGE_MAINS="${MERGE_MAINS:-true}"   # true|false

# Dry run (show what would be pushed)
DRY_RUN="${DRY_RUN:-false}"          # true|false
# Predeclare array to avoid set -u 'unbound variable' on macOS Bash 3.2
declare -a active_branches=()

# Map: per-namespace default branch (main vs master) — Bash 3 compatible
# v1 and docs use 'master', others use 'main'
main_branch_for() {
  case "$1" in
    v1|docs) echo master ;;
    *) echo main ;;
  esac
}

# Merge order for building monorepo main
MERGE_ORDER=(v1 v2 v3 docs app)

# ============================
# Helpers
# ============================

need() {
  command -v "$1" >/dev/null 2>&1 || { echo "Missing dependency: $1" >&2; exit 1; }
}

say() { printf "\n[%s] %s\n" "$(date +%T)" "$*"; }

# Compute cutoff epoch for ACTIVE_DAYS, portable across Linux/macOS
compute_cutoff_epoch() {
  local days="$1"
  if [[ "$days" -le 0 ]]; then
    echo 0
    return
  fi
  if date -d "now - ${days} days" +%s >/dev/null 2>&1; then
    date -d "now - ${days} days" +%s
  elif command -v gdate >/dev/null 2>&1; then
    gdate -d "now - ${days} days" +%s
  else
    # macOS BSD date fallback
    date -v -"${days}"d +%s
  fi
}

# Cleanup filter-repo state in a mirror repo (refs/original and metadata dir)
clean_filter_repo_state() {
  local mirror_dir="$1"
  # Remove any refs/original/* created by prior filter-repo runs
  while read -r origref; do
    [[ -n "$origref" ]] && git -C "$mirror_dir" update-ref -d "$origref" || true
  done < <(git -C "$mirror_dir" for-each-ref --format='%(refname)' refs/original || true)
  # Remove filter-repo metadata directory to avoid assertion errors when chaining runs
  rm -rf "$mirror_dir/filter-repo" || true
}

# List active branches (by HEAD commit date) in a mirror repo
list_active_branches() {
  local mirror_dir="$1" cutoff_epoch="$2"
  if [[ "$cutoff_epoch" -le 0 ]]; then
    git -C "$mirror_dir" for-each-ref --format='%(refname:short)' refs/heads
  else
    git -C "$mirror_dir" for-each-ref --format='%(refname:short) %(committerdate:unix)' refs/heads \
      | awk -v cutoff="$cutoff_epoch" '$2 >= cutoff { print $1 }'
  fi
}

push_refspec() {
  local mirror_dir="$1" ns="$2" type="$3"   # type: heads|tags
  local src="refs/${type}/*" dst="refs/${type}/${ns}/*"
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "git -C $mirror_dir push $MONOREPO_REMOTE +$src:$dst"
  else
    git -C "$mirror_dir" push "$MONOREPO_REMOTE" "+$src:$dst"
  fi
}

push_selected_branches() {
  local mirror_dir="$1"
  local ns="$2"
  shift 2
  local branches=("$@")
  local br
  for br in "${branches[@]}"; do
    local src="refs/heads/${br}"
    local dst="refs/heads/${ns}/${br}"
    if [[ "$DRY_RUN" == "true" ]]; then
      echo "git -C $mirror_dir push $MONOREPO_REMOTE +$src:$dst"
    else
      git -C "$mirror_dir" push "$MONOREPO_REMOTE" "+$src:$dst"
    fi
  done
}

# ============================
# Preflight
# ============================

need git
need git-filter-repo

mkdir -p "$WORKDIR"
say "Using monorepo remote: $MONOREPO_REMOTE"
say "Workdir: $WORKDIR"
say "Active-days filter: ${ACTIVE_DAYS} (0 = all branches)"
say "Dry run: ${DRY_RUN}"

# Ensure the monorepo exists and we can write to it
if ! git ls-remote "$MONOREPO_REMOTE" >/dev/null 2>&1; then
  echo "Cannot access monorepo remote: $MONOREPO_REMOTE" >&2
  exit 1
fi

# ============================
# Migrate each source
# ============================

cutoff_epoch="$(compute_cutoff_epoch "$ACTIVE_DAYS")"

for entry in "${SOURCES[@]}"; do
  read -r ns remote_url subdir <<<"$entry"

  say "=== Processing $ns: $remote_url → subdir '$subdir/' ==="

  mirror="${WORKDIR}/${ns}-mirror.git"
  if [[ ! -d "$mirror" ]]; then
    say "Cloning pristine mirror (no tags)..."
    git clone --mirror --no-tags "$remote_url" "$mirror" || git clone --mirror "$remote_url" "$mirror"
    # Ensure mirror will not fetch tags on subsequent fetches
    git -C "$mirror" config remote.origin.tagopt --no-tags
  else
    say "Refreshing pristine mirror with a fresh clone (no tags)..."
    rm -rf "$mirror"
    git clone --mirror --no-tags "$remote_url" "$mirror" || git clone --mirror "$remote_url" "$mirror"
    git -C "$mirror" config remote.origin.tagopt --no-tags
  fi

  # Prepare a fresh working repo from the pristine mirror (avoid double-subdir like v1/v1)
  work="${WORKDIR}/${ns}-work.git"
  rm -rf "$work"
  git clone --mirror "$mirror" "$work"

  # Clean previous filter-repo state (if any) and rewrite
  say "Rewriting history under '${subdir}/'..."
  clean_filter_repo_state "$work"
  git -C "$work" filter-repo --to-subdirectory-filter "$subdir" --force


  # Add monorepo remote to work repo if missing
  if ! git -C "$work" remote | grep -q "^monorepo$"; then
    git -C "$work" remote add monorepo "$MONOREPO_REMOTE"
  fi

  if [[ "$ACTIVE_DAYS" -le 0 ]]; then
    say "Pushing ALL branches for $ns (namespaced); skipping tags by design..."
    push_refspec "$work" "$ns" heads
  else
    say "Pushing ONLY branches updated in last ${ACTIVE_DAYS} days for $ns (tags are skipped)..."
    # Ensure array exists under set -u in Bash 3.2 and initialize for this namespace
    active_branches=()
    while IFS= read -r _br; do
      [[ -n "$_br" ]] && active_branches+=("$_br")
    done < <(list_active_branches "$work" "$cutoff_epoch")
    # Ensure the default branch is included even if inactive
    main_br="$(main_branch_for "$ns")"
    if git -C "$work" rev-parse --verify "refs/heads/${main_br}" >/dev/null 2>&1; then
      found=false
      # Iterate over active_branches safely (array is initialized above)
      for b in "${active_branches[@]}"; do
        if [[ "$b" == "$main_br" ]]; then found=true; break; fi
      done
      if [[ "$found" == "false" ]]; then
        active_branches+=("$main_br")
      fi
    fi
    # Length check (safe because array is initialized above)
    if [[ ${#active_branches[@]} -eq 0 ]]; then
      say "No active branches for $ns. Skipping branch push."
    else
      push_selected_branches "$work" "$ns" "${active_branches[@]}"
    fi
  fi
done

# ============================
# Merge mains into monorepo/main (optional)
# ============================

if [[ "$MERGE_MAINS" == "true" ]]; then
  say "Merging selected mains into monorepo 'main' (per default-branch mapping)..."
  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT

  git -C "$tmpdir" init
  git -C "$tmpdir" remote add origin "$MONOREPO_REMOTE"
  git -C "$tmpdir" fetch origin '+refs/heads/*:refs/remotes/origin/*' --prune

  # Create 'main' if missing
  if git -C "$tmpdir" rev-parse --verify origin/main >/dev/null 2>&1; then
    git -C "$tmpdir" checkout -b main origin/main
  else
    git -C "$tmpdir" checkout --orphan main
    git -C "$tmpdir" commit --allow-empty -m "Initialize monorepo main"
  fi

  # Merge all defined namespaces in MERGE_ORDER using their default branch
  for ns in "${MERGE_ORDER[@]}"; do
    main_br="$(main_branch_for "$ns")"
    if git -C "$tmpdir" rev-parse --verify "origin/${ns}/${main_br}" >/dev/null 2>&1; then
      say "Merging ${ns}/${main_br}..."
      git -C "$tmpdir" merge --no-edit --allow-unrelated-histories "origin/${ns}/${main_br}" || {
        echo
        echo "Conflict while merging ${ns}/${main_br}. Resolve in $tmpdir then run:"
        echo "  git add -A && git commit && git push origin main"
        exit 1
      }
    else
      say "No ${ns}/${main_br} found; skipping."
    fi
  done

  # Ensure go.work exists with required content
  gowork_path="$tmpdir/go.work"
  gowork_content='go 1.22
use (
  ./v1
  ./v2/backend
  ./v3
)'
  if [[ "$DRY_RUN" == "true" ]]; then
    say "Dry run: would write go.work and push 'main' to origin."
  else
    printf "%s\n" "$gowork_content" > "$gowork_path"
    git -C "$tmpdir" add go.work
    # Commit only if there are changes (go.work new or updated)
    if ! git -C "$tmpdir" diff --cached --quiet; then
      git -C "$tmpdir" commit -m "Add go.work for monorepo workspace"
    fi
    git -C "$tmpdir" push origin main
    say "Pushed merged 'main' with go.work to monorepo."
  fi
fi

say "Done."
