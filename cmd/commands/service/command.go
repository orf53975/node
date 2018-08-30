/*
 * Copyright (C) 2017 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package service

import (
	"github.com/mysterium/node/cmd"
	"github.com/mysterium/node/core/node"
	"github.com/mysterium/node/core/service"
	"github.com/mysterium/node/utils"
	"github.com/urfave/cli"
)

var (
	identityFlag = cli.StringFlag{
		Name:  "identity",
		Usage: "Keystore's identity used to provide service. If not given identity will be created automatically",
		Value: "",
	}
	identityPassphraseFlag = cli.StringFlag{
		Name:  "identity.passphrase",
		Usage: "Used to unlock keystore's identity",
		Value: "",
	}

	openvpnProtocolFlag = cli.StringFlag{
		Name:  "openvpn.protocol",
		Usage: "Openvpn protocol to use. Options: { udp, tcp }",
		Value: "udp",
	}
	openvpnPortFlag = cli.IntFlag{
		Name:  "openvpn.port",
		Usage: "Openvpn port to use. Default 1194",
		Value: 1194,
	}

	locationCountryFlag = cli.StringFlag{
		Name:  "location.country",
		Usage: "Service location country. If not given country is autodetected",
		Value: "",
	}
)

// NewCommand function creates service command
func NewCommand() *cli.Command {
	var nodeInstance *node.Node
	var serviceManager *service.Manager

	stopCommand := func() error {
		errorServiceManager := serviceManager.Kill()
		errorNode := nodeInstance.Kill()

		if errorServiceManager != nil {
			return errorServiceManager
		}
		return errorNode
	}

	return &cli.Command{
		Name:      "service",
		Usage:     "Starts and publishes service on Mysterium Network",
		ArgsUsage: " ",
		Flags: []cli.Flag{
			identityFlag, identityPassphraseFlag,
			openvpnProtocolFlag, openvpnPortFlag,
			locationCountryFlag,
		},
		Action: func(ctx *cli.Context) error {
			errorChannel := make(chan error, 1)

			nodeOptions := cmd.ParseNodeFlags(ctx)
			nodeInstance = node.NewNode(nodeOptions)
			go func() {
				if err := nodeInstance.Start(); err != nil {
					errorChannel <- err
					return
				}
				errorChannel <- nodeInstance.Wait()
			}()

			serviceOptions := service.Options{
				ctx.String("identity"),
				ctx.String("identity.passphrase"),

				ctx.String("openvpn.proto"),
				ctx.Int("openvpn.port"),

				ctx.String("location.country"),
			}
			serviceManager = service.NewManager(nodeOptions, serviceOptions)
			go func() {
				if err := serviceManager.Start(); err != nil {
					errorChannel <- err
					return
				}
				errorChannel <- serviceManager.Wait()
			}()

			cmd.RegisterSignalCallback(utils.SoftKiller(stopCommand))

			return <-errorChannel
		},
		After: func(ctx *cli.Context) error {
			return stopCommand()
		},
	}
}