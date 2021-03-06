package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"kuaishangtong/navi/cmd/navi-cli"
	"strings"
)

var createCmd = &cobra.Command{

	Use:     "create package_path",
	Aliases: []string{"c"},
	Short:   "Create a project with runnable HTTP server and thrift server",

	Example: "navi_builder create package/path/to/YourService\n" +
		"'YourService' *MUST* be a CamelCase string",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("invalid args")
		}

		tempStr := strings.Trim(args[0],"/")
		servicename := tempStr[strings.LastIndex(tempStr,"/") + 1:]
		if navicli.IsNotCamelCase(servicename) {
			return errors.New("[" + servicename + "] is not a CamelCase string")
		}

		if len(RpcType) == 0 || (RpcType != "grpc" && RpcType != "thrift") {
			return errors.New("invalid value for -r, should be grpc or thrift")
		}
		g := navicli.Creator{
			RpcType: RpcType,
			PkgPath: args[0],
		}

		g.CreateProject(servicename, force)
		return nil
	},
}

var force bool

func init() {
	createCmd.Flags().StringVarP(&RpcType, "rpctype", "r", "grpc", "[grpc|thrift]")
	createCmd.Flags().BoolVarP(&force, "force", "f", false, "create service and override existing files")
	RootCmd.AddCommand(createCmd)
}
