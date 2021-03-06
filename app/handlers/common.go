package handlers

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/getfider/fider/app"
	"github.com/getfider/fider/app/models"
	"github.com/getfider/fider/app/pkg/env"
	"github.com/getfider/fider/app/pkg/errors"
	"github.com/getfider/fider/app/pkg/web"
)

//Health always returns OK
func Health() web.HandlerFunc {
	return func(c web.Context) error {
		return c.Ok(web.Map{})
	}
}

//LegalPage returns a legal page with content from a file
func LegalPage(title, file string) web.HandlerFunc {
	return func(c web.Context) error {
		bytes, err := ioutil.ReadFile(env.Etc(file))
		if err != nil {
			return c.NotFound()
		}

		return c.Render(http.StatusOK, "legal.html", web.Props{
			Title: title,
			Data: web.Map{
				"Content": string(bytes),
			},
		})
	}
}

//Page returns a page without properties
func Page(title, description string) web.HandlerFunc {
	return func(c web.Context) error {
		return c.Page(web.Props{
			Title:       title,
			Description: description,
		})
	}
}

func validateKey(kind models.EmailVerificationKind, c web.Context) (*models.EmailVerification, error) {
	key := c.QueryParam("k")

	//If key has been used, return NotFound
	result, err := c.Services().Tenants.FindVerificationByKey(kind, key)
	if err != nil {
		if errors.Cause(err) == app.ErrNotFound {
			return nil, c.NotFound()
		}
		return nil, c.Failure(err)
	}

	//If key has been used, return Gone
	if result.VerifiedOn != nil {
		return nil, c.Gone()
	}

	//If key expired, return Gone
	if time.Now().After(result.ExpiresOn) {
		err = c.Services().Tenants.SetKeyAsVerified(key)
		if err != nil {
			return nil, c.Failure(err)
		}
		return nil, c.Gone()
	}

	return result, nil
}
