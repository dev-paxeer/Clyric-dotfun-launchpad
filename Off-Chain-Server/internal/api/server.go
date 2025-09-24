package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Server { return &Server{DB: db} }

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/health":
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	case r.Method == http.MethodGet && r.URL.Path == "/pools":
		s.handleListPools(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/pools/") && strings.HasSuffix(r.URL.Path, "/state"):
		s.handlePoolState(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/pools/") && strings.HasSuffix(r.URL.Path, "/price-updates"):
		s.handlePriceUpdates(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/pools/") && strings.HasSuffix(r.URL.Path, "/candles"):
		s.handleCandles(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/pools/") && strings.HasSuffix(r.URL.Path, "/swaps"):
		s.handleSwaps(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
	}
}

func (s *Server) handleListPools(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query(r.Context(), `SELECT pool_address, token_address, oracle_address, created_block, created_tx, created_time, reserve_usdc, reserve_token, spot_x18, floor_x18 FROM pools ORDER BY created_block DESC`)
	if err != nil {
		writeErr(w, err)
		return
	}
	type row struct {
		Pool      string     `json:"pool"`
		Token     string     `json:"token"`
		Oracle    string     `json:"oracle"`
		Block     int64      `json:"createdBlock"`
		Tx        string     `json:"createdTx"`
		Time      *time.Time `json:"createdTime,omitempty"`
		ReserveUS string     `json:"reserveUSDC,omitempty"`
		ReserveT  string     `json:"reserveToken,omitempty"`
		SpotX18   string     `json:"spotX18,omitempty"`
		FloorX18  string     `json:"floorX18,omitempty"`
	}
	var out []row
	for rows.Next() {
		var rr row
		_ = rows.Scan(&rr.Pool, &rr.Token, &rr.Oracle, &rr.Block, &rr.Tx, &rr.Time, &rr.ReserveUS, &rr.ReserveT, &rr.SpotX18, &rr.FloorX18)
		out = append(out, rr)
	}
	out = []row{}
	for rows.Next() {
		var rr row
		_ = rows.Scan(&rr.Pool, &rr.Token, &rr.Oracle, &rr.Block, &rr.Tx, &rr.Time, &rr.ReserveUS, &rr.ReserveT, &rr.SpotX18, &rr.FloorX18)
		out = append(out, rr)
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (s *Server) handlePoolState(w http.ResponseWriter, r *http.Request) {
	pool := extractBetween(r.URL.Path, "/pools/", "/state")
	var rr struct {
		Pool      string     `json:"pool"`
		Token     string     `json:"token"`
		Oracle    string     `json:"oracle"`
		Block     int64      `json:"createdBlock"`
		Tx        string     `json:"createdTx"`
		Time      *time.Time `json:"createdTime,omitempty"`
		ReserveUS string     `json:"reserveUSDC,omitempty"`
		ReserveT  string     `json:"reserveToken,omitempty"`
		SpotX18   string     `json:"spotX18,omitempty"`
		FloorX18  string     `json:"floorX18,omitempty"`
	}
	err := s.DB.QueryRow(r.Context(), `SELECT pool_address, token_address, oracle_address, created_block, created_tx, created_time, reserve_usdc, reserve_token, spot_x18, floor_x18 FROM pools WHERE pool_address = $1`, pool).Scan(&rr.Pool, &rr.Token, &rr.Oracle, &rr.Block, &rr.Tx, &rr.Time, &rr.ReserveUS, &rr.ReserveT, &rr.SpotX18, &rr.FloorX18)
	if err != nil {
		writeErr(w, err)
		return
	}
	_ = json.NewEncoder(w).Encode(rr)
}

func (s *Server) handlePriceUpdates(w http.ResponseWriter, r *http.Request) {
	pool := extractBetween(r.URL.Path, "/pools/", "/price-updates")
	limit := parseIntDefault(r.URL.Query().Get("limit"), 200)
	fromBlock := parseIntDefault(r.URL.Query().Get("fromBlock"), 0)
	rows, err := s.DB.Query(r.Context(), `SELECT price_x18, floor_x18, block_number, tx_hash, log_index, block_time FROM price_updates WHERE pool_address = $1 AND block_number >= $2 ORDER BY block_number DESC, log_index DESC LIMIT $3`, pool, fromBlock, limit)
	if err != nil {
		writeErr(w, err)
		return
	}
	type row struct{
		PriceX18 string `json:"priceX18"`
		FloorX18 string `json:"floorX18"`
		Block int64 `json:"blockNumber"`
		Tx string `json:"txHash"`
		LogIndex int `json:"logIndex"`
		Time *time.Time `json:"blockTime,omitempty"`
	}
	var out []row
	out = []row{}
	for rows.Next() {
		var rr row
		_ = rows.Scan(&rr.PriceX18, &rr.FloorX18, &rr.Block, &rr.Tx, &rr.LogIndex, &rr.Time)
		out = append(out, rr)
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (s *Server) handleSwaps(w http.ResponseWriter, r *http.Request) {
	pool := extractBetween(r.URL.Path, "/pools/", "/swaps")
	limit := parseIntDefault(r.URL.Query().Get("limit"), 100)
	rows, err := s.DB.Query(r.Context(), `SELECT sender, usdc_to_token, amount_in, amount_out, recipient, block_number, tx_hash, log_index, block_time FROM swaps WHERE pool_address = $1 ORDER BY block_number DESC, log_index DESC LIMIT $2`, pool, limit)
	if err != nil { writeErr(w, err); return }
	type row struct{
		Sender string `json:"sender"`
		USDCToToken bool `json:"usdcToToken"`
		AmountIn string `json:"amountIn"`
		AmountOut string `json:"amountOut"`
		Recipient string `json:"recipient"`
		Block int64 `json:"blockNumber"`
		Tx string `json:"txHash"`
		LogIndex int `json:"logIndex"`
		Time *time.Time `json:"blockTime,omitempty"`
	}
	var out []row
	out = []row{}
	for rows.Next() {
		var rr row
		_ = rows.Scan(&rr.Sender, &rr.USDCToToken, &rr.AmountIn, &rr.AmountOut, &rr.Recipient, &rr.Block, &rr.Tx, &rr.LogIndex, &rr.Time)
		out = append(out, rr)
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (s *Server) handleCandles(w http.ResponseWriter, r *http.Request) {
	pool := extractBetween(r.URL.Path, "/pools/", "/candles")
	interval := r.URL.Query().Get("interval")
	if interval == "" { interval = "5m" }
	bucket, ok := bucketSeconds(interval)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid interval"})
		return
	}
	limit := parseIntDefault(r.URL.Query().Get("limit"), 200)
	t := struct{
		Bucket int64
		Pool string
		Limit int
	}{bucket, pool, limit}
	query := `
	WITH b AS (
		SELECT to_timestamp(floor(extract(epoch from coalesce(block_time, now())) / $1) * $1) AS bucket_time,
		       price_x18, block_time
		FROM price_updates WHERE pool_address = $2
	), o AS (
		SELECT bucket_time,
			(ARRAY_AGG(price_x18 ORDER BY block_time ASC))[1] AS open,
			MAX(price_x18) AS high,
			MIN(price_x18) AS low,
			(ARRAY_AGG(price_x18 ORDER BY block_time DESC))[1] AS close
		FROM b GROUP BY bucket_time ORDER BY bucket_time DESC LIMIT $3
	)
	SELECT bucket_time, open, high, low, close FROM o ORDER BY bucket_time ASC`
	rows, err := s.DB.Query(context.Background(), query, t.Bucket, t.Pool, t.Limit)
	if err != nil { writeErr(w, err); return }
	type row struct{
		BucketTime time.Time `json:"bucketTime"`
		Open string `json:"open"`
		High string `json:"high"`
		Low string `json:"low"`
		Close string `json:"close"`
	}
	var out []row
	out = []row{}
	for rows.Next() {
		var rr row
		_ = rows.Scan(&rr.BucketTime, &rr.Open, &rr.High, &rr.Low, &rr.Close)
		out = append(out, rr)
	}
	_ = json.NewEncoder(w).Encode(out)
}

func writeErr(w http.ResponseWriter, err error) {
	log.Println("api error:", err)
	writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func extractBetween(s, prefix, suffix string) string {
	start := strings.Index(s, prefix)
	if start == -1 { return "" }
	start += len(prefix)
	end := strings.LastIndex(s, suffix)
	if end == -1 || end < start { return "" }
	return s[start:end]
}

func parseIntDefault(s string, def int) int {
	if s == "" { return def }
	i, err := strconv.Atoi(s); if err != nil { return def }
	return i
}

var bucketRe = regexp.MustCompile(`^(\d+)([mhd])$`)

func bucketSeconds(s string) (int64, bool) {
	m := bucketRe.FindStringSubmatch(s)
	if len(m) != 3 { return 0, false }
	n, _ := strconv.Atoi(m[1])
	switch m[2] {
	case "m": return int64(n) * 60, true
	case "h": return int64(n) * 3600, true
	case "d": return int64(n) * 86400, true
	}
	return 0, false
}
