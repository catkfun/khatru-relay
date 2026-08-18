#!/bin/bash
# 在 VPS ~/khatru-relay 内执行：完成远端创建与推送
set -e
TOKEN=$(cat /tmp/gh_token.secure)
REPO=khatru-relay

# 获取用户名
USER=$(curl -s -H "Authorization: Bearer $TOKEN" https://api.github.com/user \
  | sed -n 's/.*"login": "\([^"]*\)".*/\1/p')
if [ -z "$USER" ]; then echo "FAILED to resolve user" >&2; exit 1; fi
echo "GH_USER=$USER"

# 创建仓库（若已存在则忽略）
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/vnd.github+json" \
  -d "{\"name\":\"$REPO\",\"private\":false}" \
  https://api.github.com/user/repos >/dev/null || true

# 配置远端并推送
git remote remove origin 2>/dev/null || true
git remote add origin "https://${USER}:${TOKEN}@github.com/${USER}/${REPO}.git"
git push -u origin HEAD:main

echo "PUSH_DONE -> https://github.com/${USER}/${REPO}"