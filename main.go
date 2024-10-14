package main

import (
	"crypto/sha1"
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

var (
	//go:embed index.html
	index embed.FS
)

// 启动一个 http server , 监听 789 端口
func startHttpServer() {
	// 文件上传 router
	http.HandleFunc("/upload", uploadFileHandler)
	// 文件下载 router
	http.HandleFunc("/download/", downloadFileHandler)
	// 添加 index 目录
	http.Handle("/", http.FileServer(http.FS(index)))
	http.ListenAndServe(":789", nil)
}

// 上传文件 http handler
func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	println("upload file handler")
	defer r.Body.Close()
	// 解析表单
	err := r.ParseMultipartForm(1024 * 1024 * 1024 * 10)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	// 获取上传的文件名字
	file, header, err := r.FormFile("file")
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	// 创建文件
	f, err := os.Create("/tmp/" + header.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	var buffer = make([]byte, 1024*1024*4)
	// 创建 SHA1 哈希对象
	hasher := sha1.New()

	// 使用 MultiWriter 同时写入文件和计算 SHA1
	multiWriter := io.MultiWriter(f, hasher)
	if _, err := io.CopyBuffer(multiWriter, file, buffer); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取最终的 SHA1 哈希值
	sha1sum := fmt.Sprintf("%x", hasher.Sum(nil))
	// 返回 json ,包含文件下载地址和 sha1 值
	w.Write([]byte(`{"sha1":"` + sha1sum + `","url":"http://localhost:789/download/` + header.Filename + `"}`))
}

// 文件下载接口, 更具请求路径，返回 /tmp 目录下的文件
func downloadFileHandler(w http.ResponseWriter, r *http.Request) {
	println("download file handler")
	filename := r.URL.Path[len("/download/"):]
	http.ServeFile(w, r, "/tmp/"+filename)
}

func main() {
	// 启动 http server
	go startHttpServer()
	// 程序启动以后，浏览器打开 http://localhost:789
	// 调用 /usr/bin/open http://localhost:789
	go exec.Command("/usr/bin/open", "http://localhost:789").Start()
	println("服务器启动: http://localhost:789")

	select {}
}
