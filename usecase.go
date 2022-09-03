package app

import (
	"context"
	"github.com/adamluzsi/frameless/pkg/txs"
	"github.com/adamluzsi/frameless/ports/crud"
	"github.com/adamluzsi/testcase/random"
	"log"
)

type UseCase struct {
	Service1 SomeService
	Service2 SomeService
	Service3 FlakyService
}

func (uc UseCase) Do(ctx context.Context, ent Entity) (rErr error) {
	ctx, _ = txs.Begin(ctx)
	defer txs.Finish(&rErr, ctx)

	if err := uc.Service1.Do(ctx, ent); err != nil {
		return err // trigger RollbackTx on error
	}

	if err := uc.Service2.Do(ctx, ent); err != nil {
		return err // trigger RollbackTx on error, including Service1 rollback steps
	}

	if err := uc.Service3.Do(ctx, ent); err != nil {
		return err // trigger RollbackTx on error, including Service1 rollback steps
	}

	return nil
}

type SomeService struct {
	EntityRepository EntityRepository
}

type EntityRepository interface {
	crud.Creator[Entity]
	crud.Finder[Entity, string]
	crud.Updater[Entity]
	crud.Deleter[string]
}

func (service SomeService) Do(ctx context.Context, ent Entity) (rErr error) {
	ctx, _ = txs.Begin(ctx)
	defer txs.Finish(&rErr, ctx)

	if err := service.EntityRepository.Create(ctx, &ent); err != nil {
		return err
	}

	_ = txs.OnRollback(ctx, func() error {
		log.Println("INFO", "rollback entity: "+ent.ID)
		return service.EntityRepository.DeleteByID(context.Background(), ent.ID)
	})

	return nil
}

type FlakyService struct{}

func (service FlakyService) Do(ctx context.Context, ent Entity) (rErr error) {
	_ = ent
	ctx, _ = txs.Begin(ctx)
	defer txs.Finish(&rErr, ctx)

	_ = txs.OnRollback(ctx, func() {
		log.Println("INFO", "FlakyService is flaky again...")
	})

	rnd := random.New(random.CryptoSeed{})
	if rnd.Bool() {
		return rnd.Error()
	}

	return nil
}
