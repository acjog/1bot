<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE TS>
<TS version="2.1" language="zh_TW" sourcelanguage="en_US">
<context>
    <name>QObject</name>
    <message>
        <location filename="main.cpp" line="302"/>
        <source>CodeBuddy custom providers (models.json)</source>
        <translation>CodeBuddy 自訂供應商（models.json）</translation>
    </message>
    <message>
        <location filename="main.cpp" line="304"/>
        <source>Back up models.json, then append an entry to its models array without removing existing providers. vendor identifies the provider; name is the display label; id must match a model supported by the service. url is the full API endpoint, including /v1/chat/completions, not just the base URL. Replace API_KEY with your service key. The local address below assumes port 8787; change it to your gateway port. Set token limits and capability flags only according to actual model and gateway support. Save and restart CodeBuddy, then select the custom model in its model selector. Never route an upstream CodeBuddy CLI back to its own gateway.</source>
        <translation>先備份 models.json，再向 models 陣列追加項目，不要刪除既有供應商。vendor 識別供應商，name 是顯示名稱，id 應填寫服務支援的模型 ID。url 是完整介面位址，包含 /v1/chat/completions，而不只是基礎位址。將 API_KEY 替換為服務金鑰。下方本機位址假設連接埠為 8787，請按實際閘道連接埠修改。Token 上限和能力標記應按模型及閘道的實際支援情況設定。儲存並重新啟動 CodeBuddy 後，在模型選擇器中選擇自訂模型。不要讓上游 CodeBuddy CLI 回連自己的閘道。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="306"/>
        <source>Optional fields: maxInputTokens / maxOutputTokens are token limits; supportsToolCall / supportsImages / supportsReasoning describe model capabilities. Configure reasoning only when supported by the provider. Do not copy another model&apos;s limits or declare unsupported capabilities. API keys in this file are plaintext: protect it and do not share or commit it.</source>
        <translation>選用欄位：maxInputTokens / maxOutputTokens 表示 Token 上限；supportsToolCall / supportsImages / supportsReasoning 表示工具呼叫、圖片輸入和推理能力。僅在供應商支援時設定 reasoning。不要照抄其他模型的上限或宣告不支援的能力。此檔案中的 API 金鑰為明文，請妥善保護，不要分享或提交到程式碼儲存庫。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="44"/>
        <source>CLI installation is in progress. Wait before exiting.</source>
        <translation>CLI 正在安裝，請等待安裝結束後退出。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="49"/>
        <source>Running · </source>
        <translation>執行中 · </translation>
    </message>
    <message>
        <location filename="main.cpp" line="49"/>
        <source>Gateway stopped</source>
        <translation>閘道已停止</translation>
    </message>
    <message>
        <location filename="main.cpp" line="49"/>
        <source>Status unknown: port conflict, authentication failure, or network error.</source>
        <translation>無法確認狀態：連接埠衝突、驗證失敗或網路異常。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="95"/>
        <source>Gateway API key cannot be empty.</source>
        <translation>閘道 API 金鑰不能為空。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="96"/>
        <source>Wait for CLI installation before starting the gateway.</source>
        <translation>請等待 CLI 安裝完成後再啟動閘道。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="98"/>
        <source>Add and select a plugin on the Plugins page first.</source>
        <translation>請先在外掛頁新增並選擇外掛。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="102"/>
        <source>Gateway or plugin file not found. Run build.bat first.</source>
        <translation>找不到閘道或外掛檔案，請先執行 build.bat。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="104"/>
        <source>Port is already in use. A second gateway will not be started.</source>
        <translation>連接埠已被佔用，不會啟動第二個閘道。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="118"/>
        <source>Failed to start the gateway. Check the executable and CLI installation.</source>
        <translation>閘道啟動失敗，請檢查執行檔及 CLI 安裝。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="122"/>
        <source>Starting the gateway. Exiting the control panel will not stop it.</source>
        <translation>正在啟動閘道；結束控制台不會停止閘道。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="128"/>
        <source>Gateway stop requested. Active requests will be interrupted.</source>
        <translation>已請求停止閘道，進行中的請求將中斷。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="128"/>
        <source>Failed to stop. Check the gateway status.</source>
        <translation>停止失敗，請檢查閘道狀態。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="133"/>
        <source>Wait for CLI installation before installing updates.</source>
        <translation>請等待 CLI 安裝完成後再安裝更新。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="135"/>
        <source>Install Update</source>
        <translation>安裝更新</translation>
    </message>
    <message>
        <location filename="main.cpp" line="135"/>
        <source>Installation will stop the gateway and exit the control panel. Continue?</source>
        <translation>安裝將停止閘道並結束控制台。繼續？</translation>
    </message>
    <message>
        <location filename="main.cpp" line="142"/>
        <source>Could not safely stop the gateway. Update canceled.</source>
        <translation>無法安全停止閘道，已取消更新。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="146"/>
        <source>Failed to start the external updater.</source>
        <translation>無法啟動外部更新程式。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="159"/>
        <source>Wait for CLI installation to finish before exiting.</source>
        <translation>請等待 CLI 安裝完成後再退出。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="164"/>
        <location filename="main.cpp" line="374"/>
        <source>1bot · Gateway Control Panel</source>
        <translation>1bot · 閘道控制台</translation>
    </message>
    <message>
        <location filename="main.cpp" line="178"/>
        <location filename="main.cpp" line="375"/>
        <source>1bot / Local API Gateway</source>
        <translation>1bot / 本機 API 閘道</translation>
    </message>
    <message>
        <location filename="main.cpp" line="49"/>
        <location filename="main.cpp" line="181"/>
        <source>Checking gateway...</source>
        <translation>正在檢查閘道…</translation>
    </message>
    <message>
        <location filename="main.cpp" line="188"/>
        <location filename="main.cpp" line="376"/>
        <source>Control panel navigation</source>
        <translation>控制台頁面導覽</translation>
    </message>
    <message>
        <location filename="main.cpp" line="193"/>
        <location filename="main.cpp" line="229"/>
        <location filename="main.cpp" line="377"/>
        <location filename="main.cpp" line="385"/>
        <source>Gateway</source>
        <translation>閘道管理</translation>
    </message>
    <message>
        <location filename="main.cpp" line="194"/>
        <location filename="main.cpp" line="377"/>
        <source>Plugins</source>
        <translation>外掛</translation>
    </message>
    <message>
        <location filename="main.cpp" line="195"/>
        <location filename="main.cpp" line="378"/>
        <source>Agent Configuration</source>
        <translation>Agent 設定</translation>
    </message>
    <message>
        <location filename="main.cpp" line="196"/>
        <location filename="main.cpp" line="311"/>
        <location filename="main.cpp" line="378"/>
        <location filename="main.cpp" line="394"/>
        <source>Software Updates</source>
        <translation>軟體更新</translation>
    </message>
    <message>
        <location filename="main.cpp" line="237"/>
        <location filename="main.cpp" line="389"/>
        <source>Edit this plugin&apos;s CLI command or path on the Plugins page.</source>
        <translation>在外掛頁編輯此插件的 CLI 指令或路徑。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="253"/>
        <location filename="main.cpp" line="386"/>
        <source>Gateway executable</source>
        <translation>閘道執行檔</translation>
    </message>
    <message>
        <location filename="main.cpp" line="253"/>
        <location filename="main.cpp" line="386"/>
        <source>Upstream plugin</source>
        <translation>上游外掛</translation>
    </message>
    <message>
        <location filename="main.cpp" line="254"/>
        <location filename="main.cpp" line="387"/>
        <source>CLI executable / path</source>
        <translation>CLI 程式 / 路徑</translation>
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
        <translation>本機連接埠</translation>
    </message>
    <message>
        <location filename="main.cpp" line="255"/>
        <location filename="main.cpp" line="388"/>
        <source>Gateway API key</source>
        <translation>閘道 API 金鑰</translation>
    </message>
    <message>
        <location filename="main.cpp" line="258"/>
        <location filename="main.cpp" line="364"/>
        <location filename="main.cpp" line="390"/>
        <location filename="main.cpp" line="406"/>
        <source>Start Gateway</source>
        <translation>啟動閘道</translation>
    </message>
    <message>
        <location filename="main.cpp" line="258"/>
        <location filename="main.cpp" line="365"/>
        <location filename="main.cpp" line="391"/>
        <location filename="main.cpp" line="407"/>
        <source>Stop Gateway</source>
        <translation>停止閘道</translation>
    </message>
    <message>
        <location filename="main.cpp" line="261"/>
        <location filename="main.cpp" line="392"/>
        <source>Copy Gateway Key</source>
        <translation>複製閘道金鑰</translation>
    </message>
    <message>
        <location filename="main.cpp" line="313"/>
        <location filename="main.cpp" line="395"/>
        <source>Current version</source>
        <translation>目前版本</translation>
    </message>
    <message>
        <location filename="main.cpp" line="292"/>
        <source>Manual Agent Configuration

