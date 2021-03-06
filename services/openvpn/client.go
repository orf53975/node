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

package openvpn

import (
	"errors"

	"github.com/mysteriumnetwork/go-openvpn/openvpn"
	"github.com/mysteriumnetwork/node/core/connection"
)

// ErrProcessNotStarted represents the error we return when the process is not started yet
var ErrProcessNotStarted = errors.New("process not started yet")

// processFactory creates a new openvpn process
type processFactory func(options connection.ConnectOptions) (openvpn.Process, error)

// Client takes in the openvpn process and works with it
type Client struct {
	process        openvpn.Process
	processFactory processFactory
}

// Start starts the connection
func (c *Client) Start(options connection.ConnectOptions) error {
	proc, err := c.processFactory(options)
	if err != nil {
		return err
	}
	c.process = proc
	return c.process.Start()
}

// Wait waits for the connection to exit
func (c *Client) Wait() error {
	if c.process == nil {
		return ErrProcessNotStarted
	}
	return c.process.Wait()
}

// Stop stops the connection
func (c *Client) Stop() {
	if c.process != nil {
		c.process.Stop()
	}
}

// GetConfig returns the consumer-side configuration. In openvpn case - it doesn't return anything
func (c *Client) GetConfig() (connection.ConsumerConfig, error) {
	return nil, nil
}

//VPNConfig structure represents VPN configuration options for given session
type VPNConfig struct {
	RemoteIP        string `json:"remote"`
	RemotePort      int    `json:"port"`
	RemoteProtocol  string `json:"protocol"`
	TLSPresharedKey string `json:"TLSPresharedKey"`
	CACertificate   string `json:"CACertificate"`
}
