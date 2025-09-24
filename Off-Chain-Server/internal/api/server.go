package api

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	DB *pgxpool.Pool
}

// POST /profiles/bootstrap (unauthenticated)
// Creates a minimal profile if one does not exist yet for the given address.
func (s *Server) handleProfilesBootstrap(w http.ResponseWriter, r *http.Request) {
    var in struct{
        Address string `json:"address"`
        Username *string `json:"username"`
    }
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil { writeErr(w, err); return }
    addr := strings.ToLower(strings.TrimSpace(in.Address))
    if matched := regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`).MatchString(addr); !matched {
        writeJSON(w, http.StatusBadRequest, map[string]any{"error":"invalid address"}); return
    }
    uname := "user-" + strings.ToLower(strings.TrimPrefix(addr, "0x"))
    if len(uname) > 12 { uname = uname[:12] }
    if in.Username != nil && strings.TrimSpace(*in.Username) != "" { uname = strings.TrimSpace(*in.Username) }
    avatar := "https://api.dicebear.com/7.x/identicon/svg?seed=" + url.QueryEscape(addr)
    _, _ = s.DB.Exec(r.Context(), `INSERT INTO profiles(address, username, avatar_url) VALUES($1,$2,$3) ON CONFLICT(address) DO NOTHING`, addr, uname, avatar)
    _ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "address": addr})
}

// Serve static uploaded files: GET /uploads/{filename}
func (s *Server) handleStaticUpload(w http.ResponseWriter, r *http.Request) {
    rel := strings.TrimPrefix(r.URL.Path, "/uploads/")
    if rel == "" {
        w.WriteHeader(http.StatusNotFound)
        _ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
        return
    }
    // Prevent path traversal
    rel = filepath.Clean(rel)
    if strings.Contains(rel, "..") {
        w.WriteHeader(http.StatusForbidden)
        _ = json.NewEncoder(w).Encode(map[string]any{"error": "forbidden"})
        return
    }
    abs := filepath.Join("uploads", rel)
    fi, err := os.Stat(abs)
    if err != nil || fi.IsDir() {
        w.WriteHeader(http.StatusNotFound)
        _ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
        return
    }
    f, err := os.Open(abs)
    if err != nil {
        writeErr(w, err)
        return
    }
    defer f.Close()
    // Detect content type
    buf := make([]byte, 512)
    n, _ := f.Read(buf)
    ctype := http.DetectContentType(buf[:n])
    if _, err := f.Seek(0, io.SeekStart); err != nil {
        writeErr(w, err)
        return
    }
    w.Header().Set("Content-Type", ctype)
    _, _ = io.Copy(w, f)
}

// ===== Profiles =====
func (s *Server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
    addr := strings.TrimPrefix(r.URL.Path, "/profiles/")
    if addr == "" { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"missing address"}); return }
    var out struct{
        Address string `json:"address"`
        Username *string `json:"username,omitempty"`
        Bio *string `json:"bio,omitempty"`
        AvatarURL *string `json:"avatarUrl,omitempty"`
        BannerURL *string `json:"bannerUrl,omitempty"`
        WebsiteURL *string `json:"websiteUrl,omitempty"`
        TwitterURL *string `json:"twitterUrl,omitempty"`
        TelegramURL *string `json:"telegramUrl,omitempty"`
        UpdatedAt *time.Time `json:"updatedAt,omitempty"`
    }
    row := s.DB.QueryRow(r.Context(), `SELECT address, username, bio, avatar_url, banner_url, website_url, twitter_url, telegram_url, updated_at FROM profiles WHERE address = $1`, addr)
    if err := row.Scan(&out.Address, &out.Username, &out.Bio, &out.AvatarURL, &out.BannerURL, &out.WebsiteURL, &out.TwitterURL, &out.TelegramURL, &out.UpdatedAt); err != nil {
        writeJSON(w, http.StatusOK, map[string]any{"address": addr})
        return
    }
    _ = json.NewEncoder(w).Encode(out)
}

func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
    addr, ok := s.requireAuth(w, r)
    if !ok { return }
    var in struct{
        Username *string `json:"username"`
        Bio *string `json:"bio"`
        AvatarURL *string `json:"avatarUrl"`
        BannerURL *string `json:"bannerUrl"`
        WebsiteURL *string `json:"websiteUrl"`
        TwitterURL *string `json:"twitterUrl"`
        TelegramURL *string `json:"telegramUrl"`
    }
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil { writeErr(w, err); return }
    if in.Username != nil && len(*in.Username) > 30 { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"username too long"}); return }
    if in.Bio != nil && len(*in.Bio) > 160 { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"bio too long"}); return }
    _, err := s.DB.Exec(r.Context(), `INSERT INTO profiles(address, username, bio, avatar_url, banner_url, website_url, twitter_url, telegram_url)
        VALUES($1,$2,$3,$4,$5,$6,$7,$8)
        ON CONFLICT(address) DO UPDATE SET username=COALESCE(EXCLUDED.username, profiles.username), bio=COALESCE(EXCLUDED.bio, profiles.bio), avatar_url=COALESCE(EXCLUDED.avatar_url, profiles.avatar_url), banner_url=COALESCE(EXCLUDED.banner_url, profiles.banner_url), website_url=COALESCE(EXCLUDED.website_url, profiles.website_url), twitter_url=COALESCE(EXCLUDED.twitter_url, profiles.twitter_url), telegram_url=COALESCE(EXCLUDED.telegram_url, profiles.telegram_url), updated_at=NOW()`,
        addr, in.Username, in.Bio, in.AvatarURL, in.BannerURL, in.WebsiteURL, in.TwitterURL, in.TelegramURL)
    if err != nil { writeErr(w, err); return }
    writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}


// ===== Auth =====
func (s *Server) handleAuthNonce(w http.ResponseWriter, r *http.Request) {
    nonce, err := randomHex(16)
    if err != nil { writeErr(w, err); return }
    if _, err := s.DB.Exec(r.Context(), `INSERT INTO auth_nonces(nonce) VALUES($1) ON CONFLICT DO NOTHING`, nonce); err != nil { writeErr(w, err); return }
    _ = json.NewEncoder(w).Encode(map[string]string{"nonce": nonce})
}

func (s *Server) handleAuthVerify2(w http.ResponseWriter, r *http.Request) {
    var in struct{ Address, Signature, Nonce string }
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil { writeErr(w, err); return }
    if in.Address == "" || in.Signature == "" || in.Nonce == "" { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"missing fields"}); return }
    var used bool
    if err := s.DB.QueryRow(r.Context(), `SELECT used FROM auth_nonces WHERE nonce=$1`, in.Nonce).Scan(&used); err != nil { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"invalid nonce"}); return }
    if used { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"nonce used"}); return }
    msg := "paxeer-login:" + in.Nonce
    ok, err := verifyPersonalSign(in.Address, in.Signature, msg)
    if err != nil || !ok { writeJSON(w, http.StatusUnauthorized, map[string]any{"error":"invalid signature"}); return }
    token, err := randomHex(32)
    if err != nil { writeErr(w, err); return }
    expires := time.Now().Add(30 * 24 * time.Hour)
    if _, err := s.DB.Exec(r.Context(), `INSERT INTO sessions(token, address, expires_at) VALUES($1,$2,$3)`, token, strings.ToLower(in.Address), expires); err != nil { writeErr(w, err); return }
    if _, err := s.DB.Exec(r.Context(), `UPDATE auth_nonces SET used=true WHERE nonce=$1`, in.Nonce); err != nil { writeErr(w, err); return }
    http.SetCookie(w, &http.Cookie{Name: "sid", Value: token, Path: "/", Expires: expires, HttpOnly: true, SameSite: http.SameSiteLaxMode})
    addrLower := strings.ToLower(in.Address)
    defaultUsername := "user-" + strings.ToLower(strings.TrimPrefix(in.Address, "0x"))
    if len(defaultUsername) > 12 { defaultUsername = defaultUsername[:12] }
    defaultAvatar := "https://api.dicebear.com/7.x/identicon/svg?seed=" + url.QueryEscape(in.Address)
    _, _ = s.DB.Exec(r.Context(), `INSERT INTO profiles(address, username, avatar_url) VALUES($1,$2,$3) ON CONFLICT(address) DO NOTHING`, addrLower, defaultUsername, defaultAvatar)
    _ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "address": addrLower})
}

func (s *Server) handleAuthMe2(w http.ResponseWriter, r *http.Request) {
    addr, ok := s.getSessionAddress(r)
    if !ok { writeJSON(w, http.StatusUnauthorized, map[string]any{"error":"unauthorized"}); return }
    _ = json.NewEncoder(w).Encode(map[string]any{"address": addr})
}

func (s *Server) handleAuthLogout2(w http.ResponseWriter, r *http.Request) {
    if c, err := r.Cookie("sid"); err == nil {
        _, _ = s.DB.Exec(r.Context(), `DELETE FROM sessions WHERE token=$1`, c.Value)
        http.SetCookie(w, &http.Cookie{Name: "sid", Value: "", Path: "/", Expires: time.Unix(0,0), HttpOnly: true, SameSite: http.SameSiteLaxMode})
    }
    writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// ===== Explorer Proxies =====
const paxscanBase = "https://paxscan.paxeer.app/api/v2"

func (s *Server) handleProxyHolders2(w http.ResponseWriter, r *http.Request) {
    contract := r.URL.Query().Get("contract")
    limit := r.URL.Query().Get("limit")
    if contract == "" { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"missing contract"}); return }
    if limit == "" { limit = "25" }
    target := paxscanBase + "/tokens/" + url.PathEscape(contract) + "/holders?limit=" + url.QueryEscape(limit)
    s.proxyJSON(w, r, target)
}

func (s *Server) handleProxyAccountTxs2(w http.ResponseWriter, r *http.Request) {
    addr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/explorer/account/"), "/transactions")
    if addr == "" { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"missing address"}); return }
    target := paxscanBase + "/addresses/" + url.PathEscape(addr) + "/transactions"
    if q := r.URL.RawQuery; q != "" { target += "?" + q }
    s.proxyJSON(w, r, target)
}

func (s *Server) handleProxyAccountTokenTxs2(w http.ResponseWriter, r *http.Request) {
    path := strings.TrimPrefix(r.URL.Path, "/explorer/account/")
    parts := strings.Split(path, "/")
    if len(parts) < 2 { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"bad path"}); return }
    addr := parts[0]
    contract := r.URL.Query().Get("contract")
    if contract == "" { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"missing contract"}); return }
    q := r.URL.Query(); q.Del("contract")
    target := paxscanBase + "/addresses/" + url.PathEscape(addr) + "/token-transfers?token=" + url.QueryEscape(contract)
    if qs := q.Encode(); qs != "" { target += "&" + qs }
    s.proxyJSON(w, r, target)
}

func (s *Server) proxyJSON2(w http.ResponseWriter, r *http.Request, target string) {
    req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target, nil)
    if err != nil { writeErr(w, err); return }
    resp, err := http.DefaultClient.Do(req)
    if err != nil { writeErr(w, err); return }
    defer resp.Body.Close()
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(resp.StatusCode)
    _, _ = io.Copy(w, resp.Body)
}

// ===== Pool metadata =====
func (s *Server) handleGetPoolMetadata(w http.ResponseWriter, r *http.Request) {
    pool := extractBetween(r.URL.Path, "/pools/", "/metadata")
    if pool == "" { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"missing pool"}); return }
    var out struct{ Pool, Token string; Name, Symbol, Description, WebsiteURL, TwitterURL, TelegramURL, LogoURL, BannerURL, CreatedBy *string; UpdatedAt *time.Time }
    err := s.DB.QueryRow(r.Context(), `SELECT pool_address, token_address, name, symbol, description, website_url, twitter_url, telegram_url, logo_url, banner_url, created_by, updated_at FROM pool_metadata WHERE pool_address=$1`, pool).
        Scan(&out.Pool, &out.Token, &out.Name, &out.Symbol, &out.Description, &out.WebsiteURL, &out.TwitterURL, &out.TelegramURL, &out.LogoURL, &out.BannerURL, &out.CreatedBy, &out.UpdatedAt)
    if err != nil { writeJSON(w, http.StatusOK, map[string]any{"pool": pool}); return }
    _ = json.NewEncoder(w).Encode(out)
}

func (s *Server) handleUpsertPoolMetadata(w http.ResponseWriter, r *http.Request) {
    addr, ok := s.requireAuth(w, r)
    if !ok { return }
    pool := extractBetween(r.URL.Path, "/pools/", "/metadata")
    if pool == "" { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"missing pool"}); return }
    var in struct{ Token string; Name, Symbol, Description, WebsiteURL, TwitterURL, TelegramURL, LogoURL, BannerURL *string }
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil { writeErr(w, err); return }
    if in.Token == "" { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"missing token"}); return }
    if in.Name != nil && len(*in.Name) > 50 { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"name too long"}); return }
    if in.Symbol != nil && !regexp.MustCompile(`^[A-Z]{3,10}$`).MatchString(*in.Symbol) { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"invalid symbol"}); return }
    if in.Description != nil && len(*in.Description) > 500 { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"description too long"}); return }
    _, err := s.DB.Exec(r.Context(), `INSERT INTO pool_metadata(pool_address, token_address, name, symbol, description, website_url, twitter_url, telegram_url, logo_url, banner_url, created_by)
        VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
        ON CONFLICT(pool_address) DO UPDATE SET name=COALESCE(EXCLUDED.name, pool_metadata.name), symbol=COALESCE(EXCLUDED.symbol, pool_metadata.symbol), description=COALESCE(EXCLUDED.description, pool_metadata.description), website_url=COALESCE(EXCLUDED.website_url, pool_metadata.website_url), twitter_url=COALESCE(EXCLUDED.twitter_url, pool_metadata.twitter_url), telegram_url=COALESCE(EXCLUDED.telegram_url, pool_metadata.telegram_url), logo_url=COALESCE(EXCLUDED.logo_url, pool_metadata.logo_url), banner_url=COALESCE(EXCLUDED.banner_url, pool_metadata.banner_url), updated_at=NOW()`,
        pool, in.Token, in.Name, in.Symbol, in.Description, in.WebsiteURL, in.TwitterURL, in.TelegramURL, in.LogoURL, in.BannerURL, addr)
    if err != nil { writeErr(w, err); return }
    writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// GET /metadata/symbols?symbol=SYMB
func (s *Server) handleSymbolExists(w http.ResponseWriter, r *http.Request) {
    sym := strings.TrimSpace(r.URL.Query().Get("symbol"))
    if sym == "" {
        writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing symbol"})
        return
    }
    var exists bool
    err := s.DB.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM pool_metadata WHERE UPPER(symbol) = UPPER($1))`, sym).Scan(&exists)
    if err != nil {
        writeErr(w, err)
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"exists": exists})
}

