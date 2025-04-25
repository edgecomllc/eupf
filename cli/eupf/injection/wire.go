//go:build wireinject
// +build wireinject

package injection

import (
	"github.com/edgecomllc/eupf/cli/config"
	"github.com/edgecomllc/eupf/cli/eupf/delivery/cli"
	"github.com/edgecomllc/eupf/cli/eupf/repository/file"
	"github.com/edgecomllc/eupf/cli/eupf/repository/http"
	"github.com/edgecomllc/eupf/cli/eupf/usecase"
	"github.com/google/wire"
)

func InitEupfCLI(baseURL string, cfg *config.Config) *cli.CLI {
	wire.Build(
		http.NewEupfHttpRepository,
		file.NewFileRepository,
		usecase.NewEupf,
		cli.NewCLI,
	)

	return &cli.CLI{}
}
