package heandler

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"mxshop_srvs/user_srv/global"
	"mxshop_srvs/user_srv/model"
	"mxshop_srvs/user_srv/proto"
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

func (u *UserServer) GetUserList(ctx context.Context, req *proto.PageInfo, opts ...grpc.CallOption) (*proto.UserListResponse, error) {
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
