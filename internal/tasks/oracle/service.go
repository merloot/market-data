package oracle

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/merloot/market-data/internal/realtime"
)

type Service struct {
	log *slog.Logger
	currencyRepository CurrencyRepository
	marketDataHistoryRepository MarketDataHistoryRepository
	provider MarketDataProvider
	events realtime.EventPublisher
}

func NewService(
	log *slog.Logger,
	currencyRepository CurrencyRepository,
	marketDataHistoryRepository MarketDataHistoryRepository,
	provider MarketDataProvider,
	events realtime.EventPublisher,
) *Service {
	return &Service{
		log: log,
		currencyRepository: currencyRepository,
		marketDataHistoryRepository: marketDataHistoryRepository,
		provider: provider,
		events: events,
	}
}

func (s *Service) Execute(ctx context.Context) error {
	currencies, err := s.currencyRepository.List(ctx)
	if err != nil {
		return fmt.Errorf("List currencies: %w", err)
	}
	if len(currencies) == 0 {
		s.log.Warn("No currencies to process")
		return nil
	}

	history, err := s.collectHistory(ctx, currencies)
	if err != nil {
		return  fmt.Errorf("Collect history: %w", err)
	}

	actual, err :=s.collectActual(ctx, currencies)
	if err != nil {
		return fmt.Errorf("Collect actual: %w", err)
	}

	all := append(history, actual...)
	if err := s.marketDataHistoryRepository.Create(ctx, all); err != nil {
		return fmt.Errorf("Save: %w", err)
	}

	if err := s.publishUpdates(ctx,actual); err != nil {
		s.log.Warn("Publish updates", "err", err)
	}

	if err := s.fetchMissingLogos(ctx); err != nil {
		s.log.Warn("Fetch logos", "err", err)
	}

	return nil
}