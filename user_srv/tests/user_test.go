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
	"strings"
	"testing"
)

var userClient proto.UserClient

func init() {
	conn, err := grpc.NewClient("127.0.0.1:8088", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	userClient = proto.NewUserClient(conn)
}

func TestGetUserList(t *testing.T) {
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
	salt, encodedPwd := password.Encode("123456", options)
	pwd := fmt.Sprintf("$sha512$%s$%s", salt, encodedPwd)
	user.Password = pwd
	tx := global.DB.Create(&user)
	if tx.Error != nil {
		fmt.Printf("创建用户失败 %v", tx.Error)
	}
}

// 校验密码
func TestPassWordCheck(t *testing.T) {
	InPwd := "123456"
	options := &password.Options{16, 100, 32, sha512.New}
	salt, encodedPwd := password.Encode(InPwd, options)
	// 秘文
	pwd := fmt.Sprintf("$sha512$%s$%s", salt, encodedPwd)
	fmt.Println(pwd)
	pwdInfo := strings.Split(pwd, "$")
	// 解密
	verify := password.Verify("1111", pwdInfo[2], pwdInfo[3], options)
	fmt.Printf("verify:%v\n", verify)
	//
	result, err := userClient.CheckUserPasswd(context.Background(), &proto.PasswordCheckInfo{
		Password:          "1234",
		EncryptedPassword: pwd,
	})
	if err != nil {
		panic(err)
	}
	//
	fmt.Printf("result:%v\n", result)

}
