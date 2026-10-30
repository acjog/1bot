#include "updater.h"
#include <QtConcurrent>
#include <curl/curl.h>
#include <memory>
#include <limits>

struct CurlRuntime {
    CURLcode result = curl_global_init(CURL_GLOBAL_DEFAULT);
    ~CurlRuntime() { if (result == CURLE_OK) curl_global_cleanup(); }
};
struct DownloadSink {
    QIODevice &output;
    qint64 remaining;
    QString error;
};
static size_t receiveData(char *data, size_t size, size_t count, void *context) {
    auto &sink = *static_cast<DownloadSink *>(context);
    if (size && count > (std::numeric_limits<size_t>::max)() / size) return 0;
    const size_t bytes = size * count;
    if (quint64(bytes) > quint64(sink.remaining)) {
        sink.error = QObject::tr("Update file exceeds the size limit."); return 0;
    }
    if (sink.output.write(data, qint64(bytes)) != qint64(bytes)) {
        sink.error = QObject::tr("Cannot write update data."); return 0;
    }
    sink.remaining -= qint64(bytes);
    return bytes;
}
static QString fetch(const QUrl &url, const QByteArray &body, QIODevice &output, qint64 limit) {
    if (url.scheme() != "https" || url.host().isEmpty() || !url.userInfo().isEmpty()) return QObject::tr("Update downloads require HTTPS.");
    static CurlRuntime runtime;
    if (runtime.result != CURLE_OK) return QObject::tr("Cannot initialize libcurl.");
    const auto version = curl_version_info(CURLVERSION_NOW);
    if (!version || !version->ssl_version || !QByteArray(version->ssl_version).startsWith("OpenSSL/"))
        return QObject::tr("The updater requires libcurl with the OpenSSL backend.");
    const QString caPath = QCoreApplication::applicationDirPath() + "/curl-ca-bundle.crt";
    if (!QFileInfo::exists(caPath)) return QObject::tr("Missing curl-ca-bundle.crt CA bundle.");
    std::unique_ptr<CURL, decltype(&curl_easy_cleanup)> handle(curl_easy_init(), curl_easy_cleanup);
    if (!handle) return QObject::tr("Cannot create an HTTPS request.");
    CURL *curl = handle.get();
    const QByteArray address = url.toEncoded(), ca = QFile::encodeName(caPath);
    DownloadSink sink{output, limit, {}};
    bool configured = true;
    auto set = [&](CURLoption option, auto value) {
        if (curl_easy_setopt(curl, option, value) != CURLE_OK) configured = false;
    };
    set(CURLOPT_URL, address.constData());
    set(CURLOPT_PROTOCOLS, long(CURLPROTO_HTTPS));
    set(CURLOPT_FOLLOWLOCATION, 0L);
    set(CURLOPT_SSL_VERIFYPEER, 1L);
    set(CURLOPT_SSL_VERIFYHOST, 2L);
    set(CURLOPT_SSLVERSION, long(CURL_SSLVERSION_TLSv1_2));
    set(CURLOPT_CAINFO, ca.constData());
    set(CURLOPT_CONNECTTIMEOUT, 10L);
    set(CURLOPT_TIMEOUT, 300L);
    set(CURLOPT_LOW_SPEED_LIMIT, 1L);
    set(CURLOPT_LOW_SPEED_TIME, 30L);
    set(CURLOPT_NOSIGNAL, 1L);
    set(CURLOPT_USERAGENT, "1bot-control/" APP_VERSION);
    set(CURLOPT_WRITEFUNCTION, &receiveData);
    set(CURLOPT_WRITEDATA, &sink);
    std::unique_ptr<curl_slist, decltype(&curl_slist_free_all)> headers(
        curl_slist_append(nullptr, "Content-Type: application/json"), curl_slist_free_all);
    if (!headers) return QObject::tr("Cannot create request headers.");
    set(CURLOPT_HTTPHEADER, headers.get());
    if (!body.isEmpty()) {
        set(CURLOPT_POST, 1L);
        set(CURLOPT_POSTFIELDS, body.constData());
        set(CURLOPT_POSTFIELDSIZE_LARGE, curl_off_t(body.size()));
    }
    if (!configured) return QObject::tr("Cannot configure secure HTTPS request options.");
    const CURLcode result = curl_easy_perform(curl);
    if (!sink.error.isEmpty()) return sink.error;
    if (result != CURLE_OK)
        return QObject::tr("HTTPS request failed: ") + QString::fromUtf8(curl_easy_strerror(result));
    long status = 0;
    if (curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &status) != CURLE_OK || status != 200)
        return QObject::tr("Update server HTTP status %1 (redirects are not accepted).").arg(status);
    return {};
}
void Updater::check(const QUrl &url) {
    if (busy) return;
    if (!url.isValid() || url.scheme() != "https" || url.host().isEmpty() || !url.userInfo().isEmpty()) {
        emit message(QObject::tr("Configure a dedicated 1bot HTTPS update URL.")); return;
    }
    busy = true;
    files = {};
    const auto body = QJsonDocument(QJsonObject{
        {"client_version", APP_VERSION}, {"system", "windows"}, {"product", "1bot-control"}, {"arch", "x86"}
    }).toJson(QJsonDocument::Compact);
    auto watcher = new QFutureWatcher<QPair<QByteArray, QString>>(this);
    connect(watcher, &QFutureWatcherBase::finished, this, [this, watcher] {
        busy = false;
        const auto result = watcher->result();
        watcher->deleteLater();
        const auto data = result.first;
        if (!result.second.isEmpty()) { emit message(result.second); return; }
        QJsonParseError parse;
        auto doc = QJsonDocument::fromJson(data, &parse);
        if (parse.error != QJsonParseError::NoError || !doc.isObject()) {
            emit message(QObject::tr("Invalid update manifest.")); return;
        }
        auto manifest = doc.object();
        const QString version = manifest["latest_version"].toString();
        if (!manifest["has_update"].toBool() || QVersionNumber::fromString(version) <= QVersionNumber::fromString(APP_VERSION)) {
            emit message(QObject::tr("No updates available.")); return;
        }
        auto list = manifest["update_files"].toArray();
        QSet<QString> names;
        for (const auto &entry : list) {
            auto f = entry.toObject();
            QString name = f["file_name"].toString();
            QUrl url(f["download_url"].toString());
            QString hash = f["file_sha256"].toString();
            if (!QRegularExpression("^[a-zA-Z0-9_-][a-zA-Z0-9_.-]*$").match(name).hasMatch() ||
                name.contains("..") || name.compare("1bot-updater.exe", Qt::CaseInsensitive) == 0 ||
                names.contains(name.toLower()) || url.scheme() != "https" || url.host().isEmpty() ||
                !url.userInfo().isEmpty() || !QRegularExpression("^[a-fA-F0-9]{64}$").match(hash).hasMatch() ||
                f["file_size"].toDouble() <= 0 || f["file_size"].toDouble() > 256 * 1024 * 1024) {
                emit message(QObject::tr("The update manifest contains unsafe files, invalid sizes, or missing SHA-256 hashes.")); return;
            }
            names.insert(name.toLower());
        }
        if (list.isEmpty() || !names.contains("1bot-control.exe")) {
            emit message(QObject::tr("The update package is missing 1bot-control.exe.")); return;
        }
        files = list;
        emit available(version, manifest["release_notes"].toString());
    });
    watcher->setFuture(QtConcurrent::run([url, body] {
        QByteArray data;
        QBuffer buffer(&data); buffer.open(QIODevice::WriteOnly);
        const QString error = fetch(url, body, buffer, 1024 * 1024);
        return qMakePair(data, error);
    }));
}
void Updater::download() {
    if (busy || files.isEmpty()) return;
    directory = QStandardPaths::writableLocation(QStandardPaths::AppLocalDataLocation) +
        "/updates/" + QUuid::createUuid().toString(QUuid::WithoutBraces);
    if (!QDir().mkpath(directory)) { emit message(QObject::tr("Cannot create the update directory.")); return; }
    busy = true; index = 0; next();
}
void Updater::next() {
    if (index == files.size()) {
        QSaveFile plan(directory + "/manifest.json");
        const auto data = QJsonDocument(QJsonObject{{"update_files", files}}).toJson();
        if (!plan.open(QIODevice::WriteOnly) || plan.write(data) != data.size() || !plan.commit()) {
            busy = false; emit message(QObject::tr("Cannot save the update manifest.")); return;
        }
        busy = false; emit ready(directory); return;
    }
    auto f = files[index].toObject();
    emit message(QObject::tr("Downloading update %1/%2: %3").arg(index + 1).arg(files.size()).arg(f["file_name"].toString()));
    const QString path = directory + "/" + f["file_name"].toString();
    auto watcher = new QFutureWatcher<QString>(this);
    connect(watcher, &QFutureWatcherBase::finished, this, [this, watcher] {
        const auto error = watcher->result(); watcher->deleteLater();
        if (!error.isEmpty()) { busy = false; emit message(error); return; }
        ++index; next();
    });
    watcher->setFuture(QtConcurrent::run([path, f]() -> QString {
        QSaveFile file(path);
        if (!file.open(QIODevice::WriteOnly)) return QObject::tr("Cannot create the download file.");
        const qint64 expected = qint64(f["file_size"].toDouble());
        auto error = fetch(QUrl(f["download_url"].toString()), {}, file, expected);
        if (!error.isEmpty()) return error;
        if (file.size() != expected || !file.commit()) return QObject::tr("Download size mismatch or save failure.");
        QFile input(path);
        QCryptographicHash hash(QCryptographicHash::Sha256);
        if (!input.open(QIODevice::ReadOnly) || !hash.addData(&input) ||
            hash.result().toHex() != f["file_sha256"].toString().toLatin1().toLower()) {
            input.close(); QFile::remove(path); return QObject::tr("SHA-256 verification failed. The download was deleted.");
        }
        return {};
    }));
}
