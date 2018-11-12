package main

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

//====================================================
type lz struct {
	namI    uint16
	counter uint
	flag    uint8
}

var field []lz
var mfield sync.Mutex

func PrepareField() {
	var kun lz
	field = make([]lz, 0)
	for i := 0x61; i < 0x7a; i++ {
		for j := 0x61; j < 0x7a; j++ {
			kun.namI = uint16(i<<8 + j)
			kun.counter = 0
			kun.flag = 0
			field = append(field, kun)
		}
	}
}
func GetNamI(n uint16) (rez [2]byte) {
	rez[0] = byte((n & 0xFF00) >> 8)
	rez[1] = byte(n & 0xFF)
	return
}

type zss struct {
	name uint16
	nstr int
}

const NSTROKA = 50

var stroka [NSTROKA]zss
var mstroka sync.Mutex

func FillStroka() {
	var ex int
	for k := 0; k < NSTROKA; {
		ex = 0
		for ex == 0 {
			n := rand.Intn(len(field))
			if field[n].flag != 1 {
				stroka[k].name = field[n].namI
				stroka[k].nstr = n
				field[n].flag = 1
				k++
				ex = 1
			}
		}

	}
}
func GetFromStroka() (rez uint16) {
	n := rand.Intn(NSTROKA)
	mfield.Lock()
	field[stroka[n].nstr].counter++
	mfield.Unlock()
	mstroka.Lock()
	rez = stroka[n].name
	mstroka.Unlock()
	return
}

func Generator() {
	ex := 0
	for ex == 0 {
		time.Sleep(200 * time.Millisecond)
		u := 0
		n := rand.Intn(NSTROKA)
		for u == 0 {
			p := rand.Intn(len(field))
			mfield.Lock()
			if field[p].flag != 1 {
				field[p].flag = 1
				mstroka.Lock()
				field[stroka[n].nstr].flag = 0
				stroka[n].name = field[p].namI
				stroka[n].nstr = p
				mstroka.Unlock()
			}
			u = 1
			mfield.Unlock()
		}
	}
}
func GetAdminStat(limit int) []string {
	var arez []string
	var kstr string
	var tcounter, g int
	arez = make([]string, 0)

	for i := 0; i < len(field); i++ {
		srez := ""
		if field[i].counter > 0 {
			y := GetNamI(field[i].namI)
			srez += string(y[:]) + "-"
			srez += strconv.Itoa(int(field[i].counter)) + "\r\n"
			g = len(srez)
			if tcounter+g >= limit {
				arez = append(arez, kstr)
				tcounter = g
				kstr = srez
			} else {
				tcounter += g
				kstr += srez
			}
		}
	}
	if kstr == "" {
		kstr = " "
	}
	arez = append(arez, kstr)

	return arez
}

func TCP_read(conn net.Conn, buf []byte, chm chan<- int) {
	ex := 0
	for ex != 1 {
		rlen, err := conn.Read(buf)
		if err != nil {
			//fmt.Println("Error reading:", err.Error())
			ex = 1
		} else {
			chm <- rlen
			ex = 1
		}
	}
	close(chm)
}

const SENDLIMIT = 2000
const RECVLIMIT = 2000

func handleConnection(conn net.Conn) {
	var i int
	var rstr string
	bufR := make([]byte, RECVLIMIT)
	bufS := make([]byte, SENDLIMIT)
	ch := make(chan int)
	ex := 0
	go TCP_read(conn, bufR, ch)
	for ex != 1 {
		select {
		case x, ok := <-ch:
			if ok {
				if x <= 0 {
					ex = 1
				} else {
					//==========обработчик==============
					rstr = string(bufR[:x])
					astr := strings.Split(rstr, " ")
					if len(astr) > 1 {
						if astr[0] == "GET" {
							if astr[1] == "/request" {
								f := GetFromStroka()
								u := GetNamI(f)
								_, erw := conn.Write(u[:])
								if erw != nil {
									fmt.Println("Error writing:", erw.Error())
									ex = 1
								}
							} else {
								if astr[1] == "/admin/request" {
									ra := GetAdminStat(SENDLIMIT)
									for i = 0; i < len(ra); i++ {
										bufS = []byte(ra[i])
										_, erw := conn.Write(bufS[:])
										if erw != nil {
											fmt.Println("Error writing:", erw.Error())
										}
									}
								}
							}
						}
					}
					ex = 1
					conn.Close()
					//----------------------------------
				}
			} else {
				//fmt.Println("Channel closed!")
				ex = 1
			}
		default:
		}
	}

}

const PORT = ":80"

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	PrepareField()
	FillStroka()
	go Generator()
	fmt.Println("Start to listen on ", PORT)
	ln, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Println("Bind error:", err.Error())
	} else {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Accept error:", err.Error())
			} else {

				go handleConnection(conn)
			}
		}
	}
	ln.Close()
}
