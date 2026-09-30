package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/example/placeholder-webapp/context"
	"github.com/example/placeholder-webapp/models"
)

var validSFID = regexp.MustCompile(`SF\d+P`)

func EasyAuth(header string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := readEasyAuthHeader(r, header)
			if err != nil {
				log.Printf("failed to set user: %v", err)
				next.ServeHTTP(w, r)
				return
			}

			user, err := getUser(claims)
			if err != nil {
				log.Printf("failed to set user: %v", err)
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			ctx = context.WithUser(ctx, user)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func getUser(claims []Claim) (*models.User, error) {
	var user models.User

	for _, claim := range claims {
		if claim.Type == "http://schemas.microsoft.com/identity/claims/objectidentifier" {
			user.ID = claim.Value
		}

		if claim.Type == "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress" {
			user.Email = claim.Value
		}

		if claim.Type == "name" {
			user.FullName = claim.Value
		}

		if claim.Type == "employeeid" {
			user.SFID = claim.Value
		}
	}

	if !validSFID.MatchString(user.SFID) {
		return nil, fmt.Errorf("invalid SF ID for %q (%s)", user.ID, user.FullName)
	}

	return &user, nil
}
