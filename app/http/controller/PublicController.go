package controller

import (
	"crypto/md5"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
	"go-chats/app/global/variable"
	"go-chats/app/model"
	"go-chats/app/utils/helper"
	"gorm.io/gorm"
	"net/http"
	"reflect"
	"time"
)

type PublicController struct{}

func (p *PublicController) Login(c *gin.Context) {
	if c.Request.Method == "POST" {
		type Params struct {
			Username string `form:"username" json:"username" validate:"required" label:"用户名"`
			Password string `form:"password" json:"password" validate:"required,min=6,max=20" label:"密码"`
		}

		params := &Params{
			Username: c.PostForm("username"),
			Password: c.PostForm("password"),
		}

		validate := validator.New()

		// 注册一个函数，获取struct tag里自定义的label作为字段名
		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			label := fld.Tag.Get("label")
			return label
		})

		trans, _ := ut.New(zh.New()).GetTranslator("zh")

		// 注册翻译器
		if err := zhTranslations.RegisterDefaultTranslations(validate, trans); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": err.Error()})
			return
		}

		if errs := validate.Struct(params); errs != nil {
			for _, err := range errs.(validator.ValidationErrors) {
				c.JSON(http.StatusOK, gin.H{"code": 0, "message": err.Translate(trans)})
				return
			}
		}

		// 查询数据库验证账号密码
		user := model.User{}
		if err := model.DB.Where("`username` = ?", params.Username).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusOK, gin.H{"code": 0, "message": "该用户不存在，请检查。"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": fmt.Sprintf("查询失败： %s", err.Error())})
			return
		}

		if user.Activate != 1 {
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "该用户账号已被禁用。"})
			return
		}

		if user.Password != helper.Md5(params.Password) {
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "密码不正确，请检查。"})
			return
		}

		data := make(map[string]interface{})
		data["id"] = user.Id
		data["username"] = user.Username
		data["nickname"] = user.Nickname
		data["email"] = user.Email
		data["jump"] = fmt.Sprintf("index?rand=%d", time.Now().UnixNano())

		session := sessions.Default(c)
		session.Set("user", variable.UserSessionData{Id: user.Id, Username: user.Username, Nickname: user.Nickname, Email: user.Email})
		_ = session.Save()

		// 返回结果
		c.JSON(http.StatusOK, gin.H{
			"code":    1,
			"message": "登录成功~(￣▽￣)／",
			"data":    data,
		})
	} else {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "登录",
		})
	}
}

func (p *PublicController) Logout(c *gin.Context) {
	sessions.Default(c).Clear()
	c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("login?rand=%d", time.Now().UnixNano()))
}

func (p *PublicController) Register(c *gin.Context) {
	if c.Request.Method == "POST" {
		type Params struct {
			Username  string `form:"username" json:"username" validate:"required" label:"用户名"`
			Password  string `form:"password" json:"password" validate:"required,min=6,max=20" label:"密码"`
			Nickname  string `form:"nickname" json:"nickname" validate:"required,max=15" label:"昵称"`
			Email     string `form:"email" json:"email" validate:"required,email" label:"邮箱地址"`
			Activate  uint8
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		}

		password := c.PostForm("password")
		confirmPassword := c.DefaultPostForm("confirm_password", "")
		if password != confirmPassword {
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "两次密码输入不一致"})
			return
		}

		params := &Params{
			Username:  c.PostForm("username"),
			Password:  password,
			Nickname:  c.PostForm("nickname"),
			Email:     c.PostForm("email"),
			Activate:  1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		validate := validator.New()

		// 注册一个函数，获取struct tag里自定义的label作为字段名
		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			label := fld.Tag.Get("label")
			return label
		})

		trans, _ := ut.New(zh.New()).GetTranslator("zh")

		// 注册翻译器
		if err := zhTranslations.RegisterDefaultTranslations(validate, trans); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": err.Error()})
			return
		}

		if errs := validate.Struct(params); errs != nil {
			for _, err := range errs.(validator.ValidationErrors) {
				c.JSON(http.StatusOK, gin.H{"code": 0, "message": err.Translate(trans)})
				return
			}
		}

		var userCount int64 = 0
		if err := model.DB.Table((&model.User{}).TableName()).Where("`username` = ?", params.Username).Count(&userCount).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询用户失败"})
			return
		}

		if userCount > 0 {
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "该账号已经存在，请更换"})
			return
		}

		params.Password = helper.Md5(params.Password)
		if err := model.DB.Table((&model.User{}).TableName()).Create(params).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "注册失败，请稍后再试"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    1,
			"message": "注册成功",
			"data":    map[string]string{"jump": fmt.Sprintf("login?rand=%d", time.Now().UnixNano())},
		})
	} else {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"title": "创建帐号",
		})
	}
}

func (p *PublicController) ResetPassword(c *gin.Context) {
	c.HTML(http.StatusOK, "reset-password.html", gin.H{
		"title": "找回密码页",
	})
}

func (p *PublicController) Test(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": string(fmt.Sprintf("%x", md5.Sum([]byte("hello")))),
	})
}
