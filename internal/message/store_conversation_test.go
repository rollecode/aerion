package message

import (
	"testing"
	"time"
)

// A thread whose messages all left the folder must read as gone, not as an
// empty conversation: the summary query aggregates, so it always yields a row.
func TestGetConversation_GoneAfterMove(t *testing.T) {
	store, accountID, inboxID := newBodyFailedTestStore(t)

	const archiveID = "folder-archive"
	if _, err := store.db.Exec(
		`INSERT INTO folders (id, account_id, name, path, folder_type) VALUES (?, ?, ?, ?, ?)`,
		archiveID, accountID, "Archive", "Archive", "archive",
	); err != nil {
		t.Fatalf("seed archive folder: %v", err)
	}

	m := &Message{
		ID:        "msg-1",
		AccountID: accountID,
		FolderID:  inboxID,
		UID:       1,
		MessageID: "<one@example.com>",
		ThreadID:  "<one@example.com>",
		Subject:   "Hello",
		Date:      time.Now(),
	}
	if err := store.Upsert(m); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	conv, err := store.GetConversation(m.ThreadID, inboxID)
	if err != nil || conv == nil || conv.MessageCount != 1 {
		t.Fatalf("before move: conv=%+v err=%v", conv, err)
	}

	if err := store.MoveMessages([]string{m.ID}, archiveID); err != nil {
		t.Fatalf("MoveMessages: %v", err)
	}

	conv, err = store.GetConversation(m.ThreadID, inboxID)
	if err != nil {
		t.Fatalf("after move: %v", err)
	}
	if conv != nil {
		t.Fatalf("after move: want nil conversation, got %+v", conv)
	}
}
