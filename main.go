package main

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"

	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Login struct {
	Login string `json:"login" binding:"required"`
}

type Jwks struct {
	Keys []JwksKeys `json:"keys"`
}
type JwksKeys struct {
	E   string `json:"e"`
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
}

var (
	publicKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		latency := os.Getenv("LATENCY")
		if latency != "" {
			i, err := strconv.ParseInt(latency, 10, 64)
			if err != nil {
				fmt.Fprintf(w, "Env LATENCY needs to be a number")
				return
			}
			time.Sleep(time.Duration(i) * time.Second)
		}
		text := os.Getenv("TEXT")
		if text == "" {
			fmt.Fprintf(w, "set env TEXT to display something")
			return
		}
		next := os.Getenv("NEXT")
		if next == "" {
			fmt.Fprintf(w, "%s", text)
		} else {
			// initialize client
			client := &http.Client{}
			req, _ := http.NewRequest("GET", "http://"+next+"/", nil)

			// get headirs
			for k, _ := range r.Header {
				for _, otHeader := range otHeaders {
					if strings.ToLower(otHeader) == strings.ToLower(k) {
						req.Header.Set(k, r.Header.Get(k))
					}
				}
			}

			// do request
			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintf(w, "Couldn't connect to http://%s/", next)
				fmt.Printf("Error: %s", err)
				return
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			fmt.Fprintf(w, "%s %s\n", text, body)
		}
	})

	// load keys for JWT
	initKeys()

	// handle login
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		// read body
		decoder := json.NewDecoder(r.Body)
		var l Login
		err := decoder.Decode(&l)
		if err != nil {
			fmt.Fprintf(w, "No login supplied")
			return
		}
		// generate jwt token
		token := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), jwt.MapClaims{
			"login":  l.Login,
			"groups": "users",
			"iss":    "http-echo@http-echo.kubernetes.newtech.academy",
			"sub":    "http-echo@http-echo.kubernetes.newtech.academy",
			"exp":    time.Now().Add(time.Hour * 72).Unix(),
			"iat":    time.Now().Unix(),
		})

		token.Header["kid"] = "mykey"

		tokenString, err := token.SignedString(signKey)

		if err != nil {
			fmt.Fprintf(w, "Could not sign jwt token")
			return
		}

		fmt.Fprintf(w, "JWT token: %s \n", tokenString)
	})

	// jwks.json
	http.HandleFunc("/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		key, err := jwk.New(publicKey)
		if err != nil {
			log.Printf("failed to create JWK: %s", err)
			return
		}

		key.Set("kid", "mykey")

		jsonbuf, err := json.MarshalIndent(key, "", "  ")
		if err != nil {
			log.Printf("failed to generate JSON: %s", err)
			return
		}

		var k JwksKeys

		if err := json.Unmarshal(jsonbuf, &k); err != nil {
			log.Printf("failed to unmarshal JSON: %s", err)
			return
		}

		j := &Jwks{Keys: []JwksKeys{k}}

		jsonbuf2, err := json.Marshal(j)
		if err != nil {
			log.Printf("failed to generate JSON: %s", err)
			return
		}

		fmt.Fprintf(w, "%s", jsonbuf2)
	})

	// start server
	fmt.Printf("Listening on port 8080\n")
	http.ListenAndServe(":8080", nil)
}

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func initKeys() {
	createKeys()

	signBytes, err := ioutil.ReadFile("private.pem")
	fatal(err)

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	fatal(err)

	publicBytes, err := ioutil.ReadFile("public.pem")
	fatal(err)

	publicKey, err = jwt.ParseRSAPublicKeyFromPEM(publicBytes)
	fatal(err)
}
