package middlewares

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type EasyAuthHeader struct {
	AuthType string  `json:"auth_typ"`
	Claims   []Claim `json:"claims"`
	NameType string  `json:"name_typ"`
	RoleType string  `json:"role_typ"`
}

type Claim struct {
	Type  string `json:"typ"`
	Value string `json:"val"`
}

func readEasyAuthHeader(r *http.Request, header string) ([]Claim, error) {
	var claims []Claim

	h := r.Header.Get(header)
	if h == "" {
		return claims, fmt.Errorf("easy auth header %q not found", header)
	}

	hd, err := base64.StdEncoding.DecodeString(h)
	if err != nil {
		return claims, fmt.Errorf("easy auth header %q decoding: %w", header, err)
	}

	var easyAuthHeader EasyAuthHeader

	err = json.Unmarshal(hd, &easyAuthHeader)
	if err != nil {
		return claims, fmt.Errorf("easy auth header %q unmarshalling: %w", header, err)
	}

	claims = easyAuthHeader.Claims

	return claims, nil
}
