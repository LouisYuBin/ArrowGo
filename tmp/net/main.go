package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var contextCancel context.CancelFunc
var ctx context.Context


//并发处理goroutine数
const concurrance = 10000
//监听端口数
const netPort = 8899

func main() {

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", netPort))
	if err != nil {
		fmt.Println("net listen error:", err)
	}
	//信号处理，进程平滑退出
	go signalHandle(listener)
	//初始化多协程等待组
	wg := &sync.WaitGroup{}
	//初始化平滑退出上下文组
	ctx, contextCancel = context.WithCancel(context.Background())
	for i := 0; i < concurrance; i++ {
		go func() {
			wg.Add(1)
			acceptHandle(listener, ctx)
			wg.Done()
		}()
	}

	wg.Wait()
	fmt.Println("quitted!")
}

//循环监听
func acceptHandle(listener net.Listener, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Accept error:", err)
				return
			}

			readAndWriteHandle(conn)
		}
	}
}

func readAndWriteHandle(conn net.Conn) {
	//读取请求内容并打印
	accetpString := make([]byte, 128)
	_, err := conn.Read(accetpString)
	if err != nil {
		fmt.Println("read error:", err)
	}
	fmt.Println(string(accetpString))
	//回复请求
	responseString := "got it! " + time.Now().Format("2006-01-02 15:04:05")
	conn.Write([]byte(responseString))
}


//信号处理
func signalHandle(listener net.Listener) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT)
	for {
		s := <-signalChan
		switch s {
		case syscall.SIGHUP:
			//Todo
		case syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			contextCancel()
			listener.Close()
		}
	}
}
