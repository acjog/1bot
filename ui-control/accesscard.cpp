#include "accesscard.h"
#include <QClipboard>
#include <QApplication>

AccessCard::AccessCard(QWidget *parent) : QWidget(parent) {
    auto layout = new QVBoxLayout(this);
    layout->setContentsMargins(0, 0, 0, 0);
    layout->setSpacing(12);

    group = new QGroupBox(QObject::tr("Connect"));
    auto g = new QVBoxLayout(group);
    g->setSpacing(8);

    auto protocolRow = new QHBoxLayout;
    protocolLabel = new QLabel;
    protocols = new QTabBar;
    protocols->setExpanding(false);
    protocols->setDrawBase(false);
    for (const auto &name : { "OpenAI", "Responses", "Anthropic", "Gemini" })
        protocols->addTab(QString::fromLatin1(name));
    protocols->setStyleSheet(
        "QTabBar::tab{padding:7px 12px;margin-right:4px;border-radius:5px;}"
        "QTabBar::tab:selected{background:palette(highlight);color:palette(highlighted-text);}");
    protocolRow->addWidget(protocolLabel);
    protocolRow->addWidget(protocols);
    protocolRow->addStretch();
    g->addLayout(protocolRow);

    hint = new QLabel;
    hint->setWordWrap(true);
    g->addWidget(hint);

    auto makeField = [&](QLabel *&label, QLineEdit *&value, QPushButton *&copyButton,
                         bool password = false) -> QHBoxLayout * {
        auto row = new QHBoxLayout;
        label = new QLabel;
        label->setMinimumWidth(90);
        value = new QLineEdit;
        value->setReadOnly(true);
        value->setStyleSheet("font-family:'Consolas','Courier New',monospace;");
        if (password) value->setEchoMode(QLineEdit::Password);
        copyButton = new QPushButton;
        row->addWidget(label);
        row->addWidget(value, 1);
        row->addWidget(copyButton);
        g->addLayout(row);
        return row;
    };

    makeField(baseUrlLabel, baseUrlValue, copyBaseUrl);
    auto *apiKeyRow = makeField(apiKeyLabel, apiKeyValue, copyApiKey, true);
    makeField(modelLabel, modelValue, copyModel);

    showApiKey = new QPushButton;
    showApiKey->setCheckable(true);
    apiKeyRow->insertWidget(apiKeyRow->indexOf(copyApiKey), showApiKey);

    auto copyToClipboard = [](const QString &text) { QApplication::clipboard()->setText(text); };

    connect(copyBaseUrl, &QPushButton::clicked, this, [this, copyToClipboard] { copyToClipboard(baseUrlValue->text()); });
    connect(copyApiKey, &QPushButton::clicked, this, [this, copyToClipboard] { copyToClipboard(apiKeyValue->text()); });
    connect(copyModel, &QPushButton::clicked, this, [this, copyToClipboard] { copyToClipboard(modelValue->text()); });
    connect(showApiKey, &QPushButton::toggled, this, [this](bool checked) {
        apiKeyValue->setEchoMode(checked ? QLineEdit::Normal : QLineEdit::Password);
        showApiKey->setText(checked ? QObject::tr("Hide") : QObject::tr("Show"));
    });

    examples = new QTabWidget;
    auto makeExampleTab = [&](QPlainTextEdit *&editor, QPushButton *&copyButton) {
        auto tab = new QWidget;
        auto tl = new QVBoxLayout(tab);
        tl->setContentsMargins(0, 8, 0, 0);
        tl->setSpacing(6);
        editor = new QPlainTextEdit;
        editor->setReadOnly(true);
        editor->setLineWrapMode(QPlainTextEdit::NoWrap);
        editor->setStyleSheet("font-family:'Consolas','Courier New',monospace;font-size:9pt;");
        editor->setMinimumHeight(150);
        tl->addWidget(editor, 1);
        copyButton = new QPushButton;
        auto btnRow = new QHBoxLayout;
        btnRow->addStretch();
        btnRow->addWidget(copyButton);
        tl->addLayout(btnRow);
        examples->addTab(tab, QString());
        connect(copyButton, &QPushButton::clicked, this, [this, editor, copyToClipboard] {
            copyToClipboard(editor->toPlainText());
        });
    };
    makeExampleTab(powerShellExample, copyPowerShell);
    makeExampleTab(cmdExample, copyCmd);
    makeExampleTab(curlExample, copyCurl);
    makeExampleTab(pythonExample, copyPython);
    makeExampleTab(nodeExample, copyNode);
    g->addWidget(examples);

    layout->addWidget(group);
    connect(protocols, &QTabBar::currentChanged, this, [this](int) { refresh(); });
    refresh();
}

