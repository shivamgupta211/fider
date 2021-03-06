package middlewares

import (
	"fmt"

	"github.com/getfider/fider/app/models"

	"github.com/getfider/fider/app"
	"github.com/getfider/fider/app/pkg/env"
	"github.com/getfider/fider/app/pkg/errors"
	"github.com/getfider/fider/app/pkg/web"
)

// Tenant adds either SingleTenant or MultiTenant to the pipeline
func Tenant() web.MiddlewareFunc {
	if env.IsSingleHostMode() {
		return SingleTenant()
	}
	return MultiTenant()
}

// SingleTenant inject default tenant into current context
func SingleTenant() web.MiddlewareFunc {
	return func(next web.HandlerFunc) web.HandlerFunc {
		return func(c web.Context) error {
			tenant, err := c.Services().Tenants.First()
			if err != nil {
				if errors.Cause(err) == app.ErrNotFound {
					return c.Redirect("/signup")
				}
				return c.Failure(err)
			}

			c.SetTenant(tenant)
			return next(c)
		}
	}
}

// MultiTenant extract tenant information from hostname and inject it into current context
func MultiTenant() web.MiddlewareFunc {
	return func(next web.HandlerFunc) web.HandlerFunc {
		return func(c web.Context) error {
			hostname := c.Request.URL.Hostname()

			// If no tenant is specified, redirect user to getfider.com
			// This is only valid for fider.io hosting
			if (env.IsProduction() && hostname == "fider.io") ||
				(env.IsDevelopment() && hostname == "dev.fider.io") {
				return c.Redirect("https://getfider.com")
			}

			tenant, err := c.Services().Tenants.GetByDomain(hostname)
			if err == nil {
				c.SetTenant(tenant)

				if tenant.CNAME != "" && !c.IsAjax() {
					baseURL := c.TenantBaseURL(tenant)
					if baseURL != c.BaseURL() {
						link := baseURL + c.Request.URL.RequestURI()
						c.Response.Header().Set("Link", fmt.Sprintf("<%s>; rel=\"canonical\"", link))
					}
				}
				return next(c)
			}

			if errors.Cause(err) == app.ErrNotFound {
				c.Logger().Debugf("Tenant not found for '%s'.", hostname)
				return c.NotFound()
			}

			return c.Failure(err)
		}
	}
}

// OnlyActiveTenants blocks requests for inactive tenants
func OnlyActiveTenants() web.MiddlewareFunc {
	return func(next web.HandlerFunc) web.HandlerFunc {
		return func(c web.Context) error {
			if c.Tenant().Status == models.TenantActive {
				return next(c)
			}
			return c.NotFound()
		}
	}
}

// CheckTenantPrivacy blocks requests of unauthenticated users for private tenants
func CheckTenantPrivacy() web.MiddlewareFunc {
	return func(next web.HandlerFunc) web.HandlerFunc {
		return func(c web.Context) error {
			if c.Tenant().IsPrivate && !c.IsAuthenticated() {
				return c.Redirect("/signin")
			}
			return next(c)
		}
	}
}
