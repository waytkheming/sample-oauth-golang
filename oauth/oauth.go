package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mercadolibre/golang-restclient/rest"
	"github.com/waytkheming/sample-oauth-golang/oauth/errors"
)

const (
	headerXPublic   = "X-Public"
	headerXClientID = "X-Client-Id"
	headerXCallerID = "X-Caller-Id"

	paramAccessToken = "access_token"
)

type oauthClient struct {
}

type oauthInterface interface {
}

var (
	oauthrestClient = rest.RequestBuilder{
		BaseURL: "https://localhost:8080",
		Timeout: 200 * time.Millisecond,
	}
)

type accessToken struct {
	ID       string `json:"id"`
	UserID   int64  `json:"user_id"`
	ClientID int64  `json:"client_id"`
}

func IsPublic(req *http.Request) bool {
	if req == nil {
		return true
	}
	return req.Header.Get(headerXPublic) == "true"
}

func GetCallerID(req *http.Request) int64 {
	if req == nil {
		return 0
	}
	callerID, err := strconv.ParseInt(req.Header.Get(headerXCallerID), 10, 64)
	if err != nil {
		return 0
	}
	return callerID
}

func GetClientID(req *http.Request) int64 {
	if req == nil {
		return 0
	}
	clientID, err := strconv.ParseInt(req.Header.Get(headerXClientID), 10, 64)
	if err != nil {
		return 0
	}
	return clientID
}

func AuthenticateRequest(req *http.Request) *errors.RestErr {
	if req == nil {
		return nil
	}
	cleanRequest(req)
	accessTokenID := strings.TrimSpace(req.URL.Query().Get(paramAccessToken))
	if accessTokenID == "" {
		return nil
	}
	at, err := getAccessToken(accessTokenID)
	if err != nil {
		return nil
	}

	req.Header.Add(headerXClientID, fmt.Sprintf("%v", at.ClientID))
	req.Header.Add(headerXCallerID, fmt.Sprintf("%v", at.UserID))
	return nil
}

func getAccessToken(accessTokenId string) (*accessToken, *errors.RestErr) {
	res := oauthrestClient.Get(fmt.Sprintf("/oauth/access_token/%s", accessTokenId))
	if res == nil || res.Response == nil {
		return nil, errors.NewInternalServerError("invalid rescli responce when trying to get access token")
	}

	if res.StatusCode > 299 {
		var restErr errors.RestErr
		err := json.Unmarshal(res.Bytes(), &restErr)
		if err != nil {
			return nil, errors.NewInternalServerError("invalid error interface when tring to get access token")
		}
		return nil, &restErr
	}

	var at accessToken
	if err := json.Unmarshal(res.Bytes(), &at); err != nil {
		return nil, errors.NewInternalServerError("error when trying to unmarshal bytes")
	}
	return &at, nil
}

func cleanRequest(req *http.Request) {
	if req == nil {
		return
	}
	req.Header.Del(headerXClientID)
	req.Header.Del(headerXCallerID)
}
