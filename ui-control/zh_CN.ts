<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE TS>
<TS version="2.1" language="zh_CN" sourcelanguage="en_US">
<context>
    <name>QObject</name>
    <message>
        <location filename="main.cpp" line="302"/>
        <source>CodeBuddy custom providers (models.json)</source>
        <translation>CodeBuddy 自定义供应商（models.json）</translation>
    </message>
    <message>
        <location filename="main.cpp" line="304"/>
        <source>Back up models.json, then append an entry to its models array without removing existing providers. vendor identifies the provider; name is the display label; id must match a model supported by the service. url is the full API endpoint, including /v1/chat/completions, not just the base URL. Replace API_KEY with your service key. The local address below assumes port 8787; change it to your gateway port. Set token limits and capability flags only according to actual model and gateway support. Save and restart CodeBuddy, then select the custom model in its model selector. Never route an upstream CodeBuddy CLI back to its own gateway.</source>
        <translation>先备份 models.json，再向 models 数组追加条目，不要删除已有供应商。vendor 标识供应商，name 是显示名称，id 应填写服务支持的模型 ID。url 是完整接口地址，包含 /v1/chat/completions，而不只是基础地址。将 API_KEY 替换为服务密钥。下方本地地址假设端口为 8787，请按实际网关端口修改。Token 上限和能力标记应按模型及网关的实际支持情况设置。保存并重启 CodeBuddy 后，在模型选择器中选择自定义模型。不要让上游 CodeBuddy CLI 回连自己的网关。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="306"/>
        <source>Optional fields: maxInputTokens / maxOutputTokens are token limits; supportsToolCall / supportsImages / supportsReasoning describe model capabilities. Configure reasoning only when supported by the provider. Do not copy another model&apos;s limits or declare unsupported capabilities. API keys in this file are plaintext: protect it and do not share or commit it.</source>
        <translation>可选字段：maxInputTokens / maxOutputTokens 表示 Token 上限；supportsToolCall / supportsImages / supportsReasoning 表示工具调用、图片输入和推理能力。仅在供应商支持时配置 reasoning。不要照抄其他模型的上限或声明不支持的能力。此文件中的 API 密钥为明文，请妥善保护，不要分享或提交到代码仓库。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="44"/>
        <source>CLI installation is in progress. Wait before exiting.</source>
        <translation>CLI 正在安装，请等待安装结束后退出。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="49"/>
        <source>Running · </source>
        <translation>运行中 · </translation>
    </message>
    <message>
        <location filename="main.cpp" line="49"/>
        <source>Gateway stopped</source>
        <translation>网关已停止</translation>
    </message>
    <message>
        <location filename="main.cpp" line="49"/>
        <source>Status unknown: port conflict, authentication failure, or network error.</source>
        <translation>无法确认状态：端口占用、鉴权失败或网络异常</translation>
    </message>
    <message>
        <location filename="main.cpp" line="95"/>
        <source>Gateway API key cannot be empty.</source>
        <translation>网关 key 不能为空</translation>
    </message>
    <message>
        <location filename="main.cpp" line="96"/>
        <source>Wait for CLI installation before starting the gateway.</source>
        <translation>请等待 CLI 安装完成后再启动网关。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="98"/>
        <source>Add and select a plugin on the Plugins page first.</source>
        <translation>请先在插件页添加并选择插件。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="102"/>
        <source>Gateway or plugin file not found. Run build.bat first.</source>
        <translation>网关或插件文件不存在，请先运行 build.bat</translation>
    </message>
    <message>
        <location filename="main.cpp" line="104"/>
        <source>Port is already in use. A second gateway will not be started.</source>
        <translation>监听端口已被占用，不会启动第二个网关</translation>
    </message>
    <message>
        <location filename="main.cpp" line="118"/>
        <source>Failed to start the gateway. Check the executable and CLI installation.</source>
        <translation>网关启动失败，请检查可执行文件和 CLI 安装</translation>
    </message>
    <message>
        <location filename="main.cpp" line="122"/>
        <source>Starting the gateway. Exiting the control panel will not stop it.</source>
        <translation>正在启动网关；退出控制面不会停止网关</translation>
    </message>
    <message>
        <location filename="main.cpp" line="128"/>
        <source>Gateway stop requested. Active requests will be interrupted.</source>
        <translation>已请求停止网关，正在进行的请求会中断</translation>
    </message>
    <message>
        <location filename="main.cpp" line="128"/>
        <source>Failed to stop. Check the gateway status.</source>
        <translation>停止失败，请检查网关状态</translation>
    </message>
    <message>
        <location filename="main.cpp" line="133"/>
        <source>Wait for CLI installation before installing updates.</source>
        <translation>CLI 正在安装，请完成后再安装更新。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="135"/>
        <source>Install Update</source>
        <translation>安装更新</translation>
    </message>
    <message>
        <location filename="main.cpp" line="135"/>
        <source>Installation will stop the gateway and exit the control panel. Continue?</source>
        <translation>安装会停止网关并退出控制面。继续？</translation>
    </message>
    <message>
        <location filename="main.cpp" line="142"/>
        <source>Could not safely stop the gateway. Update canceled.</source>
        <translation>无法安全停止网关，更新已取消</translation>
    </message>
    <message>
        <location filename="main.cpp" line="146"/>
        <source>Failed to start the external updater.</source>
        <translation>无法启动独立更新器</translation>
    </message>
    <message>
        <location filename="main.cpp" line="159"/>
        <source>Wait for CLI installation to finish before exiting.</source>
        <translation>请等待 CLI 安装结束后退出。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="164"/>
        <location filename="main.cpp" line="374"/>
        <source>1bot · Gateway Control Panel</source>
        <translation>1bot · 代理网关控制面</translation>
    </message>
    <message>
        <location filename="main.cpp" line="178"/>
        <location filename="main.cpp" line="375"/>
        <source>1bot / Local API Gateway</source>
        <translation>1bot / 本地代理网关</translation>
    </message>
    <message>
        <location filename="main.cpp" line="49"/>
        <location filename="main.cpp" line="181"/>
        <source>Checking gateway...</source>
        <translation>正在检测网关…</translation>
    </message>
    <message>
        <location filename="main.cpp" line="188"/>
        <location filename="main.cpp" line="376"/>
        <source>Control panel navigation</source>
        <translation>控制面页面导航</translation>
    </message>
    <message>
        <location filename="main.cpp" line="193"/>
        <location filename="main.cpp" line="229"/>
        <location filename="main.cpp" line="377"/>
        <location filename="main.cpp" line="385"/>
        <source>Gateway</source>
        <translation>网关管理</translation>
    </message>
    <message>
        <location filename="main.cpp" line="194"/>
        <location filename="main.cpp" line="377"/>
        <source>Plugins</source>
        <translation>插件</translation>
    </message>
    <message>
        <location filename="main.cpp" line="195"/>
        <location filename="main.cpp" line="378"/>
        <source>Agent Configuration</source>
        <translation>Agent 配置</translation>
    </message>
    <message>
        <location filename="main.cpp" line="196"/>
        <location filename="main.cpp" line="311"/>
        <location filename="main.cpp" line="378"/>
        <location filename="main.cpp" line="394"/>
        <source>Software Updates</source>
        <translation>软件更新</translation>
    </message>
    <message>
        <location filename="main.cpp" line="237"/>
        <location filename="main.cpp" line="389"/>
        <source>Edit this plugin&apos;s CLI command or path on the Plugins page.</source>
        <translation>在左侧插件页编辑此插件的 CLI 命令或路径</translation>
    </message>
    <message>
        <location filename="main.cpp" line="253"/>
        <location filename="main.cpp" line="386"/>
        <source>Gateway executable</source>
        <translation>网关程序</translation>
    </message>
    <message>
        <location filename="main.cpp" line="253"/>
        <location filename="main.cpp" line="386"/>
        <source>Upstream plugin</source>
        <translation>上游插件</translation>
    </message>
    <message>
        <location filename="main.cpp" line="254"/>
        <location filename="main.cpp" line="387"/>
        <source>CLI executable / path</source>
        <translation>CLI 程序 / 路径</translation>
    </message>
    <message>
        <location filename="main.cpp" line="254"/>
        <location filename="main.cpp" line="387"/>
        <source>Model (optional)</source>
        <translation>模型（可空）</translation>
    </message>
    <message>
        <location filename="main.cpp" line="255"/>
        <location filename="main.cpp" line="388"/>
        <source>Local port</source>
        <translation>本机端口</translation>
    </message>
    <message>
        <location filename="main.cpp" line="255"/>
        <location filename="main.cpp" line="388"/>
        <source>Gateway API key</source>
        <translation>网关 API key</translation>
    </message>
    <message>
        <location filename="main.cpp" line="258"/>
        <location filename="main.cpp" line="364"/>
        <location filename="main.cpp" line="390"/>
        <location filename="main.cpp" line="406"/>
        <source>Start Gateway</source>
        <translation>启动网关</translation>
    </message>
    <message>
        <location filename="main.cpp" line="258"/>
        <location filename="main.cpp" line="365"/>
        <location filename="main.cpp" line="391"/>
        <location filename="main.cpp" line="407"/>
        <source>Stop Gateway</source>
        <translation>停止网关</translation>
    </message>
    <message>
        <location filename="main.cpp" line="261"/>
        <location filename="main.cpp" line="392"/>
        <source>Copy Gateway Key</source>
        <translation>复制网关 key</translation>
    </message>
    <message>
        <location filename="main.cpp" line="313"/>
        <location filename="main.cpp" line="395"/>
        <source>Current version</source>
        <translation>当前版本</translation>
    </message>
    <message>
        <location filename="main.cpp" line="292"/>
        <source>Manual Agent Configuration

