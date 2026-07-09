package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const imageFluxEndpoint = "https://live-api.imageflux.jp/"

var imageFluxClient = &http.Client{Timeout: 10 * time.Second}

/*
***
HLSライブ配信の暗号化鍵を取得するための要求と応答の構造体
***
*/
type getEncryptKeyRequest struct {
	Kid string `json:"kid"`
}
type getEncryptKeyResponse struct {
	EncryptKey string `json:"encrypt_key"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth())
	mux.HandleFunc("/auth_webhook_url_allow", handleWebhook(`{"allowed":true}`))
	mux.HandleFunc("/auth_webhook_url_deny", handleWebhook(`{"allowed":false,"reason":"認証に失敗しました。"}`))
	mux.HandleFunc("/encrypt_key_uri_allow", handleHLSAllow())
	mux.HandleFunc("/encrypt_key_uri_deny", handleHLSDeny())
	mux.HandleFunc("/event_webhook_url", handleEventWebhook())

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("%s番ポートでサーバを起動しました。", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

/*
***
ヘルスチェック用ハンドラ
***
*/
func handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

/*
***
Webhookの内容を読み取り、ログに出力し、指定された応答を返却する。
***
*/
func handleWebhook(response string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := readAndLogWebhookBody(r); err != nil {
			log.Printf("Webhookの内容の読み取りに失敗しました： %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(response)); err != nil {
			log.Printf("Webhookの応答の書き込みに失敗しました： %v", err)
		}
	}
}

/*
***
Webhookの内容を読み取り、ログに出力する。
***
*/
func handleEventWebhook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := readAndLogWebhookBody(r); err != nil {
			log.Printf("Webhookの内容の読み取りに失敗しました： %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

/*
***
ログ出力関数
***
*/
func readAndLogWebhookBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	log.Printf("Webhookの内容 %s %s：%s", r.Method, r.URL.Path, string(body))
	return body, nil
}

/*
***
HLSライブ配信視聴の要求を許可し、復号用の鍵を応答として返却する。
***
*/
func handleHLSAllow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logHLSRequest(r)

		setHLSCORSHeaders(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		kid := r.URL.Query().Get("kid")
		if kid == "" {
			http.Error(w, "kidは必須項目です。", http.StatusBadRequest)
			return
		}

		accessToken := os.Getenv("IMAGEFLUX_ACCESS_TOKEN")
		if accessToken == "" {
			log.Printf("環境変数 IMAGEFLUX_ACCESS_TOKEN が未設定です。")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		keyHex, err := getEncryptKeyHex(r.Context(), kid, accessToken)
		if err != nil {
			log.Printf("GetEncryptKey API 呼び出しに失敗しました: kid=%s err=%v", kid, err)
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}

		keyBin, err := hex.DecodeString(keyHex)
		if err != nil || len(keyBin) != 16 {
			log.Printf("GetEncryptKey API の鍵形式が不正です： key=%q err=%v", keyHex, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		log.Printf("HLSライブ配信視聴を許可します：%s %s kid=%s", r.Method, r.URL.Path, kid)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(keyBin); err != nil {
			log.Printf("応答の書き込みに失敗しました： %v", err)
		}
	}
}

/*
***
HLSライブ配信の暗号化鍵を取得するための関数
***
*/
func getEncryptKeyHex(ctx context.Context, kid string, accessToken string) (string, error) {
	payload := getEncryptKeyRequest{Kid: kid}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("JSONの生成に失敗しました： %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, imageFluxEndpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("API要求の作成に失敗しました： %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Sora-Target", "ImageFlux_20200707.GetEncryptKey")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := imageFluxClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API要求の実行に失敗しました： %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("API応答の読み取りに失敗しました： %w", err)
	}

	log.Printf("ImageFlux API 応答: status=%d body=%s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API応答のステータスが不正です： status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var parsed getEncryptKeyResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("API応答の解析に失敗しました： %w", err)
	}

	if parsed.EncryptKey == "" {
		return "", fmt.Errorf("API応答の鍵が空です。")
	}

	return parsed.EncryptKey, nil
}

/*
***
HLSライブ配信の暗号化鍵の取得を拒否する。
***
*/
func handleHLSDeny() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logHLSRequest(r)

		setHLSCORSHeaders(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		log.Printf("HLSライブ配信視聴を拒否します：%s %s", r.Method, r.URL.Path)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	}
}

/*
***
HLSライブ配信視聴向けの要求内容をログ出力する。
***
*/
func logHLSRequest(r *http.Request) {
	log.Printf(
		"HLSリクエスト: method=%s path=%s query=%s headers=%v",
		r.Method,
		r.URL.Path,
		r.URL.RawQuery,
		r.Header,
	)
}

/*
***
HLSライブ配信のCORSヘッダーを設定する。
***
*/
func setHLSCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
}
