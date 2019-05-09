# leetcode-spider-go 

使用 Go 编写的 leetcode 解题源码爬虫.爬取你自己的 leetcode 解题源码.

如果你也想把你在 [leetcode](https://leetcode.com/) 上提交且 accepted 的解题代码爬下来,那么本工具就是为此需求而生!爬下来的代码可以放在 github 上管理和开源出来,可以作为个人展示

## Installation

```
go get github.com/eindex/leetcode-spider-go
```

或者下载对应系统的 release 版本。

## Usage

使用前请拷贝 `settings.default.toml` 为 `setting.toml`

```
enter="golbal" # golbal/cn
username="username"
password="password"
savefile="leetcode.data.json"
out="out path"
```

- `username` 和 `password` 对应你的的 leetcode 账户.
- `enter` 对应登陆渠道 `cn` 为中国版
- `out` 表示你希望存放代码文件的目录
- `savefile` 是你 leetcode 爬去的结果，能降低下一次爬取时对服务器反复请求造成的浪费。



## Execution

```
lc-spider // 默认使用config.json为配置文件运行爬虫
```
**ac 过的题目不会再次爬取**

**这也意味着,当你在进行增量爬取时,根本不需要去指定要爬哪些题目, leetcode-spider 会自动知道哪些题目需要爬.**

举个例子,按照我们的日常使用:

* 当你昨天 A 了5道题,你用爬虫爬了下来
* 然后今天你 A 了6道题,你今天再次运行程序

程序此时会自动检查你跟上一次爬取结果相比多写了哪些题,然后把这些新增的代码爬取下来,不会重复的去爬取,你也不用手工指定你今天 AC 了哪些题.

永远只用一行命令: `lc-tool`.

此外源码对应的 leetcode 的题目,也会爬取下来,放在代码目录, markdown 格式.

爬取完成后会自动生成 README.md 文件,当你把爬下来的代码放在 github 上时,README.md 起一个介绍和导航的作用.

## 特别感谢

- [leetcode-spider](https://github.com/Ma63d/leetcode-spider)
