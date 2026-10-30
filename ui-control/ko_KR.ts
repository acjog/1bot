<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE TS>
<TS version="2.1" language="ko_KR" sourcelanguage="en_US">
<context>
    <name>QObject</name>
    <message>
        <location filename="main.cpp" line="302"/>
        <source>CodeBuddy custom providers (models.json)</source>
        <translation>CodeBuddy 사용자 지정 공급자 (models.json)</translation>
    </message>
    <message>
        <location filename="main.cpp" line="304"/>
        <source>Back up models.json, then append an entry to its models array without removing existing providers. vendor identifies the provider; name is the display label; id must match a model supported by the service. url is the full API endpoint, including /v1/chat/completions, not just the base URL. Replace API_KEY with your service key. The local address below assumes port 8787; change it to your gateway port. Set token limits and capability flags only according to actual model and gateway support. Save and restart CodeBuddy, then select the custom model in its model selector. Never route an upstream CodeBuddy CLI back to its own gateway.</source>
        <translation>models.json을 백업하고 기존 공급자를 삭제하지 않은 채 models 배열에 항목을 추가하세요. vendor는 공급자, name은 표시 이름이며 id에는 서비스에서 지원하는 모델 ID를 입력합니다. url은 기본 주소가 아니라 /v1/chat/completions를 포함한 전체 API 주소입니다. API_KEY를 서비스 키로 바꾸세요. 아래 예제는 포트 8787을 사용하므로 실제 게이트웨이 포트로 변경하세요. 토큰 제한과 기능 플래그는 모델 및 게이트웨이의 실제 지원 범위에 맞춰 설정하세요. 저장 후 CodeBuddy를 다시 시작하고 모델 선택기에서 사용자 지정 모델을 선택하세요. 업스트림 CodeBuddy CLI가 자신의 게이트웨이에 다시 연결되지 않도록 하세요.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="306"/>
        <source>Optional fields: maxInputTokens / maxOutputTokens are token limits; supportsToolCall / supportsImages / supportsReasoning describe model capabilities. Configure reasoning only when supported by the provider. Do not copy another model&apos;s limits or declare unsupported capabilities. API keys in this file are plaintext: protect it and do not share or commit it.</source>
        <translation>선택 필드: maxInputTokens / maxOutputTokens는 토큰 제한이며 supportsToolCall / supportsImages / supportsReasoning은 도구 호출, 이미지 입력, 추론 기능을 나타냅니다. reasoning은 공급자가 지원할 때만 설정하세요. 다른 모델의 제한을 그대로 복사하거나 지원하지 않는 기능을 선언하지 마세요. API 키는 평문으로 저장되므로 파일을 보호하고 공유하거나 저장소에 커밋하지 마세요.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="44"/>
        <source>CLI installation is in progress. Wait before exiting.</source>
        <translation>CLI 설치 중입니다. 완료 후 종료하세요.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="49"/>
        <source>Running · </source>
        <translation>실행 중 · </translation>
    </message>
    <message>
        <location filename="main.cpp" line="49"/>
        <source>Gateway stopped</source>
        <translation>게이트웨이 중지됨</translation>
    </message>
    <message>
        <location filename="main.cpp" line="49"/>
        <source>Status unknown: port conflict, authentication failure, or network error.</source>
        <translation>상태를 확인할 수 없습니다: 포트 충돌, 인증 실패 또는 네트워크 오류.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="95"/>
        <source>Gateway API key cannot be empty.</source>
        <translation>게이트웨이 API 키는 비워 둘 수 없습니다.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="96"/>
        <source>Wait for CLI installation before starting the gateway.</source>
        <translation>CLI 설치가 완료된 후 게이트웨이를 시작하세요.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="98"/>
        <source>Add and select a plugin on the Plugins page first.</source>
        <translation>먼저 플러그인 페이지에서 플러그인을 추가하고 선택하세요.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="102"/>
        <source>Gateway or plugin file not found. Run build.bat first.</source>
        <translation>게이트웨이 또는 플러그인 파일이 없습니다. 먼저 build.bat를 실행하세요.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="104"/>
        <source>Port is already in use. A second gateway will not be started.</source>
        <translation>포트가 사용 중입니다. 두 번째 게이트웨이를 시작하지 않습니다.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="118"/>
        <source>Failed to start the gateway. Check the executable and CLI installation.</source>
        <translation>게이트웨이 시작에 실패했습니다. 실행 파일과 CLI 설치를 확인하세요.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="122"/>
        <source>Starting the gateway. Exiting the control panel will not stop it.</source>
        <translation>게이트웨이 시작 중입니다. 제어판을 종료해도 게이트웨이는 중지되지 않습니다.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="128"/>
        <source>Gateway stop requested. Active requests will be interrupted.</source>
        <translation>게이트웨이 중지를 요청했습니다. 진행 중인 요청은 중단됩니다.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="128"/>
        <source>Failed to stop. Check the gateway status.</source>
        <translation>중지에 실패했습니다. 게이트웨이 상태를 확인하세요.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="133"/>
        <source>Wait for CLI installation before installing updates.</source>
        <translation>CLI 설치 완료 후 업데이트를 설치하세요.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="135"/>
        <source>Install Update</source>
        <translation>업데이트 설치</translation>
    </message>
    <message>
        <location filename="main.cpp" line="135"/>
        <source>Installation will stop the gateway and exit the control panel. Continue?</source>
        <translation>설치 시 게이트웨이를 중지하고 제어판을 종료합니다. 계속할까요?</translation>
    </message>
    <message>
        <location filename="main.cpp" line="142"/>
        <source>Could not safely stop the gateway. Update canceled.</source>
        <translation>게이트웨이를 안전하게 중지할 수 없어 업데이트를 취소했습니다.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="146"/>
        <source>Failed to start the external updater.</source>
        <translation>외부 업데이터를 시작할 수 없습니다.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="159"/>
        <source>Wait for CLI installation to finish before exiting.</source>
        <translation>CLI 설치 완료 후 종료하세요.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="164"/>
        <location filename="main.cpp" line="374"/>
        <source>1bot · Gateway Control Panel</source>
        <translation>1bot · 게이트웨이 제어판</translation>
    </message>
    <message>
        <location filename="main.cpp" line="178"/>
        <location filename="main.cpp" line="375"/>
        <source>1bot / Local API Gateway</source>
        <translation>1bot / 로컬 API 게이트웨이</translation>
    </message>
    <message>
        <location filename="main.cpp" line="49"/>
        <location filename="main.cpp" line="181"/>
        <source>Checking gateway...</source>
        <translation>게이트웨이 확인 중...</translation>
    </message>
    <message>
        <location filename="main.cpp" line="188"/>
        <location filename="main.cpp" line="376"/>
        <source>Control panel navigation</source>
        <translation>제어판 탐색</translation>
    </message>
    <message>
        <location filename="main.cpp" line="193"/>
        <location filename="main.cpp" line="229"/>
        <location filename="main.cpp" line="377"/>
        <location filename="main.cpp" line="385"/>
        <source>Gateway</source>
        <translation>게이트웨이</translation>
    </message>
    <message>
        <location filename="main.cpp" line="194"/>
        <location filename="main.cpp" line="377"/>
        <source>Plugins</source>
        <translation>플러그인</translation>
    </message>
    <message>
        <location filename="main.cpp" line="195"/>
        <location filename="main.cpp" line="378"/>
        <source>Agent Configuration</source>
        <translation>에이전트 설정</translation>
    </message>
    <message>
        <location filename="main.cpp" line="196"/>
        <location filename="main.cpp" line="311"/>
        <location filename="main.cpp" line="378"/>
        <location filename="main.cpp" line="394"/>
        <source>Software Updates</source>
        <translation>소프트웨어 업데이트</translation>
    </message>
    <message>
        <location filename="main.cpp" line="237"/>
        <location filename="main.cpp" line="389"/>
        <source>Edit this plugin&apos;s CLI command or path on the Plugins page.</source>
        <translation>플러그인 페이지에서 CLI 명령 또는 경로를 편집하세요.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="253"/>
        <location filename="main.cpp" line="386"/>
        <source>Gateway executable</source>
        <translation>게이트웨이 실행 파일</translation>
    </message>
    <message>
        <location filename="main.cpp" line="253"/>
        <location filename="main.cpp" line="386"/>
        <source>Upstream plugin</source>
        <translation>업스트림 플러그인</translation>
    </message>
    <message>
        <location filename="main.cpp" line="254"/>
        <location filename="main.cpp" line="387"/>
        <source>CLI executable / path</source>
        <translation>CLI 실행 파일 / 경로</translation>
    </message>
    <message>
        <location filename="main.cpp" line="254"/>
        <location filename="main.cpp" line="387"/>
        <source>Model (optional)</source>
        <translation>모델 (선택 사항)</translation>
    </message>
    <message>
        <location filename="main.cpp" line="255"/>
        <location filename="main.cpp" line="388"/>
        <source>Local port</source>
        <translation>로컬 포트</translation>
    </message>
    <message>
        <location filename="main.cpp" line="255"/>
        <location filename="main.cpp" line="388"/>
        <source>Gateway API key</source>
        <translation>게이트웨이 API 키</translation>
    </message>
    <message>
        <location filename="main.cpp" line="258"/>
        <location filename="main.cpp" line="364"/>
        <location filename="main.cpp" line="390"/>
        <location filename="main.cpp" line="406"/>
        <source>Start Gateway</source>
        <translation>게이트웨이 시작</translation>
    </message>
    <message>
        <location filename="main.cpp" line="258"/>
        <location filename="main.cpp" line="365"/>
        <location filename="main.cpp" line="391"/>
        <location filename="main.cpp" line="407"/>
        <source>Stop Gateway</source>
        <translation>게이트웨이 중지</translation>
    </message>
    <message>
        <location filename="main.cpp" line="261"/>
        <location filename="main.cpp" line="392"/>
        <source>Copy Gateway Key</source>
        <translation>게이트웨이 키 복사</translation>
    </message>
    <message>
        <location filename="main.cpp" line="313"/>
        <location filename="main.cpp" line="395"/>
        <source>Current version</source>
        <translation>현재 버전</translation>
    </message>
    <message>
        <location filename="main.cpp" line="292"/>
        <source>Manual Agent Configuration

