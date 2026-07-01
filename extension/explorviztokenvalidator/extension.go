package explorviztokenvalidator

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/ExplorViz/otel-collector/common/genproto/tokenpb"
	"github.com/ExplorViz/otel-collector/common/token"
)

type tokenValidatorExtension struct {
	logger     *zap.Logger
	tokenStore *token.InMemStore
	client     *kgo.Client
	seedBroker string
	topic      string
	cancelPoll context.CancelFunc
}

func newTokenValidatorExtension(cfg *Config, log *zap.Logger) tokenValidatorExtension {
	return tokenValidatorExtension{
		logger:     log,
		tokenStore: token.NewInMemStore(),
		seedBroker: cfg.Broker,
		topic:      cfg.Topic,
	}
}

func (e *tokenValidatorExtension) Start(_ context.Context, _ component.Host) error {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(e.seedBroker),
		kgo.ConsumeTopics(e.topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()), // replay token events from beginning
	)
	if err != nil {
		return fmt.Errorf("unable to initialize kgo client: %v", err)
	}
	e.client = cl

	pollCtx, cancel := context.WithCancel(context.Background())
	e.cancelPoll = cancel
	go runKafkaPollLoop(pollCtx, cl, e.tokenStore, e.logger)

	return nil
}

func (e *tokenValidatorExtension) Shutdown(_ context.Context) error {
	if e.cancelPoll != nil {
		e.cancelPoll()
	}

	if e.client != nil {
		e.client.Close()
	}

	return nil
}

func (e *tokenValidatorExtension) Validator() token.Validator {
	return e.tokenStore
}

// runKafkaPollLoop continuously fetches records from the given client, attempts to deserialize them
// as [tokenpb.TokenEvent]s, and updates the provided [InMemTokenStore] accordingly.
func runKafkaPollLoop(ctx context.Context, cl *kgo.Client, ts *token.InMemStore, log *zap.Logger) {
	for {
		fs := cl.PollFetches(ctx)
		if ctx.Err() != nil {
			log.Debug("exiting kafka token poll loop")
			break
		}
		fs.EachRecord(func(r *kgo.Record) {
			if r.Value == nil {
				ts.Delete(string(r.Key))
				log.Debug("received tombstone token record, deleting token", zap.String("tokenID", string(r.Key)))
				return
			}

			var t tokenpb.TokenEvent
			if err := proto.Unmarshal(r.Value, &t); err != nil {
				log.Debug("invalid protocol buffer for token event", zap.Error(err))
				return
			}

			switch t.GetType() {
			case tokenpb.EventType_EVENT_TYPE_CREATED:
				ts.Put(token.LandscapeToken{ID: string(r.Key), Secret: t.GetToken().GetSecret()})
				log.Debug("received token create event", zap.String("tokenID", t.GetToken().GetId()), zap.String("tokenSecret", t.GetToken().GetSecret()))
			case tokenpb.EventType_EVENT_TYPE_DELETED:
				ts.Delete(string(r.Key))
				log.Debug("received token delete event", zap.String("tokenID", string(r.Key)))
			default:
				log.Warn("received unhandled token event type", zap.String("eventType", tokenpb.EventType_name[int32(t.GetType())]))
			}
		})
	}

}
