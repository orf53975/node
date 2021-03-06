/*
 * Copyright (C) 2019 The "MysteriumNetwork/node" Authors.
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
	"github.com/mysteriumnetwork/node/core/service"
	service_noop "github.com/mysteriumnetwork/node/services/noop"
	service_openvpn "github.com/mysteriumnetwork/node/services/openvpn"
	service_wireguard "github.com/mysteriumnetwork/node/services/wireguard"
	"github.com/urfave/cli"
)

var (
	serviceTypesAvailable = []string{"openvpn", "wireguard", "noop"}
	serviceTypesEnabled   = []string{"openvpn", "noop"}

	serviceTypesFlagsParser = map[string]func(ctx *cli.Context) service.Options{
		service_noop.ServiceType:      parseNoopFlags,
		service_openvpn.ServiceType:   parseOpenvpnFlags,
		service_wireguard.ServiceType: parseWireguardFlags,
	}
)

// parseWireguardFlags function fills in wireguard service options from CLI context
func parseWireguardFlags(ctx *cli.Context) service.Options {
	return service.Options{
		Identity:   ctx.String(identityFlag.Name),
		Passphrase: ctx.String(identityPassphraseFlag.Name),
		Type:       service_wireguard.ServiceType,
	}
}
