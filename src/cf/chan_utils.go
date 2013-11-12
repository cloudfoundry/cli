package cf

func WaitForClose(stop chan bool) {
	for {
		_, open := <-stop
		if !open {
			break
		}
	}
}
