package goule

func HandleContext(ctx *Context) {
	if !TryAdmin(ctx) {
		if !TryService(ctx) {
			// TODO: send a nice 404 page here.
			ctx.Response.Header().Set("Content-Type", "text/plain")
			ctx.Response.Write([]byte("No forward rule found."))
		}
	}
}

func TryService(ctx *Context) bool {
	// TODO: here, check services' forward rules
	return false
}

func TryAdmin(ctx *Context) bool {
	for _, source := range ctx.Overseer.GetConfiguration().Admin.Rules {
		if source.MatchesURL(&ctx.URL) {
			adminContext := NewAdminContext(ctx, source)
			if !TrySite(adminContext) {
				TryAPI(adminContext)
			}
			return true
		}
	}
	return false
}
