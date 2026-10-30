<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE TS>
<TS version="2.1" language="ja_JP" sourcelanguage="en_US">
<context>
    <name>QObject</name>
    <message>
        <location filename="main.cpp" line="302"/>
        <source>CodeBuddy custom providers (models.json)</source>
        <translation>CodeBuddy カスタムプロバイダー（models.json）</translation>
    </message>
    <message>
        <location filename="main.cpp" line="304"/>
        <source>Back up models.json, then append an entry to its models array without removing existing providers. vendor identifies the provider; name is the display label; id must match a model supported by the service. url is the full API endpoint, including /v1/chat/completions, not just the base URL. Replace API_KEY with your service key. The local address below assumes port 8787; change it to your gateway port. Set token limits and capability flags only according to actual model and gateway support. Save and restart CodeBuddy, then select the custom model in its model selector. Never route an upstream CodeBuddy CLI back to its own gateway.</source>
        <translation>models.json をバックアップし、既存のプロバイダーを削除せずに models 配列へ項目を追加します。vendor はプロバイダー、name は表示名、id はサービスで利用できるモデル ID です。url にはベース URL だけでなく /v1/chat/completions を含む完全な API パスを指定します。API_KEY をサービスのキーに置き換えてください。以下はポート 8787 の例です。実際のゲートウェイのポートに変更してください。トークン上限や機能フラグはモデルとゲートウェイの実際の対応状況に合わせます。保存して CodeBuddy を再起動し、モデル選択画面でカスタムモデルを選びます。上流の CodeBuddy CLI を自身のゲートウェイに接続しないでください。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="306"/>
        <source>Optional fields: maxInputTokens / maxOutputTokens are token limits; supportsToolCall / supportsImages / supportsReasoning describe model capabilities. Configure reasoning only when supported by the provider. Do not copy another model&apos;s limits or declare unsupported capabilities. API keys in this file are plaintext: protect it and do not share or commit it.</source>
        <translation>任意項目：maxInputTokens / maxOutputTokens はトークン上限、supportsToolCall / supportsImages / supportsReasoning はツール呼び出し・画像入力・推論機能です。reasoning はプロバイダーが対応する場合のみ設定します。他のモデルの上限をコピーしたり、未対応の機能を宣言したりしないでください。API キーは平文で保存されるため、ファイルを保護し、共有やコミットを避けてください。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="44"/>
        <source>CLI installation is in progress. Wait before exiting.</source>
        <translation>CLI をインストール中です。完了後に終了してください。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="49"/>
        <source>Running · </source>
        <translation>稼働中 · </translation>
    </message>
    <message>
        <location filename="main.cpp" line="49"/>
        <source>Gateway stopped</source>
        <translation>ゲートウェイ停止</translation>
    </message>
    <message>
        <location filename="main.cpp" line="49"/>
        <source>Status unknown: port conflict, authentication failure, or network error.</source>
        <translation>状態不明：ポート競合、認証失敗、またはネットワークエラー。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="95"/>
        <source>Gateway API key cannot be empty.</source>
        <translation>ゲートウェイ API キーを入力してください。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="96"/>
        <source>Wait for CLI installation before starting the gateway.</source>
        <translation>CLI のインストール完了後にゲートウェイを起動してください。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="98"/>
        <source>Add and select a plugin on the Plugins page first.</source>
        <translation>まずプラグインページでプラグインを追加して選択してください。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="102"/>
        <source>Gateway or plugin file not found. Run build.bat first.</source>
        <translation>ゲートウェイまたはプラグインが見つかりません。先に build.bat を実行してください。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="104"/>
        <source>Port is already in use. A second gateway will not be started.</source>
        <translation>ポートは使用中です。2つ目のゲートウェイは起動しません。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="118"/>
        <source>Failed to start the gateway. Check the executable and CLI installation.</source>
        <translation>起動に失敗しました。実行ファイルと CLI のインストールを確認してください。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="122"/>
        <source>Starting the gateway. Exiting the control panel will not stop it.</source>
        <translation>ゲートウェイを起動中です。管理画面を終了しても停止しません。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="128"/>
        <source>Gateway stop requested. Active requests will be interrupted.</source>
        <translation>ゲートウェイの停止を要求しました。実行中のリクエストは中断されます。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="128"/>
        <source>Failed to stop. Check the gateway status.</source>
        <translation>停止に失敗しました。ゲートウェイの状態を確認してください。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="133"/>
        <source>Wait for CLI installation before installing updates.</source>
        <translation>CLI のインストール完了後に更新をインストールしてください。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="135"/>
        <source>Install Update</source>
        <translation>更新をインストール</translation>
    </message>
    <message>
        <location filename="main.cpp" line="135"/>
        <source>Installation will stop the gateway and exit the control panel. Continue?</source>
        <translation>ゲートウェイを停止し、管理画面を終了してインストールします。続行しますか？</translation>
    </message>
    <message>
        <location filename="main.cpp" line="142"/>
        <source>Could not safely stop the gateway. Update canceled.</source>
        <translation>ゲートウェイを安全に停止できません。更新を中止しました。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="146"/>
        <source>Failed to start the external updater.</source>
        <translation>外部アップデーターを起動できません。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="159"/>
        <source>Wait for CLI installation to finish before exiting.</source>
        <translation>CLI のインストール完了後に終了してください。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="164"/>
        <location filename="main.cpp" line="374"/>
        <source>1bot · Gateway Control Panel</source>
        <translation>1bot · ゲートウェイ管理</translation>
    </message>
    <message>
        <location filename="main.cpp" line="178"/>
        <location filename="main.cpp" line="375"/>
        <source>1bot / Local API Gateway</source>
        <translation>1bot / ローカル API ゲートウェイ</translation>
    </message>
    <message>
        <location filename="main.cpp" line="49"/>
        <location filename="main.cpp" line="181"/>
        <source>Checking gateway...</source>
        <translation>ゲートウェイを確認中...</translation>
    </message>
    <message>
        <location filename="main.cpp" line="188"/>
        <location filename="main.cpp" line="376"/>
        <source>Control panel navigation</source>
        <translation>管理画面ナビゲーション</translation>
    </message>
    <message>
        <location filename="main.cpp" line="193"/>
        <location filename="main.cpp" line="229"/>
        <location filename="main.cpp" line="377"/>
        <location filename="main.cpp" line="385"/>
        <source>Gateway</source>
        <translation>ゲートウェイ</translation>
    </message>
    <message>
        <location filename="main.cpp" line="194"/>
        <location filename="main.cpp" line="377"/>
        <source>Plugins</source>
        <translation>プラグイン</translation>
    </message>
    <message>
        <location filename="main.cpp" line="195"/>
        <location filename="main.cpp" line="378"/>
        <source>Agent Configuration</source>
        <translation>エージェント設定</translation>
    </message>
    <message>
        <location filename="main.cpp" line="196"/>
        <location filename="main.cpp" line="311"/>
        <location filename="main.cpp" line="378"/>
        <location filename="main.cpp" line="394"/>
        <source>Software Updates</source>
        <translation>ソフトウェア更新</translation>
    </message>
    <message>
        <location filename="main.cpp" line="237"/>
        <location filename="main.cpp" line="389"/>
        <source>Edit this plugin&apos;s CLI command or path on the Plugins page.</source>
        <translation>プラグインページで CLI コマンドまたはパスを編集してください。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="253"/>
        <location filename="main.cpp" line="386"/>
        <source>Gateway executable</source>
        <translation>ゲートウェイ実行ファイル</translation>
    </message>
    <message>
        <location filename="main.cpp" line="253"/>
        <location filename="main.cpp" line="386"/>
        <source>Upstream plugin</source>
        <translation>上流プラグイン</translation>
    </message>
    <message>
        <location filename="main.cpp" line="254"/>
        <location filename="main.cpp" line="387"/>
        <source>CLI executable / path</source>
        <translation>CLI 実行ファイル / パス</translation>
    </message>
    <message>
        <location filename="main.cpp" line="254"/>
        <location filename="main.cpp" line="387"/>
        <source>Model (optional)</source>
        <translation>モデル（任意）</translation>
    </message>
    <message>
        <location filename="main.cpp" line="255"/>
        <location filename="main.cpp" line="388"/>
        <source>Local port</source>
        <translation>ローカルポート</translation>
    </message>
    <message>
        <location filename="main.cpp" line="255"/>
        <location filename="main.cpp" line="388"/>
        <source>Gateway API key</source>
        <translation>ゲートウェイ API キー</translation>
    </message>
    <message>
        <location filename="main.cpp" line="258"/>
        <location filename="main.cpp" line="364"/>
        <location filename="main.cpp" line="390"/>
        <location filename="main.cpp" line="406"/>
        <source>Start Gateway</source>
        <translation>ゲートウェイを起動</translation>
    </message>
    <message>
        <location filename="main.cpp" line="258"/>
        <location filename="main.cpp" line="365"/>
        <location filename="main.cpp" line="391"/>
        <location filename="main.cpp" line="407"/>
        <source>Stop Gateway</source>
        <translation>ゲートウェイを停止</translation>
    </message>
    <message>
        <location filename="main.cpp" line="261"/>
        <location filename="main.cpp" line="392"/>
        <source>Copy Gateway Key</source>
        <translation>ゲートウェイキーをコピー</translation>
    </message>
    <message>
        <location filename="main.cpp" line="313"/>
        <location filename="main.cpp" line="395"/>
        <source>Current version</source>
        <translation>現在のバージョン</translation>
    </message>
    <message>
        <location filename="main.cpp" line="292"/>
        <source>Manual Agent Configuration

