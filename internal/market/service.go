package market

import "context"

type Service struct {
	impls map[string]Provider
}

func NewService(impls ...Provider) *Service {
	m := make(map[string]Provider, len(impls))
	for _, impl := range impls {
		m[impl.Provider()] = impl
	}
	return &Service{impls: m}
}

func (s *Service) impl(provider string) (Provider, error) {
	p, ok := s.impls[provider]
	if !ok {
		return nil, ErrProviderNotFound
	}
	return p, nil
}

func (s *Service) GetLogo(ctx context.Context, c Currency) (string, error) {
	p, err := s.impl(c.Provider)
	if err != nil {
		return "", err
	}
	return p.GetLogo(ctx, c.Currency)
}
