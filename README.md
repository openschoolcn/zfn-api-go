# zfn-api-go

新正方教务系统 API 接口 For Go

完善中，相关介绍请看 Python 版：
[https://github.com/openschoolcn/zfn_api](https://github.com/openschoolcn/zfn_api)

## CLI

太过无聊所以写了命令行的 CLI 程序，可以玩玩或当作 API 使用的 demo

### 编译

```bash
make init && make build
```

### 运行

```bash
# 设置教务系统URL
zfn-cli config --base_url https://xxx.com/

# 登录
zfn-cli login -u {sid} -p {password}

# 获取学生信息
zfn-cli info

# 获取成绩
zfn-cli grade --year 2022   # 2022学年整年的成绩
zfn-cli grade --year 2022 --term 1 # 2022学年第1学期的成绩
```
