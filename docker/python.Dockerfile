# 使用官方 Python 镜像作为基础镜像
FROM python:latest

# 设置工作目录
WORKDIR /app/python

# 复制所需文件到工作目录
COPY python/app.py .
COPY python/templates templates/
COPY python/static static/

# 安装依赖
RUN pip install flask

RUN pip install requests

# 暴露端口
EXPOSE 80

# 启动应用
CMD ["python", "app.py"]

