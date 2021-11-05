package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
)

const apiKey = "ORqVaoVf1TJrVnKexpWjHfjk"
const apiSecret = "mvK7p-zYF5He2eistXxXUvASoJWRGvp6eOO5TF2gn4BHI2iB"

func generateSignature(secret string, verb string, path string, expires time.Time, data interface{}) (string, error) {
	//# HEX(HMAC_SHA256(apiSecret, 'GET/api/v1/instrument1518064236'))
	//# Result is:
	//# 'c7682d435d0cfe87c16098df34ef2eb5a549d4c5a3c2b1f0f77b8af73423bf00'
	//signature = HEX(HMAC_SHA256(apiSecret, verb + path + str(expires) + data))

	var str string

	var dataStr string
	if data != nil {
		jsonStr, err := json.Marshal(data)
		if err != nil {
			return "", err
		}
		dataStr = string(jsonStr)
	}

	str = verb + path + strconv.Itoa(int(expires.Unix())) + dataStr
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(str))
	signature := hex.EncodeToString(h.Sum(nil))

	return signature, nil
}

func main() {

	t := time.Now().Add(time.Hour)
	sig, err := generateSignature(apiSecret, "GET", "/realtime", t, nil)
	if err != nil {
		panic(err)
	}

	header := make(http.Header)
	header.Set("api-expires", strconv.Itoa(int(t.Unix())))
	header.Set("api-key", apiKey)
	header.Set("api-signature", sig)

	conn, _, err := websocket.DefaultDialer.Dial("wss://testnet.bitmex.com/realtime", header)
	if err != nil {
		panic(err)
	}

	go func() {
		var message gin.H
		for {
			err = conn.ReadJSON(&message)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println(message)
		}
	}()

	r := gin.Default()
	r.GET("/subscribe", func(c *gin.Context) {
		message := gin.H{
			"op":   "subscribe",  // it was "action" in the task
			"args": "instrument", // it was "symbols" in the task
		}
		err := conn.WriteJSON(message)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, message)
	})
	r.GET("/unsubscribe", func(c *gin.Context) {
		message := gin.H{
			"op":   "unsubscribe",
			"args": "instrument",
		}
		err := conn.WriteJSON(message)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, message)

	})
	r.Run()

}