This page is read-only. 1bot does not read or modify agent configuration files.

1. Close the agent and back up its configuration yourself.
2. Open the file below in a text editor. Merge the example fields into the existing JSON; do not replace unrelated settings.
3. Replace SERVICE_URL, API_KEY and MODEL_ID with values supported by your service. Remove model if not needed. Protect API keys and never commit them to a repository.
4. Save the file and restart the agent. To undo changes, restore your own backup.</source>
        <translation>エージェントの手動設定

このページは読み取り専用です。1bot はエージェントの設定ファイルを読み取ったり変更したりしません。

1. エージェントを終了し、設定を自分でバックアップしてください。
2. 下記のファイルをテキストエディターで開き、例の項目を既存の JSON に統合してください。他の設定を上書きしないでください。
3. SERVICE_URL、API_KEY、MODEL_ID をサービスで利用できる値に置き換えてください。モデル指定が不要なら model を削除してください。キーを保護し、リポジトリにコミットしないでください。
4. 保存してエージェントを再起動してください。元に戻す場合は自分のバックアップを復元してください。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="285"/>
        <source>Selected model: </source>
        <translation>選択したモデル：</translation>
    </message>
    <message>
        <location filename="main.cpp" line="293"/>
        <source>Compatibility: the current gateway provides OpenAI Chat Completions, not Anthropic Messages. Claude Code requires a compatible Anthropic service. CodeBuddy endpoints must match its CLI version and protocol. Do not point the gateway&apos;s upstream CLI back to the same gateway; this creates a request loop.</source>
        <translation>互換性：現在のゲートウェイは OpenAI Chat Completions を提供しますが、Anthropic Messages は提供しません。Claude Code には Anthropic 互換サービスが必要です。CodeBuddy のエンドポイントは CLI のバージョンとプロトコルに合わせてください。上流 CLI を同じゲートウェイに接続するとリクエストが循環します。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="294"/>
        <source>User configuration paths are shown below. Project settings, environment variables or command-line options may override them. Examples are templates, not a guarantee of compatibility.</source>
        <translation>ユーザー設定のパスを以下に示します。プロジェクト設定、環境変数、コマンドラインオプションで上書きされる場合があります。例はテンプレートであり、互換性を保証しません。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="295"/>
        <source>CodeBuddy uses the top-level endpoint setting. Check the documentation for your installed CLI version:</source>
        <translation>CodeBuddy は最上位の endpoint 設定を使用します。インストール済み CLI バージョンのドキュメントを確認してください：</translation>
    </message>
    <message>
        <location filename="main.cpp" line="315"/>
        <location filename="main.cpp" line="398"/>
        <source>System default</source>
        <translation>システム設定</translation>
    </message>
    <message>
        <location filename="main.cpp" line="323"/>
        <location filename="main.cpp" line="395"/>
        <source>Language</source>
        <translation>言語</translation>
    </message>
    <message>
        <location filename="main.cpp" line="333"/>
        <location filename="main.cpp" line="399"/>
        <source>1bot HTTPS update endpoint</source>
        <translation>1bot HTTPS 更新エンドポイント</translation>
    </message>
    <message>
        <location filename="main.cpp" line="334"/>
        <location filename="main.cpp" line="400"/>
        <source>Check at startup and every 6 hours</source>
        <translation>起動時および6時間ごとに確認</translation>
    </message>
    <message>
        <location filename="main.cpp" line="336"/>
        <location filename="main.cpp" line="401"/>
        <source>Check for Updates</source>
        <translation>更新を確認</translation>
    </message>
    <message>
        <location filename="main.cpp" line="337"/>
        <location filename="main.cpp" line="396"/>
        <source>Update URL</source>
        <translation>更新 URL</translation>
    </message>
    <message>
        <location filename="main.cpp" line="338"/>
        <location filename="main.cpp" line="396"/>
        <source>Check policy</source>
        <translation>確認ポリシー</translation>
    </message>
    <message>
        <location filename="main.cpp" line="344"/>
        <location filename="main.cpp" line="402"/>
        <source>You will be asked before downloading a new version. Installation requires confirmation and stops the gateway.</source>
        <translation>新バージョンのダウンロード前に確認します。インストールは確認後にゲートウェイを停止します。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="348"/>
        <location filename="main.cpp" line="403"/>
        <source>Closing the window minimizes it to the tray. The gateway runs independently.</source>
        <translation>ウィンドウを閉じるとトレイに最小化します。ゲートウェイは独立して動作します。</translation>
    </message>
    <message>
        <location filename="main.cpp" line="353"/>
        <source>New Version </source>
        <translation>新バージョン </translation>
    </message>
    <message>
        <location filename="main.cpp" line="353"/>
        <source>
