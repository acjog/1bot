#pragma once
#include <QtWidgets>

class AccessCard : public QWidget {
public:
    explicit AccessCard(QWidget *parent = nullptr);
    void setConfig(int port, const QString &apiKey, const QString &model);
protected:
    void changeEvent(QEvent *event) override;
private:
    void refresh();
    QGroupBox *group = nullptr;
    QLabel *hint = nullptr;
    QLabel *protocolLabel = nullptr;
    QTabBar *protocols = nullptr;
    QLabel *baseUrlLabel = nullptr, *apiKeyLabel = nullptr, *modelLabel = nullptr;
    QLineEdit *baseUrlValue = nullptr, *apiKeyValue = nullptr, *modelValue = nullptr;
    QPushButton *copyBaseUrl = nullptr, *copyApiKey = nullptr, *copyModel = nullptr, *showApiKey = nullptr;
    QTabWidget *examples = nullptr;
    QPlainTextEdit *powerShellExample = nullptr, *cmdExample = nullptr, *curlExample = nullptr;
    QPushButton *copyPowerShell = nullptr, *copyCmd = nullptr, *copyCurl = nullptr;
    QPlainTextEdit *pythonExample = nullptr, *nodeExample = nullptr;
    QPushButton *copyPython = nullptr, *copyNode = nullptr;
    int port = 8787;
    QString apiKey, model;
};
