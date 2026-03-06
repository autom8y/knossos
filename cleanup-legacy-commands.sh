#!/usr/bin/env bash
# Remove orphaned legacy session command files from .claude/commands/
# These have no source mena after commit 3cf6c15 (legacy dromena removal).
# Safe: all targets are gitignored materialized output.

LEGACY=(start park continue wrap)

PROJECTS=(
  ~/Code/knossos
  ~/code/a8
  ~/code/autom8
  ~/code/autom8y
  ~/code/autom8y-ads
  ~/code/autom8y-asana
  ~/code/autom8y-data
  ~/code/autom8y-hello-world
  ~/code/autom8y-sms
)

removed=0

for project in "${PROJECTS[@]}"; do
  dir="$project/.claude/commands"
  if [ ! -d "$dir" ]; then
    continue
  fi

  header_printed=false
  for cmd in "${LEGACY[@]}"; do
    if [ -f "$dir/$cmd.md" ]; then
      if [ "$header_printed" = false ]; then
        echo "==> $project"
        header_printed=true
      fi
      rm "$dir/$cmd.md"
      echo "    rm $cmd.md"
      removed=$((removed + 1))
    fi
    if [ -d "$dir/$cmd" ]; then
      if [ "$header_printed" = false ]; then
        echo "==> $project"
        header_printed=true
      fi
      count=$(find "$dir/$cmd" -type f | wc -l | tr -d ' ')
      rm -rf "$dir/$cmd"
      echo "    rm -rf $cmd/ ($count files)"
      removed=$((removed + count))
    fi
  done
  if [ "$header_printed" = true ]; then
    echo ""
  fi
done

echo "Done. Removed $removed files across ${#PROJECTS[@]} projects."
