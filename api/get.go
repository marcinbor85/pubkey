package api

import (
	"io"
	"net/http"
	"strings"
	"errors"

	"crypto/rsa"
	"encoding/json"

	"github.com/marcinbor85/pubkey/crypto"
)

type GetPublicKeyResponse struct {
	Username string		`json:"username"`
	PublicKey string	`json:"public_key"`
}

func (client *Client) GetPublicKeyByUsername(username string) (*rsa.PublicKey, *RequestError) {
	url := strings.Join([]string{client.Address, "user", username}, "/")
	resp, err := http.Get(url)
	if err != nil {
		return nil, &RequestError{500, err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, &RequestError{resp.StatusCode, err}
		}

		return nil, &RequestError{resp.StatusCode, errors.New(string(bodyBytes))}
	}

	var data GetPublicKeyResponse 

	err = json.NewDecoder(resp.Body).Decode(&data);
	if  err != nil {
		return nil, &RequestError{resp.StatusCode, err}
	}

	if username != data.Username {
		return nil, &RequestError{resp.StatusCode, errors.New("username mismatch")}
	}

	key, err := crypto.DecodePublicKey(data.PublicKey)
	if  err != nil {
		return nil, &RequestError{resp.StatusCode, err}
	}

	return key, nil
}