This page is read-only. 1bot does not read or modify agent configuration files.

1. Close the agent and back up its configuration yourself.
2. Open the file below in a text editor. Merge the example fields into the existing JSON; do not replace unrelated settings.
3. Replace SERVICE_URL, API_KEY and MODEL_ID with values supported by your service. Remove model if not needed. Protect API keys and never commit them to a repository.
4. Save the file and restart the agent. To undo changes, restore your own backup.</source>
        <translation>手动配置 Agent

此页面仅提供只读说明。1bot 不读取或修改 Agent 配置文件。

1. 关闭 Agent，自行备份配置文件。
2. 用文本编辑器打开下方文件，将示例字段合并到现有 JSON 中，不要覆盖其他设置。
3. 将 SERVICE_URL、API_KEY 和 MODEL_ID 替换为服务支持的地址、密钥和模型；不需要指定模型时删除 model。妥善保管密钥，不要提交到代码仓库。
4. 保存文件并重启 Agent。如需撤销修改，请自行恢复备份。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="285"/>
        <source>Selected model: </source>
        <translation>已选择模型：</translation>
    </message>
    <message>
        <location filename="main.cpp" line="293"/>
        <source>Compatibility: the current gateway provides OpenAI Chat Completions, not Anthropic Messages. Claude Code requires a compatible Anthropic service. CodeBuddy endpoints must match its CLI version and protocol. Do not point the gateway&apos;s upstream CLI back to the same gateway; this creates a request loop.</source>
        <translation>兼容性：当前网关提供 OpenAI Chat Completions，不提供 Anthropic Messages。Claude Code 需要兼容 Anthropic 协议的服务。CodeBuddy 的服务端点必须匹配 CLI 版本和协议。不要让网关上游 CLI 指向同一个网关，否则会形成循环请求。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="294"/>
        <source>User configuration paths are shown below. Project settings, environment variables or command-line options may override them. Examples are templates, not a guarantee of compatibility.</source>
        <translation>下方显示用户级配置路径。项目设置、环境变量或命令行选项可能覆盖这些配置。示例仅为模板，不保证协议兼容。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="295"/>
        <source>CodeBuddy uses the top-level endpoint setting. Check the documentation for your installed CLI version:</source>
        <translation>CodeBuddy 使用顶层 endpoint 设置。请核对已安装 CLI 版本的文档：</translation>
    </message>
    <message>
        <location filename="main.cpp" line="315"/>
        <location filename="main.cpp" line="398"/>
        <source>System default</source>
        <translation>跟随系统</translation>
    </message>
    <message>
        <location filename="main.cpp" line="323"/>
        <location filename="main.cpp" line="395"/>
        <source>Language</source>
        <translation>语言</translation>
    </message>
    <message>
        <location filename="main.cpp" line="333"/>
        <location filename="main.cpp" line="399"/>
        <source>1bot HTTPS update endpoint</source>
        <translation>1bot HTTPS 更新接口</translation>
    </message>
    <message>
        <location filename="main.cpp" line="334"/>
        <location filename="main.cpp" line="400"/>
        <source>Check at startup and every 6 hours</source>
        <translation>启动时及每 6 小时自动检查</translation>
    </message>
    <message>
        <location filename="main.cpp" line="336"/>
        <location filename="main.cpp" line="401"/>
        <source>Check for Updates</source>
        <translation>检查更新</translation>
    </message>
    <message>
        <location filename="main.cpp" line="337"/>
        <location filename="main.cpp" line="396"/>
        <source>Update URL</source>
        <translation>更新地址</translation>
    </message>
    <message>
        <location filename="main.cpp" line="338"/>
        <location filename="main.cpp" line="396"/>
        <source>Check policy</source>
        <translation>检查策略</translation>
    </message>
    <message>
        <location filename="main.cpp" line="344"/>
        <location filename="main.cpp" line="402"/>
        <source>You will be asked before downloading a new version. Installation requires confirmation and stops the gateway.</source>
        <translation>发现新版本后会询问是否下载；安装前会请求确认并停止网关。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="348"/>
        <location filename="main.cpp" line="403"/>
        <source>Closing the window minimizes it to the tray. The gateway runs independently.</source>
        <translation>关闭窗口后驻留托盘；网关与控制面独立运行。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="353"/>
        <source>New Version </source>
        <translation>发现新版本 </translation>
    </message>
    <message>
        <location filename="main.cpp" line="353"/>
        <source>
