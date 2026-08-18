# khatru-relay — Nostr 中继

基于 [khatru](https://github.com/fiatjaf/khatru) 框架、SQLite 持久化存储的 Nostr 中继，监听 8080 端口，部署在 VPS `3.17.63.33`。

## 编译（GitHub Actions）

代码推送到远程后，在 GitHub 上运行 `Build khatru-relay` workflow（或手动触发），会产出 `khatru-relay` linux/amd64 静态二进制，作为 artifact 上传，可下载。

```bash
git add .
git commit -m "khatru relay"
git push origin main
```

下载 artifact 得到 `khatru-relay` 可执行文件。

## 部署到 VPS

1. 上传二进制：
   ```powershell
   scp -i "C:\Users\Administrator\Downloads\my-debian-key.pem" .\khatru-relay admin@3.17.63.33:~/
   ```
2. 上传部署脚本并执行安装：
   ```powershell
   scp -i "C:\Users\Administrator\Downloads\my-debian-key.pem" .\deploy.sh admin@3.17.63.33:~/
   ssh -i "C:\Users\Administrator\Downloads\my-debian-key.pem" admin@3.17.63.33 "chmod +x ~/deploy.sh && sudo ~/deploy.sh install"
   ```

## 管理

```bash
sudo ~/deploy.sh status     # 查看状态
sudo ~/deploy.sh restart    # 重启
sudo ~/deploy.sh uninstall  # 卸载
```

数据库保存在 `/var/lib/khatru/db`，日志用 `journalctl -u khatru-relay -f` 查看。

## 配置（环境变量）

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `KHATRU_ADDR` | `:8080` | 监听地址 |
| `KHATRU_DB` | `/var/lib/khatru/db` | SQLite 数据库路径 |

## 网络访问

- 中继 WS 端点：
  - 加密：`wss://relay.catk.fun`（**推荐对外，自动 TLS**）
  - 明文：`ws://3.17.63.33:8080`（内网/测试）
- **端口放行需两层**，缺一不可：
  1. AWS 安全组 `launch-wizard-1`（sg-0b4e0b4435d814960）放行入站 TCP **80 / 443 / 8080**（8080 明文、80/443 用于 ACME 挑战与 wss TLS）。
  2. VPS 本机 ufw 放行：`sudo ufw allow 80/tcp && sudo ufw allow 443/tcp && sudo ufw allow 8080/tcp`（ufw 默认 deny incoming）。
- 已验证（外网可达）：TCP 8080 通、NIP-11 返回 200、WebSocket 握手 101。

## wss（TLS）— Caddy 反代

- 域名 `relay.catk.fun` A 记录 → `3.17.63.33`（腾讯云 DNSPod）。
- Caddy v2.11.4 反代 `443 → 127.0.0.1:8080`，自动申请/续期 Let's Encrypt 证书，配置见 `Caddyfile`（已装到 VPS `/etc/caddy/Caddyfile`）。
- 放行时注意：ACME 的 http-01 挑战需要外网能访问 **80 端口**，否则证书签不下来（曾踩坑）。
- 验证：证书由 Let's Encrypt 颁发且受信任；`wss://relay.catk.fun` 端到端测试通过（见 `ws_e2e_wss.py`），EVENT/REQ/EOSE 正常。

## 身份与状态

- NIP-11 中继身份：`name=wt khatru relay`，contact=运营者邮箱，pubkey=运营者 Nostr 公钥（[main.go](main.go) 中的 `Relay.Info`）。
- 实时状态页：`https://relay.catk.fun/status`（在线/运行时长/事件计数/NIP 列表）。
- 中继 pubkey 使用**运营者本人账号公钥**（`f6c24a8a...` 对应 npub `npub17mpy4...`），中继无独立密钥体系。

## 公开收录（PR 进度）

- **目标仓库**：[CodyTseng/awesome-nostr-relays](https://github.com/CodyTseng/awesome-nostr-relays)（注：`relay.directory` 域名已失效/被劫持，弃用）。
- 已将 `wss://relay.catk.fun/` 加入其 `data/collections.yaml` 的 **global** 集合，提交 PR：**[awesome-nostr-relays#12](https://github.com/CodyTseng/awesome-nostr-relays/pull/12)**（状态 open，待维护者合并）。
- 提交流程脚本留存于 [scripts/submit_relay_pr.sh](scripts/submit_relay_pr.sh)（fork→改集合→push→创建 PR，全自动）。
- nostr.watch 无手动提交入口，靠自动扫描收录，未单独处理。

## ⚠️ 安全提醒

- 建中继期间，一个旧 GitHub PAT（含 `repo`/`admin:org` 等高层级权限）曾在 SSH 命令输出中**明文出现在对话记录**，并用于 PR 提交流程。旧 token 已**轮换**。
- **已处理**：VPS 上 khatru-relay 的 git remote 凭据已更换为**新 token**，新 token 验证有效（`login=catkfun`），旧 token 已作废。轮换脚本留存 `scripts/rotate_token.sh`。
- 后续：避免在命令行/对话中回显 token（采用脚本内引用脱敏）；建议最终改用最小权限 fine-grained token 或 SSH key 部署。