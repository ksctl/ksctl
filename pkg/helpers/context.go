package helpers

import (
	"context"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"regexp"
)

var (
	contextVars = [...]string{
		consts.KsctlTestFlag:        `true`,
		consts.ContextModuleNameKey: `^[A-Za-z-]+$`,
	}
)

func IsContextPresent(ctx context.Context, key consts.KsctlContextKeyType) (val string, isPresent bool) {
	_val := ctx.Value(key)
	expectedPattern := contextVars[key]

	if gotV, ok := _val.(string); ok {
		if _ok, err := regexp.MatchString(expectedPattern, gotV); err != nil {
			return "", false
		} else {
			if _ok {
				return gotV, true
			}
		}
	}
	return "", false
}
