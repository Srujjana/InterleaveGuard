package scenario

func SendWithoutReceiver(ch chan<- int) {
    ch <- 10
}
