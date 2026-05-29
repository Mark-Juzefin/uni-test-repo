package product

import (
	"context"
	"errors"
	"testing"

	"uni-test-repo/pkg/postgres"
	"uni-test-repo/services/products/internal/outbox"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// fakeTransactor runs the body with a nil Executor — repos are faked, so no DB
// is needed to exercise the in-transaction code path.
type fakeTransactor struct{}

func (fakeTransactor) InTransaction(_ context.Context, _ pgx.TxIsoLevel, fn func(postgres.Executor) error) error {
	return fn(nil)
}

type fakeRepo struct {
	created   []Product
	createErr error

	deletedID uuid.UUID
	deleteErr error

	listItems []Product
	listTotal int
	gotLimit  int
	gotOffset int
}

func (f *fakeRepo) Create(_ context.Context, p Product) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, p)
	return nil
}

func (f *fakeRepo) Delete(_ context.Context, id uuid.UUID) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deletedID = id
	return nil
}

func (f *fakeRepo) List(_ context.Context, limit, offset int) ([]Product, int, error) {
	f.gotLimit, f.gotOffset = limit, offset
	return f.listItems, f.listTotal, nil
}

type fakeOutbox struct {
	events    []outbox.Event
	createErr error
}

func (f *fakeOutbox) Create(_ context.Context, e outbox.Event) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.events = append(f.events, e)
	return nil
}

func (f *fakeOutbox) FetchUnpublished(_ context.Context, _ int) ([]outbox.Event, error) {
	return nil, nil
}

func (f *fakeOutbox) MarkPublished(_ context.Context, _ []uuid.UUID) error {
	return nil
}

func newTestService(repo *fakeRepo, ob *fakeOutbox) *ProductService {
	return NewProductService(
		repo,
		fakeTransactor{},
		func(postgres.Executor) ProductRepo { return repo },
		func(postgres.Executor) outbox.Store { return ob },
	)
}

func TestCreate_PersistsProductAndEmitsEvent(t *testing.T) {
	repo := &fakeRepo{}
	ob := &fakeOutbox{}
	svc := newTestService(repo, ob)

	got, err := svc.Create(context.Background(), CreateProductRequest{
		Name: "Widget", Description: "a widget", Price: 1999,
	})
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}
	if got.ID == uuid.Nil {
		t.Error("expected a generated ID")
	}
	if got.Name != "Widget" || got.Price != 1999 {
		t.Errorf("unexpected product: %+v", got)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Error("expected timestamps to be set")
	}

	if len(repo.created) != 1 || repo.created[0].ID != got.ID {
		t.Fatalf("product not persisted: %+v", repo.created)
	}
	if len(ob.events) != 1 {
		t.Fatalf("expected 1 outbox event, got %d", len(ob.events))
	}
	if ev := ob.events[0]; ev.EventType != outbox.EventProductCreated || ev.AggregateID != got.ID {
		t.Errorf("unexpected event: type=%q aggregate=%s", ev.EventType, ev.AggregateID)
	}
}

func TestCreate_RepoError_EmitsNoEvent(t *testing.T) {
	repo := &fakeRepo{createErr: errors.New("db down")}
	ob := &fakeOutbox{}
	svc := newTestService(repo, ob)

	if _, err := svc.Create(context.Background(), CreateProductRequest{Name: "X", Price: 1}); err == nil {
		t.Fatal("expected an error")
	}
	if len(ob.events) != 0 {
		t.Errorf("no outbox event must be emitted when the product write fails, got %d", len(ob.events))
	}
}

func TestDelete_RemovesProductAndEmitsEvent(t *testing.T) {
	repo := &fakeRepo{}
	ob := &fakeOutbox{}
	svc := newTestService(repo, ob)

	id := uuid.New()
	if err := svc.Delete(context.Background(), id); err != nil {
		t.Fatalf("Delete: unexpected error: %v", err)
	}
	if repo.deletedID != id {
		t.Errorf("deleted id = %s, want %s", repo.deletedID, id)
	}
	if len(ob.events) != 1 {
		t.Fatalf("expected 1 outbox event, got %d", len(ob.events))
	}
	if ev := ob.events[0]; ev.EventType != outbox.EventProductDeleted || ev.AggregateID != id {
		t.Errorf("unexpected event: type=%q aggregate=%s", ev.EventType, ev.AggregateID)
	}
}

func TestDelete_NotFound_EmitsNoEvent(t *testing.T) {
	repo := &fakeRepo{deleteErr: ErrNotFound}
	ob := &fakeOutbox{}
	svc := newTestService(repo, ob)

	err := svc.Delete(context.Background(), uuid.New())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if len(ob.events) != 0 {
		t.Errorf("no outbox event must be emitted when the product is missing, got %d", len(ob.events))
	}
}

func TestList_ClampsPagination(t *testing.T) {
	tests := []struct {
		name                  string
		reqLimit, reqOffset   int
		wantLimit, wantOffset int
	}{
		{"zero limit -> default", 0, 0, defaultListLimit, 0},
		{"negative offset -> zero", 0, -5, defaultListLimit, 0},
		{"limit over max -> max", 500, 10, maxListLimit, 10},
		{"in-range passes through", 30, 40, 30, 40},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{listItems: []Product{{Name: "p"}}, listTotal: 1}
			svc := newTestService(repo, &fakeOutbox{})

			res, err := svc.List(context.Background(), ListProductsRequest{Limit: tt.reqLimit, Offset: tt.reqOffset})
			if err != nil {
				t.Fatalf("List: unexpected error: %v", err)
			}
			if repo.gotLimit != tt.wantLimit || repo.gotOffset != tt.wantOffset {
				t.Errorf("repo got (limit=%d, offset=%d), want (%d, %d)", repo.gotLimit, repo.gotOffset, tt.wantLimit, tt.wantOffset)
			}
			if res.Limit != tt.wantLimit || res.Offset != tt.wantOffset {
				t.Errorf("response (limit=%d, offset=%d), want (%d, %d)", res.Limit, res.Offset, tt.wantLimit, tt.wantOffset)
			}
			if res.Total != 1 || len(res.Items) != 1 {
				t.Errorf("unexpected items/total: %+v", res)
			}
		})
	}
}
