# 一个课程大作业而已
# 构建
```
docker-compose -p blockchain up --build -d
```
	我是将它命名为blockchain，实际如下即可
```
docker-compose up --build -d
```
# 效果图
代码有点问题，主要是在登录/注册页部分，已注册或者密码错误这样的提示信息不会显示。。。
## 主页（登陆前）
![[Pasted image 20231212100702.png]]
## 注册/登录页
咳咳，这是一道[题目所用的网页](https://www.qsnctf.com/challenges#%E7%99%BB%E5%BD%95%E5%90%8E%E5%8F%B0-142)，然后我直接拔下来了。。。。
![[Pasted image 20231212100727.png]]
![[Pasted image 20231212100803.png]]
## 主页（登录后）
![[Pasted image 20231212101057.png]]
## 写日记
提交后会自动刷新页面，然后会更新信息
![[Pasted image 20231212101145.png]]
提交后👇
![[Pasted image 20231212101208.png]]
## 验证
他是通过比对本区快记录的前一区块哈希与前一区块的哈希来判断（写的很水。。。。）
![[Pasted image 20231212101225.png]]
## 错误页
这个是在一个[开源网站](https://codepen.io/)里找到的
![[Pasted image 20231213094401.png]]
