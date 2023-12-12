package db

import (
	"bytes"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gohutool/boot4go-docker-ui/model"
	. "github.com/gohutool/boot4go-util"
)

/**
* golang-sample源代码，版权归锦翰科技（深圳）有限公司所有。
* <p>
* 文件名称 : user.go
* 文件路径 :
* 作者 : DavidLiu
× Email: david.liu@ginghan.com
*
* 创建日期 : 2022/5/12 21:27
* 修改历史 : 1. [2022/5/12 21:27] 创建文件 by LongYong
*/

const (
	INSERT_USER           = `insert into t_user (userid, username, password, createtime) values(?,?,?,?)`
	SELECT_USER_BY_NAME   = `select * from t_user where userid=?`
	UPDATE_PWD_USER_BY_ID = `update t_user set password=? where userid=?`

	INSERT_REPOS       = `insert into t_repos (reposid, name, description, endpoint, username, password, createtime) values(?, ?,?,?,?,?,?)`
	SELECT_REPOS       = `select * from t_repos where reposid=?`
	SELECT_ALL_REPOS   = `select * from t_repos where 1=1 `
	UPDATE_REPOS_BY_ID = `update t_repos set password=?, username=?, endpoint=?, name=?, description=? where reposid=?`

	INSERT_UPLOADS     = `insert into t_uploads (Algorithm_Name,Downloads,Algorithm_Version,createtime,Author_Name,Introduction,Function,Space) values(?,?,?,?,?,?,?,?)`
	DELECT_UPLOADS     = `delete from t_uploads where id=?`
	SELECT_ALL_UPLOADS = `select * from t_uploads where 1=1`
)

func InitAdminUser() {
	c, err := dbPlus.QueryCount("select count(1) from t_user where username=?", "ginghan")
	if err != nil {
		panic(err)
	}

	if c == 0 {
		err = CreateUser("ginghan", "123456")

		if err != nil {
			panic(err)
		}
	}
}

func CreateUser(username, password string) error {
	userId := MD5(username)
	password = SaltMd5(password, userId)

	createtime := time.Now()

	_, _, err := dbPlus.Exec(INSERT_USER, userId, username, password, createtime)

	//stm, err := _db.Prepare(INSERT_USER)

	//stm.Exec(userId, username, password, createtime)

	//stm.Close()

	return err
}

func UpdatePwd(userid, passwd string) error {
	passwd = SaltMd5(passwd, userid)
	_, _, err := dbPlus.Exec(UPDATE_PWD_USER_BY_ID, userid, passwd)
	return err
}

func GetUser(userid string) *model.User {
	user, err := dbPlus.QueryOne(SELECT_USER_BY_NAME, userid)

	if err != nil || user == nil {
		return nil
	}

	rtn := &model.User{
		UserName:     GetMapValue2(user, "username", ""),
		UserID:       GetMapValue2(user, "userid", ""),
		UserPassword: GetMapValue2(user, "password", ""),

		//	CreateTime: *GetITime(GetMapValue2(user, "createtime", ""), "yyyy", nil),
	}
	//2022-05-13 13:54:40.5049073+08:00

	return rtn
}

type Upload struct {
	Algorithm_Name    string `form:"Algorithm_Name"`
	Downloads         string `form:"Downloads"`
	Algorithm_Version string `form:"Algorithm_Version"`
	Space             string `form:"Space"`
	Author_Name       string `form:"Author_Name"`
	Function          string `form:"Function"`
	Introduction      string `form:"Introduction"`
}