This page is read-only. 1bot does not read or modify agent configuration files.

1. Close the agent and back up its configuration yourself.
2. Open the file below in a text editor. Merge the example fields into the existing JSON; do not replace unrelated settings.
3. Replace SERVICE_URL, API_KEY and MODEL_ID with values supported by your service. Remove model if not needed. Protect API keys and never commit them to a repository.
4. Save the file and restart the agent. To undo changes, restore your own backup.</source>
        <translation>에이전트 수동 설정

이 페이지는 읽기 전용 안내입니다. 1bot은 에이전트 설정 파일을 읽거나 수정하지 않습니다.

1. 에이전트를 종료하고 설정 파일을 직접 백업하세요.
2. 아래 파일을 텍스트 편집기로 열고 예제 필드를 기존 JSON에 병합하세요. 다른 설정은 덮어쓰지 마세요.
3. SERVICE_URL, API_KEY, MODEL_ID를 서비스에서 지원하는 값으로 바꾸세요. 모델 지정이 필요 없으면 model을 삭제하세요. 키를 안전하게 보관하고 저장소에 커밋하지 마세요.
4. 파일을 저장하고 에이전트를 다시 시작하세요. 변경을 취소하려면 직접 백업을 복원하세요.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="285"/>
        <source>Selected model: </source>
        <translation>선택한 모델：</translation>
    </message>
    <message>
        <location filename="main.cpp" line="293"/>
        <source>Compatibility: the current gateway provides OpenAI Chat Completions, not Anthropic Messages. Claude Code requires a compatible Anthropic service. CodeBuddy endpoints must match its CLI version and protocol. Do not point the gateway&apos;s upstream CLI back to the same gateway; this creates a request loop.</source>
        <translation>호환성: 현재 게이트웨이는 OpenAI Chat Completions를 제공하며 Anthropic Messages는 제공하지 않습니다. Claude Code에는 Anthropic 호환 서비스가 필요합니다. CodeBuddy 엔드포인트는 CLI 버전 및 프로토콜과 일치해야 합니다. 업스트림 CLI가 같은 게이트웨이를 가리키면 요청 루프가 발생합니다.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="294"/>
        <source>User configuration paths are shown below. Project settings, environment variables or command-line options may override them. Examples are templates, not a guarantee of compatibility.</source>
        <translation>사용자 설정 경로는 아래에 표시됩니다. 프로젝트 설정, 환경 변수 또는 명령줄 옵션이 이를 재정의할 수 있습니다. 예제는 템플릿이며 호환성을 보장하지 않습니다.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="295"/>
        <source>CodeBuddy uses the top-level endpoint setting. Check the documentation for your installed CLI version:</source>
        <translation>CodeBuddy는 최상위 endpoint 설정을 사용합니다. 설치된 CLI 버전의 문서를 확인하세요:</translation>
    </message>
    <message>
        <location filename="main.cpp" line="315"/>
        <location filename="main.cpp" line="398"/>
        <source>System default</source>
        <translation>시스템 기본값</translation>
    </message>
    <message>
        <location filename="main.cpp" line="323"/>
        <location filename="main.cpp" line="395"/>
        <source>Language</source>
        <translation>언어</translation>
    </message>
    <message>
        <location filename="main.cpp" line="333"/>
        <location filename="main.cpp" line="399"/>
        <source>1bot HTTPS update endpoint</source>
        <translation>1bot HTTPS 업데이트 엔드포인트</translation>
    </message>
    <message>
        <location filename="main.cpp" line="334"/>
        <location filename="main.cpp" line="400"/>
        <source>Check at startup and every 6 hours</source>
        <translation>시작 시 및 6시간마다 확인</translation>
    </message>
    <message>
        <location filename="main.cpp" line="336"/>
        <location filename="main.cpp" line="401"/>
        <source>Check for Updates</source>
        <translation>업데이트 확인</translation>
    </message>
    <message>
        <location filename="main.cpp" line="337"/>
        <location filename="main.cpp" line="396"/>
        <source>Update URL</source>
        <translation>업데이트 URL</translation>
    </message>
    <message>
        <location filename="main.cpp" line="338"/>
        <location filename="main.cpp" line="396"/>
        <source>Check policy</source>
        <translation>확인 정책</translation>
    </message>
    <message>
        <location filename="main.cpp" line="344"/>
        <location filename="main.cpp" line="402"/>
        <source>You will be asked before downloading a new version. Installation requires confirmation and stops the gateway.</source>
        <translation>새 버전 다운로드 전에 확인합니다. 설치에는 확인이 필요하며 게이트웨이가 중지됩니다.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="348"/>
        <location filename="main.cpp" line="403"/>
        <source>Closing the window minimizes it to the tray. The gateway runs independently.</source>
        <translation>창을 닫으면 트레이로 최소화됩니다. 게이트웨이는 독립적으로 실행됩니다.</translation>
    </message>
    <message>
        <location filename="main.cpp" line="353"/>
        <source>New Version </source>
        <translation>새 버전 </translation>
    </message>
    <message>
        <location filename="main.cpp" line="353"/>
        <source>
