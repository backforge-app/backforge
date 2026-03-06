package transactor

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTransactor_WithinTx_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPool := NewMockPool(ctrl)
	mockTx := NewMockTx(ctrl)

	mockPool.EXPECT().
		Begin(gomock.Any()).
		Return(mockTx, nil)

	mockTx.EXPECT().
		Commit(gomock.Any()).
		Return(nil)

	mockTx.EXPECT().
		Rollback(gomock.Any()).
		Return(pgx.ErrTxClosed).
		AnyTimes()

	tnx := NewTransactor(mockPool, zap.NewNop().Sugar())

	err := tnx.WithinTx(context.Background(), func(ctx context.Context) error {
		return nil
	})

	require.NoError(t, err)
}

func TestTransactor_WithinTx_BeginError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPool := NewMockPool(ctrl)

	beginErr := errors.New("begin error")

	mockPool.EXPECT().
		Begin(gomock.Any()).
		Return(nil, beginErr)

	tnx := NewTransactor(mockPool, zap.NewNop().Sugar())

	err := tnx.WithinTx(context.Background(), func(ctx context.Context) error {
		return nil
	})

	require.ErrorIs(t, err, beginErr)
}

func TestTransactor_WithinTx_FnReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPool := NewMockPool(ctrl)
	mockTx := NewMockTx(ctrl)

	fnErr := errors.New("fn error")

	mockPool.EXPECT().
		Begin(gomock.Any()).
		Return(mockTx, nil)

	mockTx.EXPECT().
		Rollback(gomock.Any()).
		Return(nil)

	tnx := NewTransactor(mockPool, zap.NewNop().Sugar())

	err := tnx.WithinTx(context.Background(), func(ctx context.Context) error {
		return fnErr
	})

	require.ErrorIs(t, err, fnErr)
}

func TestTransactor_WithinTx_RollbackErrorLogged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPool := NewMockPool(ctrl)
	mockTx := NewMockTx(ctrl)

	fnErr := errors.New("fn error")
	rollbackErr := errors.New("rollback failed")

	mockPool.EXPECT().
		Begin(gomock.Any()).
		Return(mockTx, nil)

	mockTx.EXPECT().
		Rollback(gomock.Any()).
		Return(rollbackErr)

	tnx := NewTransactor(mockPool, zap.NewNop().Sugar())

	err := tnx.WithinTx(context.Background(), func(ctx context.Context) error {
		return fnErr
	})

	require.ErrorIs(t, err, fnErr)
}

func TestGetDB_ReturnsTxFromContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTx := NewMockTx(ctrl)

	ctx := contextWithTx(context.Background(), mockTx)

	db := GetDB(ctx, nil)

	require.Equal(t, mockTx, db)
}

func TestGetDB_ReturnsPoolIfNoTx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTx := NewMockTx(ctrl)

	db := GetDB(context.Background(), mockTx)

	require.Equal(t, mockTx, db)
}
