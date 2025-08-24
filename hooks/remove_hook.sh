#!/usr/bin/env bash

_remove_hook() {
  local rc_file=$1
  if [ -f "$rc_file" ]; then
    echo "🔍 Checking $rc_file for the hook..."
    local tmp_file=$(mktemp)
    sed '/### >>> logger start >>>/,/### <<< logger end <<</d' "$rc_file" > "$tmp_file"
    
    if ! cmp -s "$rc_file" "$tmp_file"; then
        mv "$tmp_file" "$rc_file"
        echo "✅ Hook removed from $rc_file."
    else
        echo "🤔 Hook not found in $rc_file."
        echo "✅ Nothing to remove!"
        rm "$tmp_file"
    fi
  fi
}