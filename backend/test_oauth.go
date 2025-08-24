package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// テスト用のOAuth設定
func main() {
	// credentials.jsonを読み込み
	data, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		fmt.Printf("Error reading credentials: %v\n", err)
		return
	}

	var creds struct {
		Web struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
			TokenURI     string `json:"token_uri"`
		} `json:"web"`
	}

	if err := json.Unmarshal(data, &creds); err != nil {
		fmt.Printf("Error parsing credentials: %v\n", err)
		return
	}

	fmt.Printf("Client ID: %s\n", creds.Web.ClientID)
	fmt.Printf("Client Secret: %s...\n", creds.Web.ClientSecret[:10])
	fmt.Printf("Token URI: %s\n", creds.Web.TokenURI)

	// テスト用のリフレッシュトークン（無効なもの）
	testRefreshToken := "test_refresh_token"

	// リクエストデータの準備
	requestData := url.Values{}
	requestData.Set("client_id", creds.Web.ClientID)
	requestData.Set("client_secret", creds.Web.ClientSecret)
	requestData.Set("refresh_token", testRefreshToken)
	requestData.Set("grant_type", "refresh_token")

	// HTTPリクエストの作成
	req, err := http.NewRequest("POST", creds.Web.TokenURI, bytes.NewBufferString(requestData.Encode()))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// リクエストの実行
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// レスポンスの読み取り
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Printf("\nResponse Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Body: %s\n", string(body))

	// JSONパース
	var tokenResp map[string]interface{}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	// エラーの詳細を表示
	if errorType, exists := tokenResp["error"]; exists {
		fmt.Printf("\nError Type: %v\n", errorType)
		if errorDesc, exists := tokenResp["error_description"]; exists {
			fmt.Printf("Error Description: %v\n", errorDesc)
		}
	}
}
