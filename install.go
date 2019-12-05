package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)


var engine *gin.Engine
var dow_update_flag = false
var dow_update_msg string
func init() {
	engine = gin.Default()
	engine.Use(cors())
	t, err := loadTemplate()
	if err != nil {
		panic(err)
	}
	engine.SetHTMLTemplate(t)
}
func installationGuide() error {
	engine.GET("/",index)
	engine.GET("/authorize",authorize)
	engine.GET("/install_service",install_service)
	engine.GET("/uninstall_service",uninstall_service)
	engine.GET("/checkForUpdates",checkForUpdates)

	err := engine.Run(":50000")
	return err
}

func index(c *gin.Context)  {
	status, i := Skynet_service.Status()
	user_conf := UserConf()
	Info.Println("status",status)
	if len(user_conf) > 0  {
		user_conf["code"] = 200
	}else {
		user_conf["code"] = 500
	}
	user_conf["version"] = Version
	user_conf["msg"] = ""
	user_conf["status"] = status
	if i != nil{
		err := i.Error()
		if strings.Contains(err,"Access is denied."){//win
			user_conf["msg"] = "请使用管理员用户启动!"
		}else if strings.Contains(err,"permission denied") {
			user_conf["msg"] = "请使用管理员用户启动!"
		}
		Error.Println(i)
	}
	user_conf["version"] = Version
	c.HTML(http.StatusOK,"/html/index.html",user_conf)
}

func checkForUpdates(c *gin.Context)  {
	if dow_update_flag {
		c.JSON(http.StatusOK,gin.H{
			"code":200,
			"msg":"正在更新，请勿重复点击",
		})
	}else {
		resp, err := http.Get("https://skynet-beijing.oss-cn-beijing.aliyuncs.com/skynet.client.version/skynet.client.json")
		if err != nil {
			c.JSON(http.StatusOK,gin.H{
				"code":500,
				"msg":"网络异常,请稍后再试!",
			})
		}else {
			defer resp.Body.Close()
			readAll, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				c.JSON(http.StatusOK,gin.H{
					"code":500,
					"msg":"数据解析异常,请稍后再试!",
				})
			}else {
				versionInfo := make(map[string]interface{})
				json.Unmarshal(readAll,&versionInfo)
				if len(versionInfo) > 0 {
					version := versionInfo["version"].(string)
					version = version[1:]
					version = strings.ReplaceAll(version,".","")
					local_version := Version
					local_version = local_version[1:]
					local_version = strings.ReplaceAll(local_version,".","")
					var version_,_ =  strconv.Atoi(version)
					var local_version_,_ =  strconv.Atoi(local_version)
					if version_ > local_version_{
						c.JSON(http.StatusOK,gin.H{
							"code":200,
							"msg":"发现新版本,正在后台更新,请稍后",
						})
						go func() {
							dow_update_flag = true
							dow_update_msg = "正在努力更新中,请耐心等待!"
							dowloadExe := dowload_exe(versionInfo["downlodUrl"].(string))
							if dowloadExe != nil {
								if strings.Contains(dowloadExe.Error(),"The service has not been started") {
									dow_update_msg = "请重启服务器"
								}else {
									dow_update_msg = "更新失败-msg:"+dowloadExe.Error()
								}
								Error.Println("更新失败!",dowloadExe)
							}
							dow_update_flag = false
						}()
					}else {
						c.JSON(http.StatusOK,gin.H{
							"code":200,
							"msg":"已经是最新版本",
						})
					}
				}else {
					c.JSON(http.StatusOK,gin.H{
						"code":200,
						"msg":"已经是最新版本",
					})
				}
			}
		}
	}
}

func dowload_exe(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	readAll, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	oldFileName := ExeHomeDir
	exeFileName := ExeHomeDir
	newFileName := ExeHomeDir
	goos := runtime.GOOS
	if strings.Contains(goos, "win") {
		oldFileName += "/old_skynet.exe"
		exeFileName += "/skynet.exe"
		newFileName += "/new_skynet.exe"
	}else {
		oldFileName += "/old_skynet"
		exeFileName += "/skynet"
		newFileName += "/new_skynet"
	}
	err = ioutil.WriteFile(newFileName, readAll, os.ModePerm)
	if err != nil {
		return err
	}

	old_err := os.Rename(exeFileName, oldFileName)
	if old_err != nil {
		return old_err
	}
	new_err := os.Rename(newFileName, exeFileName)
	if new_err != nil {
		return new_err
	}
	restart_err := Skynet_service.Restart()
	if restart_err != nil {
		return restart_err
	}
	return nil
}


