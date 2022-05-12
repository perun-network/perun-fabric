package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	adj "github.com/perun-network/perun-fabric/adjudicator"
	pkgjson "github.com/perun-network/perun-fabric/pkg/json"
	"github.com/perun-network/perun-fabric/tests"
	"math/big"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
	"time"
)

// TODO: Unite constants with funder
const (
	txStateReg = "StateReg"
	txRegister = "Register"
)

// Adjudicator provides methods for dispute resolution on the ledger.
type Adjudicator struct {
	adj     *client.Contract
	polling time.Duration
}

type AdjudicatorOpt func(*Adjudicator)

func AdjudicatorPollingIntervalOpt(d time.Duration) AdjudicatorOpt {
	return func(a *Adjudicator) {
		a.polling = d
	}
}

func NewAdjudicator(network *client.Network, adjudicator string, opts ...AdjudicatorOpt) *Adjudicator {
	a := &Adjudicator{
		adj:     network.GetContract(adjudicator),
		polling: defaultPollingInterval,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Register registers the given ledger channel state on-chain.
// If the channel has locked funds into sub-channels, the corresponding
// signed sub-channel states must be provided.
func (a *Adjudicator) Register(ctx context.Context, req channel.AdjudicatorReq, subChannels []channel.SignedState) error {
	if len(subChannels) > 0 {
		return fmt.Errorf("subchannels not supported")
	}
	sigCh := a.createSignedChannel(req) // Repackaging - TODO: Check for a better way here
	return a.register(sigCh)
}

// Withdraw concludes and withdraws the registered state, so that the
// final outcome is set on the asset holders and funds are withdrawn.
// If the channel has locked funds in sub-channels, the states of the
// corresponding sub-channels need to be supplied additionally.
func (a *Adjudicator) Withdraw(ctx context.Context, req channel.AdjudicatorReq, subStates channel.StateMap) error {
	if len(subStates) > 0 {
		return fmt.Errorf("subchannels not supported")
	}
	// TODO: Check how ctx can be used
	sigCh := a.createSignedChannel(req) // Repackaging - TODO: Check for a better way here
	err := a.register(sigCh)            // TODO: Does this trigger a dispute?
	if err != nil {
		return fmt.Errorf("concluding: %w", err)
	}

	_, err = a.withdraw(req.Params.ID())
	return err
}

// Progress progresses the state of a previously registered channel on-chain.
// The signatures for the old state can be nil as the state is already
// registered on the adjudicator.
func (a *Adjudicator) Progress(ctx context.Context, req channel.ProgressReq) error {
	return fmt.Errorf("unsupported")
}

// Subscribe returns an AdjudicatorEvent subscription.
//
// The context should only be used to establish the subscription. The
// framework will call Close on the subscription once the respective channel
// controller shuts down.
func (a *Adjudicator) Subscribe(ctx context.Context, ch channel.ID) (channel.AdjudicatorSubscription, error) {
	sub := NewEventSubscription(a, ch)
	return sub, nil
}

func (a *Adjudicator) createSignedChannel(req channel.AdjudicatorReq) *adj.SignedChannel {
	p := req.Params.Clone()
	params := adj.Params{
		ChallengeDuration: p.ChallengeDuration,
		Parts:             p.Parts,
		Nonce:             p.Nonce,
	}

	s := req.Tx.State.Clone()
	state := adj.State{
		ID:       s.ID,
		Version:  s.Version,
		Balances: s.Balances[0], // We only support a single asset
	}

	return &adj.SignedChannel{
		Params: params,
		State:  state,
		Sigs:   req.Tx.Sigs,
	}
}

func (a *Adjudicator) deposit(id channel.ID, amount *big.Int) error {
	args, err := pkgjson.MultiMarshal(id, amount)
	if err != nil {
		return err
	}
	_, err = a.adj.SubmitTransaction(txDeposit, args...)
	return err
}

func (a *Adjudicator) holding(id channel.ID, addr wallet.Address) (*big.Int, error) {
	args, err := pkgjson.MultiMarshal(id, addr)
	if err != nil {
		return nil, err
	}
	return tests.BigIntWithError(a.adj.SubmitTransaction(txHolding, args...))
}

func (a *Adjudicator) totalHolding(id channel.ID, addrs []wallet.Address) (*big.Int, error) {
	args, err := pkgjson.MultiMarshal(id, addrs)
	if err != nil {
		return nil, err
	}
	return tests.BigIntWithError(a.adj.SubmitTransaction(txTotalHolding, args...))
}

func (a *Adjudicator) register(ch *adj.SignedChannel) error {
	arg, err := json.Marshal(ch)
	if err != nil {
		return err
	}
	_, err = a.adj.SubmitTransaction(txRegister, string(arg))
	return err
}

func (a *Adjudicator) stateReg(id channel.ID) (*adj.StateReg, error) {
	arg, err := json.Marshal(id)
	if err != nil {
		return nil, err
	}
	regJson, err := a.adj.SubmitTransaction(txStateReg, string(arg))
	if err != nil {
		return nil, err
	}
	var reg adj.StateReg
	return &reg, json.Unmarshal(regJson, &reg)
}

func (a *Adjudicator) withdraw(id channel.ID) (*big.Int, error) {
	args, err := pkgjson.MultiMarshal(id)
	if err != nil {
		return nil, err
	}
	return tests.BigIntWithError(a.adj.SubmitTransaction(txWithdraw, args...))
}
