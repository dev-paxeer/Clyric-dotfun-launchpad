package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// PaxscanTokenData represents token info from Paxscan API
type PaxscanTokenData struct {
	AddressHash         string  `json:"address_hash"`
	CirculatingMarketCap *string `json:"circulating_market_cap"`
	Decimals            string  `json:"decimals"`
	ExchangeRate        *string `json:"exchange_rate"`
	HoldersCount        string  `json:"holders_count"`
	IconURL             *string `json:"icon_url"`
	Name                string  `json:"name"`
	Reputation          string  `json:"reputation"`
	Symbol              string  `json:"symbol"`
	TotalSupply         string  `json:"total_supply"`
	Type                string  `json:"type"`
	Volume24h           *string `json:"volume_24h"`
}

// PaxscanHoldersResponse represents holders response from Paxscan API
type PaxscanHoldersResponse struct {
	Items []struct {
		Address struct {
			Hash string `json:"hash"`
		} `json:"address"`
		Value string `json:"value"`
	} `json:"items"`
	NextPageParams interface{} `json:"next_page_params"`
}

// GET /tokens/{token}/paxscan-sync - Fetch and cache Paxscan data
func (s *Server) handlePaxscanSync(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/tokens/")
	token = strings.TrimSuffix(token, "/paxscan-sync")
	if token == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing token address"})
		return
	}

	// Fetch token data from Paxscan
	tokenData, err := s.fetchPaxscanToken(token)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to fetch token data", "details": err.Error()})
		return
	}

	// Fetch holders data from Paxscan (for validation, not stored separately)
	_, err = s.fetchPaxscanHolders(token)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to fetch holders data", "details": err.Error()})
		return
	}

	// Store/update in database
	holdersCount := 0
	if tokenData.HoldersCount != "" {
		if hc, err := strconv.Atoi(tokenData.HoldersCount); err == nil {
			holdersCount = hc
		}
	}

	// Find pool address for this token
	var poolAddress string
	err = s.DB.QueryRow(r.Context(), `SELECT pool_address FROM pools WHERE token_address = $1 LIMIT 1`, strings.ToLower(token)).Scan(&poolAddress)
	if err != nil {
		// Token not found in pools, still cache the data
		poolAddress = ""
	}

	// Upsert token metadata with Paxscan data
	if poolAddress != "" {
		_, err = s.DB.Exec(r.Context(), `
			INSERT INTO pool_metadata(pool_address, token_address, name, symbol, description, logo_url, created_by)
			VALUES($1, $2, $3, $4, $5, $6, 'paxscan-sync')
			ON CONFLICT(pool_address) DO UPDATE SET 
				name = COALESCE(pool_metadata.name, EXCLUDED.name),
				symbol = COALESCE(pool_metadata.symbol, EXCLUDED.symbol),
				logo_url = COALESCE(pool_metadata.logo_url, EXCLUDED.logo_url),
				updated_at = NOW()`,
			poolAddress, strings.ToLower(token), tokenData.Name, tokenData.Symbol, 
			fmt.Sprintf("Token with %d holders", holdersCount), tokenData.IconURL)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to store metadata", "details": err.Error()})
			return
		}
	}

	// Store cached Paxscan data
	_, err = s.DB.Exec(r.Context(), `
		INSERT INTO paxscan_cache(token_address, name, symbol, holders_count, total_supply, decimals, icon_url, cached_at)
		VALUES($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT(token_address) DO UPDATE SET
			name = EXCLUDED.name,
			symbol = EXCLUDED.symbol,
			holders_count = EXCLUDED.holders_count,
			total_supply = EXCLUDED.total_supply,
			decimals = EXCLUDED.decimals,
			icon_url = EXCLUDED.icon_url,
			cached_at = NOW()`,
		strings.ToLower(token), tokenData.Name, tokenData.Symbol, holdersCount, 
		tokenData.TotalSupply, tokenData.Decimals, tokenData.IconURL)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to cache data", "details": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true,
		"token": token,
		"name": tokenData.Name,
		"symbol": tokenData.Symbol,
		"holders": holdersCount,
		"poolAddress": poolAddress,
		"cached": true,
	})
}

// GET /tokens/{token}/cached - Get cached Paxscan data
func (s *Server) handleGetCachedTokenData(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/tokens/")
	token = strings.TrimSuffix(token, "/cached")
	if token == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing token address"})
		return
	}

	var result struct {
		TokenAddress string     `json:"tokenAddress"`
		Name         *string    `json:"name"`
		Symbol       *string    `json:"symbol"`
		HoldersCount *int       `json:"holdersCount"`
		TotalSupply  *string    `json:"totalSupply"`
		Decimals     *string    `json:"decimals"`
		IconURL      *string    `json:"iconUrl"`
		CachedAt     *time.Time `json:"cachedAt"`
	}

	err := s.DB.QueryRow(r.Context(), `
		SELECT token_address, name, symbol, holders_count, total_supply, decimals, icon_url, cached_at
		FROM paxscan_cache WHERE token_address = $1`,
		strings.ToLower(token)).Scan(
		&result.TokenAddress, &result.Name, &result.Symbol, &result.HoldersCount,
		&result.TotalSupply, &result.Decimals, &result.IconURL, &result.CachedAt)

	if err != nil {
		// Not cached, return empty result
		writeJSON(w, http.StatusOK, map[string]any{
			"tokenAddress": strings.ToLower(token),
			"cached": false,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tokenAddress": result.TokenAddress,
		"name": result.Name,
		"symbol": result.Symbol,
		"holdersCount": result.HoldersCount,
		"totalSupply": result.TotalSupply,
		"decimals": result.Decimals,
		"iconUrl": result.IconURL,
		"cachedAt": result.CachedAt,
		"cached": true,
	})
}

// Helper functions to fetch from Paxscan API
func (s *Server) fetchPaxscanToken(tokenAddress string) (*PaxscanTokenData, error) {
	url := fmt.Sprintf("https://paxscan.paxeer.app/api/v2/tokens/%s", url.PathEscape(tokenAddress))
	
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("paxscan API returned status %d", resp.StatusCode)
	}

	var tokenData PaxscanTokenData
	if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
		return nil, err
	}

	return &tokenData, nil
}

func (s *Server) fetchPaxscanHolders(tokenAddress string) (*PaxscanHoldersResponse, error) {
	url := fmt.Sprintf("https://paxscan.paxeer.app/api/v2/tokens/%s/holders", url.PathEscape(tokenAddress))
	
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("paxscan API returned status %d", resp.StatusCode)
	}

	var holdersData PaxscanHoldersResponse
	if err := json.NewDecoder(resp.Body).Decode(&holdersData); err != nil {
		return nil, err
	}

	return &holdersData, nil
}