Download the update?</source>
        <translation>
更新をダウンロードしますか？</translation>
    </message>
    <message>
        <location filename="main.cpp" line="363"/>
        <location filename="main.cpp" line="405"/>
        <source>Open Control Panel</source>
        <translation>管理画面を開く</translation>
    </message>
    <message>
        <location filename="main.cpp" line="367"/>
        <location filename="main.cpp" line="408"/>
        <source>Exit Control Panel (Keep Gateway Running)</source>
        <translation>管理画面を終了（ゲートウェイは継続）</translation>
    </message>
    <message>
        <location filename="main.cpp" line="431"/>
        <source>The control panel is already running. Open it from the system tray.</source>
        <translation>管理画面は既に起動しています。システムトレイから開いてください。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="7"/>
        <location filename="modelspage.cpp" line="39"/>
        <source>Models Supported by the Current Plugin</source>
        <translation>現在のプラグインが対応するモデル</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="9"/>
        <location filename="modelspage.cpp" line="40"/>
        <source>Refresh List</source>
        <translation>一覧を更新</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="14"/>
        <source>Start the gateway first.</source>
        <translation>先にゲートウェイを起動してください。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="19"/>
        <location filename="modelspage.cpp" line="41"/>
        <source>Model ID</source>
        <translation>モデル ID</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="19"/>
        <location filename="modelspage.cpp" line="41"/>
        <source>Plugin</source>
        <translation>プラグイン</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="28"/>
        <location filename="modelspage.cpp" line="42"/>
        <source>This catalog comes from the running gateway, not the default model argument. This page does not switch models.</source>
        <translation>実行中のゲートウェイから取得した一覧です。既定モデルの引数ではありません。この画面ではモデルを切り替えません。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="67"/>
        <source>Click Refresh to retrieve the current plugin&apos;s models.</source>
        <translation>更新をクリックして現在のプラグインのモデルを取得してください。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="44"/>
        <location filename="modelspage.cpp" line="67"/>
        <source>The gateway is offline or cannot be authenticated. Start it and check the connection on the Gateway page.</source>
        <translation>ゲートウェイが停止中、または認証できません。ゲートウェイページで起動し接続を確認してください。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="43"/>
        <location filename="modelspage.cpp" line="81"/>
        <source>Fetching model catalog...</source>
        <translation>モデル一覧を取得中...</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="104"/>
        <source>This plugin or CLI version does not support model discovery.</source>
        <translation>このプラグインまたは CLI はモデル一覧の取得に対応していません。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="105"/>
        <source>Authentication failed. Check the gateway API key.</source>
        <translation>認証に失敗しました。ゲートウェイ API キーを確認してください。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="108"/>
        <source>The request timed out or exceeded the response limit. Try refreshing later.</source>
        <translation>タイムアウトまたは応答サイズの上限超過です。後で更新してください。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="109"/>
        <source>Plugin query failed. Check the CLI status and retry.</source>
        <translation>プラグインへの問い合わせに失敗しました。CLI の状態を確認して再試行してください。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="110"/>
        <source>Model endpoint unavailable. Update the gateway and plugins.</source>
        <translation>モデルエンドポイントがありません。ゲートウェイとプラグインを更新してください。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="111"/>
        <source>Could not retrieve models. Check the gateway connection and retry.</source>
        <translation>モデルを取得できません。ゲートウェイ接続を確認して再試行してください。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="120"/>
        <source>The gateway returned an invalid model catalog.</source>
        <translation>ゲートウェイが無効なモデル一覧を返しました。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="130"/>
        <source>The catalog contains invalid entries. Results were not displayed.</source>
        <translation>モデル一覧に無効な項目があります。結果は表示しません。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="46"/>
        <location filename="modelspage.cpp" line="142"/>
        <source>The current plugin returned an empty model catalog.</source>
        <translation>現在のプラグインが空のモデル一覧を返しました。</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="47"/>
        <location filename="modelspage.cpp" line="143"/>
        <source>%1 models · Updated at %2</source>
        <translation>%1 モデル · 更新時刻 %2</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="43"/>
        <location filename="pluginspage.cpp" line="145"/>
        <source>Application Plugins</source>
        <translation>アプリケーションプラグイン</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="46"/>
        <location filename="pluginspage.cpp" line="146"/>
        <source>Plugins must implement 1bot stdio JSON-RPC and accept gateway arguments such as --bin / --model. Installation only installs the CLI dependency, not a plugin. Relative plugin paths are resolved from the gateway directory.</source>
        <translation>プラグインは 1bot stdio JSON-RPC と --bin / --model 等に対応する必要があります。インストール対象は CLI 依存関係のみで、プラグインは生成しません。相対パスはゲートウェイのディレクトリ基準です。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="54"/>
        <location filename="pluginspage.cpp" line="147"/>
        <source>Add Plugin</source>
        <translation>プラグインを追加</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="55"/>
        <location filename="pluginspage.cpp" line="148"/>
        <source>Delete Configuration</source>
        <translation>設定を削除</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="66"/>
        <location filename="pluginspage.cpp" line="152"/>
        <source>Enter the official PowerShell installation command. Unverified CLI commands are left blank.</source>
        <translation>公式の PowerShell インストールコマンドを入力してください。未確認の CLI コマンドは空欄です。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="69"/>
        <location filename="pluginspage.cpp" line="149"/>
        <source>Browse</source>
        <translation>参照</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="71"/>
        <location filename="pluginspage.cpp" line="153"/>
        <source>Display name</source>
        <translation>表示名</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="72"/>
        <location filename="pluginspage.cpp" line="153"/>
        <source>Plugin executable</source>
        <translation>プラグイン実行ファイル</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="73"/>
        <location filename="pluginspage.cpp" line="154"/>
        <source>CLI command / path</source>
        <translation>CLI コマンド / パス</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="74"/>
        <location filename="pluginspage.cpp" line="154"/>
        <source>CLI dependency installation command</source>
        <translation>CLI 依存関係のインストールコマンド</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="76"/>
        <location filename="pluginspage.cpp" line="150"/>
        <source>Save Configuration</source>
        <translation>設定を保存</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="77"/>
        <location filename="pluginspage.cpp" line="151"/>
        <source>Install Dependencies</source>
        <translation>依存関係をインストール</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="81"/>
        <source>Select a plugin to view and edit its CLI installation command.</source>
        <translation>プラグインを選択して CLI インストールコマンドを確認・編集できます。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="91"/>
        <source>Select a 1bot-compatible Plugin</source>
        <translation>1bot 対応プラグインを選択</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="91"/>
        <source>Executables (*.exe)</source>
        <translation>実行ファイル (*.exe)</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="99"/>
        <source>Save the new plugin to add it to the Gateway page selector.</source>
        <translation>新しいプラグインを保存するとゲートウェイページで選択できます。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="103"/>
        <source>Delete Plugin Configuration</source>
        <translation>プラグイン設定を削除</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="103"/>
        <source>Only the configuration will be removed. The CLI and plugin files will remain. Continue?</source>
        <translation>設定のみ削除します。CLI とプラグインファイルは残ります。続行しますか？</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="119"/>
        <source>Failed to start installation: </source>
        <translation>インストールを開始できません: </translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="126"/>
        <source>Installation command completed successfully. CLI login may still be required. Restart the control panel if PATH changed.</source>
        <translation>インストールコマンドが完了しました。CLI ログインが必要な場合があります。PATH が変わった場合は管理画面を再起動してください。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="127"/>
        <source>Installation failed (exit code %1). Check the output.</source>
        <translation>インストール失敗（終了コード %1）。出力を確認してください。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="185"/>
        <source>Requires CodeBuddy Code CLI and Node.js 18.20+ for npm installation. Official package: @tencent-ai/codebuddy-code.</source>
        <translation>CodeBuddy Code CLI が必要です。npm インストールには Node.js 18.20+ が必要です。公式パッケージ: @tencent-ai/codebuddy-code。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="187"/>
        <source>Requires ZCode CLI (zcode-app-cli) and Node.js/npm. Follow the CLI instructions to sign in after installation.</source>
        <translation>ZCode CLI (zcode-app-cli) と Node.js/npm が必要です。インストール後は CLI の案内に従ってログインしてください。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="189"/>
        <source>Requires Codex CLI and Node.js/npm. Follow the official sign-in instructions after installation.</source>
        <translation>Codex CLI と Node.js/npm が必要です。インストール後は公式の案内に従ってログインしてください。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="190"/>
        <source>Configure the CLI dependency and its official installation command.</source>
        <translation>CLI 依存関係と公式インストールコマンドを設定してください。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="201"/>
        <source>Failed to persist configuration. Check user permissions.</source>
        <translation>設定を保存できません。ユーザー権限を確認してください。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="207"/>
        <source>Name, plugin executable, and CLI command cannot be empty.</source>
        <translation>名前、プラグイン実行ファイル、CLI コマンドは必須です。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="218"/>
        <source>Configuration saved. Changes take effect after restarting the gateway; the running gateway is unchanged.</source>
        <translation>設定を保存しました。ゲートウェイの再起動後に反映され、実行中のゲートウェイには影響しません。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="231"/>
        <source>No installation command is configured. Enter one from the official CLI documentation.</source>
        <translation>インストールコマンドが未設定です。公式 CLI 文書から入力してください。</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="232"/>
        <source>Confirm Installation Command</source>
        <translation>インストールコマンドを確認</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="233"/>
        <source>The following PowerShell command will run as the current Windows user and may download software or change the environment. Run only trusted official commands.