void AccessCard::setConfig(int value, const QString &key, const QString &modelId) {
    if (port == value && apiKey == key && model == modelId) return;
    port = value;
    apiKey = key;
    model = modelId;
    refresh();
}

void AccessCard::refresh() {
    group->setTitle(QObject::tr("Connect"));
    protocolLabel->setText(QObject::tr("Protocol:"));
    protocols->setAccessibleName(QObject::tr("Protocol:"));
    hint->setText(QObject::tr("Select a protocol to view matching request examples. These examples use non-streaming requests."));
    baseUrlLabel->setText(QObject::tr("Base URL"));
    apiKeyLabel->setText(QObject::tr("API Key"));
    modelLabel->setText(QObject::tr("Model"));
    copyBaseUrl->setText(QObject::tr("Copy"));
    copyApiKey->setText(QObject::tr("Copy"));
    copyModel->setText(QObject::tr("Copy"));
    copyPowerShell->setText(QObject::tr("Copy"));
    copyCmd->setText(QObject::tr("Copy"));
    copyCurl->setText(QObject::tr("Copy"));
    showApiKey->setText(showApiKey->isChecked() ? QObject::tr("Hide") : QObject::tr("Show"));
    examples->setTabText(0, QObject::tr("PowerShell"));
    examples->setTabText(1, QObject::tr("CMD"));
    examples->setTabText(2, QObject::tr("curl"));
    examples->setTabText(3, QStringLiteral("Python"));
    examples->setTabText(4, QStringLiteral("Node.js"));
    copyPython->setText(QObject::tr("Copy"));
    copyNode->setText(QObject::tr("Copy"));

    const int protocol = protocols->currentIndex();
    const QString origin = QStringLiteral("http://127.0.0.1:%1").arg(port);
    const QString base = origin + (protocol == 2 ? QString() :
                                  protocol == 3 ? QStringLiteral("/v1beta") : QStringLiteral("/v1"));
    QString modelId = model.trimmed();
    if (modelId.isEmpty()) modelId = QStringLiteral("MODEL_ID");
    baseUrlValue->setText(base);
    apiKeyValue->setText(apiKey);
    modelValue->setText(modelId);

    const QJsonArray messages{QJsonObject{{"role", "user"}, {"content", "hi"}}};
    QJsonObject payload{{"model", modelId}, {"messages", messages}};
    QJsonObject headers{{"Content-Type", "application/json"}};
    QString url;
    if (protocol == 0) {
        url = base + "/chat/completions";
        headers.insert("Authorization", "Bearer " + apiKey);
    } else if (protocol == 1) {
        url = base + "/responses";
        headers.insert("Authorization", "Bearer " + apiKey);
        payload = QJsonObject{{"model", modelId}, {"input", "hi"}};
    } else if (protocol == 2) {
        url = base + "/v1/messages";
        headers.insert("x-api-key", apiKey);
        headers.insert("anthropic-version", "2023-06-01");
        payload.insert("max_tokens", 1024);
    } else {
        QString resource = modelId;
        if (resource.startsWith("models/")) resource.remove(0, 7);
        url = base + "/models/" + QString::fromLatin1(QUrl::toPercentEncoding(resource, "/")) + ":generateContent";
        headers.insert("x-goog-api-key", apiKey);
        const QJsonArray parts{QJsonObject{{"text", "hi"}}};
        const QJsonArray contents{QJsonObject{{"role", "user"}, {"parts", parts}}};
        payload = QJsonObject{{"contents", contents}};
    }

    auto json = [](const QJsonObject &value) {
        return QString::fromUtf8(QJsonDocument(value).toJson(QJsonDocument::Compact));
    };
    auto quoted = [](const QString &value) {
        const QByteArray data = QJsonDocument(QJsonArray{value}).toJson(QJsonDocument::Compact);
        return QString::fromUtf8(data.mid(1, data.size() - 2));
    };
    auto psQuote = [](QString value) { return "'" + value.replace("'", "''") + "'"; };
    auto shQuote = [](QString value) { return "'" + value.replace("'", "'\"'\"'") + "'"; };
    auto cmdQuote = [](const QString &value) {
        QString out = "\"";
        int slashes = 0;
        for (QChar ch : value) {
            if (ch == '\\') { ++slashes; continue; }
            if (ch == '"') out += QString(slashes * 2 + 1, '\\') + ch;
            else out += QString(slashes, '\\') + ch;
            slashes = 0;
        }
        return out + QString(slashes * 2, '\\') + "\"";
    };
    const QString body = json(payload);
    QString ps = "$headers = @{\n";
    QString cmd = "curl.exe " + cmdQuote(url);
    QString curl = "curl " + shQuote(url);
    for (auto it = headers.constBegin(); it != headers.constEnd(); ++it) {
        ps += "  " + psQuote(it.key()) + " = " + psQuote(it.value().toString()) + "\n";
        const QString header = it.key() + ": " + it.value().toString();
        cmd += " ^\n  -H " + cmdQuote(header);
        curl += " \\\n  -H " + shQuote(header);
    }
    ps += "}\n$body = " + psQuote(body) + "\nInvoke-RestMethod -Uri " + psQuote(url) +
          " -Method Post -Headers $headers -Body ([System.Text.Encoding]::UTF8.GetBytes($body))";
    powerShellExample->setPlainText(ps);
    cmdExample->setPlainText(cmd + " ^\n  --data-raw " + cmdQuote(body));
    curlExample->setPlainText(curl + " \\\n  --data-raw " + shQuote(body));

    if (protocol < 2) {
        const QString pythonCall = protocol == 0 ?
            "r = client.chat.completions.create(model=" + quoted(modelId) +
                ", messages=[{\"role\": \"user\", \"content\": \"hi\"}])\nprint(r.choices[0].message.content)\n" :
            "r = client.responses.create(model=" + quoted(modelId) + ", input=\"hi\")\nprint(r.output_text)\n";
        const QString nodeCall = protocol == 0 ?
            "const r = await client.chat.completions.create({ model: " + quoted(modelId) +
                ", messages: [{ role: \"user\", content: \"hi\" }] });\nconsole.log(r.choices[0].message.content);\n" :
            "const r = await client.responses.create({ model: " + quoted(modelId) +
                ", input: \"hi\" });\nconsole.log(r.output_text);\n";
        pythonExample->setPlainText(
            "# python -m pip install openai\nfrom openai import OpenAI\n\nclient = OpenAI(base_url=" +
            quoted(base) + ", api_key=" + quoted(apiKey) + ")\n" + pythonCall);
        nodeExample->setPlainText(
            "// npm install openai\n// Save as example.mjs; run: node example.mjs\n"
            "import OpenAI from \"openai\";\n\nconst client = new OpenAI({ baseURL: " +
            quoted(base) + ", apiKey: " + quoted(apiKey) + " });\n" + nodeCall);
    } else {
        const QString pythonOutput = protocol == 2 ?
            "print(\"\".join(b.get(\"text\", \"\") for b in r[\"content\"]))\n" :
            "print(\"\".join(p.get(\"text\", \"\") for p in r[\"candidates\"][0][\"content\"][\"parts\"]))\n";
        const QString nodeOutput = protocol == 2 ?
            "console.log(r.content.map(b => b.text ?? \"\").join(\"\"));\n" :
            "console.log(r.candidates[0].content.parts.map(p => p.text ?? \"\").join(\"\"));\n";
        pythonExample->setPlainText(
            "import json\nfrom urllib.request import Request, urlopen\n\n"
            "request = Request(\n    " + quoted(url) + ",\n    data=" + quoted(body) +
            ".encode(\"utf-8\"),\n    headers=" + json(headers) +
            ",\n    method=\"POST\",\n)\nwith urlopen(request) as response:\n    r = json.load(response)\n" + pythonOutput);
        nodeExample->setPlainText(
            "// Node.js 18+; save as example.mjs; run: node example.mjs\n"
            "const response = await fetch(" + quoted(url) + ", {\n  method: \"POST\",\n  headers: " +
            json(headers) + ",\n  body: JSON.stringify(" + body +
            "),\n});\nif (!response.ok) throw new Error(await response.text());\n"
            "const r = await response.json();\n" + nodeOutput);
    }
}

void AccessCard::changeEvent(QEvent *event) {
    QWidget::changeEvent(event);
    if (event->type() == QEvent::LanguageChange) refresh();
}
