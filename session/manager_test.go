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

package session

import (
	"testing"

	"github.com/mysteriumnetwork/node/identity"
	"github.com/mysteriumnetwork/node/market"
	"github.com/stretchr/testify/assert"
)

var (
	currentProposalID = 68
	currentProposal   = market.ServiceProposal{
		ID: currentProposalID,
	}
	expectedID      = ID("mocked-id")
	expectedSession = Session{
		ID:         expectedID,
		Config:     expectedSessionConfig,
		ConsumerID: identity.FromAddress("deadbeef"),
	}
	lastSession Session
)

const expectedSessionConfig = "config_string"

func generateSessionID() (ID, error) {
	return expectedID, nil
}

type fakePromiseProcessor struct {
	started  bool
	proposal market.ServiceProposal
}

func (processor *fakePromiseProcessor) Start(proposal market.ServiceProposal) error {
	processor.started = true
	processor.proposal = proposal
	return nil
}

func (processor *fakePromiseProcessor) Stop() error {
	processor.started = false
	return nil
}

type mockPaymentOrchestrator struct {
	errChan chan error
}

func (m mockPaymentOrchestrator) Start() <-chan error {
	return m.errChan
}

func (m mockPaymentOrchestrator) Stop() {

}

func mockPaymentOrchestratorFactory() PaymentOrchestrator {
	return &mockPaymentOrchestrator{}
}

func TestManager_Create_StoresSession(t *testing.T) {
	sessionStore := NewStorageMemory()
	manager := NewManager(currentProposal, generateSessionID, sessionStore, &fakePromiseProcessor{}, mockPaymentOrchestratorFactory)

	sessionInstance, err := manager.Create(identity.FromAddress("deadbeef"), currentProposalID, expectedSessionConfig, nil)
	assert.NoError(t, err)
	assert.Equal(t, expectedSession, sessionInstance)
}

func TestManager_Create_RejectsUnknownProposal(t *testing.T) {
	sessionStore := NewStorageMemory()
	manager := NewManager(currentProposal, generateSessionID, sessionStore, &fakePromiseProcessor{}, mockPaymentOrchestratorFactory)

	sessionInstance, err := manager.Create(identity.FromAddress("deadbeef"), 69, expectedSessionConfig, nil)
	assert.Exactly(t, err, ErrorInvalidProposal)
	assert.Exactly(t, Session{}, sessionInstance)
}

func TestManager_Create_StartsPromiseProcessor(t *testing.T) {
	promiseProcessor := &fakePromiseProcessor{}
	sessionStore := NewStorageMemory()
	manager := NewManager(currentProposal, generateSessionID, sessionStore, promiseProcessor, mockPaymentOrchestratorFactory)

	_, err := manager.Create(identity.FromAddress("deadbeef"), currentProposalID, expectedSessionConfig, nil)
	assert.NoError(t, err)
	assert.True(t, promiseProcessor.started)
	assert.Exactly(t, currentProposal, promiseProcessor.proposal)
}
