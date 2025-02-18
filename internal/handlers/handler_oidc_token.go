package handlers

import (
	"net/http"

	"github.com/ory/fosite"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func oidcToken(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	var (
		requester fosite.AccessRequester
		responder fosite.AccessResponder
		err       error
	)

	oidcSession := oidc.NewSession()

	if requester, err = ctx.Providers.OpenIDConnect.Fosite.NewAccessRequest(ctx, req, oidcSession); err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)

		ctx.Logger.Errorf("Access Request failed with error: %+v", rfc)

		ctx.Providers.OpenIDConnect.Fosite.WriteAccessError(rw, requester, err)

		return
	}

	client := requester.GetClient()

	ctx.Logger.Debugf("Access Request with id '%s' on client with id '%s' is being processed", requester.GetID(), client.GetID())

	// If this is a client_credentials grant, grant all scopes the client is allowed to perform.
	if requester.GetGrantTypes().ExactOne("client_credentials") {
		for _, scope := range requester.GetRequestedScopes() {
			if fosite.HierarchicScopeStrategy(client.GetScopes(), scope) {
				requester.GrantScope(scope)
			}
		}
	}

	if responder, err = ctx.Providers.OpenIDConnect.Fosite.NewAccessResponse(ctx, requester); err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)

		ctx.Logger.Errorf("Access Response for Request with id '%s' failed to be created with error: %+v", requester.GetID(), rfc)

		ctx.Providers.OpenIDConnect.Fosite.WriteAccessError(rw, requester, err)

		return
	}

	ctx.Logger.Debugf("Access Request with id '%s' on client with id '%s' has successfully been processed", requester.GetID(), client.GetID())

	ctx.Logger.Tracef("Access Request with id '%s' on client with id '%s' produced the following claims: %+v", requester.GetID(), client.GetID(), responder.ToMap())

	ctx.Providers.OpenIDConnect.Fosite.WriteAccessResponse(rw, requester, responder)
}
