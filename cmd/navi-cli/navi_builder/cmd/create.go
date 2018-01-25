/*
 * Copyright © 2017 Xiao Zhang <zzxx513@gmail.com>.
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file.
 */
package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"git.oschina.net/kuaishangtong/navi/cmd/navi-cli"
)

var createCmd = &cobra.Command{
	Use:     "create package_path ServiceName",
	Aliases: []string{"c"},
	Short:   "Create a project with runnable HTTP server and gRPC/thrift server",
	Example: "turbo create package/path/to/yourservice YourService -r grpc\n" +
		"'ServiceName' *MUST* be a CamelCase string",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("invalid args")
		}
		//if navicli.IsNotCamelCase(args[1]) {
		//	return errors.New("[" + args[1] + "] is not a CamelCase string")
		//}
		if navicli.IsNotCamelCase(args[0]) {
			return errors.New("[" + args[1] + "] is not a CamelCase string")
		}
		//if len(RpcType) == 0 || (RpcType != "grpc" && RpcType != "thrift") {
		//	return errors.New("invalid value for -r, should be grpc or thrift")
		//}
		RpcType = "thrift"
		g := navicli.Creator{
			RpcType: RpcType,
			PkgPath: args[0],
		}

		g.CreateProject(args[0], force)
		return nil
	},
}

var force bool

func init() {
	//createCmd.Flags().StringVarP(&RpcType, "rpctype", "r", "grpc", "[grpc|thrift]")
	createCmd.Flags().BoolVarP(&force, "force", "f", false, "create service and override existing files")
	RootCmd.AddCommand(createCmd)
}
