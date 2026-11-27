package scenario

func WorkerA(aToB chan<- int, bToA <-chan int) {
    aToB <- 1
    <-bToA
}

func WorkerB(aToB <-chan int, bToA chan<- int) {
    bToA <- 2
    <-aToB
}