This page is read-only. 1bot does not read or modify agent configuration files.

1. Close the agent and back up its configuration yourself.
2. Open the file below in a text editor. Merge the example fields into the existing JSON; do not replace unrelated settings.
3. Replace SERVICE_URL, API_KEY and MODEL_ID with values supported by your service. Remove model if not needed. Protect API keys and never commit them to a repository.
4. Save the file and restart the agent. To undo changes, restore your own backup.</source>
        <translation>手動設定 Agent

此頁面僅提供唯讀說明。1bot 不讀取或修改 Agent 設定檔。

1. 關閉 Agent，自行備份設定檔。
2. 用文字編輯器開啟下方檔案，將範例欄位合併到現有 JSON 中，不要覆蓋其他設定。
3. 將 SERVICE_URL、API_KEY 和 MODEL_ID 替換為服務支援的位址、金鑰和模型；不需指定模型時刪除 model。妥善保管金鑰，不要提交到程式碼儲存庫。
4. 儲存檔案並重新啟動 Agent。如需撤銷修改，請自行還原備份。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="285"/>
        <source>Selected model: </source>
        <translation>已選擇模型：</translation>
    </message>
    <message>
        <location filename="main.cpp" line="293"/>
        <source>Compatibility: the current gateway provides OpenAI Chat Completions, not Anthropic Messages. Claude Code requires a compatible Anthropic service. CodeBuddy endpoints must match its CLI version and protocol. Do not point the gateway&apos;s upstream CLI back to the same gateway; this creates a request loop.</source>
        <translation>相容性：目前閘道提供 OpenAI Chat Completions，不提供 Anthropic Messages。Claude Code 需要相容 Anthropic 協定的服務。CodeBuddy 的服務端點必須符合 CLI 版本和協定。不要讓閘道上游 CLI 指向同一個閘道，否則會形成循環請求。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="294"/>
        <source>User configuration paths are shown below. Project settings, environment variables or command-line options may override them. Examples are templates, not a guarantee of compatibility.</source>
        <translation>下方顯示使用者層級設定路徑。專案設定、環境變數或命令列選項可能覆蓋這些設定。範例僅為範本，不保證協定相容。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="295"/>
        <source>CodeBuddy uses the top-level endpoint setting. Check the documentation for your installed CLI version:</source>
        <translation>CodeBuddy 使用頂層 endpoint 設定。請核對已安裝 CLI 版本的文件：</translation>
    </message>
    <message>
        <location filename="main.cpp" line="315"/>
        <location filename="main.cpp" line="398"/>
        <source>System default</source>
        <translation>跟隨系統</translation>
    </message>
    <message>
        <location filename="main.cpp" line="323"/>
        <location filename="main.cpp" line="395"/>
        <source>Language</source>
        <translation>語言</translation>
    </message>
    <message>
        <location filename="main.cpp" line="333"/>
        <location filename="main.cpp" line="399"/>
        <source>1bot HTTPS update endpoint</source>
        <translation>1bot HTTPS 更新端點</translation>
    </message>
    <message>
        <location filename="main.cpp" line="334"/>
        <location filename="main.cpp" line="400"/>
        <source>Check at startup and every 6 hours</source>
        <translation>啟動時及每 6 小時自動檢查</translation>
    </message>
    <message>
        <location filename="main.cpp" line="336"/>
        <location filename="main.cpp" line="401"/>
        <source>Check for Updates</source>
        <translation>檢查更新</translation>
    </message>
    <message>
        <location filename="main.cpp" line="337"/>
        <location filename="main.cpp" line="396"/>
        <source>Update URL</source>
        <translation>更新網址</translation>
    </message>
    <message>
        <location filename="main.cpp" line="338"/>
        <location filename="main.cpp" line="396"/>
        <source>Check policy</source>
        <translation>檢查策略</translation>
    </message>
    <message>
        <location filename="main.cpp" line="344"/>
        <location filename="main.cpp" line="402"/>
        <source>You will be asked before downloading a new version. Installation requires confirmation and stops the gateway.</source>
        <translation>下載新版本前會詢問；安裝需確認並停止閘道。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="348"/>
        <location filename="main.cpp" line="403"/>
        <source>Closing the window minimizes it to the tray. The gateway runs independently.</source>
        <translation>關閉視窗後縮小至系統匣，閘道獨立執行。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="353"/>
        <source>New Version </source>
        <translation>發現新版本 </translation>
    </message>
    <message>
        <location filename="main.cpp" line="353"/>
        <source>
