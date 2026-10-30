#include "modelspage.h"

ModelsPage::ModelsPage(QWidget *parent) : QWidget(parent) {
    auto layout = new QVBoxLayout(this);
    layout->setContentsMargins(0, 0, 8, 0);
    auto heading = new QHBoxLayout;
    auto title = new QLabel(QObject::tr("Models Supported by the Current Plugin"));
    title->setStyleSheet("font-size:16px;font-weight:600;");
    refreshButton = new QPushButton(QObject::tr("Refresh List"));
    heading->addWidget(title);
    heading->addStretch();
    heading->addWidget(refreshButton);
    layout->addLayout(heading);
    summary = new QLabel(QObject::tr("Start the gateway first."));
    summary->setWordWrap(true);
    summary->setTextFormat(Qt::PlainText);
    layout->addWidget(summary);
    table = new QTableWidget(0, 2);
    table->setHorizontalHeaderLabels({QObject::tr("Model ID"), QObject::tr("Plugin")});
    table->setEditTriggers(QAbstractItemView::NoEditTriggers);
    table->setSelectionBehavior(QAbstractItemView::SelectRows);
    table->setAlternatingRowColors(true);
    table->verticalHeader()->hide();
    table->horizontalHeader()->setSectionResizeMode(0, QHeaderView::Stretch);
    table->horizontalHeader()->setSectionResizeMode(1, QHeaderView::ResizeToContents);
    table->setSortingEnabled(true);
    layout->addWidget(table, 1);
    auto hint = new QLabel(QObject::tr("This catalog comes from the running gateway, not the default model argument. This page does not switch models."));
    hint->setWordWrap(true);
    layout->addWidget(hint);
    refreshButton->setEnabled(false);
    connect(refreshButton, &QPushButton::clicked, this, &ModelsPage::refresh);
    connect(table, &QTableWidget::cellClicked, this, [this](int row, int) {
        if (row < 0 || row >= table->rowCount()) return;
        const auto item = table->item(row, 0);
        if (item && !item->text().trimmed().isEmpty()) emit modelSelected(item->text());
    });
    retranslate = [this, title, hint] {
        title->setText(QObject::tr("Models Supported by the Current Plugin"));
        refreshButton->setText(QObject::tr("Refresh List"));
        table->setHorizontalHeaderLabels({QObject::tr("Model ID"), QObject::tr("Plugin")});
        hint->setText(QObject::tr("This catalog comes from the running gateway, not the default model argument. This page does not switch models."));
        if (pending) summary->setText(QObject::tr("Fetching model catalog..."));
        else if (!online) summary->setText(QObject::tr("The gateway is offline or cannot be authenticated. Start it and check the connection on the Gateway page."));
        else if (!updatedAt.isEmpty())
            summary->setText(table->rowCount() == 0 ? QObject::tr("The current plugin returned an empty model catalog.") :
                QObject::tr("%1 models · Updated at %2").arg(table->rowCount()).arg(updatedAt));
    };
}

void ModelsPage::changeEvent(QEvent *event) {
    QWidget::changeEvent(event);
    if (event->type() == QEvent::LanguageChange && retranslate) retranslate();
}

void ModelsPage::setGateway(const QString &address, const QByteArray &key, qint64 pid, bool running) {
    if (endpoint == address && apiKey == key && gatewayPid == pid && online == running) return;
    ++generation;
    if (pending) { pending->abort(); pending.clear(); }
    endpoint = address;
    apiKey = key;
    gatewayPid = pid;
    online = running;
    updatedAt.clear();
    table->setRowCount(0);
    refreshButton->setEnabled(online);
    summary->setText(online ? QObject::tr("Click Refresh to retrieve the current plugin's models.") : QObject::tr("The gateway is offline or cannot be authenticated. Start it and check the connection on the Gateway page."));
    if (online && isVisible()) refresh();
}

