package dubbo

// import (
// 	"context"
// 	"time"

// 	hessian "github.com/apache/dubbo-go-hessian2"
// 	"github.com/apache/dubbo-go/config"
// 	"github.com/sirupsen/logrus"
// )

// func init() {
// 	config.SetProviderService(new(UserProvider))
// 	// ------for hessian2------
// 	hessian.RegisterPOJO(&User{})
// }

// type User struct {
// 	Id   string
// 	Name string
// 	Age  int32
// 	Time time.Time
// }

// type UserProvider struct {
// }

// func (u *UserProvider) GetUser(ctx context.Context, req []interface{}) (*User, error) {
// 	logrus.Info("req: ", req)
// 	rsp := User{"A001", "Alex Stocks", 18, time.Now()}
// 	logrus.Info("rsp ", rsp)
// 	return &rsp, nil
// }

// func (u *UserProvider) Reference() string {
// 	return "UserProvider"
// }

// func (u User) JavaClassName() string {
// 	return "com.ikurento.user.User"
// }
