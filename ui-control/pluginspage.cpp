#include "pluginspage.h"

PluginsPage::PluginsPage(QWidget *parent) : QWidget(parent) {
    entries = {
        {"codebuddy", "CodeBuddy", "plugin-codebuddy.exe", "codebuddy", "npm.cmd install -g @tencent-ai/codebuddy-code@latest"},
        {"codex", "Codex", "plugin-codex.exe", "codex", "npm.cmd install -g @openai/codex"},
        {"zcode", "ZCode", "plugin-zcode.exe", "zcode", "npm.cmd install -g zcode-app-cli@latest"}
    };
    if (settings.contains("plugins/catalog")) {
        QJsonParseError error;
        auto doc = QJsonDocument::fromJson(settings.value("plugins/catalog").toByteArray(), &error);
        if (error.error == QJsonParseError::NoError && doc.isArray()) {
            QList<PluginDefinition> loaded;
            QSet<QString> ids;
            for (auto value : doc.array()) {
                auto o = value.toObject();
                const QString id = o["id"].toString();
                if (id.isEmpty() || ids.contains(id)) continue;
                ids.insert(id);
                loaded.append({id, o["name"].toString(), o["executable"].toString(),
                    o["cli"].toString(), o["installCommand"].toString()});
            }
            entries = loaded;
        }
    } else if (settings.contains("cli")) {
        const QString selected = settings.value("plugin", "codebuddy").toString();
        for (auto &entry : entries) if (entry.id == selected) entry.cli = settings.value("cli").toString();
    }
    if (!settings.value("plugins/codebuddyDependencyV1", false).toBool()) {
        for (auto &entry : entries) {
            if (entry.id == "codebuddy" && entry.cli == "codebuddy" && entry.installCommand.trimmed().isEmpty())
                entry.installCommand = "npm.cmd install -g @tencent-ai/codebuddy-code@latest";
        }
    }
    if (!settings.value("plugins/zcodeDependencyV1", false).toBool()) {
        for (auto &entry : entries) {
            if (entry.id == "zcode" && entry.installCommand.trimmed().isEmpty())
                entry.installCommand = "npm.cmd install -g zcode-app-cli@latest";
        }
    }
    auto layout = new QVBoxLayout(this);
    layout->setContentsMargins(0, 0, 8, 0);
    auto title = new QLabel(QObject::tr("Application Plugins"));
    title->setStyleSheet("font-size:16px;font-weight:600;");
    layout->addWidget(title);
    auto hint = new QLabel(QObject::tr("Plugins must implement 1bot stdio JSON-RPC and accept gateway arguments such as --bin / --model. Installation only installs the CLI dependency, not a plugin. Relative plugin paths are resolved from the gateway directory."));
    hint->setWordWrap(true);
    layout->addWidget(hint);
    list = new QListWidget;
    list->setMaximumHeight(105);
    for (const auto &entry : entries) list->addItem(entry.name);
    layout->addWidget(list);
    auto actions = new QHBoxLayout;
    addButton = new QPushButton(QObject::tr("Add Plugin"));
    deleteButton = new QPushButton(QObject::tr("Delete Configuration"));
    actions->addWidget(addButton); actions->addWidget(deleteButton); actions->addStretch();
    layout->addLayout(actions);
    editor = new QWidget;
    auto form = new QFormLayout(editor);
    form->setContentsMargins(0, 0, 0, 0);
    name = new QLineEdit;
    executable = new QLineEdit;
    cli = new QLineEdit;
    command = new QPlainTextEdit;
    command->setMaximumHeight(70);
    command->setPlaceholderText(QObject::tr("Enter the official PowerShell installation command. Unverified CLI commands are left blank."));
    auto pathRow = new QHBoxLayout;
    pathRow->addWidget(executable);
    auto browse = new QPushButton(QObject::tr("Browse"));
    pathRow->addWidget(browse);
    form->addRow(QObject::tr("Display name"), name);
    form->addRow(QObject::tr("Plugin executable"), pathRow);
    form->addRow(QObject::tr("CLI command / path"), cli);
    form->addRow(QObject::tr("CLI dependency installation command"), command);
    auto buttons = new QHBoxLayout;
    auto save = new QPushButton(QObject::tr("Save Configuration"));
    installButton = new QPushButton(QObject::tr("Install Dependencies"));
    buttons->addWidget(save); buttons->addWidget(installButton); buttons->addStretch();
    form->addRow(buttons);
    layout->addWidget(editor);
    status = new QLabel(QObject::tr("Select a plugin to view and edit its CLI installation command."));
    status->setWordWrap(true); status->setTextFormat(Qt::PlainText);
    layout->addWidget(status);
    output = new QPlainTextEdit;
    output->setReadOnly(true);
    output->setMaximumBlockCount(1500);
    output->setMinimumHeight(90);
    layout->addWidget(output, 1);
    connect(list, &QListWidget::currentRowChanged, this, &PluginsPage::select);
    connect(browse, &QPushButton::clicked, this, [this] {
        const auto path = QFileDialog::getOpenFileName(this, QObject::tr("Select a 1bot-compatible Plugin"), {}, QObject::tr("Executables (*.exe)"));
        if (!path.isEmpty()) executable->setText(path);
    });
    connect(save, &QPushButton::clicked, this, &PluginsPage::saveEntry);
    connect(addButton, &QPushButton::clicked, this, [this] {
        list->setCurrentRow(-1); editingId.clear();
        name->clear(); executable->clear(); cli->clear(); command->clear();
        editor->setEnabled(true); deleteButton->setEnabled(false); name->setFocus();
        status->setText(QObject::tr("Save the new plugin to add it to the Gateway page selector."));
    });
    connect(deleteButton, &QPushButton::clicked, this, [this] {
        const int row = list->currentRow();
        if (row < 0 || QMessageBox::question(this, QObject::tr("Delete Plugin Configuration"), QObject::tr("Only the configuration will be removed. The CLI and plugin files will remain. Continue?")) != QMessageBox::Yes) return;
        entries.removeAt(row); delete list->takeItem(row); persist(); emit definitionsChanged();
        if (entries.isEmpty()) select(-1);
    });
    connect(installButton, &QPushButton::clicked, this, &PluginsPage::install);
    process.setProcessChannelMode(QProcess::MergedChannels);
    connect(&process, &QProcess::started, this, [this] { process.closeWriteChannel(); });
    connect(&process, &QProcess::readyReadStandardOutput, this, [this] {
        const QByteArray bytes = process.readAllStandardOutput();
        output->moveCursor(QTextCursor::End);
        output->insertPlainText(QString::fromLocal8Bit(bytes));
        output->ensureCursorVisible();
    });
    connect(&process, &QProcess::errorOccurred, this, [this](QProcess::ProcessError error) {
        if (error == QProcess::FailedToStart) {
            setBusy(false);
            status->setText(QObject::tr("Failed to start installation: ") + process.errorString());
        }
    });
    connect(&process, QOverload<int, QProcess::ExitStatus>::of(&QProcess::finished), this,
        [this](int code, QProcess::ExitStatus exit) {
            setBusy(false);
            status->setText(exit == QProcess::NormalExit && code == 0
                ? QObject::tr("Installation command completed successfully. CLI login may still be required. Restart the control panel if PATH changed.")
                : QObject::tr("Installation failed (exit code %1). Check the output.").arg(code));
        });
    if (!settings.value("plugins/codebuddyDependencyV1", false).toBool()) {
        persist();
        if (settings.status() == QSettings::NoError) {
            settings.setValue("plugins/codebuddyDependencyV1", true);
            settings.sync();
        }
    }
    if (!settings.value("plugins/zcodeDependencyV1", false).toBool()) {
        persist();
        if (settings.status() == QSettings::NoError) {
            settings.setValue("plugins/zcodeDependencyV1", true);
            settings.sync();
        }
    }
    if (!entries.isEmpty()) list->setCurrentRow(0); else select(-1);
    retranslate = [this, title, hint, browse, save, form] {
        title->setText(QObject::tr("Application Plugins"));
        hint->setText(QObject::tr("Plugins must implement 1bot stdio JSON-RPC and accept gateway arguments such as --bin / --model. Installation only installs the CLI dependency, not a plugin. Relative plugin paths are resolved from the gateway directory."));
        addButton->setText(QObject::tr("Add Plugin"));
        deleteButton->setText(QObject::tr("Delete Configuration"));
        browse->setText(QObject::tr("Browse"));
        save->setText(QObject::tr("Save Configuration"));
        installButton->setText(QObject::tr("Install Dependencies"));
        command->setPlaceholderText(QObject::tr("Enter the official PowerShell installation command. Unverified CLI commands are left blank."));
        const QStringList labels{QObject::tr("Display name"), QObject::tr("Plugin executable"),
            QObject::tr("CLI command / path"), QObject::tr("CLI dependency installation command")};
        for (int i = 0; i < labels.size(); ++i)
            qobject_cast<QLabel *>(form->itemAt(i, QFormLayout::LabelRole)->widget())->setText(labels[i]);
        if (installing())
            status->setText(QObject::tr("Installing. Do not exit the control panel. Run interactive or administrator installations in a terminal."));
        else if (showingDependency) updateDependencyText();
    };
}

