package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type counterEvent struct {
	ClientID  string
	Action    counterActionType
	CreatedAt time.Time
}

const (
	eventQueueSize    = 10000
	eventBatchSize    = 10
	eventFlushTimeout = 1000 * time.Millisecond
)

var eventQueue chan counterEvent

func enqueueCounterEvent(event counterEvent) {
	eventQueue <- event
}

func startCounterEventWorker(db *sql.DB, queue <-chan counterEvent) {
	go func() {
		ticker := time.NewTicker(eventFlushTimeout)
		defer ticker.Stop()
		batch := make([]counterEvent, 0, eventBatchSize)

		flush := func() {
			if len(batch) == 0 {
				return
			}
			start := time.Now()
			if err := insertCounterEvents(db, batch); err != nil {
				log.Printf("counter event batch failed: count=%d err=%v", len(batch), err)
			} else {
				log.Printf("counter event batch ok: count=%d ms=%d", len(batch), time.Since(start).Milliseconds())
			}
			batch = batch[:0]
		}

		for {
			select {
			case event := <-queue:
				batch = append(batch, event)
				if len(batch) >= eventBatchSize {
					flush()
				}
			case <-ticker.C:
				flush()
			}
		}
	}()
}

func insertCounterEvents(db *sql.DB, events []counterEvent) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO counter_events (client_id, action, created_at) VALUES (?, ?, ?)")
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("prepare insert failed: %w (rollback: %v)", err, rollbackErr)
		}
		return err
	}
	defer stmt.Close()

	for _, event := range events {
		if _, err := stmt.Exec(event.ClientID, string(event.Action), event.CreatedAt); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return fmt.Errorf("insert failed: %w (rollback: %v)", err, rollbackErr)
			}
			return err
		}
	}
	return tx.Commit()
}
