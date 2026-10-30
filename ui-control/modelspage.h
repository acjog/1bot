#pragma once
#include <QtWidgets>
#include <functional>
#include <QtNetwork>

class ModelsPage : public QWidget {
    Q_OBJECT
public:
    explicit ModelsPage(QWidget *parent = nullptr);
    void setGateway(const QString &endpoint, const QByteArray &key, qint64 pid, bool online);
    void refresh();
signals:
    void modelSelected(const QString &id);
protected:
    void showEvent(QShowEvent *event) override;
    void changeEvent(QEvent *event) override;
private:
    std::function<void()> retranslate;
    QNetworkAccessManager network;
    QPointer<QNetworkReply> pending;
    QTableWidget *table;
    QLabel *summary;
    QPushButton *refreshButton;
    QString updatedAt;
    QString endpoint;
    QByteArray apiKey;
    qint64 gatewayPid = 0;
    bool online = false;
    quint64 generation = 0;
};
