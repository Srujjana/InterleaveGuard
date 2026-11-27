package scenario

type Accounts struct {
    A int
    B int
}

func Transfer(accounts *Accounts, debited chan<- struct{}, continueTransfer <-chan struct{}) {
    accounts.A -= 10
    debited <- struct{}{}

    <-continueTransfer

    accounts.B += 10
}