// ===== Uploads =====
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseMultipartForm(6 << 20); err != nil { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"invalid form"}); return }
    f, fh, err := r.FormFile("file")
    if err != nil { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"missing file"}); return }
    defer f.Close()
    ftype := r.FormValue("type")
    if ftype != "logo" && ftype != "banner" { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"invalid type"}); return }
    // sniff
    buf := make([]byte, 512)
    n, _ := f.Read(buf)
    ctype := http.DetectContentType(buf[:n])
    if _, err := f.Seek(0, io.SeekStart); err != nil { writeErr(w, err); return }
    if ctype != "image/png" && ctype != "image/jpeg" { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"only PNG/JPEG allowed"}); return }
    var max int64 = 2 << 20
    if ftype == "banner" { max = 3 << 20 }
    if fh.Size > max { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"file too large"}); return }
    if err := os.MkdirAll("uploads", 0o755); err != nil { writeErr(w, err); return }
    ext := ".png"; if ctype == "image/jpeg" { ext = ".jpg" }
    name := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), ftype, ext)
    dst := filepath.Join("uploads", name)
    out, err := os.Create(dst); if err != nil { writeErr(w, err); return }
    defer out.Close()
    if _, err := io.Copy(out, io.LimitReader(f, max)); err != nil { writeErr(w, err); return }
    _ = json.NewEncoder(w).Encode(map[string]any{"url": "/" + dst})
}

