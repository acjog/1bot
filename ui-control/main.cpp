#include <QtWidgets>
#include <QtNetwork>
#include <functional>
#include "wiring.h"
#include "updater.h"
#include "modelspage.h"
#include "pluginspage.h"
#include "accesscard.h"
#include "localization.h"

class Panel : public QWidget {
    Localization &localization;
    std::function<void()> retranslate;
    bool statusKnown = false, gatewayLive = false, gatewayStopped = false, hasNotice = false;
    QSettings settings;
    QNetworkAccessManager network;
    QSystemTrayIcon *tray;
    QLineEdit *gateway, *cli, *model, *gatewayKey, *updateUrl;
    QComboBox *plugin;
    QSpinBox *port;
    QLabel *status, *notice;
    QPushButton *startButton, *stopButton;
    QCheckBox *autoCheck;
    QByteArray controlKey;
    Updater updater;
    bool probing = false, starting = false, updating = false;
    qint64 gatewayPid = 0;
    QString staged;
    QGroupBox *gatewayGroup;
    ModelsPage *modelsPage;
    PluginsPage *pluginsPage;
    AccessCard *accessCard;
    void reloadPlugins() {
        QString selected = plugin->currentData().toString();
        if (selected.isEmpty()) selected = settings.value("plugin", "codebuddy").toString();
        QSignalBlocker blocker(plugin);
        plugin->clear();
        for (const auto &entry : pluginsPage->definitions()) plugin->addItem(entry.name, entry.id);
        int index = plugin->findData(selected);
        plugin->setCurrentIndex(index >= 0 ? index : (plugin->count() ? 0 : -1));
        cli->setText(pluginsPage->definition(plugin->currentData().toString()).cli);
    }
    void quitControl() {
        if (pluginsPage->installing()) { report(QObject::tr("CLI installation is in progress. Wait before exiting.")); showNormal(); return; }
        save(); qApp->quit();
    }
    void report(const QString &text) { hasNotice = true; notice->setText(text); }
    void updateStatusText() {
        status->setText(!statusKnown ? QObject::tr("Checking gateway...") : gatewayLive ? QObject::tr("Running · ") + endpoint() : gatewayStopped ? QObject::tr("Gateway stopped") : QObject::tr("Status unknown: port conflict, authentication failure, or network error."));
        tray->setToolTip("1bot · " + status->text());
    }
    QString endpoint() const { return "http://127.0.0.1:" + QString::number(port->value()); }
    QNetworkRequest controlRequest(const QString &path) {
        QNetworkRequest r(QUrl(endpoint() + path));
        r.setRawHeader("Authorization", "Bearer " + controlKey);
        r.setTransferTimeout(2000);
        return r;
    }
    void save() {
        settings.setValue("gateway", gateway->text());
        settings.setValue("plugin", plugin->currentData().toString());
        settings.setValue("cli", cli->text());
        settings.setValue("model", model->text());
        settings.setValue("port", port->value());
        settings.setValue("gatewayKey", protectSecret(gatewayKey->text().toUtf8()));
        settings.setValue("updateUrl", updateUrl->text());
        settings.setValue("autoCheck", autoCheck->isChecked());
        settings.sync();
    }
    void probe() {
        if (probing) return;
        probing = true;
        auto reply = network.get(controlRequest("/control/status"));
        connect(reply, &QNetworkReply::finished, this, [this, reply] {
            probing = false;
            bool live = reply->error() == QNetworkReply::NoError;
            bool stopped = reply->error() == QNetworkReply::ConnectionRefusedError;
            auto body = QJsonDocument::fromJson(reply->readAll()).object();
            if (live) {
                gatewayPid = qint64(body["pid"].toDouble());
                live = body["running"].toBool() && gatewayPid > 0;
            }
            statusKnown = true; gatewayLive = live; gatewayStopped = stopped;
            updateStatusText();
            startButton->setEnabled(stopped && !starting && !updating);
            stopButton->setEnabled(live && !updating);
            gatewayGroup->setEnabled(!live && !starting && !updating);
            modelsPage->setGateway(endpoint(), gatewayKey->text().toUtf8(), live ? gatewayPid : 0, live);
            tray->setToolTip("1bot · " + status->text());
            reply->deleteLater();
        });
    }
    void startGateway() {
        if (starting || updating) return;
        if (gatewayKey->text().isEmpty() || controlKey.isEmpty()) { report(QObject::tr("Gateway API key cannot be empty.")); return; }
        if (pluginsPage->installing()) { report(QObject::tr("Wait for CLI installation before starting the gateway.")); return; }
        const auto selected = pluginsPage->definition(plugin->currentData().toString());
        if (selected.id.isEmpty()) { report(QObject::tr("Add and select a plugin on the Plugins page first.")); return; }
        QString executable = gateway->text();
        QString pluginPath = QFileInfo(selected.executable).isAbsolute() ? selected.executable :
            QDir(QFileInfo(executable).absolutePath()).absoluteFilePath(selected.executable);
        if (!QFileInfo::exists(executable) || !QFileInfo::exists(pluginPath)) { report(QObject::tr("Gateway or plugin file not found. Run build.bat first.")); return; }
        QTcpServer guard;
        if (!guard.listen(QHostAddress::LocalHost, quint16(port->value()))) { report(QObject::tr("Port is already in use. A second gateway will not be started.")); return; }
        guard.close();
        save();
        QProcess process;
        process.setProgram(executable);
        process.setArguments({"-plugin", pluginPath, "-addr", "127.0.0.1:" + QString::number(port->value()),
                              "-bin", cli->text(), "-model", model->text()});
        auto env = QProcessEnvironment::systemEnvironment();
        env.insert("ONEBOT_CONTROL_KEY", QString::fromUtf8(controlKey));
        env.insert("ONEBOT_API_KEY", gatewayKey->text());
        process.setProcessEnvironment(env);
        process.setWorkingDirectory(QFileInfo(executable).absolutePath());
        process.setStandardOutputFile(QProcess::nullDevice());
        process.setStandardErrorFile(QProcess::nullDevice());
        if (!process.startDetached(&gatewayPid)) { report(QObject::tr("Failed to start the gateway. Check the executable and CLI installation.")); return; }
        starting = true;
        startButton->setEnabled(false);
        gatewayGroup->setEnabled(false);
        report(QObject::tr("Starting the gateway. Exiting the control panel will not stop it."));
        QTimer::singleShot(4000, this, [this] { starting = false; probe(); });
    }
    void stopGateway() {
        auto reply = network.post(controlRequest("/control/stop"), QByteArray());
        connect(reply, &QNetworkReply::finished, this, [this, reply] {
            report(reply->error() == QNetworkReply::NoError ? QObject::tr("Gateway stop requested. Active requests will be interrupted.") : QObject::tr("Failed to stop. Check the gateway status."));
            reply->deleteLater(); probe();
        });
    }
    void installUpdate() {
        if (pluginsPage->installing()) { report(QObject::tr("Wait for CLI installation before installing updates.")); return; }
        if (staged.isEmpty()) return;
        if (QMessageBox::question(this, QObject::tr("Install Update"), QObject::tr("Installation will stop the gateway and exit the control panel. Continue?")) != QMessageBox::Yes) return;
        updating = true;
        startButton->setEnabled(false); stopButton->setEnabled(false);
        auto reply = network.post(controlRequest("/control/stop"), QByteArray());
        connect(reply, &QNetworkReply::finished, this, [this, reply] {
            bool ok = reply->error() == QNetworkReply::NoError || reply->error() == QNetworkReply::ConnectionRefusedError;
            reply->deleteLater();
            if (!ok) { updating = false; report(QObject::tr("Could not safely stop the gateway. Update canceled.")); return; }
            QString appDir = QCoreApplication::applicationDirPath();
            QStringList args{appDir, staged, QString::number(QCoreApplication::applicationPid()), QString::number(gatewayPid)};
            if (!QProcess::startDetached(appDir + "/1bot-updater.exe", args, appDir)) {
                updating = false; report(QObject::tr("Failed to start the external updater.")); return;
            }
            save(); qApp->quit();
        });
    }
protected:
    void changeEvent(QEvent *event) override {
        QWidget::changeEvent(event);
        if (event->type() == QEvent::LanguageChange && retranslate) retranslate();
    }
    void closeEvent(QCloseEvent *event) override {
        save();
        if (tray->isVisible()) { hide(); event->ignore(); }
        else if (pluginsPage->installing()) { report(QObject::tr("Wait for CLI installation to finish before exiting.")); event->ignore(); }
        else event->accept();
    }
public:
    explicit Panel(Localization &translations) : localization(translations) {
        setWindowTitle(QObject::tr("1bot · Gateway Control Panel"));
        resize(900, 620);
        setMinimumSize(760, 520);
        setStyleSheet(
            "QWidget{font-family:'Microsoft YaHei UI';font-size:10pt;}"
            "QGroupBox{font-weight:600;margin-top:12px;padding-top:15px;}"
            "QLineEdit,QComboBox,QSpinBox{padding:5px;}"
            "QPushButton{padding:7px 15px;} QLabel{font-weight:normal;}"
            "QListWidget#navigation{border:0;background:palette(window);outline:0;}"
            "QListWidget#navigation::item{padding:13px 10px;margin:3px 0;border-radius:6px;}"
            "QListWidget#navigation::item:selected{background:palette(highlight);color:palette(highlighted-text);}");
        auto root = new QVBoxLayout(this);
        root->setContentsMargins(20, 18, 20, 16);
        root->setSpacing(12);
        auto title = new QLabel(QObject::tr("1bot / Local API Gateway"));
        title->setStyleSheet("font-size:20px;font-weight:600;");
        root->addWidget(title);
        status = new QLabel(QObject::tr("Checking gateway..."));
        status->setWordWrap(true);
        root->addWidget(status);
        auto body = new QHBoxLayout;
        body->setSpacing(18);
        auto navigation = new QListWidget;
        navigation->setObjectName("navigation");
        navigation->setAccessibleName(QObject::tr("Control panel navigation"));
        navigation->setFixedWidth(170);
        navigation->setSelectionMode(QAbstractItemView::SingleSelection);
        navigation->setEditTriggers(QAbstractItemView::NoEditTriggers);
        navigation->setHorizontalScrollBarPolicy(Qt::ScrollBarAlwaysOff);
        new QListWidgetItem(style()->standardIcon(QStyle::SP_ComputerIcon), QObject::tr("Gateway"), navigation);
        new QListWidgetItem(style()->standardIcon(QStyle::SP_DirIcon), QObject::tr("Plugins"), navigation);
        new QListWidgetItem(style()->standardIcon(QStyle::SP_FileDialogDetailedView), QObject::tr("Agent Configuration"), navigation);
        new QListWidgetItem(style()->standardIcon(QStyle::SP_BrowserReload), QObject::tr("Software Updates"), navigation);
        auto separator = new QFrame;
        separator->setFrameShape(QFrame::VLine);
        separator->setFrameShadow(QFrame::Sunken);
        auto pages = new QStackedWidget;
        auto makePage = [pages]() {
            auto page = new QWidget;
            auto layout = new QVBoxLayout(page);
            layout->setContentsMargins(0, 0, 8, 0);
            layout->setSpacing(14);
            auto scroll = new QScrollArea;
            scroll->setFrameShape(QFrame::NoFrame);
            scroll->setWidgetResizable(true);
            scroll->setWidget(page);
            pages->addWidget(scroll);
            return layout;
        };
        auto gatewayPage = makePage();
        modelsPage = new ModelsPage;
        pluginsPage = new PluginsPage;
        auto pluginsScroll = new QScrollArea;
        pluginsScroll->setFrameShape(QFrame::NoFrame);
        pluginsScroll->setWidgetResizable(true);
        pluginsScroll->setWidget(pluginsPage);
        pages->addWidget(pluginsScroll);
        auto agentPage = makePage();
        auto updatePage = makePage();
        connect(navigation, &QListWidget::currentRowChanged, pages, &QStackedWidget::setCurrentIndex);
        navigation->setCurrentRow(0);
        body->addWidget(navigation);
        body->addWidget(separator);
        body->addWidget(pages, 1);
        root->addLayout(body, 1);
        gatewayGroup = new QGroupBox(QObject::tr("Gateway"));
        auto form = new QFormLayout(gatewayGroup);
        auto line = [](const QString &value) { return new QLineEdit(value); };
        QString dir = QCoreApplication::applicationDirPath();
        gateway = line(settings.value("gateway", dir + "/gateway.exe").toString());
        plugin = new QComboBox;
        cli = line("");
        cli->setReadOnly(true);
        cli->setToolTip(QObject::tr("Edit this plugin's CLI command or path on the Plugins page."));
        reloadPlugins();
        connect(plugin, QOverload<int>::of(&QComboBox::currentIndexChanged), this, [this](int) {
            cli->setText(pluginsPage->definition(plugin->currentData().toString()).cli);
        });
        connect(pluginsPage, &PluginsPage::definitionsChanged, this, &Panel::reloadPlugins);
        model = line(settings.value("model").toString());
        port = new QSpinBox; port->setRange(1024, 65535); port->setValue(settings.value("port", 8787).toInt());
        controlKey = unprotectSecret(settings.value("controlKey").toByteArray());
        if (controlKey.isEmpty()) {
            controlKey = QUuid::createUuid().toString(QUuid::WithoutBraces).toUtf8();
            settings.setValue("controlKey", protectSecret(controlKey));
        }
        gatewayKey = line(QString::fromUtf8(unprotectSecret(settings.value("gatewayKey").toByteArray())));
        if (gatewayKey->text().isEmpty()) gatewayKey->setText(QUuid::createUuid().toString(QUuid::WithoutBraces));
        gatewayKey->setEchoMode(QLineEdit::Password);
        form->addRow(QObject::tr("Gateway executable"), gateway); form->addRow(QObject::tr("Upstream plugin"), plugin);
        form->addRow(QObject::tr("CLI executable / path"), cli); form->addRow(QObject::tr("Model (optional)"), model);
        form->addRow(QObject::tr("Local port"), port); form->addRow(QObject::tr("Gateway API key"), gatewayKey);
        gatewayPage->addWidget(gatewayGroup);
        auto actions = new QHBoxLayout;
        startButton = new QPushButton(QObject::tr("Start Gateway")); stopButton = new QPushButton(QObject::tr("Stop Gateway"));
        startButton->setEnabled(false); stopButton->setEnabled(false);
        actions->addWidget(startButton); actions->addWidget(stopButton);
        auto copy = new QPushButton(QObject::tr("Copy Gateway Key")); actions->addWidget(copy);
        connect(copy, &QPushButton::clicked, this, [this] {
            const QString secret = gatewayKey->text();
            QApplication::clipboard()->setText(secret);
            QTimer::singleShot(30000, this, [secret] { if (QApplication::clipboard()->text() == secret) QApplication::clipboard()->clear(); });
        });
        connect(startButton, &QPushButton::clicked, this, &Panel::startGateway);
        connect(stopButton, &QPushButton::clicked, this, &Panel::stopGateway);
        gatewayPage->addLayout(actions);
        accessCard = new AccessCard;
        accessCard->setConfig(port->value(), gatewayKey->text(), model->text());
        gatewayPage->addWidget(accessCard);
        connect(port, QOverload<int>::of(&QSpinBox::valueChanged), this, [this](int) {
            accessCard->setConfig(port->value(), gatewayKey->text(), model->text());
        });
        connect(gatewayKey, &QLineEdit::textChanged, this, [this](const QString &) {
            accessCard->setConfig(port->value(), gatewayKey->text(), model->text());
        });
        connect(model, &QLineEdit::textChanged, this, [this](const QString &) {
            accessCard->setConfig(port->value(), gatewayKey->text(), model->text());
        });
        gatewayPage->addWidget(modelsPage, 1);
        connect(modelsPage, &ModelsPage::modelSelected, this, [this](const QString &id) {
            model->setText(id);
            report(QObject::tr("Selected model: ") + id);
        });
        auto guide = new QLabel;
        guide->setWordWrap(true);
        guide->setTextFormat(Qt::PlainText);
        guide->setTextInteractionFlags(Qt::TextSelectableByMouse | Qt::TextSelectableByKeyboard);
        auto refreshGuide = [guide] {
            const QString steps = QObject::tr("Manual Agent Configuration\n\nThis page is read-only. 1bot does not read or modify agent configuration files.\n\n1. Close the agent and back up its configuration yourself.\n2. Open the file below in a text editor. Merge the example fields into the existing JSON; do not replace unrelated settings.\n3. Replace SERVICE_URL, API_KEY and MODEL_ID with values supported by your service. Remove model if not needed. Protect API keys and never commit them to a repository.\n4. Save the file and restart the agent. To undo changes, restore your own backup.");
            const QString warning = QObject::tr("Compatibility: the current gateway provides OpenAI Chat Completions, not Anthropic Messages. Claude Code requires a compatible Anthropic service. CodeBuddy endpoints must match its CLI version and protocol. Do not point the gateway's upstream CLI back to the same gateway; this creates a request loop.");
            const QString paths = QObject::tr("User configuration paths are shown below. Project settings, environment variables or command-line options may override them. Examples are templates, not a guarantee of compatibility.");
            const QString buddy = QObject::tr("CodeBuddy uses the top-level endpoint setting. Check the documentation for your installed CLI version:");
            guide->setText(steps + "\n\n" + warning + "\n\n" + paths +
                "\n\nClaude Code\n" + QDir::toNativeSeparators(QDir::homePath() + "/.claude/settings.json") +
                "\n\n{\n  \"env\": {\n    \"ANTHROPIC_BASE_URL\": \"SERVICE_URL\",\n    \"ANTHROPIC_AUTH_TOKEN\": \"API_KEY\"\n  },\n  \"model\": \"MODEL_ID\"\n}\n\nCodeBuddy\n" +
                QDir::toNativeSeparators(QDir::homePath() + "/.codebuddy/settings.json") +
                "\n\n{\n  \"endpoint\": \"SERVICE_URL\",\n  \"env\": {\n    \"CODEBUDDY_API_KEY\": \"API_KEY\"\n  },\n  \"model\": \"MODEL_ID\"\n}\n\n" +
                buddy + "\nhttps://www.codebuddy.cn/docs/cli/settings\n\n" +
                QObject::tr("CodeBuddy custom providers (models.json)") + "\n" +
                QDir::toNativeSeparators(QDir::homePath() + "/.codebuddy/models.json") + "\n\n" +
                QObject::tr("Back up models.json, then append an entry to its models array without removing existing providers. vendor identifies the provider; name is the display label; id must match a model supported by the service. url is the full API endpoint, including /v1/chat/completions, not just the base URL. Replace API_KEY with your service key. The local address below assumes port 8787; change it to your gateway port. Set token limits and capability flags only according to actual model and gateway support. Save and restart CodeBuddy, then select the custom model in its model selector. Never route an upstream CodeBuddy CLI back to its own gateway.") +
                "\n\n{\n  \"models\": [\n    {\n      \"vendor\": \"1bot\",\n      \"name\": \"My custom model\",\n      \"id\": \"MODEL_ID\",\n      \"url\": \"http://127.0.0.1:8787/v1/chat/completions\",\n      \"apiKey\": \"API_KEY\"\n    }\n  ]\n}\n\n" +
                QObject::tr("Optional fields: maxInputTokens / maxOutputTokens are token limits; supportsToolCall / supportsImages / supportsReasoning describe model capabilities. Configure reasoning only when supported by the provider. Do not copy another model's limits or declare unsupported capabilities. API keys in this file are plaintext: protect it and do not share or commit it."));
        };
        refreshGuide();
        agentPage->addWidget(guide);
        agentPage->addStretch();
        auto updateGroup = new QGroupBox(QObject::tr("Software Updates"));
        auto updates = new QFormLayout(updateGroup);
        updates->addRow(QObject::tr("Current version"), new QLabel(QCoreApplication::applicationVersion()));
        auto language = new QComboBox;
        language->addItem(QObject::tr("System default"), "system");
        language->addItem(QStringLiteral("简体中文"), "zh_CN");
        language->addItem(QStringLiteral("English"), "en_US");
        language->addItem(QStringLiteral("日本語"), "ja_JP");
        language->addItem(QStringLiteral("한국어"), "ko_KR");
        language->addItem(QStringLiteral("繁體中文"), "zh_TW");
        int languageIndex = language->findData(settings.value("ui/language", "system"));
        language->setCurrentIndex(languageIndex < 0 ? 0 : languageIndex);
        updates->addRow(QObject::tr("Language"), language);
        connect(language, QOverload<int>::of(&QComboBox::currentIndexChanged), this, [this, language](int) {
            if (!localization.setLanguage(language->currentData().toString())) {
                QSignalBlocker blocker(language);
                language->setCurrentIndex(language->findData(settings.value("ui/language", "system")));
                return;
            }
            settings.setValue("ui/language", language->currentData());
            settings.sync();
        });
        updateUrl = line(settings.value("updateUrl").toString()); updateUrl->setPlaceholderText(QObject::tr("1bot HTTPS update endpoint"));
        autoCheck = new QCheckBox(QObject::tr("Check at startup and every 6 hours"));
        autoCheck->setChecked(settings.value("autoCheck", true).toBool());
        auto check = new QPushButton(QObject::tr("Check for Updates"));
        updates->addRow(QObject::tr("Update URL"), updateUrl);
        updates->addRow(QObject::tr("Check policy"), autoCheck);
        auto updateActions = new QHBoxLayout;
        updateActions->addWidget(check);
        updateActions->addStretch();
        updates->addRow(updateActions);
        updatePage->addWidget(updateGroup);
        auto updateHint = new QLabel(QObject::tr("You will be asked before downloading a new version. Installation requires confirmation and stops the gateway."));
        updateHint->setWordWrap(true);
        updatePage->addWidget(updateHint);
        updatePage->addStretch();
        notice = new QLabel(QObject::tr("Closing the window minimizes it to the tray. The gateway runs independently."));
        notice->setWordWrap(true); notice->setTextInteractionFlags(Qt::TextSelectableByMouse); root->addWidget(notice);
        connect(check, &QPushButton::clicked, this, [this] { save(); updater.check(QUrl(updateUrl->text())); });
        connect(&updater, &Updater::message, this, &Panel::report);
        connect(&updater, &Updater::available, this, [this](const QString &version, const QString &notes) {
            if (QMessageBox::question(this, QObject::tr("New Version ") + version, notes.left(2000) + QObject::tr("\nDownload the update?")) == QMessageBox::Yes) updater.download();
        });
        connect(&updater, &Updater::ready, this, [this](const QString &path) { staged = path; installUpdate(); });
        QPixmap pix(32, 32); pix.fill(Qt::transparent);
        QPainter painter(&pix); painter.setRenderHint(QPainter::Antialiasing);
        painter.setBrush(QColor("#2563eb")); painter.setPen(Qt::NoPen); painter.drawRoundedRect(0, 0, 32, 32, 8, 8);
        painter.setPen(Qt::white); painter.setFont(QFont("Arial", 17, QFont::Bold)); painter.drawText(pix.rect(), Qt::AlignCenter, "1"); painter.end();
        setWindowIcon(QIcon(pix));
        tray = new QSystemTrayIcon(QIcon(pix), this);
        auto menu = new QMenu(this);
        menu->addAction(QObject::tr("Open Control Panel"), this, [this] { showNormal(); raise(); activateWindow(); });
        menu->addAction(QObject::tr("Start Gateway"), startButton, &QPushButton::click);
        menu->addAction(QObject::tr("Stop Gateway"), stopButton, &QPushButton::click);
        menu->addSeparator();
        menu->addAction(QObject::tr("Exit Control Panel (Keep Gateway Running)"), this, &Panel::quitControl);
        tray->setContextMenu(menu);
        connect(tray, &QSystemTrayIcon::activated, this, [this](QSystemTrayIcon::ActivationReason reason) {
            if (reason == QSystemTrayIcon::Trigger || reason == QSystemTrayIcon::DoubleClick) { showNormal(); raise(); activateWindow(); }
        });
        retranslate = [this, title, navigation, form, copy, refreshGuide,
                       updateGroup, updates, language, check, updateHint, menu] {
            setWindowTitle(QObject::tr("1bot · Gateway Control Panel"));
            title->setText(QObject::tr("1bot / Local API Gateway"));
            navigation->setAccessibleName(QObject::tr("Control panel navigation"));
            const QStringList tabs{QObject::tr("Gateway"), QObject::tr("Plugins"),
                QObject::tr("Agent Configuration"), QObject::tr("Software Updates")};
            for (int i = 0; i < tabs.size(); ++i) navigation->item(i)->setText(tabs[i]);
            auto labels = [](QFormLayout *layout, const QStringList &texts) {
                for (int i = 0; i < texts.size(); ++i)
                    if (auto label = qobject_cast<QLabel *>(layout->itemAt(i, QFormLayout::LabelRole)->widget()))
                        label->setText(texts[i]);
            };
            gatewayGroup->setTitle(QObject::tr("Gateway"));
            labels(form, {QObject::tr("Gateway executable"), QObject::tr("Upstream plugin"),
                QObject::tr("CLI executable / path"), QObject::tr("Model (optional)"),
                QObject::tr("Local port"), QObject::tr("Gateway API key")});
            cli->setToolTip(QObject::tr("Edit this plugin's CLI command or path on the Plugins page."));
            startButton->setText(QObject::tr("Start Gateway"));
            stopButton->setText(QObject::tr("Stop Gateway"));
            copy->setText(QObject::tr("Copy Gateway Key"));
            refreshGuide();
            updateGroup->setTitle(QObject::tr("Software Updates"));
            labels(updates, {QObject::tr("Current version"), QObject::tr("Language"),
                QObject::tr("Update URL"), QObject::tr("Check policy")});
            QSignalBlocker blocker(language);
            language->setItemText(0, QObject::tr("System default"));
            updateUrl->setPlaceholderText(QObject::tr("1bot HTTPS update endpoint"));
            autoCheck->setText(QObject::tr("Check at startup and every 6 hours"));
            check->setText(QObject::tr("Check for Updates"));
            updateHint->setText(QObject::tr("You will be asked before downloading a new version. Installation requires confirmation and stops the gateway."));
            if (!hasNotice) notice->setText(QObject::tr("Closing the window minimizes it to the tray. The gateway runs independently."));
            const auto actions = menu->actions();
            actions[0]->setText(QObject::tr("Open Control Panel"));
            actions[1]->setText(QObject::tr("Start Gateway"));
            actions[2]->setText(QObject::tr("Stop Gateway"));
            actions[4]->setText(QObject::tr("Exit Control Panel (Keep Gateway Running)"));
            updateStatusText();
        };
        if (QSystemTrayIcon::isSystemTrayAvailable()) tray->show();
        auto timer = new QTimer(this); connect(timer, &QTimer::timeout, this, &Panel::probe); timer->start(2000);
        QTimer::singleShot(0, this, &Panel::probe);
        auto updateTimer = new QTimer(this);
        connect(updateTimer, &QTimer::timeout, this, [this] { if (autoCheck->isChecked() && !updateUrl->text().isEmpty()) updater.check(QUrl(updateUrl->text())); });
        updateTimer->start(6 * 60 * 60 * 1000);
        QTimer::singleShot(3000, this, [this] { if (autoCheck->isChecked() && !updateUrl->text().isEmpty()) updater.check(QUrl(updateUrl->text())); });
    }
};

int main(int argc, char **argv) {
    QCoreApplication::setOrganizationName("1bot");
    QCoreApplication::setApplicationName("1bot-control");
    QCoreApplication::setApplicationVersion(APP_VERSION);
    if (argc > 1 && QByteArray(argv[1]) == "--wire") return 2;
    QApplication app(argc, argv);
    Localization localization;
    auto dataDir = QStandardPaths::writableLocation(QStandardPaths::AppLocalDataLocation);
    QDir().mkpath(dataDir);
    QLockFile lock(dataDir + "/control.lock");
    if (!lock.tryLock()) { QMessageBox::information(nullptr, "1bot", QObject::tr("The control panel is already running. Open it from the system tray.")); return 0; }
    app.setQuitOnLastWindowClosed(!QSystemTrayIcon::isSystemTrayAvailable());
    Panel panel(localization); panel.show();
    return app.exec();
}
