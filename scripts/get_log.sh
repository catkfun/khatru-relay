#!/bin/bash
# 获取指定 run 的失败日志
set -e
TOKEN=$(cat /tmp/gh_token.secure)
RUN_ID="${1:-}"
if [ -z "$RUN_ID" ]; then echo "usage: $0 <run_id>"; exit 1; fi
curl -sL -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/vnd.github+json" \
  "https://api.github.com/repos/catkfun/khatru-relay/actions/runs/$RUN_ID/logs" \
  -o /tmp/run_log.zip
echo "log size: $(stat -c%s /tmp/run_log.zip 2>/dev/null || echo '?')"
mkdir -p /tmp/run_log && rm -rf /tmp/run_log/*
unzip -o -q /tmp/run_log.zip -d /tmp/run_log 2>/dev/null || echo "unzip failed"
# 打印包含 error/fail/build 相关行
grep -rinE "error|cannot|failed|no required|go: cannot|could not" /tmp/run_log 2>/dev/null | grep -viE "0 error" | head -40 || echo "(no error lines)"