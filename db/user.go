package db

import (
	"bytes"
	"fmt"
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

	INSERT_UPLOADS     = `insert into t_uploads (Algorithm_Name,Downloads,Algorithm_Version,createtime) values(?,?,?,?)`
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
}

func Createupload(c *gin.Context) {
	upload := Upload{}
	c.Bind(&upload)
	createtime := time.Now()
	_, _, err := dbPlus.Exec(INSERT_UPLOADS, upload.Algorithm_Name, upload.Downloads, upload.Algorithm_Version, createtime)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"ret":        0,
			"upload":     upload,
			"createtime": createtime,
		})
	} else {
		fmt.Fprintln(os.Stderr, err)
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
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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
	dockerfile := "File/Dockerfile"
	f, err := os.Create(dockerfile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create Dockerfile"})
		return
	}
	defer f.Close()
	_, err = f.WriteString("FROM " + system_Name + ":" + system_Version + "\n")
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
	cmd := exec.Command("docker", "build", "-t ", system_Name, ":", system_Version, ".") // 构建镜像

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": stderr.String(),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
			"data":  "docker" + "build" + "-t " + system_Name + ":" + system_Version + " .",
			"path":  "/" + filepath.Dir(dockerfile) + "/",
		})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Image build successfully!"})
	}
}

// 定义一个函数，接受一个docker镜像名称和一个安装路径作为参数，返回一个错误值
func pullAndDeploy(image string, path string) error {
	// 使用exec.Command函数创建一个命令对象，指定要执行的命令和参数
	cmd := exec.Command("docker", "pull", image)
	// 将命令的标准输出和标准错误连接到当前进程的标准输出和标准错误
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// 执行命令，并等待命令完成
	err := cmd.Run()
	if err != nil {
		// 如果命令执行出错，返回错误值
		return err
	}
	// 如果命令执行成功，创建一个新的命令对象，指定要执行的命令和参数
	// 添加-v参数，使用安装路径作为挂载卷的一部分
	cmd = exec.Command("docker", "run", "-d", "-v", path+":/app", image)
	// 将命令的标准输出和标准错误连接到当前进程的标准输出和标准错误
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// 执行命令，并等待命令完成
	err = cmd.Run()
	if err != nil {
		// 如果命令执行出错，返回错误值
		return err
	}
	// 如果命令执行成功，返回nil
	return nil
}

func Deploy(c *gin.Context) {
	// form, err := c.MultipartForm()
	// if err != nil {
	// 	c.String(http.StatusBadRequest, "get form err: %s", err.Error())
	// 	return
	// }

	// image := "crasl/image1:1.34"
	image := c.PostForm("imageName")
	// 从请求中获取安装路径
	path := c.PostForm("installPath")
	// 调用pullAndDeploy函数，传入docker镜像名称和安装路径
	err := pullAndDeploy(image, path)
	if err != nil {
		// 如果函数返回错误值，打印错误信息并退出程序
		fmt.Println("Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed."})
		return
	}
	// 如果函数返回nil，打印成功信息

	c.JSON(http.StatusOK, gin.H{"message": "Successfully pulled and deployed" + image + " to " + path})

}
