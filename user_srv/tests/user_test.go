package tests

import (
	"context"
	"crypto/sha512"
	"fmt"
	"github.com/anaskhan96/go-password-encoder"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"mxshop_srvs/user_srv/global"
	"mxshop_srvs/user_srv/model"
	"mxshop_srvs/user_srv/proto"
	"testing"
)

var userClient proto.UserClient

func Init() {
	conn, err := grpc.NewClient("127.0.0.1:8088", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	userClient = proto.NewUserClient(conn)
}

func TestGetUserList(t *testing.T) {
	Init()
	rsp, err := userClient.GetUserList(context.Background(), &proto.PageInfo{
		Pn:    1,
		PSize: 10,
	})
	if err != nil {
		panic(err)
	}
	if len(rsp.Data) < 1 {
		fmt.Println("没有用户")
		return
	}
	for _, value := range rsp.Data {
		fmt.Println(value)
	}
}

// 注册用户
func TestCreateUser(t *testing.T) {
	var user model.User
	user.Nickname = "test"
	user.Mobile = "13888888888"
	options := &password.Options{16, 100, 32, sha512.New}
	salt, encodedPwd := password.Encode("test", options)
	pwd := fmt.Sprintf("sha512$%s$%s", salt, encodedPwd)
	user.Password = pwd
	tx := global.DB.Create(&user)
	if tx.Error != nil {
		fmt.Printf("创建用户失败 %v", tx.Error)
	}
}
