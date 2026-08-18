#!/bin/bash
# 查询 khatru-relay 最近的 workflow 运行状态
set -e
TOKEN=$(cat /tmp/gh_token.secure)
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/vnd.github+json" \
  "https://api.github.com/repos/catkfun/khatru-relay/actions/runs?per_page=5" \
  | python3 -c "import json,sys; d=json.load(sys.stdin); [print(r['id'], r['status'], r['conclusion'], r['created_at']) for r in d['workflow_runs']]"