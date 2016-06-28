package main

import (
	"log"
	"time"

	"github.com/mattn/go-easyplugin"
)

func main() {
	// アプリケーション foobar のプラグインを起動
	ps, err := easyplugin.New("foobar")
	if err != nil {
		log.Fatal(err)
	}
	// 終了時には皆殺し
	defer ps.Unload()

	// client-xxx からの通知を受け取る
	ps.Handle(func(data string) {
		log.Println(data)
	})
	// 3秒後には死ぬ
	time.AfterFunc(3*time.Second, func() {
		ps.Stop()
	})

	var res struct {
		C int
	}
	err = ps.CallFor("server-app1", "Calc.Add", &res, struct {
		A, B int
	}{1, 3})
	log.Println(res) // 4 が返る

	// 通知がある間は待つ
	ps.ListenAndServe()
}
