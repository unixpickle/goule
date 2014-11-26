package handler

import "github.com/unixpickle/goule/pkg/overseer"

func Handle(ctx *overseer.Context) {
	if !TryAdmin(ctx) {
		if !TryService(ctx) {
			// TODO: send a nice 404 page here.
			ctx.Response.Header().Set("Content-Type", "text/plain")
			ctx.Response.Write([]byte("No forward rule found."))
		}
	}
}

func TryService(ctx *overseer.Context) bool {
	// TODO: here, check services' forward rules
	return false
}

func TryAdmin(ctx *overseer.Context) bool {
	for _, source := range ctx.Overseer.GetConfiguration().Admin.Rules {
		if source.MatchesURL(&ctx.URL) {
			adminContext := NewContext(ctx, source)
			if !TrySite(adminContext) {
				TryAPI(adminContext)
			}
			return true
		}
	}
	return false
}
