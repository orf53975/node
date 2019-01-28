/*
 * Copyright (C) 2018 The "MysteriumNetwork/node" Authors.
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
	"encoding/json"
	"sync"
	"time"

	log "github.com/cihub/seelog"
	"github.com/mysteriumnetwork/node/identity"
	"github.com/mysteriumnetwork/node/market"
	"github.com/mysteriumnetwork/node/money"
	"github.com/mysteriumnetwork/node/nat"
	wg "github.com/mysteriumnetwork/node/services/wireguard"
	"github.com/mysteriumnetwork/node/services/wireguard/endpoint"
	"github.com/mysteriumnetwork/node/services/wireguard/resources"
	"github.com/mysteriumnetwork/node/session"
)

const logPrefix = "[service-wireguard] "

// NewManager creates new instance of Wireguard service
func NewManager(publicIP, outIP, country string) *Manager {
	resourceAllocator := resources.NewAllocator()
	return &Manager{
		natService: nat.NewService(),

		publicIP:        publicIP,
		outboundIP:      outIP,
		currentLocation: country,

		connectionEndpointFactory: func() (wg.ConnectionEndpoint, error) {
			return endpoint.NewConnectionEndpoint(publicIP, &resourceAllocator)
		},
	}
}

// Manager represents an instance of Wireguard service
type Manager struct {
	wg         sync.WaitGroup
	natService nat.NATService

	connectionEndpointFactory func() (wg.ConnectionEndpoint, error)

	publicIP        string
	outboundIP      string
	currentLocation string
}

// ProvideConfig provides the config for consumer
func (manager *Manager) ProvideConfig(publicKey json.RawMessage) (session.ServiceConfiguration, session.DestroyCallback, error) {
	key := &wg.ConsumerConfig{}
	err := json.Unmarshal(publicKey, key)
	if err != nil {
		return nil, nil, err
	}

	connectionEndpoint, err := manager.connectionEndpointFactory()
	if err != nil {
		return nil, nil, err
	}

	if err := connectionEndpoint.Start(nil); err != nil {
		return nil, nil, err
	}

	if err := connectionEndpoint.AddPeer(key.PublicKey, nil); err != nil {
		return nil, nil, err
	}

	config, err := connectionEndpoint.Config()
	if err != nil {
		return nil, nil, err
	}

	manager.natService.Add(nat.RuleForwarding{
		SourceAddress: config.Consumer.IPAddress.String(),
		TargetIP:      manager.outboundIP,
	})
	if err := manager.natService.Start(); err != nil {
		return nil, nil, err
	}

	destroyCallback := func(dryRun bool) error {
		if !dryRun {
			return connectionEndpoint.Stop()
		}

		if isSessionAlive(connectionEndpoint) {
			return session.ErrSessionStileAlive
		}
		return nil
	}

	return config, destroyCallback, nil
}

func isSessionAlive(ce wg.ConnectionEndpoint) bool {
	_, lastHandshake, err := ce.PeerStats()
	if err != nil {
		log.Error(logPrefix, "Failed to check if the session active: ", err)
		return false
	}

	if lastHandshake > 0 && time.Since(time.Unix(int64(lastHandshake), 0)) > 2*time.Hour {
		return false
	}
	return true
}

// Serve starts service - does block
func (manager *Manager) Serve(providerID identity.Identity) error {
	manager.wg.Add(1)
	log.Info(logPrefix, "Wireguard service started successfully")

	manager.wg.Wait()
	return nil
}

// GetProposal returns the proposal for wireguard service
func GetProposal(country string) market.ServiceProposal {
	return market.ServiceProposal{
		ServiceType: wg.ServiceType,
		ServiceDefinition: wg.ServiceDefinition{
			Location: market.Location{Country: country},
		},
		PaymentMethodType: wg.PaymentMethod,
		PaymentMethod: wg.Payment{
			Price: money.NewMoney(0, money.CURRENCY_MYST),
		},
	}
}

// Stop stops service.
func (manager *Manager) Stop() error {
	manager.wg.Done()
	manager.natService.Stop()

	log.Info(logPrefix, "Wireguard service stopped")
	return nil
}
