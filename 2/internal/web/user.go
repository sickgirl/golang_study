package web

import (
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"golang_study/2/internal/domain"
	"golang_study/2/internal/service"
	"log"
	"net/http"
	"time"
)

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`

	userIdKey = "userId"
	bizLogin  = "login"
)

var _ handler = &UserHandler{}

type UserHandler struct {
	svc              service.UserService
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
	// 只有在使用 JWT 的时候才有用
	jwtKey string
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{
		svc:              svc,
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
	}
}

func (c *UserHandler) RegisterRoutes(server *gin.Engine) {
	// 直接注册
	//server.POST("/users/signup", c.SignUp)
	//server.POST("/users/login", c.Login)
	//server.POST("/users/edit", c.Edit)
	//server.GET("/users/profile", c.Profile)

	// 分组注册
	ug := server.Group("/users")
	ug.POST("/signup", c.SignUp)
	// session 机制
	//ug.POST("/login", c.Login)
	// JWT 机制
	ug.POST("/login", c.LoginJWT)
	ug.POST("/edit", c.Edit)
	//ug.GET("/profile", c.Profile)
	ug.GET("/profile", c.ProfileJWT)
	ug.POST("/login_sms/code/send", c.SendSMSLoginCode)
	ug.POST("/login_sms", c.LoginSMS)
}

func (c *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 验证码是对的
	// 登录或者注册用户
	u, err := c.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "系统错误"})
		return
	}
	err = c.setJWTToken(ctx, u.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "登录成功"})
}

// SendSMSLoginCode 发送短信验证码
func (c *UserHandler) SendSMSLoginCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 你也可以用正则表达式校验是不是合法的手机号
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "请输入手机号码"})
		return
	}

}

// SignUp 用户注册接口
func (c *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	// 当我们调用 Bind 方法的时候，如果有问题，Bind 方法已经直接写响应回去了
	if err := ctx.Bind(&req); err != nil {
		return
	}

	isEmail, err := c.emailRegexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "邮箱不正确")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入的密码不相同")
		return
	}

	isPassword, err := c.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK,
			"密码必须包含数字、特殊字符，并且长度不能小于 8 位")
		return
	}

	err = c.svc.Signup(ctx.Request.Context(),
		domain.User{Email: req.Email, Password: req.ConfirmPassword})

	if err == service.ErrUserDuplicateEmail {
		ctx.String(http.StatusOK, "重复邮箱，请换一个邮箱")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "服务器异常，注册失败")
		return
	}
	ctx.String(http.StatusOK, "hello, 注册成功")
}

// LoginJWT 用户登录接口，使用的是 JWT，如果你想要测试 JWT，就启用这个
func (c *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	// 当我们调用 Bind 方法的时候，如果有问题，Bind 方法已经直接写响应回去了
	if err := ctx.Bind(&req); err != nil {
		return
	}
	u, err := c.svc.Login(ctx.Request.Context(), req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或者密码不正确，请重试")
		return
	}
	err = c.setJWTToken(ctx, u.Id)
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	ctx.String(http.StatusOK, "登录成功")
}

func (c *UserHandler) setJWTToken(ctx *gin.Context, uid int64) error {
	a := time.Now().Add(time.Minute * 30)

	fmt.Printf("过期时间:%v", a)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaims{
		Id:        uid,
		UserAgent: ctx.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			// 演示目的设置为一分钟过期
			//ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
			// 在压测的时候，要将过期时间设置更长一些
			ExpiresAt: jwt.NewNumericDate(a),
		},
	})
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		return err
	}
	fmt.Printf("当前token:%v", tokenStr)

	log.Println(tokenStr)
	println("1")

	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

// Login 用户登录接口
func (c *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	// 当我们调用 Bind 方法的时候，如果有问题，Bind 方法已经直接写响应回去了
	if err := ctx.Bind(&req); err != nil {
		return
	}
	u, err := c.svc.Login(ctx.Request.Context(), req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或者密码不正确，请重试")
		return
	}
	sess := sessions.Default(ctx)
	sess.Set(userIdKey, u.Id)
	sess.Options(sessions.Options{
		// 60 秒过期
		MaxAge: 1200,
	})
	err = sess.Save()
	if err != nil {
		ctx.String(http.StatusOK, "服务器异常")
		return
	}
	ctx.String(http.StatusOK, "登录成功")
}

// Edit 用户编译信息
func (c *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		BirthDay string `json:"birth_day"`
		NickName string `json:"nick_name"`
		Profile  string `json:"profile"`
	}

	var req EditReq
	// 当我们调用 Bind 方法的时候，如果有问题，Bind 方法已经直接写响应回去了
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(UserClaims)
	// 输入的时间字符串
	dateString := "2023-08-09"

	// 定义时间字符串的格式
	layout := "2006-01-02"

	// 使用 time.Parse() 函数将字符串转换为 time.Time 类型
	_, err1 := time.Parse(layout, dateString)
	if err1 != nil {
		ctx.String(http.StatusOK, "用户生日输入异常,请重试")
		return
	}

	err := c.svc.Edit(ctx.Request.Context(), req.NickName, req.Profile, dateString, uc.Id)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或者密码不正确，请重试")
		return
	}

	ctx.String(http.StatusOK, "修改成功")
}

// ProfileJWT 用户详情, JWT 版本
func (c *UserHandler) ProfileJWT(ctx *gin.Context) {
	type Profile struct {
		Email    string
		Phone    string
		BirthDay string
		NickName string
		Profile  string
	}
	uc := ctx.MustGet("user").(UserClaims)
	u, err := c.svc.Profile(ctx, uc.Id)
	if err != nil {
		// 按照道理来说，这边 id 对应的数据肯定存在，所以要是没找到，
		// 那就说明是系统出了问题。
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Profile{
		Email:    u.Email,
		Phone:    u.Phone,
		BirthDay: u.BirthDay,
		NickName: u.NickName,
		Profile:  u.Profile,
	})
}

// Profile 用户详情
func (c *UserHandler) Profile(ctx *gin.Context) {
	type Profile struct {
		Email string
	}
	sess := sessions.Default(ctx)
	id := sess.Get(userIdKey).(int64)
	u, err := c.svc.Profile(ctx, id)
	if err != nil {
		// 按照道理来说，这边 id 对应的数据肯定存在，所以要是没找到，
		// 那就说明是系统出了问题。
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Profile{
		Email: u.Email,
	})
}