Download the update?</source>
        <translation>
업데이트를 다운로드할까요?</translation>
    </message>
    <message>
        <location filename="main.cpp" line="363"/>
        <location filename="main.cpp" line="405"/>
        <source>Open Control Panel</source>
        <translation>제어판 열기</translation>
    </message>
    <message>
        <location filename="main.cpp" line="367"/>
        <location filename="main.cpp" line="408"/>
        <source>Exit Control Panel (Keep Gateway Running)</source>
        <translation>제어판 종료 (게이트웨이 유지)</translation>
    </message>
    <message>
        <location filename="main.cpp" line="431"/>
        <source>The control panel is already running. Open it from the system tray.</source>
        <translation>제어판이 이미 실행 중입니다. 시스템 트레이에서 여세요.</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="7"/>
        <location filename="modelspage.cpp" line="39"/>
        <source>Models Supported by the Current Plugin</source>
        <translation>현재 플러그인이 지원하는 모델</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="9"/>
        <location filename="modelspage.cpp" line="40"/>
        <source>Refresh List</source>
        <translation>목록 새로 고침</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="14"/>
        <source>Start the gateway first.</source>
        <translation>먼저 게이트웨이를 시작하세요.</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="19"/>
        <location filename="modelspage.cpp" line="41"/>
        <source>Model ID</source>
        <translation>모델 ID</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="19"/>
        <location filename="modelspage.cpp" line="41"/>
        <source>Plugin</source>
        <translation>플러그인</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="28"/>
        <location filename="modelspage.cpp" line="42"/>
        <source>This catalog comes from the running gateway, not the default model argument. This page does not switch models.</source>
        <translation>목록은 기본 모델 인수가 아닌 실행 중인 게이트웨이에서 가져옵니다. 이 페이지에서는 모델을 전환하지 않습니다.</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="67"/>
        <source>Click Refresh to retrieve the current plugin&apos;s models.</source>
        <translation>새로 고침을 클릭하여 현재 플러그인의 모델을 가져오세요.</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="44"/>
        <location filename="modelspage.cpp" line="67"/>
        <source>The gateway is offline or cannot be authenticated. Start it and check the connection on the Gateway page.</source>
        <translation>게이트웨이가 오프라인이거나 인증할 수 없습니다. 게이트웨이 페이지에서 시작하고 연결을 확인하세요.</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="43"/>
        <location filename="modelspage.cpp" line="81"/>
        <source>Fetching model catalog...</source>
        <translation>모델 목록을 가져오는 중...</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="104"/>
        <source>This plugin or CLI version does not support model discovery.</source>
        <translation>이 플러그인 또는 CLI 버전은 모델 검색을 지원하지 않습니다.</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="105"/>
        <source>Authentication failed. Check the gateway API key.</source>
        <translation>인증에 실패했습니다. 게이트웨이 API 키를 확인하세요.</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="108"/>
        <source>The request timed out or exceeded the response limit. Try refreshing later.</source>
        <translation>요청 시간이 초과되었거나 응답 제한을 초과했습니다. 나중에 새로 고침하세요.</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="109"/>
        <source>Plugin query failed. Check the CLI status and retry.</source>
        <translation>플러그인 조회에 실패했습니다. CLI 상태를 확인하고 다시 시도하세요.</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="110"/>
        <source>Model endpoint unavailable. Update the gateway and plugins.</source>
        <translation>모델 엔드포인트를 사용할 수 없습니다. 게이트웨이와 플러그인을 업데이트하세요.</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="111"/>
        <source>Could not retrieve models. Check the gateway connection and retry.</source>
        <translation>모델을 가져올 수 없습니다. 게이트웨이 연결을 확인하고 다시 시도하세요.</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="120"/>
        <source>The gateway returned an invalid model catalog.</source>
        <translation>게이트웨이가 잘못된 모델 목록을 반환했습니다.</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="130"/>
        <source>The catalog contains invalid entries. Results were not displayed.</source>
        <translation>모델 목록에 잘못된 항목이 있습니다. 결과를 표시하지 않았습니다.</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="46"/>
        <location filename="modelspage.cpp" line="142"/>
        <source>The current plugin returned an empty model catalog.</source>
        <translation>현재 플러그인이 빈 모델 목록을 반환했습니다.</translation>
    </message>
    <message>
        <location filename="modelspage.cpp" line="47"/>
        <location filename="modelspage.cpp" line="143"/>
        <source>%1 models · Updated at %2</source>
        <translation>모델 %1개 · 업데이트 %2</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="43"/>
        <location filename="pluginspage.cpp" line="145"/>
        <source>Application Plugins</source>
        <translation>애플리케이션 플러그인</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="46"/>
        <location filename="pluginspage.cpp" line="146"/>
        <source>Plugins must implement 1bot stdio JSON-RPC and accept gateway arguments such as --bin / --model. Installation only installs the CLI dependency, not a plugin. Relative plugin paths are resolved from the gateway directory.</source>
        <translation>플러그인은 1bot stdio JSON-RPC와 --bin / --model 등의 인수를 지원해야 합니다. 설치는 CLI 종속성만 설치하며 플러그인을 생성하지 않습니다. 상대 경로는 게이트웨이 디렉터리 기준입니다.</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="54"/>
        <location filename="pluginspage.cpp" line="147"/>
        <source>Add Plugin</source>
        <translation>플러그인 추가</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="55"/>
        <location filename="pluginspage.cpp" line="148"/>
        <source>Delete Configuration</source>
        <translation>설정 삭제</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="66"/>
        <location filename="pluginspage.cpp" line="152"/>
        <source>Enter the official PowerShell installation command. Unverified CLI commands are left blank.</source>
        <translation>공식 PowerShell 설치 명령을 입력하세요. 확인되지 않은 CLI 명령은 비워 둡니다.</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="69"/>
        <location filename="pluginspage.cpp" line="149"/>
        <source>Browse</source>
        <translation>찾아보기</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="71"/>
        <location filename="pluginspage.cpp" line="153"/>
        <source>Display name</source>
        <translation>표시 이름</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="72"/>
        <location filename="pluginspage.cpp" line="153"/>
        <source>Plugin executable</source>
        <translation>플러그인 실행 파일</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="73"/>
        <location filename="pluginspage.cpp" line="154"/>
        <source>CLI command / path</source>
        <translation>CLI 명령 / 경로</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="74"/>
        <location filename="pluginspage.cpp" line="154"/>
        <source>CLI dependency installation command</source>
        <translation>CLI 종속성 설치 명령</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="76"/>
        <location filename="pluginspage.cpp" line="150"/>
        <source>Save Configuration</source>
        <translation>설정 저장</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="77"/>
        <location filename="pluginspage.cpp" line="151"/>
        <source>Install Dependencies</source>
        <translation>종속성 설치</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="81"/>
        <source>Select a plugin to view and edit its CLI installation command.</source>
        <translation>플러그인을 선택하여 CLI 설치 명령을 확인하고 편집하세요.</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="91"/>
        <source>Select a 1bot-compatible Plugin</source>
        <translation>1bot 호환 플러그인 선택</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="91"/>
        <source>Executables (*.exe)</source>
        <translation>실행 파일 (*.exe)</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="99"/>
        <source>Save the new plugin to add it to the Gateway page selector.</source>
        <translation>새 플러그인을 저장하면 게이트웨이 페이지 선택 목록에 추가됩니다.</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="103"/>
        <source>Delete Plugin Configuration</source>
        <translation>플러그인 설정 삭제</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="103"/>
        <source>Only the configuration will be removed. The CLI and plugin files will remain. Continue?</source>
        <translation>설정만 삭제합니다. CLI와 플러그인 파일은 유지됩니다. 계속할까요?</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="119"/>
        <source>Failed to start installation: </source>
        <translation>설치를 시작할 수 없습니다: </translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="126"/>
        <source>Installation command completed successfully. CLI login may still be required. Restart the control panel if PATH changed.</source>
        <translation>설치 명령이 완료되었습니다. CLI 로그인이 필요할 수 있습니다. PATH가 변경되었다면 제어판을 다시 시작하세요.</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="127"/>
        <source>Installation failed (exit code %1). Check the output.</source>
        <translation>설치 실패 (종료 코드 %1). 출력을 확인하세요.</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="185"/>
        <source>Requires CodeBuddy Code CLI and Node.js 18.20+ for npm installation. Official package: @tencent-ai/codebuddy-code.</source>
        <translation>CodeBuddy Code CLI가 필요합니다. npm 설치에는 Node.js 18.20+가 필요합니다. 공식 패키지: @tencent-ai/codebuddy-code.</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="187"/>
        <source>Requires ZCode CLI (zcode-app-cli) and Node.js/npm. Follow the CLI instructions to sign in after installation.</source>
        <translation>ZCode CLI (zcode-app-cli)와 Node.js/npm이 필요합니다. 설치 후 CLI 안내에 따라 로그인하세요.</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="189"/>
        <source>Requires Codex CLI and Node.js/npm. Follow the official sign-in instructions after installation.</source>
        <translation>Codex CLI와 Node.js/npm이 필요합니다. 설치 후 공식 안내에 따라 로그인하세요.</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="190"/>
        <source>Configure the CLI dependency and its official installation command.</source>
        <translation>CLI 종속성과 공식 설치 명령을 설정하세요.</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="201"/>
        <source>Failed to persist configuration. Check user permissions.</source>
        <translation>설정을 저장할 수 없습니다. 사용자 권한을 확인하세요.</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="207"/>
        <source>Name, plugin executable, and CLI command cannot be empty.</source>
        <translation>이름, 플러그인 실행 파일 및 CLI 명령은 비워 둘 수 없습니다.</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="218"/>
        <source>Configuration saved. Changes take effect after restarting the gateway; the running gateway is unchanged.</source>
        <translation>설정을 저장했습니다. 게이트웨이를 다시 시작하면 적용되며 실행 중인 게이트웨이에는 영향이 없습니다.</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="231"/>
        <source>No installation command is configured. Enter one from the official CLI documentation.</source>
        <translation>설치 명령이 설정되지 않았습니다. 공식 CLI 문서에서 확인하여 입력하세요.</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="232"/>
        <source>Confirm Installation Command</source>
        <translation>설치 명령 확인</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="233"/>
        <source>The following PowerShell command will run as the current Windows user and may download software or change the environment. Run only trusted official commands.