// ===== Helpers =====
func (s *Server) getSessionAddress2(r *http.Request) (string, bool) {
    c, err := r.Cookie("sid"); if err != nil { return "", false }
    var addr string; var exp time.Time
    if err := s.DB.QueryRow(r.Context(), `SELECT address, expires_at FROM sessions WHERE token=$1`, c.Value).Scan(&addr, &exp); err != nil { return "", false }
    if time.Now().After(exp) { return "", false }
    return addr, true
}

func (s *Server) requireAuth2(w http.ResponseWriter, r *http.Request) (string, bool) {
    addr, ok := s.getSessionAddress(r)
    if !ok { writeJSON(w, http.StatusUnauthorized, map[string]any{"error":"unauthorized"}); return "", false }
    return addr, true
}

func randomHex2(n int) (string, error) {
    b := make([]byte, n); if _, err := rand.Read(b); err != nil { return "", err }
    return hex.EncodeToString(b), nil
}

func verifyPersonalSign2(address, signatureHex, message string) (bool, error) {
    sig, err := hex.DecodeString(strings.TrimPrefix(signatureHex, "0x")); if err != nil { return false, err }
    if len(sig) != 65 { return false, errors.New("invalid signature length") }
    prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(message))
    hash := crypto.Keccak256Hash([]byte(prefix + message))
    if sig[64] >= 27 { sig[64] -= 27 }
    pub, err := crypto.SigToPub(hash.Bytes(), sig); if err != nil { return false, err }
    rec := crypto.PubkeyToAddress(*pub)
    return strings.EqualFold(rec.Hex(), address), nil
}

