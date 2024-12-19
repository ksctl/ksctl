package ssh

import (
	"context"
	"fmt"
	"github.com/gookit/goutil/dump"
	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"gotest.tools/v3/assert"
	"os"
	"path/filepath"
	"testing"
)

var (
	dir                        = filepath.Join(os.TempDir(), "ksctl-k3s-test")
	log          logger.Logger = logger.NewStructuredLogger(-1, os.Stdout)
	mainStateDoc               = &storageTypes.StorageDocument{}
	dummyCtx                   = context.WithValue(context.TODO(), consts.KsctlTestFlagKey, "true")
)

func TestCreateSSHKeyPair(t *testing.T) {
	err := CreateSSHKeyPair(dummyCtx, log, mainStateDoc)
	if err != nil {
		t.Fatal(err)
	}
	dump.Println(mainStateDoc.SSHKeyPair)
}

func TestSSHExecute(t *testing.T) {

	var sshTest SSHCollection = &SSHPayload{
		ctx: dummyCtx,
		log: log,
	}
	testSimulator := NewScriptCollection()
	testSimulator.Append(types.Script{
		Name:           "test",
		CanRetry:       false,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
cat /etc/os-releases
`,
	})
	testSimulator.Append(types.Script{
		Name:           "testhaving retry",
		CanRetry:       true,
		MaxRetries:     3,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
suao apt install ...
`,
	})
	sshTest.Username("fake")
	sshTest.PrivateKey(mainStateDoc.SSHKeyPair.PrivateKey)
	sshT := sshTest.Flag(consts.UtilExecWithoutOutput).Script(testSimulator).
		IPv4("A.A.A.A").
		FastMode(true).SSHExecute()
	assert.Assert(t, sshT == nil, fmt.Sprintf("ssh should fail, got: %v, exepected ! nil", sshT))

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}

}

func TestScriptCollection(t *testing.T) {
	var scripts *Scripts = func() *Scripts {
		o := NewScriptCollection()
		switch v := o.(type) {
		case *Scripts:
			return v
		default:
			return nil
		}
	}()

	t.Run("init state test", func(t *testing.T) {
		assert.Equal(t, scripts.currIdx, -1, "must be initialized with -1")
		assert.Assert(t, scripts.mu != nil, "the mutext variable should be initialized!")
		assert.Assert(t, scripts.data == nil)
		assert.Equal(t, scripts.IsCompleted(), false, "must be empty")
	})

	t.Run("append scripts", func(t *testing.T) {
		datas := []types.Script{
			{
				ScriptExecutor: consts.LinuxBash,
				CanRetry:       false,
				Name:           "test",
				MaxRetries:     0,
				ShellScript:    "script",
			},
			{
				ScriptExecutor: consts.LinuxSh,
				CanRetry:       true,
				Name:           "x test",
				MaxRetries:     9,
				ShellScript:    "demo",
			},
		}

		for idx, data := range datas {
			scripts.Append(data)
			data.ShellScript = "#!" + string(data.ScriptExecutor) + "\n" + data.ShellScript

			assert.Equal(t, scripts.currIdx, 0, "the first element added so index should be 0")
			assert.DeepEqual(t, scripts.data[idx], data)
		}

	})

	t.Run("get script", func(t *testing.T) {
		v := scripts.NextScript()

		expected := &types.Script{
			ScriptExecutor: consts.LinuxBash,
			CanRetry:       false,
			Name:           "test",
			MaxRetries:     0,
			ShellScript:    "#!/bin/bash\nscript",
		}

		assert.DeepEqual(t, v, expected)
		assert.Equal(t, scripts.currIdx, 1, "the index must increment")
	})
}
