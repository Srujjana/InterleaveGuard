package scenario

import "testing"

func TestTransferBreaksBalanceInvariant(t *testing.T) {
    accounts := &Accounts{A: 100, B: 50}

    debited := make(chan struct{})
    continueTransfer := make(chan struct{})
    finished := make(chan struct{})

    go func() {
        Transfer(accounts, debited, continueTransfer)
        close(finished)
    }()

    <-debited

    total := accounts.A + accounts.B

    close(continueTransfer)
    <-finished

    if total == 150 {
        t.Fatal("expected to observe a broken balance invariant")
    }

    if total != 140 {
        t.Fatalf("expected temporary total 140, got %d", total)
    }
}
