package db_hooks

import (
	"context"
	"gorm.io/gorm"

	"github.com/bazilio91/sferra-cloud/pkg/proto"
	"github.com/bazilio91/sferra-cloud/pkg/types"
)

// registerDataRecognitionTaskHooks registers hooks for DataRecognitionTaskORM
func (sm *StateMachine) registerDataRecognitionTaskHooks() {
	sm.db.Callback().Create().After("gorm:commit_or_rollback_transaction").Register("state_machine:create", sm.processTaskState)
	sm.db.Callback().Update().After("gorm:commit_or_rollback_transaction").Register("state_machine:update", sm.processTaskState)
}

func (sm *StateMachine) processTaskState(tx *gorm.DB) {
	if tx.Statement.Schema == nil || tx.Statement.Schema.ModelType.Name() != "DataRecognitionTaskORM" {
		return
	}

	if task, ok := tx.Statement.Dest.(*proto.DataRecognitionTaskORM); ok {
		// Skip state machine processing if we're already in a terminal state
		if types.IsTerminalState(proto.Status(task.Status)) {
			return
		}
		if err := sm.Process(context.Background(), task); err != nil {
			panic(err)
		}

		return
	}
}
