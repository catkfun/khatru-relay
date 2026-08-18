package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

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
	relay.Info.SupportedNIPs = []int{1, 2, 9, 11, 15, 16, 20, 22, 28, 33, 40, 42}

	// 存储回调：入库 / 查询 / 计数 / 删除 / 替换
	relay.StoreEvent = append(relay.StoreEvent, db.SaveEvent)
	relay.QueryEvents = append(relay.QueryEvents, db.QueryEvents)
	relay.CountEvents = append(relay.CountEvents, db.CountEvents)
	relay.DeleteEvent = append(relay.DeleteEvent, db.DeleteEvent)
	relay.ReplaceEvent = append(relay.ReplaceEvent, db.ReplaceEvent)

	// 注入一组安全、合理的默认策略
	policies.ApplySaneDefaults(relay)

	// 自定义 HTTP 处理器（如欢迎页、健康检查）
	mux := relay.Router()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/plain")
		fmt.Fprintf(w, "wt khatru relay ready.\n")
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// 事件入库后日志（可选）
	relay.OnEventSaved = append(relay.OnEventSaved, func(ctx context.Context, evt *nostr.Event) {
		log.Printf("saved id=%s kind=%d pubkey=%s", evt.ID[:8], evt.Kind, evt.PubKey[:8])
	})

	log.Printf("khatru relay running on %s (db=%s)", addr, dbPath)
	if err := http.ListenAndServe(addr, relay); err != nil {
		log.Fatalf("listen: %v", err)
	}
}