void PluginsPage::changeEvent(QEvent *event) {
    QWidget::changeEvent(event);
    if (event->type() == QEvent::LanguageChange && retranslate) retranslate();
}
PluginDefinition PluginsPage::definition(const QString &id) const {
    for (const auto &entry : entries) if (entry.id == id) return entry;
    return {};
}
void PluginsPage::select(int row) {
    showingDependency = false;
    editor->setEnabled(row >= 0);
    deleteButton->setEnabled(row >= 0);
    if (row < 0 || row >= entries.size()) { editingId.clear(); return; }
    const auto &entry = entries[row];
    editingId = entry.id; name->setText(entry.name); executable->setText(entry.executable);
    cli->setText(entry.cli); command->setPlainText(entry.installCommand);
    showingDependency = true;
    updateDependencyText();
}
void PluginsPage::updateDependencyText() {
    const auto entry = definition(editingId);
    if (entry.id == "codebuddy")
        status->setText(QObject::tr("Requires CodeBuddy Code CLI and Node.js 18.20+ for npm installation. Official package: @tencent-ai/codebuddy-code."));
    else if (entry.id == "zcode")
        status->setText(QObject::tr("Requires ZCode CLI (zcode-app-cli) and Node.js/npm. Follow the CLI instructions to sign in after installation."));
    else if (entry.id == "codex")
        status->setText(QObject::tr("Requires Codex CLI and Node.js/npm. Follow the official sign-in instructions after installation."));
    else status->setText(QObject::tr("Configure the CLI dependency and its official installation command."));
}
void PluginsPage::persist() {
    QJsonArray array;
    for (const auto &entry : entries) array.append(QJsonObject{
        {"id",entry.id},{"name",entry.name},{"executable",entry.executable},
        {"cli",entry.cli},{"installCommand",entry.installCommand}});
    settings.setValue("plugins/catalog", QJsonDocument(array).toJson(QJsonDocument::Compact));
    settings.sync();
    if (settings.status() != QSettings::NoError) {
        showingDependency = false;
        status->setText(QObject::tr("Failed to persist configuration. Check user permissions."));
    }
}
void PluginsPage::saveEntry() {
    showingDependency = false;
    if (name->text().trimmed().isEmpty() || executable->text().trimmed().isEmpty() || cli->text().trimmed().isEmpty()) {
        status->setText(QObject::tr("Name, plugin executable, and CLI command cannot be empty.")); return;
    }
    if (editingId.isEmpty()) editingId = QUuid::createUuid().toString(QUuid::WithoutBraces);
    PluginDefinition entry{editingId,name->text().trimmed(),executable->text().trimmed(),
        cli->text().trimmed(),command->toPlainText().trimmed()};
    int row = -1;
    for (int i = 0; i < entries.size(); ++i) if (entries[i].id == editingId) row = i;
    if (row < 0) { row = entries.size(); entries.append(entry); list->addItem(entry.name); }
    else { entries[row] = entry; list->item(row)->setText(entry.name); }
    list->setCurrentRow(row); deleteButton->setEnabled(true);
    showingDependency = false;
    status->setText(QObject::tr("Configuration saved. Changes take effect after restarting the gateway; the running gateway is unchanged."));
    persist(); emit definitionsChanged();
}
void PluginsPage::setBusy(bool busy) {
    showingDependency = false;
    list->setEnabled(!busy); editor->setEnabled(!busy);
    addButton->setEnabled(!busy);
    deleteButton->setEnabled(!busy && list->currentRow() >= 0);
}
void PluginsPage::install() {
    if (installing()) return;
    showingDependency = false;
    const QString script = command->toPlainText().trimmed();
    if (script.isEmpty()) { status->setText(QObject::tr("No installation command is configured. Enter one from the official CLI documentation.")); return; }
    QMessageBox confirm(QMessageBox::Warning, QObject::tr("Confirm Installation Command"),
        QObject::tr("The following PowerShell command will run as the current Windows user and may download software or change the environment. Run only trusted official commands.\n\n") + script,
        QMessageBox::Yes | QMessageBox::No, this);
    confirm.setTextFormat(Qt::PlainText);
    confirm.setDefaultButton(QMessageBox::No);
    confirm.setDetailedText(QObject::tr("Working directory: ") + QCoreApplication::applicationDirPath() + "\n\n" + script);
    if (confirm.exec() != QMessageBox::Yes) return;
    output->clear();
    status->setText(QObject::tr("Installing. Do not exit the control panel. Run interactive or administrator installations in a terminal."));
    setBusy(true);
    const QString shell = qEnvironmentVariable("SystemRoot") + "/System32/WindowsPowerShell/v1.0/powershell.exe";
    process.setProgram(shell);
    process.setArguments({"-NoLogo", "-NoProfile", "-NonInteractive", "-Command",
        "$ErrorActionPreference = 'Stop'; $global:LASTEXITCODE = 0; & {\n" + script +
        "\n}; if (-not $?) { exit 1 }; exit $LASTEXITCODE"});
    process.setWorkingDirectory(QCoreApplication::applicationDirPath());
    process.start();
}
