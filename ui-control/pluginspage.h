#pragma once
#include <QtWidgets>
#include <functional>

struct PluginDefinition {
    QString id, name, executable, cli, installCommand;
};
class PluginsPage : public QWidget {
    Q_OBJECT
public:
    explicit PluginsPage(QWidget *parent = nullptr);
    QList<PluginDefinition> definitions() const { return entries; }
    PluginDefinition definition(const QString &id) const;
    bool installing() const { return process.state() != QProcess::NotRunning; }
signals:
    void definitionsChanged();
protected:
    void changeEvent(QEvent *event) override;
private:
    std::function<void()> retranslate;
    bool showingDependency = false;
    void updateDependencyText();
    QList<PluginDefinition> entries;
    QSettings settings;
    QListWidget *list;
    QWidget *editor;
    QLineEdit *name, *executable, *cli;
    QPlainTextEdit *command, *output;
    QLabel *status;
    QPushButton *addButton, *deleteButton, *installButton;
    QProcess process;
    QString editingId;
    void select(int row);
    void saveEntry();
    void persist();
    void install();
    void setBusy(bool busy);
};
