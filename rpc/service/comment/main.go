package main

import (
	comment "douyin/rpc/kitex_gen/comment/commentservice"
	"log"
)

func main() {
	svr := comment.NewServer(new(CommentServiceImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
