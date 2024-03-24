package helpers

import (
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/resources"
	"gotest.tools/v3/assert"
	"testing"
)

func HelperTestTemplate(t *testing.T, testData []resources.Script, f func() resources.ScriptCollection) {

	var expectedScripts *helpers.Scripts = func() *helpers.Scripts {
		o := helpers.NewScriptCollection()
		switch v := o.(type) {
		case *helpers.Scripts:
			return v
		default:
			return nil
		}
	}()

	for _, script := range testData {
		expectedScripts.Append(script)
	}

	var actualScripts *helpers.Scripts = func() *helpers.Scripts {
		o := f()
		switch v := o.(type) {
		case *helpers.Scripts:
			return v
		default:
			panic("unable to conver the interface type to concerete type")
		}
	}()
	assert.DeepEqual(t, actualScripts.TestAllScripts(), expectedScripts.TestAllScripts())
	assert.Equal(t, actualScripts.TestLen(), expectedScripts.TestLen())
}