func Createupload(c *gin.Context) {
	upload := Upload{}
	c.Bind(&upload)
	createtime := time.Now()
	_, _, err := dbPlus.Exec(INSERT_UPLOADS, upload.Algorithm_Name, upload.Downloads,
		upload.Algorithm_Version, createtime, upload.Author_Name, upload.Introduction, upload.Function, upload.Space)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"ret":        0,
			"upload":     upload,
			"createtime": createtime,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"ret":  1,
			"data": "",
		})
	}
}
func Deleteuploads(c *gin.Context) {
	id := c.PostForm("id")
	_, _, err := dbPlus.Exec(DELECT_UPLOADS, id)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"ret":  0,
			"data": id,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"ret":  1,
			"data": id,
		})
	}
}
func UploadFile(c *gin.Context) {
	form, err := c.MultipartForm()
	files := form.File["files"]
	if err != nil {
		c.String(http.StatusBadRequest, "get form err: %s", err.Error())
		return
	}
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		if err := c.SaveUploadedFile(file, "File/"+filename); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"ret": 1,
				"msg": "upload file err: " + err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"ret": 0,
		"msg": "Uploaded successfully " + strconv.Itoa(len(files)) + " files",
	})
}
func BulidImage(c *gin.Context) {
	form, err := c.MultipartForm()
	//选择哪个仓库地址
	repositoryId := c.PostForm("repositoryId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//由操作系统/语言环境去选择基础镜像 默认tag latest
	basicImageName := c.PostForm("basicImageName")
	files := form.File["files"]
	system_Name := c.PostForm("system_Name")
	system_Version := c.PostForm("system_Version")
	//Language := c.PostForm("Language")
	//Language_Version := c.PostForm("Language_Version")
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		if err := c.SaveUploadedFile(file, "File/"+filename); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"ret": 0,
				"msg": "upload file err: " + err.Error(),
			})
		}
	}
	currentdir, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get currentdir"})
		return
	}
	dockerfile := "File/Dockerfile"
	f, err := os.Create(dockerfile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create Dockerfile"})
		return
	}
	defer f.Close()
	_, err = f.WriteString("FROM " + basicImageName + "\n")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to write to Dockerfile"})
		return
	}
	_, err = f.WriteString("ADD . /app/File\n")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to write to Dockerfile"})
		return
	}
	_, err = f.WriteString("WORKDIR /app\n")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to write to Dockerfile"})
		return
	}
	_, err = f.WriteString("EXPOSE 8080\n")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to write to Dockerfile"})
		return
	}
	_, err = f.WriteString("ENTRYPOINT " + "[" + "./main" + "]\n")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to write to Dockerfile"})
		return
	}
	err = os.Chdir(filepath.Dir(dockerfile)) // 切换到Dockerfile所在目录
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to change directory"})
		return
	}
	// "-t" , system_Name, ":", system_Version, " ",
	//cmd := exec.Command("docker", "build", ".") // 构建镜像
	ImageNameTag := c.PostForm("ImageNameTag")
	cmd := exec.Command("docker", "build", "-t", ImageNameTag, ".") // 构建镜像
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	err1 := os.Chdir(currentdir)
	if err1 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to change directory"})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": stderr.String(),
		})
		return
	}
	os.Chdir("..") //回到有docker环境的目录
	if repositoryId == "default" {
		cmd = exec.Command("docker", "login") //登陆一次在json里自动保存账户密码
		err = cmd.Run()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"loginerror": err,
			})
			return
		}
		//自定义仓库名
		Repositorybase := "crasl/"
		PushImageName := Repositorybase + system_Name + ":" + system_Version
		//打包镜像
		Tagcmd := exec.Command("docker", "tag", ImageNameTag, PushImageName)
		err1 = Tagcmd.Run()
		if err1 != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Tagerror": err,
			})
			return
		}
		//Push
		Pushcmd := exec.Command("docker", "push", PushImageName)
		err2 := Pushcmd.Run()
		if err2 != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Pusherror": err,
			})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"ret":  0,
				"data": "OK",
			})
		}
	} else if repositoryId == "registry.cn-hangzhou.aliyuncs.com" {
		loginCmd := exec.Command("docker", "login", "-u", "aicrazy", "-p", "Zufe1234!", "registry.cn-hangzhou.aliyuncs.com")
		err := loginCmd.Run()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err,
			})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"msg": "Login Success!",
			})
		}
		Repositorybase := "registry.cn-hangzhou.aliyuncs.com/zufe_123/docker_image:"
		PushImageName := Repositorybase + "dockerimage"
		Tagcmd := exec.Command("docker", "tag", ImageNameTag, PushImageName)
		err1 = Tagcmd.Run()
		if err1 != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err,
			})
			return
		}
		Pushcmd := exec.Command("docker", "push", PushImageName)
		err2 := Pushcmd.Run()
		if err2 != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err,
			})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"msg": "Push Success!",
			})
		}
	}
}
func Dockerpush(c *gin.Context) {
	os.Chdir("..")                         //回到有docker环境的目录
	cmd := exec.Command("docker", "login") //登陆一次在json里自动保存账户密码
	err := cmd.Run()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Login Success!",
		})
	}
	//自定义仓库名
	Repositorybase := "crasl/"
	ImageNameTag := "mysql" + ":" + "latest"
	PushImageName := Repositorybase + "Algorithm_Name" + ":" + "1.01"
	//打包镜像
	Tagcmd := exec.Command("docker", "tag", ImageNameTag, PushImageName)
	err1 := Tagcmd.Run()
	if err1 != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Tag Success!",
		})
	}
	//Push
	Pushcmd := exec.Command("docker", "push", PushImageName)
	err2 := Pushcmd.Run()
	if err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Push Success!",
		})
	}
}
func DockerLogin(c *gin.Context) {
	loginCmd := exec.Command("docker", "login", "-u", "aicrazy", "-p", "Zufe1234!", "registry.cn-hangzhou.aliyuncs.com")
	loginCmd.Stdin = os.Stdin
	loginCmd.Stdout = os.Stdout
	loginCmd.Stderr = os.Stderr
	err := loginCmd.Run()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Login Success!",
		})
	}
	Pushcmd := exec.Command("docker", "push", "registry.cn-hangzhou.aliyuncs.com/zufe_123/docker_image:mysql")
	err2 := Pushcmd.Run()
	if err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Push Success!",
		})
	}
}
