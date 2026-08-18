package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/fiatjaf/eventstore/sqlite3"
	khatru "github.com/fiatjaf/khatru"
	"github.com/fiatjaf/khatru/policies"
	"github.com/nbd-wtf/go-nostr"
)

func main() {
	// 数据库存放路径，可用环境变量覆盖
	dbPath := os.Getenv("KHATRU_DB")
	if dbPath == "" {
		dbPath = "/var/lib/khatru/db"
	}
	addr := os.Getenv("KHATRU_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	// 持久化存储（SQLite，单机性能好、低维护）
	db := &sqlite3.SQLite3Backend{DatabaseURL: dbPath}
	if err := db.Init(); err != nil {
		log.Fatalf("init sqlite: %v", err)
	}

	relay := khatru.NewRelay()
	relay.Info.Name = "wt khatru relay"
	relay.Info.Description = "Nostr relay powered by khatru for wt geohash channel"
	relay.Info.Software = "github.com/fiatjaf/khatru"
	relay.Info.Contact = "kai19961201@gmail.com"
	relay.Info.PubKey = "f6c24a8a503917702514e9812b4b91b692c13937e0d24cbffa4e9bd814ad9091"
	relay.Info.SupportedNIPs = []any{1, 2, 9, 11, 15, 16, 20, 22, 28, 33, 40, 42}

	// 实时状态统计：进程启动时间 + 已保存事件计数（原子，避免并发竞争）
	var (
		startedAt       = time.Now()
		savedEventCount int64
	)

	// 存储回调：入库 / 查询 / 计数 / 删除 / 替换
	relay.StoreEvent = append(relay.StoreEvent, db.SaveEvent)
	relay.QueryEvents = append(relay.QueryEvents, db.QueryEvents)
	relay.CountEvents = append(relay.CountEvents, db.CountEvents)
	relay.DeleteEvent = append(relay.DeleteEvent, db.DeleteEvent)
	relay.ReplaceEvent = append(relay.ReplaceEvent, db.ReplaceEvent)

	// 注入一组安全、合理的默认策略
	policies.ApplySaneDefaults(relay)

	// 额外 HTTP 处理器（根路径由 khatru 原生处理：WebSocket 升级 + NIP-11）
	mux := relay.Router()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// 事件入库后日志（可选） + 递增实时状态计数
	relay.OnEventSaved = append(relay.OnEventSaved, func(ctx context.Context, evt *nostr.Event) {
		atomic.AddInt64(&savedEventCount, 1)
		log.Printf("saved id=%s kind=%d pubkey=%s", evt.ID[:8], evt.Kind, evt.PubKey[:8])
	})

	// 实时状态页 /status：运行时长、监听地址、事件计数、NIP-11 信息、最近开机时间
	statusTmpl := template.Must(template.New("status").Parse(`<!DOCTYPE html>
<html lang="zh">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Name}} - 中继状态</title>
<style>
  body{font-family:system-ui,-apple-system,"Segoe UI",sans-serif;background:#0f172a;color:#e2e8f0;margin:0;padding:0}
  .wrap{max-width:640px;margin:48px auto;padding:0 20px}
  h1{font-size:1.5rem;margin:0 0 4px}
  .sub{color:#94a3b8;font-size:.9rem;margin-bottom:28px;word-break:break-all}
  .grid{display:grid;grid-template-columns:1fr 1fr;gap:16px}
  .card{background:#1e293b;border:1px solid #334155;border-radius:12px;padding:20px}
  .card.full{grid-column:1 / -1}
  .label{font-size:.75rem;text-transform:uppercase;letter-spacing:.06em;color:#64748b;margin-bottom:8px}
  .value{font-size:1.6rem;font-weight:600}
  .value.sm{font-size:1rem;font-weight:400;line-height:1.6;color:#cbd5e1}
  .badge{display:inline-block;background:#065f46;color:#6ee7b7;border:1px solid #047857;border-radius:999px;padding:2px 10px;font-size:.8rem;margin:0 4px 4px 0}
  .ok{color:#4ade80}.warn{color:#fbbf24}
  footer{color:#475569;font-size:.8rem;text-align:center;margin-top:24px}
</style>
</head>
<body>
<div class="wrap">
  <h1>{{.Name}}</h1>
  <div class="sub">{{.Description}}</div>
  <div class="grid">
    <div class="card"><div class="label">状态</div><div class="value ok">● 在线</div></div>
    <div class="card"><div class="label">运行时长</div><div class="value">{{.Uptime}}</div></div>
    <div class="card"><div class="label">已保存事件</div><div class="value">{{.Events}}</div></div>
    <div class="card"><div class="label">监听</div><div class="value sm">{{.Addr}}</div></div>
    <div class="card full"><div class="label">数据库</div><div class="value sm">{{.DB}}</div></div>
    <div class="card full"><div class="label">支持的 NIP</div><div>{{range .NIPs}}<span class="badge">{{.}}</span>{{end}}</div></div>
  </div>
  <footer>软件 {{.Software}} · 自启时间 {{.Started}}</footer>
</div>
</body>
</html>`))
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, r *http.Request) {
		started := startedAt
		uptime := time.Since(started)
		h := map[string]any{
			"Name":        relay.Info.Name,
			"Description": relay.Info.Description,
			"Addr":        addr,
			"DB":          dbPath,
			"Software":    relay.Info.Software,
			"Uptime":      fmt.Sprintf("%dd %02dh %02dm", int(uptime.Hours())/24, int(uptime.Hours())%24, int(uptime.Minutes())%60),
			"Events":      atomic.LoadInt64(&savedEventCount),
			"NIPs":        relay.Info.SupportedNIPs,
			"Started":     started.Format("2006-01-02 15:04:05 MST"),
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := statusTmpl.Execute(w, h); err != nil {
			log.Printf("status render err: %v", err)
		}
	})

	log.Printf("khatru relay running on %s (db=%s)", addr, dbPath)
	if err := http.ListenAndServe(addr, relay); err != nil {
		log.Fatalf("listen: %v", err)
	}
}