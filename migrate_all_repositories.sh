#!/usr/bin/env bash
set -euo pipefail

# ============================
# CONFIG (edit to your taste)
# ============================

# Destination monorepo remote
MONOREPO_REMOTE="${MONOREPO_REMOTE:-git@github.com:gobitfly/beaconchain-monorepo.git}"

# Workspace dir for mirrors (kept so re-runs are incremental)
WORKDIR="${WORKDIR:-./_monorepo_migration}"

# Map: namespace => "remote_url subdir_name"
# - namespace is used to prefix refs in the monorepo: heads/<ns>/*, tags/<ns>/*
# - subdir_name is the folder under which history will be placed.
declare -A SOURCES=(
  [v1]="git@github.com:gobitfly/eth2-beaconchain-explorer.git v1"
  [v2]="git@github.com:gobitfly/beaconchain.git v2"
  [v3]="git@github.com:gobitfly/beaconchain-backend.git v3"
  [docs]="git@github.com:gobitfly/eth2-knowledge-base.git docs"
  [app]="git@github.com:gobitfly/eth2-beaconchain-explorer-app.git app"
)

# Push only branches updated within the last N days (set 0 to push ALL)
ACTIVE_DAYS="${ACTIVE_DAYS:-0}"   # 0 = all branches

# Also merge selected mains into monorepo's main?
MERGE_MAINS="${MERGE_MAINS:-true}"   # true|false

# Dry run (show what would be pushed)
DRY_RUN="${DRY_RUN:-false}"          # true|false

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
  local mirror_dir="$1" ns="$2" branches=("$@"); shift 2
  for br in "${branches[@]:2}"; do
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

for ns in "${!SOURCES[@]}"; do
  read -r remote_url subdir <<<"${SOURCES[$ns]}"

  say "=== Processing $ns: $remote_url → subdir '$subdir/' ==="

  mirror="${WORKDIR}/${ns}-mirror.git"
  if [[ ! -d "$mirror" ]]; then
    say "Cloning mirror..."
    git clone --mirror "$remote_url" "$mirror"
  else
    say "Mirror exists, fetching updates..."
    git -C "$mirror" fetch --all --prune --tags
  fi

  # Clean previous filter-repo refs (if any) and rewrite
  say "Rewriting history under '${subdir}/'..."
  # Remove any refs/original/* created by prior filter-repo runs
  while read -r origref; do
    [[ -n "$origref" ]] && git -C "$mirror" update-ref -d "$origref" || true
  done < <(git -C "$mirror" for-each-ref --format='%(refname)' refs/original || true)
  git -C "$mirror" filter-repo --to-subdirectory-filter "$subdir" --force

  # For Go code repos, rewrite import paths to the new monorepo modules
  case "$ns" in
    v1)
      say "Rewriting imports for v1 to module github.com/gobitfly/beaconchain-monorepo ..."
      mapfile_v1="$(mktemp)"; trap 'rm -f "$mapfile_v1"' RETURN
      {
        echo "github.com/gobitfly/eth2-beaconchain-explorer/ ==> github.com/gobitfly/beaconchain-monorepo/"
        echo "\bmodule[[:space:]]+github.com/gobitfly/eth2-beaconchain-explorer\b ==> module github.com/gobitfly/beaconchain-monorepo"
      } > "$mapfile_v1"
      git -C "$mirror" filter-repo --replace-text "$mapfile_v1" --force
      ;;
    v2)
      say "Rewriting imports for v2 to module github.com/gobitfly/beaconchain-monorepo/v2 ..."
      mapfile_v2="$(mktemp)"; trap 'rm -f "$mapfile_v2"' RETURN
      {
        echo "github.com/gobitfly/beaconchain/ ==> github.com/gobitfly/beaconchain-monorepo/v2/"
        echo "\bmodule[[:space:]]+github.com/gobitfly/beaconchain\b ==> module github.com/gobitfly/beaconchain-monorepo/v2"
      } > "$mapfile_v2"
      git -C "$mirror" filter-repo --replace-text "$mapfile_v2" --force
      ;;
    v3)
      say "Rewriting imports for v3 to module github.com/gobitfly/beaconchain-monorepo/v3 ..."
      mapfile_v3="$(mktemp)"; trap 'rm -f "$mapfile_v3"' RETURN
      {
        echo "github.com/gobitfly/beaconchain-backend/ ==> github.com/gobitfly/beaconchain-monorepo/v3/"
        echo "\bmodule[[:space:]]+github.com/gobitfly/beaconchain-backend\b ==> module github.com/gobitfly/beaconchain-monorepo/v3"
      } > "$mapfile_v3"
      git -C "$mirror" filter-repo --replace-text "$mapfile_v3" --force
      ;;
  esac

  # Add monorepo remote if missing
  if ! git -C "$mirror" remote | grep -q "^monorepo$"; then
    git -C "$mirror" remote add monorepo "$MONOREPO_REMOTE"
  fi

  if [[ "$ACTIVE_DAYS" -le 0 ]]; then
    say "Pushing ALL branches and tags for $ns (namespaced)..."
    push_refspec "$mirror" "$ns" heads
    push_refspec "$mirror" "$ns" tags
  else
    say "Pushing ONLY branches updated in last ${ACTIVE_DAYS} days for $ns..."
    mapfile -t active_branches < <(list_active_branches "$mirror" "$cutoff_epoch")
    if [[ "${#active_branches[@]}" -eq 0 ]]; then
      say "No active branches for $ns. Skipping branch push."
    else
      push_selected_branches "$mirror" "$ns" "${active_branches[@]}"
    fi
    say "Pushing ALL tags for $ns (namespaced) to preserve release history..."
    push_refspec "$mirror" "$ns" tags
  fi
done

# ============================
# Merge mains into monorepo/main (optional)
# ============================

if [[ "$MERGE_MAINS" == "true" ]]; then
  say "Merging v1/main + v2/main + v3/main into monorepo 'main'..."
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

  # Merge the three mains (namespaced)
  for ns in v1 v2 v3; do
    if git -C "$tmpdir" rev-parse --verify "origin/${ns}/main" >/dev/null 2>&1; then
      say "Merging ${ns}/main..."
      git -C "$tmpdir" merge --no-edit --allow-unrelated-histories "origin/${ns}/main" || {
        echo
        echo "Conflict while merging ${ns}/main. Resolve in $tmpdir then run:"
        echo "  git add -A && git commit && git push origin main"
        exit 1
      }
    else
      say "No ${ns}/main found; skipping."
    fi
  done

  if [[ "$DRY_RUN" == "true" ]]; then
    say "Dry run: would push 'main' to origin."
  else
    git -C "$tmpdir" push origin main
    say "Pushed merged 'main' to monorepo."
  fi
fi

say "Done."
