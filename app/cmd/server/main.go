package main

import (
	"log"
	"net/http"
	"os"

	"github.com/heidericklucas/go-security-swe-bench/app/auth"
	"github.com/heidericklucas/go-security-swe-bench/app/clock"
	"github.com/heidericklucas/go-security-swe-bench/app/fetch"
	"github.com/heidericklucas/go-security-swe-bench/app/httpapi"
	"github.com/heidericklucas/go-security-swe-bench/app/seed"
	"github.com/heidericklucas/go-security-swe-bench/app/store"
)

func main() {
	addr := envOr("APP_ADDR", ":8080")
	filesDir := envOr("APP_FILES_DIR", "/seed/files")

	clk := clock.System{}
	st, err := store.New(envOr("APP_DB", ":memory:"), clk)
	if err != nil {
		log.Fatal(err)
	}
	keys := seed.FixedRSAKeys()
	if _, err := seed.Load(st, filesDir); err != nil {
		log.Fatal(err)
	}
	v := auth.NewVerifier(keys, clk)
	h := httpapi.New(st, v, fetch.New(), filesDir)
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, h))
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