Download the update?</source>
        <translation>
下載更新？</translation>
    </message>
    <message>
        <location filename="main.cpp" line="363"/>
        <location filename="main.cpp" line="405"/>
        <source>Open Control Panel</source>
        <translation>開啟控制台</translation>
    </message>
    <message>
        <location filename="main.cpp" line="367"/>
        <location filename="main.cpp" line="408"/>
        <source>Exit Control Panel (Keep Gateway Running)</source>
        <translation>結束控制台（保留閘道）</translation>
    </message>
    <message>
        <location filename="main.cpp" line="431"/>
        <source>The control panel is already running. Open it from the system tray.</source>
        <translation>控制台已在執行，請從系統匣開啟。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="7"/>
        <location filename="modelspage.cpp" line="39"/>
        <source>Models Supported by the Current Plugin</source>
        <translation>目前外掛支援的模型</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="9"/>
        <location filename="modelspage.cpp" line="40"/>
        <source>Refresh List</source>
        <translation>重新整理清單</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="14"/>
        <source>Start the gateway first.</source>
        <translation>請先啟動閘道。</translation>
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
        <translation>所屬外掛</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="28"/>
        <location filename="modelspage.cpp" line="42"/>
        <source>This catalog comes from the running gateway, not the default model argument. This page does not switch models.</source>
        <translation>清單來自執行中的閘道，而非預設模型參數。此頁不切換模型。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="67"/>
        <source>Click Refresh to retrieve the current plugin&apos;s models.</source>
        <translation>按重新整理以取得目前外掛的模型清單。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="44"/>
        <location filename="modelspage.cpp" line="67"/>
        <source>The gateway is offline or cannot be authenticated. Start it and check the connection on the Gateway page.</source>
        <translation>閘道未執行或無法驗證，請在閘道頁啟動並檢查連線。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="43"/>
        <location filename="modelspage.cpp" line="81"/>
        <source>Fetching model catalog...</source>
        <translation>正在查詢模型清單…</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="104"/>
        <source>This plugin or CLI version does not support model discovery.</source>
        <translation>此外掛或 CLI 版本不支援模型探索。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="105"/>
        <source>Authentication failed. Check the gateway API key.</source>
        <translation>驗證失敗，請檢查閘道 API 金鑰。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="108"/>
        <source>The request timed out or exceeded the response limit. Try refreshing later.</source>
        <translation>查詢逾時或回應超出限制，請稍後重新整理。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="109"/>
        <source>Plugin query failed. Check the CLI status and retry.</source>
        <translation>外掛查詢失敗，請檢查 CLI 狀態後重試。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="110"/>
        <source>Model endpoint unavailable. Update the gateway and plugins.</source>
        <translation>模型端點無法使用，請更新閘道與外掛。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="111"/>
        <source>Could not retrieve models. Check the gateway connection and retry.</source>
        <translation>無法取得模型，請檢查閘道連線後重試。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="120"/>
        <source>The gateway returned an invalid model catalog.</source>
        <translation>閘道回傳無效的模型清單。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="130"/>
        <source>The catalog contains invalid entries. Results were not displayed.</source>
        <translation>模型清單含無效項目，未顯示結果。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="46"/>
        <location filename="modelspage.cpp" line="142"/>
        <source>The current plugin returned an empty model catalog.</source>
        <translation>目前外掛回傳空模型清單。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="47"/>
        <location filename="modelspage.cpp" line="143"/>
        <source>%1 models · Updated at %2</source>
        <translation>共 %1 個模型 · 更新於 %2</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="43"/>
        <location filename="pluginspage.cpp" line="145"/>
        <source>Application Plugins</source>
        <translation>應用程式外掛</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="46"/>
        <location filename="pluginspage.cpp" line="146"/>
        <source>Plugins must implement 1bot stdio JSON-RPC and accept gateway arguments such as --bin / --model. Installation only installs the CLI dependency, not a plugin. Relative plugin paths are resolved from the gateway directory.</source>
        <translation>外掛須實作 1bot stdio JSON-RPC 並接受 --bin / --model 等參數。安裝只安裝 CLI 相依套件，不會產生外掛。相對路徑以閘道目錄為基準。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="54"/>
        <location filename="pluginspage.cpp" line="147"/>
        <source>Add Plugin</source>
        <translation>新增外掛</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="55"/>
        <location filename="pluginspage.cpp" line="148"/>
        <source>Delete Configuration</source>
        <translation>刪除設定</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="66"/>
        <location filename="pluginspage.cpp" line="152"/>
        <source>Enter the official PowerShell installation command. Unverified CLI commands are left blank.</source>
        <translation>填入官方 PowerShell 安裝指令，未確認的 CLI 指令留空。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="69"/>
        <location filename="pluginspage.cpp" line="149"/>
        <source>Browse</source>
        <translation>瀏覽</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="71"/>
        <location filename="pluginspage.cpp" line="153"/>
        <source>Display name</source>
        <translation>顯示名稱</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="72"/>
        <location filename="pluginspage.cpp" line="153"/>
        <source>Plugin executable</source>
        <translation>外掛執行檔</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="73"/>
        <location filename="pluginspage.cpp" line="154"/>
        <source>CLI command / path</source>
        <translation>CLI 指令 / 路徑</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="74"/>
        <location filename="pluginspage.cpp" line="154"/>
        <source>CLI dependency installation command</source>
        <translation>CLI 相依套件安裝指令</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="76"/>
        <location filename="pluginspage.cpp" line="150"/>
        <source>Save Configuration</source>
        <translation>儲存設定</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="77"/>
        <location filename="pluginspage.cpp" line="151"/>
        <source>Install Dependencies</source>
        <translation>安裝相依套件</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="81"/>
        <source>Select a plugin to view and edit its CLI installation command.</source>
        <translation>選擇外掛以檢視及編輯 CLI 安裝指令。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="91"/>
        <source>Select a 1bot-compatible Plugin</source>
        <translation>選擇相容 1bot 的外掛</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="91"/>
        <source>Executables (*.exe)</source>
        <translation>執行檔 (*.exe)</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="99"/>
        <source>Save the new plugin to add it to the Gateway page selector.</source>
        <translation>儲存新外掛後會加入閘道頁的選擇清單。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="103"/>
        <source>Delete Plugin Configuration</source>
        <translation>刪除外掛設定</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="103"/>
        <source>Only the configuration will be removed. The CLI and plugin files will remain. Continue?</source>
        <translation>只移除設定，不移除 CLI 或外掛檔案。繼續？</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="119"/>
        <source>Failed to start installation: </source>
        <translation>無法啟動安裝：</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="126"/>
        <source>Installation command completed successfully. CLI login may still be required. Restart the control panel if PATH changed.</source>
        <translation>安裝指令成功結束，可能仍需 CLI 登入；若 PATH 變更，請重啟控制台。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="127"/>
        <source>Installation failed (exit code %1). Check the output.</source>
        <translation>安裝失敗（結束代碼 %1），請查看輸出。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="185"/>
        <source>Requires CodeBuddy Code CLI and Node.js 18.20+ for npm installation. Official package: @tencent-ai/codebuddy-code.</source>
        <translation>需要 CodeBuddy Code CLI；npm 安裝需 Node.js 18.20+。官方套件：@tencent-ai/codebuddy-code。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="187"/>
        <source>Requires ZCode CLI (zcode-app-cli) and Node.js/npm. Follow the CLI instructions to sign in after installation.</source>
        <translation>需要 ZCode CLI（zcode-app-cli）及 Node.js/npm，安裝後依 CLI 指引登入。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="189"/>
        <source>Requires Codex CLI and Node.js/npm. Follow the official sign-in instructions after installation.</source>
        <translation>需要 Codex CLI 及 Node.js/npm，安裝後依官方指引登入。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="190"/>
        <source>Configure the CLI dependency and its official installation command.</source>
        <translation>請設定 CLI 相依套件及其官方安裝指令。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="201"/>
        <source>Failed to persist configuration. Check user permissions.</source>
        <translation>無法儲存設定，請檢查使用者權限。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="207"/>
        <source>Name, plugin executable, and CLI command cannot be empty.</source>
        <translation>名稱、外掛執行檔及 CLI 指令不能為空。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="218"/>
        <source>Configuration saved. Changes take effect after restarting the gateway; the running gateway is unchanged.</source>
        <translation>設定已儲存，重啟閘道後生效；執行中的閘道不受影響。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="231"/>
        <source>No installation command is configured. Enter one from the official CLI documentation.</source>
        <translation>尚未設定安裝指令，請參考 CLI 官方文件填入。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="232"/>
        <source>Confirm Installation Command</source>
        <translation>確認執行安裝指令</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="233"/>
        <source>The following PowerShell command will run as the current Windows user and may download software or change the environment. Run only trusted official commands.

