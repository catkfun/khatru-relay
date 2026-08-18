#!/bin/bash
# khatru-relay 部署管理脚本（在 VPS 上使用）
# 用法: ./deploy.sh {install|start|stop|restart|status|uninstall}
set -e

BIN=/usr/local/bin/khatru-relay
SVC=/etc/systemd/system/khatru-relay.service
SVC_NAME=khatru-relay

case "${1:-}" in
  install)
    sudo mkdir -p /var/lib/khatru
    sudo install -m 0755 ./khatru-relay "$BIN"
    echo "[unit]" | sudo tee "$SVC" >/dev/null
    echo "Description=khatru Nostr relay" | sudo tee -a "$SVC" >/dev/null
    echo "After=network.target" | sudo tee -a "$SVC" >/dev/null
    echo "[Service]" | sudo tee -a "$SVC" >/dev/null
    echo "ExecStart=$BIN" | sudo tee -a "$SVC" >/dev/null
    echo "Restart=on-failure" | sudo tee -a "$SVC" >/dev/null
    echo "RestartSec=3" | sudo tee -a "$SVC" >/dev/null
    # 预留环境变量，按需取消注释
    # echo "Environment=KHATRU_ADDR=:8080" | sudo tee -a "$SVC" >/dev/null
    # echo "Environment=KHATRU_DB=/var/lib/khatru/db" | sudo tee -a "$SVC" >/dev/null
    echo "[Install]" | sudo tee -a "$SVC" >/dev/null
    echo "WantedBy=multi-user.target" | sudo tee -a "$SVC" >/dev/null
    sudo systemctl daemon-reload
    sudo systemctl enable --now "$SVC_NAME"
    echo "[install] done"
    ;;
  start)   sudo systemctl start "$SVC_NAME" ;;
  stop)    sudo systemctl stop "$SVC_NAME" ;;
  restart) sudo systemctl restart "$SVC_NAME" ;;
  status)  sudo systemctl status "$SVC_NAME" --no-pager ;;
  uninstall)
    sudo systemctl disable --now "$SVC_NAME"
    sudo rm -f "$SVC" "$BIN"
    sudo systemctl daemon-reload
    echo "[uninstall] done"
    ;;
  *)
    echo "usage: $0 {install|start|stop|restart|status|uninstall}"
    exit 1
    ;;
esac