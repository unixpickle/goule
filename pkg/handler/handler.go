package handler

import (
	"github.com/unixpickle/goule/pkg/overseer"
	"github.com/unixpickle/goule/pkg/proxy"
	"net/http"
)

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
	cfg := ctx.Overseer.GetConfiguration()
	for _, service := range cfg.Services {
		for _, rule := range service.ForwardRules {
			if rule.From.MatchesURL(&ctx.URL) {
				dest := rule.Apply(&ctx.URL)
				context := proxy.Context{ctx.Request, ctx.Response, &ctx.URL,
					dest, &cfg.Proxy}
				proxy.ProxyRequest(&context, &http.Client{})
				return true
			}
		}
	}
	return false
}

func TryAdmin(ctx *overseer.Context) bool {
	for _, source := range ctx.Overseer.GetConfiguration().Admin.Rules {
		if source.MatchesURL(&ctx.URL) {
			adminContext := NewContext(ctx, source)
			if !TryStatic(adminContext) {
				TryAPI(adminContext)
			}
			return true
		}
	}
	return false
}
