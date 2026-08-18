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

- 中继 WS 端点：`ws://3.17.63.33:8080`（纯文本）或 `wss://`（需反向代理 TLS）。
- AWS 安全组需放行入站 TCP **8080**（`launch-wizard-1`）。