</source>
        <translation>以下 PowerShell 指令將以目前 Windows 使用者權限執行，可能下載軟體或修改環境。只執行信任的官方指令。

</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="237"/>
        <source>Working directory: </source>
        <translation>工作目錄：</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="158"/>
        <location filename="pluginspage.cpp" line="240"/>
        <source>Installing. Do not exit the control panel. Run interactive or administrator installations in a terminal.</source>
        <translation>安裝中，請勿結束控制台。需互動或管理員權限的安裝請在終端執行。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="21"/>
        <source>Update file exceeds the size limit.</source>
        <translation>更新檔案超過大小限制。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="24"/>
        <source>Cannot write update data.</source>
        <translation>無法寫入更新資料。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="30"/>
        <source>Update downloads require HTTPS.</source>
        <translation>更新下載必須使用 HTTPS。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="32"/>
        <source>Cannot initialize libcurl.</source>
        <translation>無法初始化 libcurl。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="35"/>
        <source>The updater requires libcurl with the OpenSSL backend.</source>
        <translation>更新程式需要使用 OpenSSL 後端的 libcurl。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="37"/>
        <source>Missing curl-ca-bundle.crt CA bundle.</source>
        <translation>缺少 curl-ca-bundle.crt 根憑證套件。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="39"/>
        <source>Cannot create an HTTPS request.</source>
        <translation>無法建立 HTTPS 請求。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="64"/>
        <source>Cannot create request headers.</source>
        <translation>無法建立請求標頭。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="71"/>
        <source>Cannot configure secure HTTPS request options.</source>
        <translation>無法設定安全 HTTPS 請求選項。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="75"/>
        <source>HTTPS request failed: </source>
        <translation>HTTPS 請求失敗：</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="78"/>
        <source>Update server HTTP status %1 (redirects are not accepted).</source>
        <translation>更新伺服器 HTTP 狀態 %1（不接受重新導向）。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="84"/>
        <source>Configure a dedicated 1bot HTTPS update URL.</source>
        <translation>請設定 1bot 專用 HTTPS 更新網址。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="101"/>
        <source>Invalid update manifest.</source>
        <translation>無效的更新清單。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="106"/>
        <source>No updates available.</source>
        <translation>目前沒有更新。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="120"/>
        <source>The update manifest contains unsafe files, invalid sizes, or missing SHA-256 hashes.</source>
        <translation>更新清單包含不安全檔案、無效大小或缺少 SHA-256。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="125"/>
        <source>The update package is missing 1bot-control.exe.</source>
        <translation>更新套件缺少 1bot-control.exe。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="141"/>
        <source>Cannot create the update directory.</source>
        <translation>無法建立更新目錄。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="149"/>
        <source>Cannot save the update manifest.</source>
        <translation>無法儲存更新清單。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="154"/>
        <source>Downloading update %1/%2: %3</source>
        <translation>正在下載更新 %1/%2：%3</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="164"/>
        <source>Cannot create the download file.</source>
        <translation>無法建立下載檔案。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="168"/>
        <source>Download size mismatch or save failure.</source>
        <translation>下載大小不符或無法儲存。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="173"/>
        <source>SHA-256 verification failed. The download was deleted.</source>
        <translation>SHA-256 驗證失敗，已刪除下載檔案。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="33"/>
        <source>Cannot read file: </source>
        <translation>無法讀取檔案：</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="39"/>
        <source>Cannot write atomically: </source>
        <translation>無法原子寫入：</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="50"/>
        <source>Invalid request.</source>
        <translation>無效請求。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="53"/>
        <source>The configuration path must be absolute and cannot be a symbolic link.</source>
        <translation>設定路徑必須為絕對路徑且不能為符號連結。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="56"/>
        <source>Another control panel is modifying this configuration.</source>
        <translation>另一個控制台正在修改此設定。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="64"/>
        <source>Cannot decrypt the backup.</source>
        <translation>無法解密備份。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="67"/>
        <source>The configuration was changed externally. Overwrite refused; merge changes manually first.</source>
        <translation>設定已被外部修改，拒絕覆寫；請先手動合併。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="70"/>
        <source>No backup is available to restore.</source>
        <translation>沒有可還原的備份。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="72"/>
        <source>Cannot remove the newly created configuration.</source>
        <translation>無法刪除新建設定。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="73"/>
        <source>Configuration restored, but backup cleanup failed.</source>
        <translation>設定已還原，但備份清理失敗。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="76"/>
        <source>Only strict JSON object configurations are supported. The original file was not changed.</source>
        <translation>僅支援嚴格 JSON 物件設定，原始檔案未變更。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="78"/>
        <source>env must be a JSON object.</source>
        <translation>env 必須為 JSON 物件。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="89"/>
        <source>Unsupported agent.</source>
        <translation>不支援的 Agent。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="93"/>
        <source>The API URL must use HTTPS or local HTTP.</source>
        <translation>API 網址必須使用 HTTPS 或本機 HTTP。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="94"/>
        <source>API key cannot be empty.</source>
        <translation>API 金鑰不能為空。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="107"/>
        <source>Cannot encrypt the backup.</source>
        <translation>無法加密備份。</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="10"/>
        <location filename="accesscard.cpp" line="111"/>
        <source>Connect</source>
        <translation>接入</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="65"/>
        <location filename="accesscard.cpp" line="124"/>
        <source>Hide</source>
        <translation>隱藏</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="65"/>
        <location filename="accesscard.cpp" line="124"/>
        <source>Show</source>
        <translation>顯示</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="112"/>
        <location filename="accesscard.cpp" line="113"/>
        <source>Protocol:</source>
        <translation>協定：</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="114"/>
        <source>Select a protocol to view matching request examples. These examples use non-streaming requests.</source>
        <translation>選擇協定查看對應的請求範例。以下範例均使用非串流請求。</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="115"/>
        <source>Base URL</source>
        <translation>基礎位址</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="116"/>
        <source>API Key</source>
        <translation>API 金鑰</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="117"/>
        <source>Model</source>
        <translation>模型</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="118"/>
        <location filename="accesscard.cpp" line="119"/>
        <location filename="accesscard.cpp" line="120"/>
        <location filename="accesscard.cpp" line="121"/>
        <location filename="accesscard.cpp" line="122"/>
        <location filename="accesscard.cpp" line="123"/>
        <location filename="accesscard.cpp" line="130"/>
        <location filename="accesscard.cpp" line="131"/>
        <source>Copy</source>
        <translation>複製</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="125"/>
        <source>PowerShell</source>
        <translation>PowerShell</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="126"/>
        <source>CMD</source>
        <translation>CMD</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="127"/>
        <source>curl</source>
        <translation>curl</translation>
    </message>
</context>
</TS>
