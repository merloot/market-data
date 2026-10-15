package oracle

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/merloot/market-data/internal/realtime"
)

type Service struct {
	log                         *slog.Logger
	currencyRepository          CurrencyRepository
	marketDataHistoryRepository MarketDataHistoryRepository
	provider                    MarketDataProvider
	events                      realtime.EventPublisher
}

func NewService(
	log *slog.Logger,
	currencyRepository CurrencyRepository,
	marketDataHistoryRepository MarketDataHistoryRepository,
	provider MarketDataProvider,
	events realtime.EventPublisher,
) *Service {
	return &Service{
		log:                         log,
		currencyRepository:          currencyRepository,
		marketDataHistoryRepository: marketDataHistoryRepository,
		provider:                    provider,
		events:                      events,
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

	var errs []error

	history, err := s.collectHistory(ctx, currencies)
	if err != nil {
		s.log.Warn("Collect history partial", "err", err)
		errs = append(errs, fmt.Errorf("Collect history: %w", err))
	}

	actual, err := s.collectActual(ctx, currencies)
	if err != nil {
		s.log.Warn("Collect actual partial", "err", err)
		errs = append(errs, fmt.Errorf("Collect actual: %w", err))
	}

	all := append(history, actual...)
	if len(all) > 0 {
		if err := s.marketDataHistoryRepository.Create(ctx, all); err != nil {
			s.log.Error("Save failed", "err", err, "count", len(all))
			errs = append(errs, fmt.Errorf("Save: %w", err))
		} else {
			s.log.Info("Saved", "count", len(all))
		}
	} else {
		s.log.Warn("Nothing to save")
	}

	if err := s.publishUpdates(ctx, actual); err != nil {
		s.log.Warn("Publish updates", "err", err)
	}

	if err := s.fetchMissingLogos(ctx); err != nil {
		s.log.Warn("Fetch logos", "err", err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
