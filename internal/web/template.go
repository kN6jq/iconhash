package web

// HTMLTemplate 是Web模式的HTML模板
const HTMLTemplate = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Favicon Hash 计算器</title>
    <link rel="stylesheet" href="/static/css/all.min.css">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }

        .container {
            background: rgba(255, 255, 255, 0.95);
            border-radius: 20px;
            padding: 40px;
            box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
            backdrop-filter: blur(10px);
            max-width: 600px;
            width: 100%;
            transition: transform 0.3s ease;
        }

        .container:hover {
            transform: translateY(-5px);
        }

        .header {
            text-align: center;
            margin-bottom: 40px;
        }

        .title {
            color: #333;
            font-size: 2.5rem;
            font-weight: 700;
            margin-bottom: 10px;
            background: linear-gradient(135deg, #667eea, #764ba2);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }

        .subtitle {
            color: #666;
            font-size: 1.1rem;
            font-weight: 400;
        }

        .input-group {
            position: relative;
            margin-bottom: 30px;
        }

        .input-label {
            position: absolute;
            top: -10px;
            left: 15px;
            background: #fff;
            padding: 0 10px;
            color: #667eea;
            font-size: 0.9rem;
            font-weight: 600;
            z-index: 1;
        }

        .url-input {
            width: 100%;
            padding: 18px 20px;
            border: 2px solid #e1e5e9;
            border-radius: 15px;
            font-size: 1rem;
            transition: all 0.3s ease;
            outline: none;
            background: #f8f9fa;
        }

        .url-input:focus {
            border-color: #667eea;
            background: #fff;
            box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
        }

        .calculate-btn {
            width: 100%;
            padding: 18px;
            background: linear-gradient(135deg, #667eea, #764ba2);
            color: white;
            border: none;
            border-radius: 15px;
            font-size: 1.1rem;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s ease;
            box-shadow: 0 8px 15px rgba(102, 126, 234, 0.3);
            position: relative;
            overflow: hidden;
        }

        .calculate-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 12px 20px rgba(102, 126, 234, 0.4);
        }

        .calculate-btn:active {
            transform: translateY(0);
        }

        .calculate-btn.loading {
            opacity: 0.8;
            cursor: not-allowed;
        }

        .loading-spinner {
            display: none;
            width: 20px;
            height: 20px;
            border: 2px solid rgba(255, 255, 255, 0.3);
            border-radius: 50%;
            border-top-color: #fff;
            animation: spin 1s ease-in-out infinite;
            margin-right: 10px;
        }

        @keyframes spin {
            to { transform: rotate(360deg); }
        }

        .result-card {
            margin-top: 30px;
            background: #f8f9fa;
            border-radius: 15px;
            padding: 25px;
            border-left: 4px solid #667eea;
            display: none;
            animation: slideIn 0.5s ease-out;
        }

        @keyframes slideIn {
            from {
                opacity: 0;
                transform: translateY(20px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }

        .result-title {
            color: #333;
            font-size: 1.3rem;
            font-weight: 600;
            margin-bottom: 20px;
            display: flex;
            align-items: center;
        }

        .result-title i {
            margin-right: 10px;
            color: #667eea;
        }

        .hash-item {
            background: #fff;
            border-radius: 12px;
            padding: 20px;
            margin-bottom: 15px;
            box-shadow: 0 3px 10px rgba(0, 0, 0, 0.05);
            transition: transform 0.2s ease;
        }

        .hash-item:hover {
            transform: translateY(-2px);
        }

        .hash-platform {
            font-weight: 600;
            color: #333;
            margin-bottom: 8px;
            display: flex;
            align-items: center;
        }

        .platform-icon {
            width: 20px;
            height: 20px;
            margin-right: 8px;
            border-radius: 4px;
        }

        .fofa-icon {
            background: linear-gradient(135deg, #ff6b6b, #ee5a52);
        }

        .hunter-icon {
            background: linear-gradient(135deg, #4ecdc4, #44a08d);
        }

        .md5-icon {
            background: linear-gradient(135deg, #feca57, #ff9ff3);
        }

        .hash-value {
            font-family: 'Courier New', monospace;
            background: #f1f3f4;
            padding: 12px;
            border-radius: 8px;
            margin: 10px 0;
            word-break: break-all;
            color: #333;
            font-size: 0.9rem;
            position: relative;
        }

        .copy-btn {
            position: absolute;
            top: 8px;
            right: 8px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 5px;
            padding: 4px 8px;
            font-size: 0.8rem;
            cursor: pointer;
            opacity: 0.7;
            transition: opacity 0.2s ease;
        }

        .copy-btn:hover {
            opacity: 1;
        }

        .search-link {
            display: inline-flex;
            align-items: center;
            color: #667eea;
            text-decoration: none;
            font-weight: 500;
            padding: 8px 12px;
            border-radius: 8px;
            background: rgba(102, 126, 234, 0.1);
            transition: all 0.2s ease;
        }

        .search-link:hover {
            background: rgba(102, 126, 234, 0.2);
            transform: translateX(3px);
        }

        .search-link i {
            margin-right: 5px;
        }

        .error-message {
            background: #ffe6e6;
            color: #d63031;
            padding: 15px;
            border-radius: 10px;
            border-left: 4px solid #d63031;
            margin-top: 20px;
            display: none;
            animation: slideIn 0.3s ease-out;
        }

        .footer {
            text-align: center;
            margin-top: 40px;
            color: #666;
            font-size: 0.9rem;
        }

        .footer a {
            color: #667eea;
            text-decoration: none;
        }

        @media (max-width: 768px) {
            .container {
                padding: 30px 20px;
                margin: 10px;
            }
            
            .title {
                font-size: 2rem;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="title">
                <i class="fas fa-fingerprint"></i>
                Favicon Hash
            </h1>
            <p class="subtitle">快速计算网站图标的哈希值，支持 FOFA 和 Hunter 搜索</p>
        </div>

        <div class="input-group">
            <label class="input-label">Favicon URL</label>
            <input 
                type="text" 
                id="urlInput" 
                class="url-input"
                placeholder="请输入 favicon URL，例如：https://example.com/favicon.ico"
                onkeypress="handleKeyPress(event)"
            >
        </div>

        <button class="calculate-btn" onclick="calculateHash()">
            <span class="loading-spinner"></span>
            <span class="btn-text">
                <i class="fas fa-calculator"></i>
                开始计算
            </span>
        </button>

        <div id="result" class="result-card">
            <h3 class="result-title">
                <i class="fas fa-check-circle"></i>
                计算结果
            </h3>
            
            <div class="hash-item">
                <div class="hash-platform">
                    <div class="platform-icon fofa-icon"></div>
                    FOFA 搜索引擎
                </div>
                <div class="hash-value" id="fofaHash">
                    <button class="copy-btn" onclick="copyToClipboard('fofaHash')">复制</button>
                </div>
                <a id="fofaLink" href="#" target="_blank" class="search-link">
                    <i class="fas fa-external-link-alt"></i>
                    在 FOFA 中搜索
                </a>
            </div>

            <div class="hash-item">
                <div class="hash-platform">
                    <div class="platform-icon hunter-icon"></div>
                    Hunter 搜索引擎
                </div>
                <div class="hash-value" id="hunterHash">
                    <button class="copy-btn" onclick="copyToClipboard('hunterHash')">复制</button>
                </div>
                <a id="hunterLink" href="#" target="_blank" class="search-link">
                    <i class="fas fa-external-link-alt"></i>
                    在 Hunter 中搜索
                </a>
            </div>

            <div class="hash-item">
                <div class="hash-platform">
                    <div class="platform-icon md5-icon"></div>
                    纯 MD5 哈希值
                </div>
                <div class="hash-value" id="md5Hash">
                    <button class="copy-btn" onclick="copyToClipboard('md5Hash')">复制</button>
                </div>
            </div>
        </div>

        <div id="error" class="error-message"></div>

        <div class="footer">
            <p>
                <i class="fas fa-heart" style="color: #e74c3c;"></i>
                一个简单实用的 Favicon Hash 计算工具
            </p>
        </div>
    </div>

    <script>
        function handleKeyPress(event) {
            if (event.key === 'Enter') {
                calculateHash();
            }
        }

        async function calculateHash() {
            const url = document.getElementById("urlInput").value.trim();
            const resultDiv = document.getElementById("result");
            const errorDiv = document.getElementById("error");
            const fofaSpan = document.getElementById("fofaHash");
            const hunterSpan = document.getElementById("hunterHash");
            const md5Span = document.getElementById("md5Hash");
            const fofaLink = document.getElementById("fofaLink");
            const hunterLink = document.getElementById("hunterLink");
            const btn = document.querySelector(".calculate-btn");
            const spinner = document.querySelector(".loading-spinner");
            const btnText = document.querySelector(".btn-text");

            // 清空之前的错误和结果
            errorDiv.style.display = "none";
            resultDiv.style.display = "none";
            errorDiv.textContent = "";

            if (!url) {
                showError("请输入有效的 favicon URL");
                return;
            }

            // 显示加载状态
            btn.classList.add("loading");
            spinner.style.display = "inline-block";
            btnText.innerHTML = '<i class="fas fa-sync fa-spin"></i> 计算中...';
            btn.disabled = true;

            try {
                const response = await fetch("/calculate?url=" + encodeURIComponent(url));
                if (!response.ok) {
                    throw new Error(await response.text());
                }
                const text = await response.text();
                const lines = text.split("\n");
                
                // 设置哈希值和链接
                fofaSpan.textContent = lines[0];
                fofaLink.href = lines[1];
                
                hunterSpan.textContent = lines[2];
                hunterLink.href = lines[3];
                
                // 设置纯 MD5 值
                if (lines[4]) {
                    md5Span.textContent = lines[4];
                }
                
                resultDiv.style.display = "block";
                
                // 添加成功动画
                resultDiv.style.animation = 'none';
                resultDiv.offsetHeight; // 触发重绘
                resultDiv.style.animation = 'slideIn 0.5s ease-out';
                
            } catch (error) {
                showError("计算失败: " + error.message);
            } finally {
                // 恢复按钮状态
                btn.classList.remove("loading");
                spinner.style.display = "none";
                btnText.innerHTML = '<i class="fas fa-calculator"></i> 开始计算';
                btn.disabled = false;
            }
        }

        function showError(message) {
            const errorDiv = document.getElementById("error");
            errorDiv.textContent = message;
            errorDiv.style.display = "block";
        }

        function copyToClipboard(elementId) {
            const element = document.getElementById(elementId);
            const text = element.textContent.replace(/复制$/, '').trim();
            
            navigator.clipboard.writeText(text).then(() => {
                const btn = element.querySelector('.copy-btn');
                const originalText = btn.textContent;
                btn.textContent = '已复制!';
                btn.style.background = '#00b894';
                
                setTimeout(() => {
                    btn.textContent = originalText;
                    btn.style.background = '#667eea';
                }, 1500);
            }).catch(() => {
                // 降级方案：使用传统方法复制
                const textArea = document.createElement('textarea');
                textArea.value = text;
                document.body.appendChild(textArea);
                textArea.select();
                document.execCommand('copy');
                document.body.removeChild(textArea);
                
                const btn = element.querySelector('.copy-btn');
                const originalText = btn.textContent;
                btn.textContent = '已复制!';
                btn.style.background = '#00b894';
                
                setTimeout(() => {
                    btn.textContent = originalText;
                    btn.style.background = '#667eea';
                }, 1500);
            });
        }

        // 页面加载完成后自动聚焦输入框
        document.addEventListener('DOMContentLoaded', function() {
            document.getElementById('urlInput').focus();
        });
    </script>
</body>
</html>
`
