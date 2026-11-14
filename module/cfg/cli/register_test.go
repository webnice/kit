package cli

import (
	"testing"

	kitTypes "github.com/webnice/kit/v4/types"
)

// TestImplRegisterCommand Регистрация команд и флагов командной строки.
func TestImplRegisterCommand(t *testing.T) {
	const (
		tplCommandNotAdded = "RegisterCommand() не добавил команду: %#v."
		tplFlagNotAdded    = "RegisterFlag() не добавил флаг: %#v."
	)
	var cli *impl

	cli = New(&kitTypes.BootstrapConfiguration{}).(*impl)
	cli.RegisterCommand(&Command{Command: "cmd", Description: "description", GroupKey: "group"})
	cli.RegisterFlag(&Flag{Flag: "flag", Description: "description", Value: new(string)})
	if len(cli.command) != 1 || cli.command[0].Command != "cmd" {
		t.Fatalf(tplCommandNotAdded, cli.command)
	}
	if len(cli.flag) != 1 || cli.flag[0].Flag != "flag" {
		t.Fatalf(tplFlagNotAdded, cli.flag)
	}
}
