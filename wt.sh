# -----------------------------------------------------------------------------
# wt.sh – a Git-Worktree helper (POSIX-compatible, no external compile needed)
#
# Usage (after sourcing in your shell):
#   wt list
#   wt add  <branch>
#   wt rm   <branch>
#   wt go   <index|branch>
#   wt dash
#
# For "go" and "dash", the function itself will perform a `cd` when possible.
# -----------------------------------------------------------------------------

wt() {
  cmd=$1
  arg=$2

  # 1) Find the Git top-level directory (repo root). If not in a git repo, print error.
  repo=$(git rev-parse --show-toplevel 2>/dev/null) || {
    printf 'wt: not inside a Git repository.\n' >&2
    return 1
  }

  # 2) Derive repo name and worktree base
  #    e.g. repo=/home/user/projects/myapp
  #         repo_name=myapp
  #         worktree_base=/home/user/projects/myapp-worktrees
  repo_name=$(basename "$repo")
  worktree_parent=$(dirname "$repo")
  worktree_base="$worktree_parent/${repo_name}-worktrees"

  # 3) Helper: build arrays of existing worktree paths and their branch names
  build_worktree_list() {
    # Parse `git worktree list --porcelain`
    # Output lines:
    #   worktree /full/path
    #   HEAD     xxxxxx...
    #   branch   refs/heads/<branch>
    #
    PATHS=()
    BRANCHES=()
    # Use a temp file (POSIX)
    tmpfile=$(mktemp) || return 1
    git -C "$repo" worktree list --porcelain >"$tmpfile"
    current_path=
    while IFS= read -r line; do
      case "$line" in
        worktree\ *)
          current_path=${line#worktree }
          ;;
        branch\ refs/heads/*)
          branch_name=${line#branch refs/heads/}
          PATHS+=("$current_path")
          BRANCHES+=("$branch_name")
          ;;
      esac
    done < "$tmpfile"
    rm -f "$tmpfile"
  }

  case "$cmd" in
    list)
      build_worktree_list
      if [ "${#PATHS[@]}" -eq 0 ]; then
        printf 'wt: no worktrees found.\n'
        return 0
      fi
      printf '%-5s %-20s %s\n' Index Branch Path
      i=0
      while [ "$i" -lt "${#PATHS[@]}" ]; do
        printf '%-5d %-20s %s\n' "$i" "${BRANCHES[$i]}" "${PATHS[$i]}"
        i=$((i + 1))
      done
      ;;

    add)
      if [ -z "$arg" ]; then
        printf 'Usage: wt add <branch>\n' >&2
        return 1
      fi
      mkdir -p "$worktree_base" || return 1
      git -C "$repo" worktree add "$worktree_base/$arg" "$arg"
      ;;

    rm)
      if [ -z "$arg" ]; then
        printf 'Usage: wt rm <branch>\n' >&2
        return 1
      fi
      git -C "$repo" worktree remove "$worktree_base/$arg"
      ;;

    go)
      if [ -z "$arg" ]; then
        printf 'Usage: wt go <index|branch>\n' >&2
        return 1
      fi
      build_worktree_list
      if [ "${#PATHS[@]}" -eq 0 ]; then
        printf 'wt: no worktrees exist.\n' >&2
        return 1
      fi

      # If arg is a non-negative integer, treat as index
      case "$arg" in
        ''|*[!0-9]*)
          # not a pure number → try matching branch
          found=0
          i=0
          while [ "$i" -lt "${#BRANCHES[@]}" ]; do
            if [ "${BRANCHES[$i]}" = "$arg" ]; then
              cd "${PATHS[$i]}" || {
                printf 'wt: failed to cd to %s\n' "${PATHS[$i]}" >&2
                return 1
              }
              found=1
              break
            fi
            i=$((i + 1))
          done
          if [ "$found" -eq 0 ]; then
            printf "wt: branch '%s' not found among worktrees.\n" "$arg" >&2
            return 1
          fi
          ;;
        *)
          # purely numeric
          idx=$((arg))
          if [ "$idx" -ge 0 ] && [ "$idx" -lt "${#PATHS[@]}" ]; then
            cd "${PATHS[$idx]}" || {
              printf 'wt: failed to cd to %s\n' "${PATHS[$idx]}" >&2
              return 1
            }
          else
            printf "wt: index %d out of range (0..%d).\n" "$idx" "$(( ${#PATHS[@]} - 1 ))" >&2
            return 1
          fi
          ;;
      esac
      ;;

    dash)
      dash_path="$repo/applications/dashboard-app"
      if [ -d "$dash_path" ]; then
        cd "$dash_path" || {
          printf 'wt: failed to cd to %s\n' "$dash_path" >&2
          return 1
        }
      else
        printf "wt: 'applications/dashboard-app' not found under repo.\n" >&2
        return 1
      fi
      ;;

    *)
      printf 'Usage: wt [list|add <branch>|rm <branch>|go <index|branch>|dash]\n' >&2
      return 1
      ;;
  esac
}
