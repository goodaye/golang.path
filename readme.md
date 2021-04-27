## Golang Path
golang常用的工具包集合体，因为很多包（特别是 `golang.org`仓库）在国内访问延迟很大，甚至有些无法访问。 所以下载到本地后，保存共享。方便其他人一起使用。 

## 1. 用法

1. 安装`golang` 开发环境

下载二进制安装包并安装，不同操作系统的安装步骤、安装方式略有不同, 自行完成。

一般安装完`golang` 环境后，会自动设置一个 `$GOPATH=/home/${user}/go`环境变量和相应的目录。 可以通过`go env ` 查看默认的 `GOPATH` 变量。 


2. 安装工具包

clone 本仓库后，将该包内容 复制到 `$GOPATH/src`目录下。 

`/bin/cp -rf golang.path/* /home/${user}/go/src/`


3. 编译常用工具

```
    go install  github.com/mdempsky/gocode 
    go install  github.com/uudashr/gopkgs/cmd/gopkgs 
    go install  github.com/ramya-rao-a/go-outline 
    go install  github.com/acroca/go-symbols 
    go install  golang.org/x/tools/cmd/guru 
    go install  golang.org/x/tools/cmd/gorename 
    go install  github.com/go-delve/delve/cmd/dlv 
    go install  github.com/rogpeppe/godef 
    go install  github.com/sqs/goreturns 
    go install  github.com/kardianos/govendor
    go install github.com/fatih/gomodifytags
    go install github.com/josharian/impl
    go install honnef.co/go/tools/cmd/staticcheck
    go install golang.org/x/tools/gopls

```

## 2. 更新仓库内容

用户本地有了更多的工具包的时候，可以更新本仓库，提供给其他人使用。 

1. 用户本地下来工具。 

    `go get pathtoway/pathy`
    
会下载工具包到 用户自己的 `$GOPATH/src`目录下


2. 复制内容到 `golang.path` 仓库

    `/bin/cp -rf /home/${user}/go/src/pathtoway golang.path/`

3. 删除git 标记。(不然无法提交)

    `find . -name ".git*" |grep -v "^./.git$" |xargs rm -rf`

4. 提交 代码到仓库

    ```
    git add -A
    git commit -m "xxx"
    git push 
    ```