Download the update?</source>
        <translation>
下载更新？</translation>
    </message>
    <message>
        <location filename="main.cpp" line="363"/>
        <location filename="main.cpp" line="405"/>
        <source>Open Control Panel</source>
        <translation>打开控制面</translation>
    </message>
    <message>
        <location filename="main.cpp" line="367"/>
        <location filename="main.cpp" line="408"/>
        <source>Exit Control Panel (Keep Gateway Running)</source>
        <translation>退出控制面（保留网关）</translation>
    </message>
    <message>
        <location filename="main.cpp" line="431"/>
        <source>The control panel is already running. Open it from the system tray.</source>
        <translation>控制面已经运行，请从托盘打开。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="7"/>
        <location filename="modelspage.cpp" line="39"/>
        <source>Models Supported by the Current Plugin</source>
        <translation>当前插件支持的模型</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="9"/>
        <location filename="modelspage.cpp" line="40"/>
        <source>Refresh List</source>
        <translation>刷新列表</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="14"/>
        <source>Start the gateway first.</source>
        <translation>请先启动网关。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="19"/>
        <location filename="modelspage.cpp" line="41"/>
        <source>Model ID</source>
        <translation>模型 ID</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="19"/>
        <location filename="modelspage.cpp" line="41"/>
        <source>Plugin</source>
        <translation>所属插件</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="28"/>
        <location filename="modelspage.cpp" line="42"/>
        <source>This catalog comes from the running gateway, not the default model argument. This page does not switch models.</source>
        <translation>列表来自正在运行的网关，不是启动参数中的默认模型。此页仅展示，不切换模型。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="67"/>
        <source>Click Refresh to retrieve the current plugin&apos;s models.</source>
        <translation>点击刷新获取当前插件的模型列表。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="44"/>
        <location filename="modelspage.cpp" line="67"/>
        <source>The gateway is offline or cannot be authenticated. Start it and check the connection on the Gateway page.</source>
        <translation>网关未运行或无法确认身份，请先在网关管理页启动并检查连接。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="43"/>
        <location filename="modelspage.cpp" line="81"/>
        <source>Fetching model catalog...</source>
        <translation>正在查询模型列表…</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="104"/>
        <source>This plugin or CLI version does not support model discovery.</source>
        <translation>当前插件或 CLI 版本不支持枚举模型。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="105"/>
        <source>Authentication failed. Check the gateway API key.</source>
        <translation>鉴权失败，请检查网关 API key。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="108"/>
        <source>The request timed out or exceeded the response limit. Try refreshing later.</source>
        <translation>查询超时或响应超出限制，请稍后刷新。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="109"/>
        <source>Plugin query failed. Check the CLI status and retry.</source>
        <translation>插件查询失败，请检查 CLI 状态后重试。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="110"/>
        <source>Model endpoint unavailable. Update the gateway and plugins.</source>
        <translation>网关未提供模型接口，请更新网关与插件。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="111"/>
        <source>Could not retrieve models. Check the gateway connection and retry.</source>
        <translation>无法获取模型列表，请检查网关连接后重试。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="120"/>
        <source>The gateway returned an invalid model catalog.</source>
        <translation>网关返回了无效的模型列表。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="130"/>
        <source>The catalog contains invalid entries. Results were not displayed.</source>
        <translation>模型列表包含无效条目，未显示结果。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="46"/>
        <location filename="modelspage.cpp" line="142"/>
        <source>The current plugin returned an empty model catalog.</source>
        <translation>当前插件返回了空模型列表。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="47"/>
        <location filename="modelspage.cpp" line="143"/>
        <source>%1 models · Updated at %2</source>
        <translation>共 %1 个模型 · 更新于 %2</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="43"/>
        <location filename="pluginspage.cpp" line="145"/>
        <source>Application Plugins</source>
        <translation>应用插件</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="46"/>
        <location filename="pluginspage.cpp" line="146"/>
        <source>Plugins must implement 1bot stdio JSON-RPC and accept gateway arguments such as --bin / --model. Installation only installs the CLI dependency, not a plugin. Relative plugin paths are resolved from the gateway directory.</source>
        <translation>插件程序须实现 1bot stdio JSON-RPC，并接受 --bin / --model 等网关参数。安装按钮只安装依赖 CLI，不会自动生成插件。相对插件路径以网关程序目录为基准。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="54"/>
        <location filename="pluginspage.cpp" line="147"/>
        <source>Add Plugin</source>
        <translation>新增插件</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="55"/>
        <location filename="pluginspage.cpp" line="148"/>
        <source>Delete Configuration</source>
        <translation>删除配置</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="66"/>
        <location filename="pluginspage.cpp" line="152"/>
        <source>Enter the official PowerShell installation command. Unverified CLI commands are left blank.</source>
        <translation>填写官方 PowerShell 安装命令；未核实的 CLI 不预填命令。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="69"/>
        <location filename="pluginspage.cpp" line="149"/>
        <source>Browse</source>
        <translation>浏览</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="71"/>
        <location filename="pluginspage.cpp" line="153"/>
        <source>Display name</source>
        <translation>显示名称</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="72"/>
        <location filename="pluginspage.cpp" line="153"/>
        <source>Plugin executable</source>
        <translation>插件程序</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="73"/>
        <location filename="pluginspage.cpp" line="154"/>
        <source>CLI command / path</source>
        <translation>CLI 命令 / 路径</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="74"/>
        <location filename="pluginspage.cpp" line="154"/>
        <source>CLI dependency installation command</source>
        <translation>CLI 依赖安装命令</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="76"/>
        <location filename="pluginspage.cpp" line="150"/>
        <source>Save Configuration</source>
        <translation>保存配置</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="77"/>
        <location filename="pluginspage.cpp" line="151"/>
        <source>Install Dependencies</source>
        <translation>安装依赖</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="81"/>
        <source>Select a plugin to view and edit its CLI installation command.</source>
        <translation>选择插件可查看和编辑 CLI 安装命令。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="91"/>
        <source>Select a 1bot-compatible Plugin</source>
        <translation>选择兼容 1bot 的插件程序</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="91"/>
        <source>Executables (*.exe)</source>
        <translation>可执行文件 (*.exe)</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="99"/>
        <source>Save the new plugin to add it to the Gateway page selector.</source>
        <translation>填写后保存，新插件会加入网关管理页的选择列表。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="103"/>
        <source>Delete Plugin Configuration</source>
        <translation>删除插件配置</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="103"/>
        <source>Only the configuration will be removed. The CLI and plugin files will remain. Continue?</source>
        <translation>只删除配置，不卸载 CLI，也不删除插件文件。继续？</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="119"/>
        <source>Failed to start installation: </source>
        <translation>安装进程启动失败：</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="126"/>
        <source>Installation command completed successfully. CLI login may still be required. Restart the control panel if PATH changed.</source>
        <translation>安装命令成功结束。首次使用可能仍需 CLI 登录；若更改 PATH，请重启控制面。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="127"/>
        <source>Installation failed (exit code %1). Check the output.</source>
        <translation>安装失败（退出码 %1），请查看输出。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="185"/>
        <source>Requires CodeBuddy Code CLI and Node.js 18.20+ for npm installation. Official package: @tencent-ai/codebuddy-code.</source>
        <translation>依赖 CodeBuddy Code CLI；npm 安装需先安装 Node.js 18.20+。官方包：@tencent-ai/codebuddy-code。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="187"/>
        <source>Requires ZCode CLI (zcode-app-cli) and Node.js/npm. Follow the CLI instructions to sign in after installation.</source>
        <translation>依赖 ZCode CLI（zcode-app-cli）；npm 安装需先安装 Node.js/npm，安装后按 CLI 指引完成登录。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="189"/>
        <source>Requires Codex CLI and Node.js/npm. Follow the official sign-in instructions after installation.</source>
        <translation>依赖 Codex CLI；npm 安装需先安装 Node.js/npm，安装后按官方指引登录。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="190"/>
        <source>Configure the CLI dependency and its official installation command.</source>
        <translation>请配置此插件依赖的 CLI 和官方安装命令。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="201"/>
        <source>Failed to persist configuration. Check user permissions.</source>
        <translation>配置持久化失败，请检查当前用户权限。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="207"/>
        <source>Name, plugin executable, and CLI command cannot be empty.</source>
        <translation>名称、插件程序和 CLI 命令不能为空。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="218"/>
        <source>Configuration saved. Changes take effect after restarting the gateway; the running gateway is unchanged.</source>
        <translation>配置已保存；运行中的网关不受影响，重启网关后生效。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="231"/>
        <source>No installation command is configured. Enter one from the official CLI documentation.</source>
        <translation>尚无已确认的安装命令，请从 CLI 官方文档填写。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="232"/>
        <source>Confirm Installation Command</source>
        <translation>确认执行安装命令</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="233"/>
        <source>The following PowerShell command will run as the current Windows user and may download software or change the environment. Run only trusted official commands.

