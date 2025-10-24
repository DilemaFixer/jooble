package loader

import "context"

type HtmlLoader interface {
	Load(url string, ctx context.Context) (string, error)
}
