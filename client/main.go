package main

func main() {
	s := Socks5Server{"127.0.0.1:1090"}
	s.listen()
}