</source>
        <translation>将使用当前 Windows 用户权限执行以下 PowerShell 命令，可能下载软件并修改系统环境。仅执行你信任的官方命令。

</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="237"/>
        <source>Working directory: </source>
        <translation>工作目录：</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="158"/>
        <location filename="pluginspage.cpp" line="240"/>
        <source>Installing. Do not exit the control panel. Run interactive or administrator installations in a terminal.</source>
        <translation>安装中，请勿退出控制面；需要交互或管理员权限的安装请在终端完成。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="21"/>
        <source>Update file exceeds the size limit.</source>
        <translation>更新文件超出大小限制</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="24"/>
        <source>Cannot write update data.</source>
        <translation>无法写入更新数据</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="30"/>
        <source>Update downloads require HTTPS.</source>
        <translation>更新下载地址必须为 HTTPS</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="32"/>
        <source>Cannot initialize libcurl.</source>
        <translation>无法初始化 libcurl</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="35"/>
        <source>The updater requires libcurl with the OpenSSL backend.</source>
        <translation>更新器要求使用 OpenSSL 后端的 libcurl</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="37"/>
        <source>Missing curl-ca-bundle.crt CA bundle.</source>
        <translation>缺少 curl-ca-bundle.crt 根证书包</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="39"/>
        <source>Cannot create an HTTPS request.</source>
        <translation>无法创建 HTTPS 请求</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="64"/>
        <source>Cannot create request headers.</source>
        <translation>无法创建请求头</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="71"/>
        <source>Cannot configure secure HTTPS request options.</source>
        <translation>无法设置安全 HTTPS 请求参数</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="75"/>
        <source>HTTPS request failed: </source>
        <translation>HTTPS 请求失败：</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="78"/>
        <source>Update server HTTP status %1 (redirects are not accepted).</source>
        <translation>更新服务 HTTP 状态 %1（不接受重定向）</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="84"/>
        <source>Configure a dedicated 1bot HTTPS update URL.</source>
        <translation>请配置 1bot 专用 HTTPS 更新地址</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="101"/>
        <source>Invalid update manifest.</source>
        <translation>无效的更新清单</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="106"/>
        <source>No updates available.</source>
        <translation>当前没有新版本</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="120"/>
        <source>The update manifest contains unsafe files, invalid sizes, or missing SHA-256 hashes.</source>
        <translation>更新清单含不安全文件、无效大小或缺少 SHA-256</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="125"/>
        <source>The update package is missing 1bot-control.exe.</source>
        <translation>更新包缺少 1bot-control.exe</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="141"/>
        <source>Cannot create the update directory.</source>
        <translation>无法创建更新目录</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="149"/>
        <source>Cannot save the update manifest.</source>
        <translation>无法保存更新清单</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="154"/>
        <source>Downloading update %1/%2: %3</source>
        <translation>正在下载更新 %1/%2：%3</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="164"/>
        <source>Cannot create the download file.</source>
        <translation>无法创建下载文件</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="168"/>
        <source>Download size mismatch or save failure.</source>
        <translation>下载大小不符或无法保存</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="173"/>
        <source>SHA-256 verification failed. The download was deleted.</source>
        <translation>SHA-256 校验失败，已删除下载</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="33"/>
        <source>Cannot read file: </source>
        <translation>无法读取文件：</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="39"/>
        <source>Cannot write atomically: </source>
        <translation>无法原子写入：</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="50"/>
        <source>Invalid request.</source>
        <translation>无效请求</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="53"/>
        <source>The configuration path must be absolute and cannot be a symbolic link.</source>
        <translation>配置必须是绝对路径且不能为符号链接</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="56"/>
        <source>Another control panel is modifying this configuration.</source>
        <translation>配置正在被其他控制面修改</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="64"/>
        <source>Cannot decrypt the backup.</source>
        <translation>备份无法解密</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="67"/>
        <source>The configuration was changed externally. Overwrite refused; merge changes manually first.</source>
        <translation>配置已被外部修改，已拒绝覆盖。请手工合并后再操作</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="70"/>
        <source>No backup is available to restore.</source>
        <translation>没有可恢复的备份</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="72"/>
        <source>Cannot remove the newly created configuration.</source>
        <translation>无法删除新建配置</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="73"/>
        <source>Configuration restored, but backup cleanup failed.</source>
        <translation>已恢复配置，但备份清理失败</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="76"/>
        <source>Only strict JSON object configurations are supported. The original file was not changed.</source>
        <translation>仅支持严格 JSON 对象配置，原文件未修改</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="78"/>
        <source>env must be a JSON object.</source>
        <translation>env 必须是 JSON 对象</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="89"/>
        <source>Unsupported agent.</source>
        <translation>不支持的 Agent</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="93"/>
        <source>The API URL must use HTTPS or local HTTP.</source>
        <translation>API 地址必须为 HTTPS 或本机 HTTP</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="94"/>
        <source>API key cannot be empty.</source>
        <translation>API key 不能为空</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="107"/>
        <source>Cannot encrypt the backup.</source>
        <translation>无法加密备份</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="10"/>
        <location filename="accesscard.cpp" line="93"/>
        <source>Connect</source>
        <translation>接入</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="50"/>
        <location filename="accesscard.cpp" line="104"/>
        <source>Hide</source>
        <translation>隐藏</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="50"/>
        <location filename="accesscard.cpp" line="104"/>
        <source>Show</source>
        <translation>显示</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="94"/>
        <source>Use these values to connect any OpenAI-compatible client to the local gateway.</source>
        <translation>使用以下信息即可将任意 OpenAI 兼容客户端接入本机网关。</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="95"/>
        <source>Base URL</source>
        <translation>基础地址</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="96"/>
        <source>API Key</source>
        <translation>API 密钥</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="97"/>
        <source>Model</source>
        <translation>模型</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="98"/>
        <location filename="accesscard.cpp" line="99"/>
        <location filename="accesscard.cpp" line="100"/>
        <location filename="accesscard.cpp" line="101"/>
        <location filename="accesscard.cpp" line="102"/>
        <location filename="accesscard.cpp" line="103"/>
        <source>Copy</source>
        <translation>复制</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="105"/>
        <source>PowerShell</source>
        <translation>PowerShell</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="106"/>
        <source>CMD</source>
        <translation>CMD</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="107"/>
        <source>curl</source>
        <translation>curl</translation>
    </message>
</context>
</TS>
