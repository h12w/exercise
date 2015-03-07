qsys使用和设计说明
==================

简介
----

qsys是对游戏排队系统的一个练习。

1. 当登陆用户的数量超过服务器上限时，服务器将用户按照登陆的先后顺序加入到等待队列
   中, 并通过等待页面实时通知用户前面等待的人数
2. 当有一个玩家退出游戏时，自动让排在等待队列中最前排的用户进入游戏
3. 处理排队用户提前退出的情况
4. 自动/压力测试命令行客户端

使用说明
--------

### 安装

1. 从源代码编译安装，需要Go语言开发环境，请参考：[Go Gettting
Started](http://golang.org/doc/install).

2. go get
```
go get h12.me/exercise/qsys
```

### 命令行参数

运行 `qsys -h` 获取关于命令行参数的帮助。简要说明如下：

* cap: 服务器玩家上限
* port: web服务器端口
* tdir: 网页模板目录，模板格式请参考[html/template](http://golang.org/pkg/html/template/)
* dbtype: 用户账户数据库类型，可以是mem (内存hash map）, file (单一gob文件），
  mongo (MongoDB)或SQL驱动名称，如postgres, mysql等
* dbsrc: 账户数据文件路径或数据库连接字符串, 格式请参考
  [httpauth](https://github.com/hailiang/httpauth)
* ckey: cookie加密密钥
* gen: 启动前自动生成测试用账户的数量，用户名为：u0, u1, u2, u3 ......
  密码和用户名相同

### 服务URL

* GET /login: 登陆页面
* POST /login: 登陆表单提交
* POST /register: 注册用户表单提交
* /logout: 用户离开
* /wait-num: 通知用户前方等待人数的websocket路径
* /: 游戏主页面或排队等待页面
* /js, /css: 静态资源


自动/压力测试工具
-----------------

测试工具源码在目录qsys/tester下。使用前注意重启服务端，并通过命令行参数gen生成需
要的测试账户信息。

### 命令行参数

* url: 服务器地址
* auto: 是否自动检查测试结果(true/false), 自动检查只能单机运行。多机压力测试需要
  设置为false
* cap: 服务器玩家上限，必须和服务端保持一致
* cnt: 登陆游戏的玩家数量

设计说明
--------

### 使用的非标准库

* Mux和session管理: 轻量级的[Gorilla Toolkit](http://www.gorillatoolkit.org)
* 用户登陆: 自己修改过的[httpauth](https://github.com/hailiang/httpauth)
* Websocket: 准标准库[x/net/websocket](https://godoc.org/golang.org/x/net/websocket)
* HTML页面解析：自己写的[html-query](https://github.com/hailiang/html-query)

### 登陆系统 (qsys/login.go)

用户登陆采用常见的cookie session登陆。提供了最基本的用户注册、登入、登出的功能。

### 用户队列 (qsys/pool.go)

PlayerPool和UserQueue实例化为两个全局变量，分别用来管理在线玩家和等待用户。为了
维护一致性，各自有自己的Mutex锁。

用户登陆后，在服务端将产生一个保存关联信息的结构（User)。用户如果服务器未到上限，
将User加入到PlayerPool中，否则加入到UserQueue中。如果用户离开，则从PlayerPool或
UserQueue中移除。

队列移动时的通知采用channel的方式。每当一个用户浏览器端的javascript脚本用
websocket连接服务器时，在服务器端 (func serveWaitNum in sys/game.go) 就会生成一个
channel，并将此channel注册到UserQueue, 之后UserQueue有任何变化，就可以通过channel ->
websocket的方式通知客户端浏览器。需要注意的是在该用户移出UserQueue时必须关闭对应
的channel, 否则无法释放websocket的goroutine.

### websocket (qsys/template/wait.html, qsys/game.go)

需要注意的是websocket也一样需要通过cookie session登陆, 浏览器会自动处理，测试
工具需要手工把cookie配置进去(qsys/tester/server.go)。

### 测试工具

Server (qsys/tester/server.go) 提供了表单提交和连接websocket的功能。
User (qsys/tester/user.go) 模拟登入，登出，注册和获取前方等待人数, 每个操作都会
用html-query解析返回的网页，保存在GamePage结构中。Websocket同样在客户端也连接了
一个channel, 用来读取等待人数。

测试脚本(qsys/tester/test.go)先登入一定数量（超过玩家上限）的用户，然后再依次
登出。重复两遍来确保登出操作在服务器正确执行。在自动模式下，会检查每一步返回
的在线玩家数量或等待数量是否正确。

并发测试和分析
--------------

由于每增加一个用户只增加了非常小的一个数据结构（User)，因此内存不是排队服务的
瓶颈。由于排队每增加一个用户，都会增加一个websocket连接，如果未经优化，
TCP连接数的限制将会是系统的瓶颈。

实际测试中，默认系统 `ulimit -n` 数量是1024, 并发连接也被限制在1000多一点。通过
初步调整把ulimit调整到10000，并发等待用户就能接近10000.

进一步的优化需要采用一些成熟的办法调整Linux系统的并发连接数。具体做法搜索关键词
`c1000k` (百万连接数优化), 由于时间有限，没有再进一步优化和测试。

