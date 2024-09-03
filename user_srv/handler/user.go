package handler

import (
	"context"
	"crypto/sha512"
	"fmt"
	"github.com/anaskhan96/go-password-encoder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"strings"

	"mxshop_srvs/user_srv/global"
	"mxshop_srvs/user_srv/model"
	"mxshop_srvs/user_srv/proto"
	"time"
)

type UserServer struct{}

// 分页
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// model to proto
func Model2Response(user model.User) *proto.UserInfoResponse {
	UserInfoRsp := &proto.UserInfoResponse{
		Id:       user.Id,
		Mobile:   user.Mobile,
		Password: user.Password,
		Nickname: user.Nickname,
		Gender:   uint32(user.Gender),
		Role:     uint32(user.Role),
	}
	if user.Birthday != nil {
		UserInfoRsp.BirthDay = uint64(user.Birthday.Unix())
	}
	return UserInfoRsp
}

func (u *UserServer) GetUserList(ctx context.Context, req *proto.PageInfo) (*proto.UserListResponse, error) {
	// 实例化用户组
	var users []model.User
	result := global.DB.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	// 初始化返回对象
	rsp := &proto.UserListResponse{}
	rsp.Total = uint32(result.RowsAffected)

	// 处理分页
	global.DB.Scopes(Paginate(int(req.Pn), int(req.PSize))).Find(&users)

	for _, user := range users {
		UserInfoRsp := Model2Response(user)
		rsp.Data = append(rsp.Data, UserInfoRsp)
	}
	return rsp, nil
}

func (u *UserServer) GetUserById(ctx context.Context, req *proto.IdRequest) (*proto.UserInfoResponse, error) {
	var user model.User
	result := global.DB.First(&user, req.Id)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}
	UserInfoRsp := Model2Response(user)
	return UserInfoRsp, nil
}

func (u *UserServer) GetUserByMobile(ctx context.Context, req *proto.MobileRequest) (*proto.UserInfoResponse, error) {
	var user model.User
	result := global.DB.Where(&model.User{Mobile: req.Mobile}).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "手机号不存在")
	}
	UserInfoRsp := Model2Response(user)
	return UserInfoRsp, nil
}

func (u *UserServer) CreateUser(ctx context.Context, req *proto.CreateUserInfo) (*proto.UserInfoResponse, error) {
	// 先查询用户是否存在
	var user model.User
	result := global.DB.Find(&model.User{Mobile: req.Mobile}).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected != 0 {
		return nil, status.Error(codes.AlreadyExists, "手机号已注册")
	}
	user.Nickname = req.Nickname
	user.UpdatedAt = time.Now()
	// 加密
	options := &password.Options{16, 100, 32, sha512.New}
	salt, encodedPwd := password.Encode(req.Password, options)
	pwd := fmt.Sprintf("sha512$%s$%s", salt, encodedPwd)

	user.Password = pwd
	tx := global.DB.Create(&user)
	if tx.Error != nil {
		return nil, result.Error
	}
	UserInfoRsp := Model2Response(user)
	return UserInfoRsp, nil
}

func (u *UserServer) UpdateUser(ctx context.Context, req *proto.UpdateUserInfo) (*emptypb.Empty, error) {
	// 修改个人中心页面
	var user model.User
	result := global.DB.First(&user, req.Id)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, status.Error(codes.AlreadyExists, "用户不存照")
	}
	birthDay := time.Unix(int64(req.BirthDay), 0)
	user.Birthday = &birthDay
	user.Mobile = req.Mobile
	user.Nickname = req.Nickname
	user.UpdatedAt = time.Now()
	user.Gender = uint8(req.Gender)
	tx := global.DB.Save(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &emptypb.Empty{}, nil
}

func (u *UserServer) CheckUserPassword(ctx context.Context, req *proto.PasswordCheckInfo) (*proto.CheckResponse, error) {
	options := &password.Options{16, 100, 32, sha512.New}
	encrypted := strings.Split(req.EncryptedPassword, "$")
	if len(encrypted) != 4 {
		return nil, status.Error(codes.Internal, "密文不正确")
	}
	Check := password.Verify(req.Password, encrypted[2], encrypted[3], options)
	return &proto.CheckResponse{
		Success: Check,
	}, nil
}
