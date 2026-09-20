package contracts

import (
	"context"

	"moneytransfer/domain"
)

type AccountRepository interface {
	Retrieve(ctx context.Context, id domain.AccountID) (*domain.Account, error)
	UpdateMut(account *domain.Account) *Mutation
}

type Mutation struct {
	Table   string
	ID      string
	Updates map[string]interface{}
}

type Plan struct {
	mutations []*Mutation
}

func NewPlan() *Plan {
	return &Plan{}
}

func (p *Plan) Add(mutation *Mutation) {
	if mutation != nil {
		p.mutations = append(p.mutations, mutation)
	}
}

func (p *Plan) Mutations() []*Mutation {
	if p == nil {
		return nil
	}

	mutations := make([]*Mutation, len(p.mutations))
	copy(mutations, p.mutations)
	return mutations
}