func fmtDurationPG2(d time.Duration) string {
    sec := int64(d.Seconds())
    if sec%(24*3600) == 0 { days := sec/(24*3600); if days==1 { return "1 day" }; return fmt.Sprintf("%d days", days) }
    if sec%3600 == 0 { h := sec/3600; if h==1 { return "1 hour" }; return fmt.Sprintf("%d hours", h) }
    if sec%60 == 0 { m := sec/60; if m==1 { return "1 minute" }; return fmt.Sprintf("%d minutes", m) }
    return fmt.Sprintf("%d seconds", sec)
}

func (s *Server) handleUpdateProfile2(w http.ResponseWriter, r *http.Request) {
	addr, ok := s.requireAuth(w, r)
	if !ok {
		return
	}
	var in struct {
		Username *string `json:"username"`
		Bio *string `json:"bio"`
		AvatarURL *string `json:"avatarUrl"`
		BannerURL *string `json:"bannerUrl"`
		WebsiteURL *string `json:"websiteUrl"`
		TwitterURL *string `json:"twitterUrl"`
		TelegramURL *string `json:"telegramUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, err)
		return
	}
	// Basic validation
	if in.Username != nil && len(*in.Username) > 30 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "username too long"})
		return
	}
	if in.Bio != nil && len(*in.Bio) > 160 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "bio too long"})
		return
	}
	// Upsert
	_, err := s.DB.Exec(r.Context(), `INSERT INTO profiles(address, username, bio, avatar_url, banner_url, website_url, twitter_url, telegram_url) 
		VALUES($1,$2,$3,$4,$5,$6,$7,$8) 
		ON CONFLICT(address) DO UPDATE SET username=COALESCE(EXCLUDED.username, profiles.username), bio=COALESCE(EXCLUDED.bio, profiles.bio), 
			avatar_url=COALESCE(EXCLUDED.avatar_url, profiles.avatar_url), banner_url=COALESCE(EXCLUDED.banner_url, profiles.banner_url), 
			website_url=COALESCE(EXCLUDED.website_url, profiles.website_url), twitter_url=COALESCE(EXCLUDED.twitter_url, profiles.twitter_url), telegram_url=COALESCE(EXCLUDED.telegram_url, profiles.telegram_url), updated_at=NOW()`,
		addr, in.Username, in.Bio, in.AvatarURL, in.BannerURL, in.WebsiteURL, in.TwitterURL, in.TelegramURL)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// ===== Comments =====
func (s *Server) handleCommentsList(w http.ResponseWriter, r *http.Request) {
	pool := strings.TrimPrefix(r.URL.Path, "/comments/")
	if pool == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing pool"})
		return
	}
	limit := parseIntDefault(r.URL.Query().Get("limit"), 200)
	rows, err := s.DB.Query(r.Context(), `SELECT author_address, message, created_at FROM comments WHERE pool_address = $1 ORDER BY created_at DESC LIMIT $2`, pool, limit)
	if err != nil {
		writeErr(w, err)
		return
	}
	type row struct {
		Author string `json:"author"`
		Message string `json:"message"`
		CreatedAt time.Time `json:"createdAt"`
	}
	var out []row
	for rows.Next() {
		var rr row
		_ = rows.Scan(&rr.Author, &rr.Message, &rr.CreatedAt)
		out = append(out, rr)
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (s *Server) handleCommentsCount(w http.ResponseWriter, r *http.Request) {
	// path: /pools/{pool}/comments/count OR /comments/{pool}/count (we matched latter)
	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	pool := ""
	for i, p := range parts {
		if p == "comments" && i+1 < len(parts) {
			pool = parts[i+1]
			break
		}
	}
	if pool == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing pool"})
		return
	}
	since := r.URL.Query().Get("since")
	// default: 24h
	dur := 24 * time.Hour
	if since != "" {
		if d, err := time.ParseDuration(since); err == nil {
			dur = d
		}
	}
	var cnt int64
	err := s.DB.QueryRow(r.Context(), `SELECT COUNT(*) FROM comments WHERE pool_address=$1 AND created_at >= NOW() - `+fmtDurationPG(dur)+`::interval`, pool).Scan(&cnt)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"count": cnt})
}

func (s *Server) handleCommentsCreate(w http.ResponseWriter, r *http.Request) {
	addr, ok := s.requireAuth(w, r)
	if !ok {
		return
	}
	pool := strings.TrimPrefix(r.URL.Path, "/comments/")
	if pool == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing pool"})
		return
	}
	var in struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, err)
		return
	}
	if strings.TrimSpace(in.Message) == "" || len(in.Message) > 280 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid message"})
		return
	}
	_, err := s.DB.Exec(r.Context(), `INSERT INTO comments(pool_address, author_address, message) VALUES($1,$2,$3)`, pool, addr, in.Message)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true})
}

// ===== Auth (nonce, verify, me, logout) =====
func (s *Server) handleAuthNonce2(w http.ResponseWriter, r *http.Request) {
	nonce, err := randomHex(16)
	if err != nil {
		writeErr(w, err)
		return
	}
	if _, err := s.DB.Exec(r.Context(), `INSERT INTO auth_nonces(nonce) VALUES($1) ON CONFLICT DO NOTHING`, nonce); err != nil {
		writeErr(w, err)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"nonce": nonce})
}

func (s *Server) handleAuthVerify(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Address, Signature, Nonce string
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, err)
		return
	}
	if in.Address == "" || in.Signature == "" || in.Nonce == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing fields"})
		return
	}
	// Check nonce unused
	var used bool
	if err := s.DB.QueryRow(r.Context(), `SELECT used FROM auth_nonces WHERE nonce=$1`, in.Nonce).Scan(&used); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid nonce"})
		return
	}
	if used {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "nonce used"})
		return
	}
	msg := "paxeer-login:" + in.Nonce
	ok, err := verifyPersonalSign(in.Address, in.Signature, msg)
	if err != nil || !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid signature"})
		return
	}
	// Create session
	token, err := randomHex(32)
	if err != nil {
		writeErr(w, err)
		return
	}
	expires := time.Now().Add(30 * 24 * time.Hour)
	if _, err := s.DB.Exec(r.Context(), `INSERT INTO sessions(token, address, expires_at) VALUES($1,$2,$3)`, token, strings.ToLower(in.Address), expires); err != nil {
		writeErr(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{ Name: "sid", Value: token, Path: "/", Expires: expires, HttpOnly: true, SameSite: http.SameSiteLaxMode })
	addrLower := strings.ToLower(in.Address)
	defaultUsername := "user-" + strings.ToLower(strings.TrimPrefix(in.Address, "0x"))
	if len(defaultUsername) > 12 { defaultUsername = defaultUsername[:12] }
	defaultAvatar := "https://api.dicebear.com/7.x/identicon/svg?seed=" + url.QueryEscape(in.Address)
	_, _ = s.DB.Exec(r.Context(), `INSERT INTO profiles(address, username, avatar_url) VALUES($1,$2,$3) ON CONFLICT(address) DO NOTHING`, addrLower, defaultUsername, defaultAvatar)
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "address": addrLower})
}

func (s *Server) handleAuthMe(w http.ResponseWriter, r *http.Request) {
    addr, ok := s.getSessionAddress(r)
    if !ok {
        writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
        return
    }
    _ = json.NewEncoder(w).Encode(map[string]any{"address": addr})
}

func (s *Server) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("sid"); err == nil {
		_, _ = s.DB.Exec(r.Context(), `DELETE FROM sessions WHERE token=$1`, c.Value)
		http.SetCookie(w, &http.Cookie{ Name: "sid", Value: "", Path: "/", Expires: time.Unix(0,0), HttpOnly: true, SameSite: http.SameSiteLaxMode })
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
const paxscanBase2 = "https://paxscan.paxeer.app/api/v2"

func (s *Server) handleProxyHolders(w http.ResponseWriter, r *http.Request) {
	contract := r.URL.Query().Get("contract")
	limitStr := r.URL.Query().Get("limit")
	if contract == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing contract"})
		return
	}
	// Clamp limit to a sensible minimum of 1, default 25
	n := 25
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			n = v
		}
	}
	path := fmt.Sprintf("/tokens/%s/holders?limit=%d", url.PathEscape(contract), n)
	s.proxyJSON(w, r, paxscanBase2+path)
}

func (s *Server) handleProxyAccountTxs(w http.ResponseWriter, r *http.Request) {
	addr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/explorer/account/"), "/transactions")
	if addr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing address"})
		return
	}
	q := r.URL.RawQuery
	urlStr := fmt.Sprintf("%s/addresses/%s/transactions", paxscanBase2, url.PathEscape(addr))
	if q != "" {
		urlStr += "?" + q
	}
	s.proxyJSON(w, r, urlStr)
}

func (s *Server) handleProxyAccountTokenTxs(w http.ResponseWriter, r *http.Request) {
	// /explorer/account/{addr}/token-transfers?contract=...
	path := strings.TrimPrefix(r.URL.Path, "/explorer/account/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "bad path"})
		return
	}
	addr := parts[0]
	contract := r.URL.Query().Get("contract")
	if contract == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing contract"})
		return
	}
	q := r.URL.Query()
	q.Del("contract")
	urlStr := fmt.Sprintf("%s/addresses/%s/token-transfers?token=%s", paxscanBase2, url.PathEscape(addr), url.QueryEscape(contract))
	if qs := q.Encode(); qs != "" {
		urlStr += "&" + qs
	}
	s.proxyJSON(w, r, urlStr)
}

func (s *Server) proxyJSON(w http.ResponseWriter, r *http.Request, target string) {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target, nil)
	if err != nil {
		writeErr(w, err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		writeErr(w, err)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

// ===== Helpers =====
func (s *Server) getSessionAddress(r *http.Request) (string, bool) {
	c, err := r.Cookie("sid")
	if err != nil {
		return "", false
	}
	var addr string
	var exp time.Time
	if err := s.DB.QueryRow(r.Context(), `SELECT address, expires_at FROM sessions WHERE token=$1`, c.Value).Scan(&addr, &exp); err != nil {
		return "", false
	}
	if time.Now().After(exp) {
		return "", false
	}
	return addr, true
}

func (s *Server) requireAuth(w http.ResponseWriter, r *http.Request) (string, bool) {
	addr, ok := s.getSessionAddress(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		return "", false
	}
	return addr, true
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func verifyPersonalSign(address, signatureHex, message string) (bool, error) {
	// Normalize signature
	sig, err := hex.DecodeString(strings.TrimPrefix(signatureHex, "0x"))
	if err != nil {
		return false, err
	}
	if len(sig) != 65 {
		return false, errors.New("invalid signature length")
	}
	// Ethereum Signed Message prefix
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(message))
	hash := crypto.Keccak256Hash([]byte(prefix + message))
	// Adjust V if needed
	if sig[64] >= 27 {
		sig[64] -= 27
	}
	pub, err := crypto.SigToPub(hash.Bytes(), sig)
	if err != nil {
		return false, err
	}
	recAddr := crypto.PubkeyToAddress(*pub)
	return strings.EqualFold(recAddr.Hex(), address), nil
}

// Ensure we reference ecdsa to avoid unused import on some platforms
var _ *ecdsa.PrivateKey

func fmtDurationPG(d time.Duration) string {
    // Return an interval literal Postgres can cast with ::interval
    seconds := int64(d.Seconds())
    if seconds%(24*3600) == 0 {
        days := seconds / (24 * 3600)
        if days == 1 {
            return "1 day"
        }
        return fmt.Sprintf("%d days", days)
    }
    if seconds%3600 == 0 {
        hours := seconds / 3600
        if hours == 1 {
            return "1 hour"
        }
        return fmt.Sprintf("%d hours", hours)
    }
    if seconds%60 == 0 {
        minutes := seconds / 60
        if minutes == 1 {
            return "1 minute"
        }
        return fmt.Sprintf("%d minutes", minutes)
    }
    return fmt.Sprintf("%d seconds", seconds)
}

func New(db *pgxpool.Pool) *Server {
	return &Server{DB: db}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // CORS
    origin := r.Header.Get("Origin")
    if origin != "" {
        // Echo explicit origin and allow credentials
        w.Header().Set("Access-Control-Allow-Origin", origin)
        w.Header().Set("Vary", "Origin")
        w.Header().Set("Access-Control-Allow-Credentials", "true")
    } else {
        // Default wildcard for SSR or same-origin requests without Origin
        w.Header().Set("Access-Control-Allow-Origin", "*")
    }
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,OPTIONS")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/uploads/"):
		s.handleStaticUpload(w, r)
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
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/pools/") && strings.HasSuffix(r.URL.Path, "/metadata"):
		s.handleGetPoolMetadata(w, r)
	case (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch) && strings.HasPrefix(r.URL.Path, "/pools/") && strings.HasSuffix(r.URL.Path, "/metadata"):
		s.handleUpsertPoolMetadata(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/metadata/symbols":
		s.handleSymbolExists(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/comments/") && !strings.HasSuffix(r.URL.Path, "/count"):
		s.handleCommentsList(w, r)
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/comments/count"):
		s.handleCommentsCount(w, r)
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/comments/"):
		s.handleCommentsCreate(w, r)
	// Profiles
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/profiles/"):
		s.handleGetProfile(w, r)
	case r.Method == http.MethodPatch && r.URL.Path == "/profiles":
		s.handleUpdateProfile(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/profiles/bootstrap":
		s.handleProfilesBootstrap(w, r)
	// Auth
	case r.Method == http.MethodGet && r.URL.Path == "/auth/nonce":
		s.handleAuthNonce(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/auth/verify":
		s.handleAuthVerify(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/auth/me":
		s.handleAuthMe(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/auth/logout":
		s.handleAuthLogout(w, r)
	// Uploads
	case r.Method == http.MethodPost && r.URL.Path == "/upload":
		s.handleUpload(w, r)
	// Explorer proxy (paxscan api v2)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/explorer/holders"):
		s.handleProxyHolders(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/explorer/account/") && strings.HasSuffix(r.URL.Path, "/transactions"):
		s.handleProxyAccountTxs(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/explorer/account/") && strings.Contains(r.URL.Path, "/token-transfers"):
		s.handleProxyAccountTokenTxs(w, r)
	// Paxscan caching endpoints
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/tokens/") && strings.HasSuffix(r.URL.Path, "/paxscan-sync"):
		s.handlePaxscanSync(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/tokens/") && strings.HasSuffix(r.URL.Path, "/cached"):
		s.handleGetCachedTokenData(w, r)
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
