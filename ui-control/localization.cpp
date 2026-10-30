#include "localization.h"

QString Localization::resolveLanguage(const QString &preference) {
    const QStringList supported{"zh_CN", "en_US", "ja_JP", "ko_KR", "zh_TW"};
    if (supported.contains(preference)) return preference;
    const QLocale system = QLocale::system();
    if (system.language() == QLocale::Chinese)
        return system.script() == QLocale::TraditionalHanScript ? "zh_TW" : "zh_CN";
    if (system.language() == QLocale::Japanese) return "ja_JP";
    if (system.language() == QLocale::Korean) return "ko_KR";
    return "en_US";
}
Localization::Localization() {
    const auto args = QCoreApplication::arguments();
    const int languageOption = args.indexOf("--language");
    const QString preference = languageOption >= 0 && languageOption + 1 < args.size()
        ? args[languageOption + 1] : QSettings().value("ui/language", "system").toString();
    setLanguage(preference);
}

bool Localization::setLanguage(const QString &preference) {
    const QString language = resolveLanguage(preference);
    if (QCoreApplication::instance()->property("uiLanguage").toString() == language) return true;
    QTranslator candidate;
    if (!candidate.load(":/i18n/" + language + ".qm")) return false;
    QCoreApplication::removeTranslator(&applicationTranslator);
    QCoreApplication::removeTranslator(&qtTranslator);
    qtTranslator.load(":/i18n/qtbase_" + language + ".qm");
    applicationTranslator.load(":/i18n/" + language + ".qm");
    QCoreApplication::instance()->setProperty("uiLanguage", language);
    QCoreApplication::installTranslator(&qtTranslator);
    QCoreApplication::installTranslator(&applicationTranslator);
    return true;
}
