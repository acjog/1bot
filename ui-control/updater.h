#pragma once
#include <QtCore>
#include <QtNetwork>
class Updater : public QObject {
    Q_OBJECT
public:
    explicit Updater(QObject *parent = nullptr) : QObject(parent) {}
    void check(const QUrl &url);
    void download();
signals:
    void available(QString version, QString notes);
    void message(QString text);
    void ready(QString directory);
private:
    QJsonArray files;
    QString directory;
    int index = 0;
    bool busy = false;
    void next();
};
