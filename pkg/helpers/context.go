package helpers

import (
	"context"
	"regexp"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

func IsContextPresent(ctx context.Context, key consts.KsctlContextKeyType) (val string, isPresent bool) {
	var contextVars = [...]string{
		consts.KsctlTestFlagKey:        `true`,
		consts.KsctlModuleNameKey:      `^[\w-]+$`,
		consts.KsctlContextUserID:      `^[\w-]+$`,
		consts.KsctlCustomDirLoc:       `^[\w-:~\\/\s]+$`,
		consts.KsctlComponentOverrides: `^([\w]+=[\w-:\.~\\/\s]+)+(,[\w]+=[\w-:\.~\\/\s]+)*$`,
	}
	_val := ctx.Value(key)
	if _val == nil {
		return "", false
	}

	expectedPattern := contextVars[key]

	gotV, ok := _val.(string)
	if ok {
		_ok, err := regexp.MatchString(expectedPattern, gotV)
		if err != nil {
			return "", false
		}
		if _ok {
			return gotV, true
		}
	}
	return "", false
}
