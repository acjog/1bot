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
    g->addWidget(examples);

    layout->addWidget(group);
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
    hint->setText(QObject::tr("Use these values to connect any OpenAI-compatible client to the local gateway."));
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

    const QString base = QStringLiteral("http://127.0.0.1:%1/v1").arg(port);
    const QString chatUrl = base + QStringLiteral("/chat/completions");
    QString modelId = model.trimmed();
    if (modelId.isEmpty()) modelId = QStringLiteral("MODEL_ID");

    baseUrlValue->setText(base);
    apiKeyValue->setText(apiKey);
    modelValue->setText(modelId);

    powerShellExample->setPlainText(QStringLiteral(
        "$headers = @{\n"
        "  \"Content-Type\"  = \"application/json\"\n"
        "  \"Authorization\" = \"Bearer %1\"\n"
        "}\n"
        "$body = '{\"model\":\"%2\",\"stream\":true,\"messages\":[{\"role\":\"user\",\"content\":\"hi\"}]}'\n"
        "Invoke-RestMethod -Uri \"%3\" -Method Post -Headers $headers -Body $body")
        .arg(apiKey, modelId, chatUrl));

    cmdExample->setPlainText(QStringLiteral(
        "curl -N %1 ^\n"
        "  -H \"Content-Type: application/json\" ^\n"
        "  -H \"Authorization: Bearer %2\" ^\n"
        "  -d \"{\\\"model\\\":\\\"%3\\\",\\\"stream\\\":true,\\\"messages\\\":[{\\\"role\\\":\\\"user\\\",\\\"content\\\":\\\"hi\\\"}]}\"")
        .arg(chatUrl, apiKey, modelId));

    curlExample->setPlainText(QStringLiteral(
        "curl -N %1 \\\n"
        "  -H \"Content-Type: application/json\" \\\n"
        "  -H \"Authorization: Bearer %2\" \\\n"
        "  -d '{\"model\":\"%3\",\"stream\":true,\"messages\":[{\"role\":\"user\",\"content\":\"hi\"}]}'")
        .arg(chatUrl, apiKey, modelId));
}

void AccessCard::changeEvent(QEvent *event) {
    QWidget::changeEvent(event);
    if (event->type() == QEvent::LanguageChange) refresh();
}