</source>
        <translation>다음 PowerShell 명령을 현재 Windows 사용자 권한으로 실행합니다. 소프트웨어를 다운로드하거나 환경을 변경할 수 있습니다. 신뢰하는 공식 명령만 실행하세요.

</translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="237"/>
        <source>Working directory: </source>
        <translation>작업 디렉터리: </translation>
    </message>
    <message>
        <location filename="pluginspage.cpp" line="158"/>
        <location filename="pluginspage.cpp" line="240"/>
        <source>Installing. Do not exit the control panel. Run interactive or administrator installations in a terminal.</source>
        <translation>설치 중입니다. 제어판을 종료하지 마세요. 상호 작용이나 관리자 권한이 필요한 설치는 터미널에서 실행하세요.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="21"/>
        <source>Update file exceeds the size limit.</source>
        <translation>업데이트 파일이 크기 제한을 초과했습니다.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="24"/>
        <source>Cannot write update data.</source>
        <translation>업데이트 데이터를 쓸 수 없습니다.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="30"/>
        <source>Update downloads require HTTPS.</source>
        <translation>업데이트 다운로드에는 HTTPS가 필요합니다.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="32"/>
        <source>Cannot initialize libcurl.</source>
        <translation>libcurl을 초기화할 수 없습니다.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="35"/>
        <source>The updater requires libcurl with the OpenSSL backend.</source>
        <translation>업데이터에는 OpenSSL 백엔드의 libcurl이 필요합니다.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="37"/>
        <source>Missing curl-ca-bundle.crt CA bundle.</source>
        <translation>curl-ca-bundle.crt CA 번들이 없습니다.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="39"/>
        <source>Cannot create an HTTPS request.</source>
        <translation>HTTPS 요청을 생성할 수 없습니다.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="64"/>
        <source>Cannot create request headers.</source>
        <translation>요청 헤더를 생성할 수 없습니다.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="71"/>
        <source>Cannot configure secure HTTPS request options.</source>
        <translation>안전한 HTTPS 요청 옵션을 설정할 수 없습니다.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="75"/>
        <source>HTTPS request failed: </source>
        <translation>HTTPS 요청 실패: </translation>
    </message>
    <message>
        <location filename="updater.cpp" line="78"/>
        <source>Update server HTTP status %1 (redirects are not accepted).</source>
        <translation>업데이트 서버 HTTP 상태 %1 (리디렉션 허용 안 함).</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="84"/>
        <source>Configure a dedicated 1bot HTTPS update URL.</source>
        <translation>1bot 전용 HTTPS 업데이트 URL을 설정하세요.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="101"/>
        <source>Invalid update manifest.</source>
        <translation>잘못된 업데이트 매니페스트입니다.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="106"/>
        <source>No updates available.</source>
        <translation>사용 가능한 업데이트가 없습니다.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="120"/>
        <source>The update manifest contains unsafe files, invalid sizes, or missing SHA-256 hashes.</source>
        <translation>업데이트 매니페스트에 안전하지 않은 파일, 잘못된 크기 또는 누락된 SHA-256이 있습니다.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="125"/>
        <source>The update package is missing 1bot-control.exe.</source>
        <translation>업데이트 패키지에 1bot-control.exe가 없습니다.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="141"/>
        <source>Cannot create the update directory.</source>
        <translation>업데이트 디렉터리를 생성할 수 없습니다.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="149"/>
        <source>Cannot save the update manifest.</source>
        <translation>업데이트 매니페스트를 저장할 수 없습니다.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="154"/>
        <source>Downloading update %1/%2: %3</source>
        <translation>업데이트 다운로드 중 %1/%2: %3</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="164"/>
        <source>Cannot create the download file.</source>
        <translation>다운로드 파일을 생성할 수 없습니다.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="168"/>
        <source>Download size mismatch or save failure.</source>
        <translation>다운로드 크기가 일치하지 않거나 저장에 실패했습니다.</translation>
    </message>
    <message>
        <location filename="updater.cpp" line="173"/>
        <source>SHA-256 verification failed. The download was deleted.</source>
        <translation>SHA-256 검증에 실패했습니다. 다운로드를 삭제했습니다.</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="33"/>
        <source>Cannot read file: </source>
        <translation>파일을 읽을 수 없습니다: </translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="39"/>
        <source>Cannot write atomically: </source>
        <translation>원자적으로 쓸 수 없습니다: </translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="50"/>
        <source>Invalid request.</source>
        <translation>잘못된 요청입니다.</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="53"/>
        <source>The configuration path must be absolute and cannot be a symbolic link.</source>
        <translation>설정 경로는 절대 경로여야 하며 심볼릭 링크일 수 없습니다.</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="56"/>
        <source>Another control panel is modifying this configuration.</source>
        <translation>다른 제어판이 이 설정을 수정하고 있습니다.</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="64"/>
        <source>Cannot decrypt the backup.</source>
        <translation>백업을 복호화할 수 없습니다.</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="67"/>
        <source>The configuration was changed externally. Overwrite refused; merge changes manually first.</source>
        <translation>설정이 외부에서 변경되었습니다. 덮어쓰기를 거부했습니다. 먼저 수동으로 병합하세요.</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="70"/>
        <source>No backup is available to restore.</source>
        <translation>복원할 백업이 없습니다.</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="72"/>
        <source>Cannot remove the newly created configuration.</source>
        <translation>새로 생성된 설정을 삭제할 수 없습니다.</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="73"/>
        <source>Configuration restored, but backup cleanup failed.</source>
        <translation>설정은 복원했지만 백업 정리에 실패했습니다.</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="76"/>
        <source>Only strict JSON object configurations are supported. The original file was not changed.</source>
        <translation>엄격한 JSON 객체 설정만 지원합니다. 원본 파일은 변경하지 않았습니다.</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="78"/>
        <source>env must be a JSON object.</source>
        <translation>env는 JSON 객체여야 합니다.</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="89"/>
        <source>Unsupported agent.</source>
        <translation>지원하지 않는 에이전트입니다.</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="93"/>
        <source>The API URL must use HTTPS or local HTTP.</source>
        <translation>API URL은 HTTPS 또는 로컬 HTTP를 사용해야 합니다.</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="94"/>
        <source>API key cannot be empty.</source>
        <translation>API 키는 비워 둘 수 없습니다.</translation>
    </message>
    <message>
        <location filename="wiring.cpp" line="107"/>
        <source>Cannot encrypt the backup.</source>
        <translation>백업을 암호화할 수 없습니다.</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="10"/>
        <location filename="accesscard.cpp" line="93"/>
        <source>Connect</source>
        <translation>연결</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="50"/>
        <location filename="accesscard.cpp" line="104"/>
        <source>Hide</source>
        <translation>숨기기</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="50"/>
        <location filename="accesscard.cpp" line="104"/>
        <source>Show</source>
        <translation>표시</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="94"/>
        <source>Use these values to connect any OpenAI-compatible client to the local gateway.</source>
        <translation>아래 값을 사용하여 모든 OpenAI 호환 클라이언트를 로컬 게이트웨이에 연결할 수 있습니다.</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="95"/>
        <source>Base URL</source>
        <translation>베이스 URL</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="96"/>
        <source>API Key</source>
        <translation>API 키</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="97"/>
        <source>Model</source>
        <translation>모델</translation>
    </message>
    <message>
        <location filename="accesscard.cpp" line="98"/>
        <location filename="accesscard.cpp" line="99"/>
        <location filename="accesscard.cpp" line="100"/>
        <location filename="accesscard.cpp" line="101"/>
        <location filename="accesscard.cpp" line="102"/>
        <location filename="accesscard.cpp" line="103"/>
        <source>Copy</source>
        <translation>복사</translation>
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