</source>
        <translation>以下の PowerShell コマンドを現在の Windows ユーザーとして実行します。ソフトウェアのダウンロードや環境変更を伴う場合があります。信頼できる公式コマンドのみ実行してください。

</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="237"/>
        <source>Working directory: </source>
        <translation>作業ディレクトリ: </translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="158"/>
        <location filename="pluginspage.cpp" line="240"/>
        <source>Installing. Do not exit the control panel. Run interactive or administrator installations in a terminal.</source>
        <translation>インストール中です。管理画面を終了しないでください。対話や管理者権限が必要な場合はターミナルで実行してください。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="21"/>
        <source>Update file exceeds the size limit.</source>
        <translation>更新ファイルがサイズ上限を超えました。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="24"/>
        <source>Cannot write update data.</source>
        <translation>更新データを書き込めません。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="30"/>
        <source>Update downloads require HTTPS.</source>
        <translation>更新のダウンロードには HTTPS が必要です。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="32"/>
        <source>Cannot initialize libcurl.</source>
        <translation>libcurl を初期化できません。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="35"/>
        <source>The updater requires libcurl with the OpenSSL backend.</source>
        <translation>アップデーターには OpenSSL バックエンドの libcurl が必要です。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="37"/>
        <source>Missing curl-ca-bundle.crt CA bundle.</source>
        <translation>CA バンドル curl-ca-bundle.crt がありません。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="39"/>
        <source>Cannot create an HTTPS request.</source>
        <translation>HTTPS リクエストを作成できません。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="64"/>
        <source>Cannot create request headers.</source>
        <translation>リクエストヘッダーを作成できません。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="71"/>
        <source>Cannot configure secure HTTPS request options.</source>
        <translation>安全な HTTPS リクエスト設定を適用できません。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="75"/>
        <source>HTTPS request failed: </source>
        <translation>HTTPS リクエスト失敗: </translation>
    </message>
    <message>
        <location filename="updater.cpp" line="78"/>
        <source>Update server HTTP status %1 (redirects are not accepted).</source>
        <translation>更新サーバーの HTTP ステータス %1（リダイレクト非対応）。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="84"/>
        <source>Configure a dedicated 1bot HTTPS update URL.</source>
        <translation>1bot 専用の HTTPS 更新 URL を設定してください。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="101"/>
        <source>Invalid update manifest.</source>
        <translation>無効な更新マニフェストです。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="106"/>
        <source>No updates available.</source>
        <translation>更新はありません。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="120"/>
        <source>The update manifest contains unsafe files, invalid sizes, or missing SHA-256 hashes.</source>
        <translation>更新マニフェストに安全でないファイル、不正なサイズ、または SHA-256 の欠落があります。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="125"/>
        <source>The update package is missing 1bot-control.exe.</source>
        <translation>更新パッケージに 1bot-control.exe がありません。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="141"/>
        <source>Cannot create the update directory.</source>
        <translation>更新ディレクトリを作成できません。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="149"/>
        <source>Cannot save the update manifest.</source>
        <translation>更新マニフェストを保存できません。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="154"/>
        <source>Downloading update %1/%2: %3</source>
        <translation>更新をダウンロード中 %1/%2: %3</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="164"/>
        <source>Cannot create the download file.</source>
        <translation>ダウンロードファイルを作成できません。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="168"/>
        <source>Download size mismatch or save failure.</source>
        <translation>ダウンロードサイズの不一致、または保存に失敗しました。</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="173"/>
        <source>SHA-256 verification failed. The download was deleted.</source>
        <translation>SHA-256 検証に失敗しました。ダウンロードを削除しました。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="33"/>
        <source>Cannot read file: </source>
        <translation>ファイルを読み込めません: </translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="39"/>
        <source>Cannot write atomically: </source>
        <translation>アトミック書き込みができません: </translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="50"/>
        <source>Invalid request.</source>
        <translation>無効なリクエストです。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="53"/>
        <source>The configuration path must be absolute and cannot be a symbolic link.</source>
        <translation>設定パスは絶対パスで指定し、シンボリックリンクは使用できません。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="56"/>
        <source>Another control panel is modifying this configuration.</source>
        <translation>別の管理画面がこの設定を変更しています。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="64"/>
        <source>Cannot decrypt the backup.</source>
        <translation>バックアップを復号できません。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="67"/>
        <source>The configuration was changed externally. Overwrite refused; merge changes manually first.</source>
        <translation>設定が外部で変更されました。上書きを拒否しました。先に手動でマージしてください。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="70"/>
        <source>No backup is available to restore.</source>
        <translation>復元できるバックアップがありません。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="72"/>
        <source>Cannot remove the newly created configuration.</source>
        <translation>新しく作成した設定を削除できません。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="73"/>
        <source>Configuration restored, but backup cleanup failed.</source>
        <translation>設定を復元しましたが、バックアップの削除に失敗しました。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="76"/>
        <source>Only strict JSON object configurations are supported. The original file was not changed.</source>
        <translation>厳密な JSON オブジェクト形式のみ対応しています。元のファイルは変更していません。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="78"/>
        <source>env must be a JSON object.</source>
        <translation>env は JSON オブジェクトである必要があります。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="89"/>
        <source>Unsupported agent.</source>
        <translation>未対応のエージェントです。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="93"/>
        <source>The API URL must use HTTPS or local HTTP.</source>
        <translation>API URL は HTTPS またはローカル HTTP を使用してください。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="94"/>
        <source>API key cannot be empty.</source>
        <translation>API キーを入力してください。</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="107"/>
        <source>Cannot encrypt the backup.</source>
        <translation>バックアップを暗号化できません。</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="10"/>
        <location filename="accesscard.cpp" line="111"/>
        <source>Connect</source>
        <translation>接続</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="65"/>
        <location filename="accesscard.cpp" line="124"/>
        <source>Hide</source>
        <translation>非表示</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="65"/>
        <location filename="accesscard.cpp" line="124"/>
        <source>Show</source>
        <translation>表示</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="112"/>
        <location filename="accesscard.cpp" line="113"/>
        <source>Protocol:</source>
        <translation>プロトコル：</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="114"/>
        <source>Select a protocol to view matching request examples. These examples use non-streaming requests.</source>
        <translation>プロトコルを選択すると、対応するリクエスト例が表示されます。これらの例は非ストリーミング方式です。</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="115"/>
        <source>Base URL</source>
        <translation>ベース URL</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="116"/>
        <source>API Key</source>
        <translation>API キー</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="117"/>
        <source>Model</source>
        <translation>モデル</translation>
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
        <translation>コピー</translation>
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
