#include <QtCore>
#ifdef Q_OS_WIN
#include <windows.h>
#endif
static bool copyAtomic(const QString &from, const QString &to) {
    QFile input(from);
    QSaveFile output(to);
    if (!input.open(QIODevice::ReadOnly) || !output.open(QIODevice::WriteOnly)) return false;
    while (!input.atEnd()) {
        auto data = input.read(1024 * 1024);
        if (data.isEmpty() && input.error() != QFile::NoError) return false;
        if (output.write(data) != data.size()) return false;
    }
    return output.commit();
}
int main(int argc, char **argv) {
    QCoreApplication app(argc, argv);
    auto args = app.arguments();
    if (args.size() != 5) return 2;
    QString target = QDir(args[1]).absolutePath(), source = QDir(args[2]).absolutePath();
    QFile log(source + "/install.log");
    log.open(QIODevice::WriteOnly | QIODevice::Append);
    auto fail = [&](const QByteArray &message) { log.write(message + "\n"); log.flush(); return 1; };
#ifdef Q_OS_WIN
    for (int i = 3; i <= 4; ++i) {
        DWORD pid = args[i].toULong();
        if (!pid) continue;
        HANDLE process = OpenProcess(SYNCHRONIZE, FALSE, pid);
        if (process) {
            DWORD result = WaitForSingleObject(process, 30000);
            CloseHandle(process);
            if (result != WAIT_OBJECT_0) return fail("Process did not exit");
        } else if (GetLastError() != ERROR_INVALID_PARAMETER) return fail("Cannot wait for process");
    }
#endif
    QLockFile lock(target + "/.1bot-update.lock");
    if (!lock.tryLock()) return fail("Another update is running");
    QFile manifest(source + "/manifest.json");
    if (!manifest.open(QIODevice::ReadOnly)) return fail("Cannot read manifest");
    auto files = QJsonDocument::fromJson(manifest.readAll()).object()["update_files"].toArray();
    if (files.isEmpty()) return fail("Empty manifest");
    QSet<QString> names;
    for (auto entry : files) {
        auto f = entry.toObject();
        QString name = f["file_name"].toString();
        if (!QRegularExpression("^[a-zA-Z0-9_-][a-zA-Z0-9_.-]*$").match(name).hasMatch() ||
            name.contains("..") || name.compare("1bot-updater.exe", Qt::CaseInsensitive) == 0 ||
            names.contains(name.toLower()) || QFileInfo(target + "/" + name).isSymLink())
            return fail("Unsafe file name");
        names.insert(name.toLower());
        QFile file(source + "/" + name);
        if (!file.open(QIODevice::ReadOnly)) return fail("Missing download");
        QCryptographicHash hash(QCryptographicHash::Sha256);
        if (!hash.addData(&file) || hash.result().toHex() != f["file_sha256"].toString().toLatin1().toLower() ||
            file.size() != qint64(f["file_size"].toDouble())) return fail("Invalid download");
    }
    if (!names.contains("1bot-control.exe")) return fail("Missing main executable");
    QString backup = source + "/rollback";
    if (!QDir().mkpath(backup)) return fail("Cannot create backup");
    QStringList changed, created;
    for (auto entry : files) {
        QString name = entry.toObject()["file_name"].toString();
        QString dest = target + "/" + name;
        bool exists = QFile::exists(dest);
        if (exists && !copyAtomic(dest, backup + "/" + name)) return fail("Cannot back up installed files");
    }
    for (auto entry : files) {
        QString name = entry.toObject()["file_name"].toString();
        QString dest = target + "/" + name;
        bool exists = QFile::exists(dest);
        if (!copyAtomic(source + "/" + name, dest)) {
            bool restored = true;
            for (const auto &item : changed) restored = copyAtomic(backup + "/" + item, target + "/" + item) && restored;
            for (const auto &item : created) restored = QFile::remove(target + "/" + item) && restored;
            return fail(restored ? "Install failed; rollback completed" : "Rollback failed; recover from rollback directory");
        }
        if (exists) changed << name; else created << name;
    }
    log.write("Update installed\n"); log.flush();
    if (!QProcess::startDetached(target + "/1bot-control.exe", {}, target)) return fail("Installed but restart failed");
    return 0;
}
