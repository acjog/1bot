#include "wiring.h"
#ifdef Q_OS_WIN
#include <windows.h>
#include <wincrypt.h>
#endif

QByteArray protectSecret(const QByteArray &value) {
#ifdef Q_OS_WIN
    DATA_BLOB in{DWORD(value.size()), reinterpret_cast<BYTE *>(const_cast<char *>(value.constData()))}, out{};
    if (!CryptProtectData(&in, L"1bot", nullptr, nullptr, nullptr, CRYPTPROTECT_UI_FORBIDDEN, &out)) return {};
    QByteArray result(reinterpret_cast<char *>(out.pbData), int(out.cbData));
    LocalFree(out.pbData);
    return result;
#else
    Q_UNUSED(value);
    return {};
#endif
}
QByteArray unprotectSecret(const QByteArray &value) {
#ifdef Q_OS_WIN
    DATA_BLOB in{DWORD(value.size()), reinterpret_cast<BYTE *>(const_cast<char *>(value.constData()))}, out{};
    if (!CryptUnprotectData(&in, nullptr, nullptr, nullptr, nullptr, CRYPTPROTECT_UI_FORBIDDEN, &out)) return {};
    QByteArray result(reinterpret_cast<char *>(out.pbData), int(out.cbData));
    LocalFree(out.pbData);
    return result;
#else
    Q_UNUSED(value);
    return {};
#endif
}
static QByteArray read(const QString &path) {
    QFile f(path);
    if (!f.open(QIODevice::ReadOnly)) throw QObject::tr("Cannot read file: ") + path;
    return f.readAll();
}
static void write(const QString &path, const QByteArray &data) {
    QSaveFile f(path);
    if (!f.open(QIODevice::WriteOnly) || f.write(data) != data.size() || !f.commit())
        throw QObject::tr("Cannot write atomically: ") + path;
}
int runWiring() {
    QFile input, output;
    input.open(stdin, QIODevice::ReadOnly);
    output.open(stdout, QIODevice::WriteOnly);
    QJsonObject result;
    int code = 0;
    try {
        QJsonParseError error;
        auto request = QJsonDocument::fromJson(input.readAll(), &error);
        if (error.error != QJsonParseError::NoError || !request.isObject()) throw QObject::tr("Invalid request.");
        auto r = request.object();
        QString path = r["path"].toString();
        if (!QFileInfo(path).isAbsolute() || QFileInfo(path).isSymLink()) throw QObject::tr("The configuration path must be absolute and cannot be a symbolic link.");
        QDir().mkpath(QFileInfo(path).absolutePath());
        QLockFile lock(path + ".1bot.lock");
        if (!lock.tryLock()) throw QObject::tr("Another control panel is modifying this configuration.");
        QString stash = path + ".1bot-backup";
        bool existed = QFile::exists(path);
        QByteArray current = existed ? read(path) : QByteArray("{}");
        QJsonObject state;
        if (QFile::exists(stash)) {
            auto saved = unprotectSecret(read(stash));
            auto doc = QJsonDocument::fromJson(saved, &error);
            if (error.error != QJsonParseError::NoError || !doc.isObject()) throw QObject::tr("Cannot decrypt the backup.");
            state = doc.object();
            if (state["applied"].toString().toLatin1() != QCryptographicHash::hash(current, QCryptographicHash::Sha256).toHex())
                throw QObject::tr("The configuration was changed externally. Overwrite refused; merge changes manually first.");
        }
        if (r["action"] == "restore") {
            if (state.isEmpty()) throw QObject::tr("No backup is available to restore.");
            if (state["existed"].toBool()) write(path, QByteArray::fromBase64(state["original"].toString().toLatin1()));
            else if (existed && !QFile::remove(path)) throw QObject::tr("Cannot remove the newly created configuration.");
            if (!QFile::remove(stash)) throw QObject::tr("Configuration restored, but backup cleanup failed.");
        } else {
            auto doc = QJsonDocument::fromJson(current, &error);
            if (error.error != QJsonParseError::NoError || !doc.isObject()) throw QObject::tr("Only strict JSON object configurations are supported. The original file was not changed.");
            auto config = doc.object();
            if (config.contains("env") && !config["env"].isObject()) throw QObject::tr("env must be a JSON object.");
            auto env = config["env"].toObject();
            QString agent = r["agent"].toString();
            QString baseName, keyName;
            if (agent == "Claude Code") {
                baseName = "ANTHROPIC_BASE_URL"; keyName = "ANTHROPIC_AUTH_TOKEN";
                env.remove("ANTHROPIC_API_KEY");
            } else if (agent == "CodeBuddy") {
                baseName = "CODEBUDDY_BASE_URL"; keyName = "CODEBUDDY_API_KEY";
                env.remove("CODEBUDDY_AUTH_TOKEN");
                config.remove("apiKeyHelper");
            } else throw QObject::tr("Unsupported agent.");
            QUrl url(r["url"].toString());
            if (!url.isValid() || url.host().isEmpty() || !url.userInfo().isEmpty() ||
                (url.scheme() != "https" && !(url.scheme() == "http" && (url.host() == "127.0.0.1" || url.host() == "localhost"))))
                throw QObject::tr("The API URL must use HTTPS or local HTTP.");
            if (r["key"].toString().isEmpty()) throw QObject::tr("API key cannot be empty.");
            env[baseName] = url.toString();
            env[keyName] = r["key"];
            config["env"] = env;
            if (!r["model"].toString().isEmpty()) config["model"] = r["model"];
            QByteArray next = QJsonDocument(config).toJson();
            if (state.isEmpty()) {
                state["existed"] = existed;
                state["original"] = QString::fromLatin1(current.toBase64());
            }
            state["applied"] = QString::fromLatin1(QCryptographicHash::hash(next, QCryptographicHash::Sha256).toHex());
            QByteArray previous = QFile::exists(stash) ? read(stash) : QByteArray();
            QByteArray encrypted = protectSecret(QJsonDocument(state).toJson());
            if (encrypted.isEmpty()) throw QObject::tr("Cannot encrypt the backup.");
            write(stash, encrypted);
            try { write(path, next); }
            catch (...) {
                if (previous.isEmpty()) QFile::remove(stash); else write(stash, previous);
                throw;
            }
        }
        result["ok"] = true;
    } catch (const QString &message) {
        result["error"] = message; code = 1;
    }
    output.write(QJsonDocument(result).toJson(QJsonDocument::Compact));
    output.flush();
    return code;
}
