package auctioneer

import (
	"encoding/json"
	"errors"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/consuladapter"
	loggingclient "code.cloudfoundry.org/diego-logging-client"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket"
	"github.com/tedsuo/ifrit"
)

const LockSchemaKey = "auctioneer_lock"

func LockSchemaPath() string {
	return locket.LockSchemaPath(LockSchemaKey)
}

type Presence struct {
	AuctioneerID      string `json:"auctioneer_id"`
	AuctioneerAddress string `json:"auctioneer_address"`
}

func NewPresence(id, address string) Presence {
	return Presence{
		AuctioneerID:      id,
		AuctioneerAddress: address,
	}
}

func (a Presence) Validate() error {
	if a.AuctioneerID == "" {
		return errors.New("auctioneer_id cannot be blank")
	}

	if a.AuctioneerAddress == "" {
		return errors.New("auctioneer_address cannot be blank")
	}

	return nil
}

type ServiceClient interface {
	NewAuctioneerLockRunner(logger lager.Logger, presence Presence, retryInterval, lockTTL time.Duration, metronClient loggingclient.IngressClient) (ifrit.Runner, error)
	CurrentAuctioneer() (Presence, error)
	CurrentAuctioneerAddress() (string, error)
}

type serviceClient struct {
	consulClient consuladapter.Client
	clock        clock.Clock
}

func NewServiceClient(consulClient consuladapter.Client, clock clock.Clock) ServiceClient {
	return serviceClient{
		consulClient: consulClient,
		clock:        clock,
	}
}

func (c serviceClient) NewAuctioneerLockRunner(logger lager.Logger, presence Presence, retryInterval, lockTTL time.Duration, metronClient loggingclient.IngressClient) (ifrit.Runner, error) {
	if err := presence.Validate(); err != nil {
		return nil, err
	}

	payload, err := json.Marshal(presence)
	if err != nil {
		return nil, err
	}
	return locket.NewLock(logger, c.consulClient, LockSchemaPath(), payload, c.clock, retryInterval, lockTTL, locket.WithMetronClient(metronClient)), nil
}

func (c serviceClient) CurrentAuctioneer() (Presence, error) {
	presence := Presence{}

	value, err := c.getAcquiredValue(LockSchemaPath())
	if err != nil {
		return presence, err
	}

	if err := json.Unmarshal(value, &presence); err != nil {
		return presence, err
	}

	if err := presence.Validate(); err != nil {
		return presence, err
	}

	return presence, nil
}

func (c serviceClient) CurrentAuctioneerAddress() (string, error) {
	presence, err := c.CurrentAuctioneer()
	return presence.AuctioneerAddress, err
}

func (c serviceClient) getAcquiredValue(key string) ([]byte, error) {
	kvPair, _, err := c.consulClient.KV().Get(key, nil)
	if err != nil {
		return nil, err
	}

	if kvPair == nil || kvPair.Session == "" {
		return nil, consuladapter.NewKeyNotFoundError(key)
	}

	return kvPair.Value, nil
}
