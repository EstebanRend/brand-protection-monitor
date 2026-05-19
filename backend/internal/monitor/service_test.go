package monitor

import (
	"context"
	"errors"
	"testing"
)

func TestRunOnceReturnsAlreadyInProgressWhenLocked(t *testing.T) {
	t.Parallel()

	service := &Service{isRunning: true}

	err := service.RunOnce(context.Background())
	if !errors.Is(err, ErrRunAlreadyInProgress) {
		t.Fatalf("RunOnce error = %v, want %v", err, ErrRunAlreadyInProgress)
	}
}

func TestComputeProcessingRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		previousTreeSize int64
		currentTreeSize  int64
		batchSize        int64
		wantStart        int64
		wantEnd          int64
		wantNextTreeSize int64
		wantHasWork      bool
	}{
		{
			name:             "initial run processes recent batch",
			previousTreeSize: 0,
			currentTreeSize:  200,
			batchSize:        100,
			wantStart:        100,
			wantEnd:          199,
			wantNextTreeSize: 200,
			wantHasWork:      true,
		},
		{
			name:             "incremental run starts from previous tree size",
			previousTreeSize: 200,
			currentTreeSize:  220,
			batchSize:        100,
			wantStart:        200,
			wantEnd:          219,
			wantNextTreeSize: 220,
			wantHasWork:      true,
		},
		{
			name:             "large backlog is capped by batch size",
			previousTreeSize: 50,
			currentTreeSize:  500,
			batchSize:        100,
			wantStart:        50,
			wantEnd:          149,
			wantNextTreeSize: 150,
			wantHasWork:      true,
		},
		{
			name:             "no new entries has no work",
			previousTreeSize: 220,
			currentTreeSize:  220,
			batchSize:        100,
			wantStart:        0,
			wantEnd:          -1,
			wantNextTreeSize: 220,
			wantHasWork:      false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			start, end, nextTreeSize, hasWork := computeProcessingRange(tt.previousTreeSize, tt.currentTreeSize, tt.batchSize)
			if start != tt.wantStart || end != tt.wantEnd || nextTreeSize != tt.wantNextTreeSize || hasWork != tt.wantHasWork {
				t.Fatalf(
					"computeProcessingRange(%d, %d, %d) = (%d, %d, %d, %t), want (%d, %d, %d, %t)",
					tt.previousTreeSize,
					tt.currentTreeSize,
					tt.batchSize,
					start,
					end,
					nextTreeSize,
					hasWork,
					tt.wantStart,
					tt.wantEnd,
					tt.wantNextTreeSize,
					tt.wantHasWork,
				)
			}
		})
	}
}
