from flask import Flask, render_template, request, redirect, url_for, session, send_from_directory
import requests
from flask import jsonify  # 引入 jsonify


app = Flask(__name__)
app.secret_key = 'why so serious ?'  # 设置的session密钥

# 后端 Go 服务器的地址
backend_url = 'http://backend:82'

# 配置静态文件夹
app.static_folder = 'static'

# 提供静态文件的路由
@app.route('/static/<path:filename>')
def serve_static(filename):
    return send_from_directory(app.static_folder, filename)

# 默认路由，显示默认页面
@app.route('/', methods=['GET'])
def default():
    if 'username' in session:
        return render_template('index.html', loggedIn=True, currentUser={'Username': session['username']})
    else:
        return render_template('index.html', loggedIn=False)


#错误处理路由
@app.route('/error', methods=['GET'])
def error():
    return render_template('error.html')

# 登录后的路由
@app.route('/loginin', methods=['GET', 'POST'])
def loginin():
    if 'username' not in session:
        return redirect(url_for('login'))  # 如果用户未登录，则重定向到登录页面

    if request.method == 'GET':
        # 处理获取日记和区块链信息的请求
        # 向后端发送获取日记和区块链信息的请求
        diary_response = requests.get(f'{backend_url}/loginin', headers={'Authorization': session['username']})
        if diary_response.status_code == 200:
            # 解析后端返回的 JSON 数据
            diary_result = diary_response.json()
            # 打印响应头信息
            #print("响应头:", diary_response.headers)
            #打印响应信息
            #print(diary_response.text)

            # 检查响应状态
            if diary_result.get('status') == 'success':
                return jsonify(diary_result)  # 返回 JSON 格式的响应
            else:
                return jsonify({'status': 'error', 'message': '获取信息失败'})
        else:
            return jsonify({'status': 'error', 'message': '获取信息失败，无法连接到后端服务'})

    elif request.method == 'POST':
        # 处理写日记的请求
        data = request.get_json()  # 使用 request.get_json() 获取 JSON 数据
        content = data.get('content')

        # 向后端发送写日记的请求
        diary_entry_response = requests.post(f'{backend_url}/loginin', json={'content': content},
                                             headers={'Authorization': session['username']})
        # 接收后端更新的数据
        diary_response = requests.get(f'{backend_url}/loginin', headers={'Authorization': session['username']})
        if diary_response.status_code == 200:
            # 解析后端返回的 JSON 数据
            entry_result = diary_response.json()

            # 检查响应状态
            if entry_result.get('status') == 'success':
                return jsonify(entry_result)  # 返回 JSON 格式的响应
            else:
                return jsonify({'status': 'error', 'message': '写日记失败'})
        else:
            return jsonify({'status': 'error', 'message': '写日记失败，无法连接到后端服务'})


# 登出的路由
@app.route('/logout', methods=['POST'])
def logout():
    # 清除会话中的用户名信息
    session.pop('username', None)

    # 重定向到默认页面或登录页面，根据你的需求
    return redirect(url_for('default'))

# 登录页面
@app.route('/login', methods=['GET', 'POST'])
def login():
    if request.method == 'POST':
        data = request.get_json()  # 使用 request.get_json() 获取 JSON 数据
        username = data.get('username')
        password = data.get('password')

        # 向后端发送登录请求
        response = requests.post(f'{backend_url}/login', json={'username': username, 'password': password})

        # 检查后端响应的状态码
        if response.status_code == 200:
            # 解析后端返回的 JSON 数据
            result = response.json()
            #print(result)
            # 检查登录状态
            if result.get('status') == 'success':
                session['username'] = username
                return redirect(url_for('default'))
            else:
                return render_template('login.html', message=result.get('message', '登录失败，请检查用户名和密码'))
        else:
            return render_template('login.html', message='登录失败，无法连接到后端服务')

    else:
        return render_template('login.html', message='')


# 注册页面
@app.route('/register', methods=['GET', 'POST'])
def register():
    if request.method == 'POST':
        data = request.get_json()  # 使用 request.get_json() 获取 JSON 数据
        username = data.get('username')
        password = data.get('password')

        # 向后端发送注册请求
        response = requests.post(f'{backend_url}/register', json={'username': username, 'password': password})

        # 检查后端响应的状态码
        if response.status_code == 200:
            # 解析后端返回的 JSON 数据
            result = response.json()

            # 检查注册状态
            if result.get('status') == 'success':
                # 注册成功后可以自动登录，这里可以根据需要设置
                session['username'] = username
                return redirect(url_for('default'))
            else:
                return render_template('register.html', message=result.get('message', '注册失败，请检查用户名和密码'))
        else:
            return render_template('register.html', message='注册失败，无法连接到后端服务')

    else:
        return render_template('register.html', message='')


# 验证
@app.route('/validate', methods=['GET'])
def validate():
    # 向后端发送获取日记和区块链信息的请求
    validate_response = requests.get(f'{backend_url}/validateBlockchain')
    if validate_response.status_code == 200:
        # 解析后端返回的 JSON 数据
        result = validate_response.json()
        if result.get('status') == 'success':
            return jsonify(result)  # 返回 JSON 格式的响应
        else:
            return jsonify(result)
    else:
        return jsonify({'status': 'error', 'message': '无法连接到后端服务'})



if __name__ == '__main__':
    app.run(host='0.0.0.0',debug=True,port=80)