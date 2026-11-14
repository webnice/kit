package cli

import (
	"testing"

	kitTypes "github.com/webnice/kit/v4/types"
)

// TestImplConstantAndCommands Окружение констант и команды до инициализации.
func TestImplConstantAndCommands(t *testing.T) {
	const (
		errErrorsNotPackage = "Errors() вернул не справочник пакета."
		errCommandNotEmpty  = "Command()/CommandFull() до Init() должны быть пустыми."
	)
	var cli *impl

	cli = New(&kitTypes.BootstrapConfiguration{}).(*impl)
	cli.Constant(ConstantEnvironmentName{{Anchor: "anchor", Destination: "DEST"}})
	if cli.Errors() != Errors() {
		t.Fatal(errErrorsNotPackage)
	}
	if cli.Command() != "" || len(cli.CommandFull()) != 0 {
		t.Fatal(errCommandNotEmpty)
	}
}