func uninstall_service(c *gin.Context)  {
	r := make(map[string]interface{})
	r["code"] = 200
	r["msg"] = ""
	uninstall := Skynet_service.Uninstall()
	if uninstall != nil {
		r["code"] = 500
		r["msg"] = uninstall.Error()
	}else {
		r["code"] = 200
		r["msg"] = "已经卸载,重启后生效"
		Info.Println("已经卸载")
	}
	c.JSON(http.StatusOK,r)
}
func install_service(c *gin.Context)  {
	r := make(map[string]interface{})
	r["code"] = 200
	r["msg"] = ""
	install := Skynet_service.Install()
	if install != nil {
		r["code"] = 500
		r["msg"] = install.Error()
	}else {
		userConf := UserConf()
		if len(userConf) > 0 {
			port, _ := strconv.Atoi(userConf["port"].(string))
			start(userConf["key"].(string), userConf["public_ip"].(string), port, nil)
			r["code"] = 200
			r["msg"] = "已经启动"
			Info.Println("已经启动")
		}else {
			r["code"] = 500
			r["msg"] = "未授权"
			Info.Println("未授权")
		}
	}
	c.JSON(http.StatusOK,r)
}
func authorize(c *gin.Context)  {
	accountNumber := c.Query("accountNumber")
	password := c.Query("password")
	if password == "" || accountNumber == "" {
		c.JSON(http.StatusOK,gin.H{
			"code":500,
			"msg":"账号或密码不能为空",
		})
	}else {
		response, err := http.Get("https://api.tbytm.com/open/ip/login?ip=" + accountNumber + "&phone=" + password)
		defer response.Body.Close()
		if err != nil {
			c.JSON(http.StatusOK,gin.H{
				"code":500,
				"msg":"网络请求错误",
			})
		}else {
			bytes, _ := ioutil.ReadAll(response.Body)
			h := gin.H{}
			json.Unmarshal(bytes,&h)
			if h["code"].(float64) != 0 {
				c.JSON(http.StatusOK,h)
			}else {
				user_conf := make(map[string]interface{})
				user_conf["key"] = password
				user_conf["port"] = "4900"
				user_conf["public_ip"] = accountNumber
				user_conf_bytes, err := json.Marshal(user_conf)
				if err != nil {
					c.JSON(http.StatusOK,gin.H{
						"code":500,
						"msg":"解析授权文件失败",
					})
				}else {
					writeFile := ioutil.WriteFile(SkynetPath+"/user.conf", user_conf_bytes, os.ModeAppend)
					if writeFile != nil {
						err := writeFile.Error()
						if strings.Contains(err,"Access is denied."){//win
							err = "没有获得管理员权限无法安装!"
						}else if strings.Contains(err,"permission denied") {
							err = "没有获得管理员权限无法安装!"
						}
						c.JSON(http.StatusOK,gin.H{
							"code":500,
							"msg":"授权失败,msg:"+err,
						})
					}else {
						c.JSON(http.StatusOK,gin.H{
							"code":200,
							"msg":"授权成功",
						})
					}
				}

			}
		}
	}
}

func openFile(fileName string) *os.File {
	openFile, e := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0666)
	if e != nil {
		fmt.Println(e)
		os.Exit(0)
	}
	return openFile
}


func loadTemplate() (*template.Template, error) {
	t := template.New("")
	for name, file := range Assets.Files {
		if file.IsDir() || !strings.HasSuffix(name, ".html") {
			continue
		}
		h, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}
		t, err = t.New(name).Parse(string(h))
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		origin := c.Request.Header.Get("Origin")
		var headerKeys []string
		for k, _ := range c.Request.Header {
			headerKeys = append(headerKeys, k)
		}
		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		} else {
			headerStr = "access-control-allow-origin, access-control-allow-headers"
		}
		if origin != "" {
			//下面的都是乱添加的-_-~
			// c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Headers", "*")
			c.Header("Access-Control-Allow-Methods", "*")
			//c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			c.Header("Access-Control-Expose-Headers", "*")
			// c.Header("Access-Control-Max-Age", "172800")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Set("content-type", "application/json")
		}

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}

		c.Next()
	}
}
