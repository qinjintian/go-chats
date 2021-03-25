package controller

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/astaxie/beego/validation"
	"go-chats/app/global/variable"
	"go-chats/app/model"
	"go-chats/app/utils/helper"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"time"
)

type PublicController struct{}

func (p *PublicController) Login(c *gin.Context) {

	if c.Request.Method == "POST" {
		// 接收参数
		username := c.DefaultPostForm("username", "")
		password := c.DefaultPostForm("password", "")

		// 校验参数
		validate := validation.Validation{}
		validate.Required(username, "username").Message("用户名不能为空，请检查")
		validate.Required(password, "password").Message("密码不能为空，请检查")
		validate.MinSize(password, 6, "password").Message("密码不能少于6位数，请检查")
		if validate.HasErrors() {
			for _, err := range validate.Errors {
				c.JSON(http.StatusOK, gin.H{"code": 0, "message": err.Error()})
				return
			}
		}

		// 数据查询
		user := model.User{}
		if err := model.DB.Where("`username` = ?", username).First(&user).Error; err != nil {
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

		if user.Password != helper.Md5(password) {
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "密码不正确，请检查。"})
			return
		}

		data := make(map[string]interface{})
		data["id"] = user.Id
		data["username"] = user.Username
		data["nickname"] = user.Nickname
		data["email"] = user.Email
		data["jump"] = fmt.Sprintf("index?t=%d", time.Now().UnixNano())

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
	c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("login?t=%d", time.Now().UnixNano()))
}

func (p *PublicController) Register(c *gin.Context) {
	if c.Request.Method == "POST" {
		// 接收参数
		username := c.DefaultPostForm("username", "")
		password := c.DefaultPostForm("password", "")
		nickname := c.DefaultPostForm("nickname", "")
		email := c.DefaultPostForm("email", "")

		// 校验参数
		validate := validation.Validation{}
		validate.Required(username, "username").Message("请输出用户名")
		validate.Required(password, "password").Message("请输入密码")
		validate.Required(nickname, "nickname").Message("请输入昵称")
		validate.Email(email, "email").Message("邮箱地址不正确")
		if validate.HasErrors() {
			for _, err := range validate.Errors {
				c.JSON(http.StatusOK, gin.H{"code": 0, "message": err.Error()})
				return
			}
		}

		confirmPassword := c.DefaultPostForm("confirm_password", "")
		if password != confirmPassword {
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "两次密码输入不一致"})
			return
		}

		var userCount int64 = 0
		if err := model.DB.Table((&model.User{}).TableName()).Where("`username` = ?", username).Count(&userCount).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询用户失败"})
			return
		}

		if userCount > 0 {
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "该账号已经存在，请更换"})
			return
		}

		user := model.User{
			Username: username,
			Password: helper.Md5(password),
			Nickname: nickname,
			Email: email,
			Activate: 1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := model.DB.Table((&model.User{}).TableName()).Create(&user).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "注册失败，请稍后再试"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    1,
			"message": "注册成功",
			"data":    map[string]string{"jump": fmt.Sprintf("login?t=%d", time.Now().UnixNano())},
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
	ConsulRegister() // 注册服务到consul
	// ConsulDeRegister() // 取消consul注册的服务
	// ConsulFindServer() // 从consul中发现服务
	// ConsulKVTest() // KV测试
}

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s\n", msg, err)
		os.Exit(1)
	}
}

// 注册服务到consul
func ConsulRegister() {
	// 创建连接consul服务配置
	config := consulapi.DefaultConfig()
	config.Address = "127.0.0.1:8500"
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("consul client error : ", err)
	}

	// 创建注册到consul的服务到
	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = "grpc"
	registration.Name = "go-consul-test"
	registration.Tags = []string{"go-consul-test"}
	registration.Port = 10086
	registration.Address = "172.18.0.16"

	// 增加consul健康检查回调函数
	check := new(consulapi.AgentServiceCheck)
	check.Timeout = "5s"
	check.Interval = "5s"
	check.GRPC = fmt.Sprintf("%v:%v/%v", registration.Address, registration.Port, "aaaa")
	check.DeregisterCriticalServiceAfter = "30s" // 故障检查失败30s后 consul自动将注册服务删除
	registration.Check = check

	// 注册服务到consul
	if err := client.Agent().ServiceRegister(registration); err != nil {
		log.Fatal("注册服务到consul失败: ", err)
	}

	fmt.Printf("注册服务到consul成功，服务ID：%s，服务名：%s", registration.ID, registration.Name)
}

// 取消consul注册的服务
func ConsulDeRegister() {
	// 创建连接consul服务配置
	config := consulapi.DefaultConfig()
	config.Address = "127.0.0.1:8500"
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("consul client error : ", err)
	}

	serverID := "grpc"
	_ = client.Agent().ServiceDeregister(serverID)

	fmt.Printf("取消consul某个注册的服务成功，服务ID：%s", serverID)
}

// 从consul中发现服务
func ConsulFindServer() {
	// 创建连接consul服务配置
	config := consulapi.DefaultConfig()
	config.Address = "127.0.0.1:8500"
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("consul client error : ", err)
	}

	fmt.Println("================ 获取所有service =================")

	// 获取所有service
	services, _ := client.Agent().Services()
	for _, value := range services {
		fmt.Println("服务 ID：", value.ID)
		fmt.Println("服务名称：", value.Service)
		fmt.Println("服务地址：", value.Address)
		fmt.Println("服务端口：", value.Port)
		fmt.Println()
	}

	fmt.Println("================ 获取指定service =================")

	// 获取指定service
	service, _, err := client.Agent().Service("CalculatorService-0", nil)
	if err == nil {
		fmt.Println("服务地址：", service.Address)
		fmt.Println("服务端口：", service.Port)
	}
}

func ConsulCheckHeath() {
	// 创建连接consul服务配置
	config := consulapi.DefaultConfig()
	config.Address = "127.0.0.1:8500"
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("consul client error : ", err)
	}

	// 健康检查
	a, b, _ := client.Agent().AgentHealthServiceByID("CalculatorService-0")
	fmt.Println(a)
	fmt.Println(b.Service)
	fmt.Println(b.AggregatedStatus)
	fmt.Println(b.Checks)
}

func ConsulKVTest() {
	// 创建连接consul服务配置
	config := consulapi.DefaultConfig()
	config.Address = "127.0.0.1:8500"
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("consul client error : ", err)
	}

	// KV, put值
	values := "grpc"
	key := "grpc/172.18.0.16:8100"
	_, _ = client.KV().Put(&consulapi.KVPair{Key: key, Flags: 0, Value: []byte(values)}, nil)

	// KV get值
	data, _, _ := client.KV().Get(key, nil)
	fmt.Println(string(data.Value))

	// KV list
	datas, _, _ := client.KV().List("g", nil)
	for _, value := range datas {
		fmt.Println(value)
	}
	keys, _, _ := client.KV().Keys("go", "", nil)
	fmt.Println(keys)
}