void ModelsPage::showEvent(QShowEvent *event) {
    QWidget::showEvent(event);
    if (online) refresh();
}

void ModelsPage::refresh() {
    if (!online || pending) return;
    const quint64 current = ++generation;
    updatedAt.clear();
    table->setRowCount(0);
    summary->setText(QObject::tr("Fetching model catalog..."));
    refreshButton->setEnabled(false);
    QNetworkRequest request(QUrl(endpoint + "/v1/models"));
    request.setRawHeader("Authorization", "Bearer " + apiKey);
    request.setAttribute(QNetworkRequest::RedirectPolicyAttribute, QNetworkRequest::ManualRedirectPolicy);
    request.setTransferTimeout(35000);
    auto reply = network.get(request);
    pending = reply;
    auto deadline = new QTimer(reply);
    deadline->setSingleShot(true);
    connect(deadline, &QTimer::timeout, reply, &QNetworkReply::abort);
    deadline->start(35000);
    connect(reply, &QNetworkReply::downloadProgress, reply, [reply](qint64 received, qint64 total) {
        if (received > 8 * 1024 * 1024 || total > 8 * 1024 * 1024) reply->abort();
    });
    connect(reply, &QNetworkReply::finished, this, [this, reply, current] {
        reply->deleteLater();
        if (current != generation) return;
        pending.clear();
        refreshButton->setEnabled(online);
        const int status = reply->attribute(QNetworkRequest::HttpStatusCodeAttribute).toInt();
        if (status != 200 || reply->error() != QNetworkReply::NoError) {
            QString message;
            if (status == 501) message = QObject::tr("This plugin or CLI version does not support model discovery.");
            else if (status == 401 || status == 403) message = QObject::tr("Authentication failed. Check the gateway API key.");
            else if (status == 504 || reply->error() == QNetworkReply::TimeoutError ||
                     reply->error() == QNetworkReply::OperationCanceledError)
                message = QObject::tr("The request timed out or exceeded the response limit. Try refreshing later.");
            else if (status == 502) message = QObject::tr("Plugin query failed. Check the CLI status and retry.");
            else if (status == 404) message = QObject::tr("Model endpoint unavailable. Update the gateway and plugins.");
            else message = QObject::tr("Could not retrieve models. Check the gateway connection and retry.");
            summary->setText(message);
            return;
        }
        QJsonParseError error;
        const auto doc = QJsonDocument::fromJson(reply->readAll(), &error);
        const auto object = doc.object();
        if (error.error != QJsonParseError::NoError || !doc.isObject() ||
            object["object"].toString() != "list" || !object["data"].isArray()) {
            summary->setText(QObject::tr("The gateway returned an invalid model catalog."));
            return;
        }
        const auto data = object["data"].toArray();
        QList<QPair<QString, QString>> rows;
        QSet<QString> seen;
        for (const auto &entry : data) {
            const auto model = entry.toObject();
            const QString id = model["id"].toString(), owner = model["owned_by"].toString();
            if (id.trimmed().isEmpty() || model["object"].toString() != "model") {
                summary->setText(QObject::tr("The catalog contains invalid entries. Results were not displayed.")); return;
            }
            if (!seen.contains(id)) { rows.append(qMakePair(id, owner)); seen.insert(id); }
        }
        table->setSortingEnabled(false);
        table->setRowCount(rows.size());
        for (int row = 0; row < rows.size(); ++row) {
            table->setItem(row, 0, new QTableWidgetItem(rows[row].first));
            table->setItem(row, 1, new QTableWidgetItem(rows[row].second));
        }
        table->setSortingEnabled(true);
        updatedAt = QTime::currentTime().toString("HH:mm:ss");
        summary->setText(rows.isEmpty() ? QObject::tr("The current plugin returned an empty model catalog.") :
            QObject::tr("%1 models · Updated at %2").arg(rows.size()).arg(updatedAt));
    });
}
